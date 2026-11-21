package workbench

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sylvinus/agent-vm/internal/remotehost"
)

func TestSSHConfigImportKeepsLocalSessionAndOriginBoundary(t *testing.T) {
	s := testServer(t)
	s.remote = newRemoteGateway(s.profile)
	t.Cleanup(s.remote.close)
	for _, method := range []string{"GET", "POST"} {
		if w := request(s, method, "/api/client/ssh-config", s.origin, "{}", false); w.Code != http.StatusUnauthorized {
			t.Fatalf("unauthenticated %s: %d", method, w.Code)
		}
	}
	if w := request(s, "POST", "/api/client/ssh-config", "https://untrusted.example", "{}", true); w.Code != http.StatusForbidden {
		t.Fatalf("cross-origin import: %d", w.Code)
	}
	if runtime.GOOS != "windows" {
		w := request(s, "GET", "/api/client/ssh-config", "", "", true)
		if w.Code == http.StatusOK || !strings.Contains(w.Body.String(), "Windows") {
			t.Fatalf("non-Windows import: %d %s", w.Code, w.Body.String())
		}
	}
}

func TestWindowsSSHConfigPreviewAndImport(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows user-profile path and named pipes")
	}
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("USERNAME", "local-user")
	if err := os.Mkdir(filepath.Join(home, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(home, ".ssh", "config")
	text := "Host dev prod\n HostName %h.example\n User remote\n IdentityFile C:/never-read/SECRET_KEY\nHost jump\n ProxyCommand command SECRET_COMMAND\n"
	if err := os.WriteFile(config, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	s := testServer(t)
	s.remote = newDesktopGateway(s.profile, nil)
	t.Cleanup(s.remote.close)
	s.account.lock()
	// Even while displaying a remote view, import is owned by the local shell.
	w := request(s, "GET", "/api/client/ssh-config", "", "", true)
	var preview remotehost.SSHConfigPreview
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &preview) != nil || len(preview.Hosts) != 2 || len(preview.Skipped) != 1 {
		t.Fatalf("preview: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "SECRET") {
		t.Fatal("preview leaked a referenced key path or command")
	}
	if _, err := os.Stat(filepath.Join(s.profile, "remote-hosts.json")); !os.IsNotExist(err) {
		t.Fatal("preview persisted profiles")
	}
	payload := func(revision string, aliases []string) string {
		data, _ := json.Marshal(map[string]any{"revision": revision, "aliases": aliases})
		return string(data)
	}
	for _, body := range []string{payload(preview.Revision, []string{"jump"}), payload(preview.Revision, []string{"dev", "dev"}), payload("stale", []string{"dev"}), `{"path":"C:/arbitrary","aliases":["dev"]}`} {
		if got := request(s, "POST", "/api/client/ssh-config", s.origin, body, true); got.Code == 200 {
			t.Fatalf("invalid selection accepted: %s", body)
		}
	}
	if _, err := os.Stat(filepath.Join(s.profile, "remote-hosts.json")); !os.IsNotExist(err) {
		t.Fatal("rejected request persisted profiles")
	}
	if err := os.WriteFile(config, []byte(text+"# changed\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := request(s, "POST", "/api/client/ssh-config", s.origin, payload(preview.Revision, []string{"dev"}), true); got.Code != 409 {
		t.Fatalf("changed config accepted: %d %s", got.Code, got.Body.String())
	}
	if err := os.WriteFile(config, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	body := payload(preview.Revision, []string{"dev"})
	w = request(s, "POST", "/api/client/ssh-config", s.origin, body, true)
	var result struct {
		Imported []remotehost.HostProfile `json:"imported"`
		Skipped  int                      `json:"skipped"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || len(result.Imported) != 1 || result.Imported[0].Auth.Kind != "agent" || result.Imported[0].Auth.IdentityFile != "" {
		t.Fatalf("import: %d %s", w.Code, w.Body.String())
	}
	w = request(s, "POST", "/api/client/ssh-config", s.origin, body, true)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || len(result.Imported) != 0 || result.Skipped != 1 {
		t.Fatalf("repeat: %d %s", w.Code, w.Body.String())
	}
	stored, err := os.ReadFile(filepath.Join(s.profile, "remote-hosts.json"))
	if err != nil || strings.Contains(string(stored), "SECRET") || strings.Contains(string(stored), "identity_file") {
		t.Fatalf("stored private material: %s %v", stored, err)
	}
	if state := s.remote.context(); state.State.State != "disconnected" || state.Candidate != nil {
		t.Fatalf("import started a connection: %+v", state)
	}
}
