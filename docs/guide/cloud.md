---
title: 云端部署与 SSH
description: 将工作环境放到自己的 Linux VPS，通过桌面 Cage 部署、连接和管理服务器上的任务。
group: 远程使用
order: 8
keywords: 云端 远程部署 VPS SSH 无KVM rootless 服务器 主机 指纹 远程项目
---
## 什么时候选择云端

希望任务在服务器上持续运行，或本机不适合运行虚拟机时，可以使用云端模式。终端、浏览器、编辑器、Claude Desktop 和网络出口位于服务器，电脑显示同一套工作台。本机虚拟机可以保持关闭。

服务器模式通过 rootless 容器运行，不需要 KVM。它仍需要满足系统、权限和资源条件，并非所有 VPS 都可以直接部署。

本机模式在电脑上运行 Debian 虚拟机，云端模式在服务器的容器里运行同一套 Linux 环境；需要服务器常驻任务时选择云端。两种方式沿用同一套工作台。本机环境见[Debian 环境](guest-systems.md)。

## 准备服务器

当前目标为 Linux x86_64，具备 systemd 用户服务、cgroup v2、rootless 用户命名空间、至少 65,536 个 subordinate UID/GID，以及 `/dev/fuse`、`/dev/net/tun`。普通 SSH 用户需要相应权限；还需 OpenSSH、curl、tar、sha256sum、iproute2 与 iptables 等基础工具。

完整前提和维护命令见[部署说明](https://github.com/Lee-xeo/Claude-Cage/blob/main/deploy/cloud/README.md)。权限缺失应修复服务器配置，不要改用特权容器或把 Docker socket 暴露给工作环境。

需要在退出 SSH 后持续运行或在服务器重启后恢复服务时，由服务器管理员配置 systemd 用户服务、linger 与相应启动策略。

<figure class="guide-shot">
  <a href="../../www/public/docs/screenshots/ssh-sidebar.jpg" target="_blank" rel="noopener" aria-label="查看原图：SSH 主机侧栏：添加自己的服务器"><img src="../../www/public/docs/screenshots/ssh-sidebar.jpg" alt="SSH 主机侧栏：添加自己的服务器" width="2560" height="1764" loading="lazy" decoding="async" /></a>
  <figcaption>SSH 主机侧栏：添加自己的服务器。截图版本：0.1.6-alpha-preview（开发构建） · macOS 原生应用 · 2026-11-07。<a href="../../www/public/docs/screenshots/ssh-sidebar.jpg" target="_blank" rel="noopener">查看原图（2560 × 1764） ↗</a></figcaption>
</figure>

## 从侧栏部署并连接

1. 打开 Cage，进入左栏“SSH 主机”，点击加号填写名称、地址、端口、用户名和本机密钥文件或 SSH agent。自定义路径可在“云端安装路径”设置。使用 SSH agent 时见下文[指定 SSH agent](#指定-ssh-agent)。
2. 点击“连接此主机”，将 SSH 主机指纹与服务器独立显示的记录核对。指纹变化时先核查原因，再继续。
3. 首次使用空配置目录时，点击“部署并连接”。Cage 传输匹配的程序和脚本，在服务器准备独立引擎与持久环境。
4. 已部署且就绪时直接进入；能确认归属的用户服务已停止时会尝试启动。安装中断后可用同一版本重试。
5. 云端工作台打开后，创建或解锁服务器自己的 Cage 账户，再配置服务器的代理和项目。

本机有任务时，连接卡片提供保留或停止本机环境的选项，停止前先保存内容。不要在未经备份的已有本机数据目录上首次部署云端实例。

### 指定 SSH agent

登录方式选“本机 SSH agent”时，“SSH agent 套接字”留空即沿用原有行为：macOS 和 Linux 使用启动 Cage 时的 `SSH_AUTH_SOCK`；Windows 先用 `SSH_AUTH_SOCK`，未设置时使用 Windows OpenSSH agent（`\\.\pipe\openssh-ssh-agent`）。

密码管理器等程序自带 agent、而 `SSH_AUTH_SOCK` 指向另一个没有密钥的 agent 时（例如 macOS 默认的 launchd agent），在这一栏填写该程序给出的 agent 位置，与 SSH 配置中 `IdentityAgent` 的值相同，但要写成完整路径：

- macOS 和 Linux：填写 Unix 套接字的绝对路径，不展开 `~` 和环境变量，长度不超过系统上限（macOS 103 字节，Linux 107 字节）。连接前核对它确实是套接字、属于当前用户，且组和其他用户没有写权限；可以是指向这样的套接字的符号链接。权限不符时按提示改为仅自己可读写（如 `chmod 600`）。
- Windows：只接受本机命名管道，格式为 `\\.\pipe\名称`，不接受网络路径和其他文件路径。

Cage 只连接你填写的位置，不在连接时自动读取 `~/.ssh/config`，也不运行 `ssh -G` 等命令推断 agent。Windows 可以通过下文的导入入口显式读取配置。连接失败时提示会区分三种情况：无法连接 agent 或指定位置不符合要求、agent 通信失败（未能读取密钥列表）、agent 中没有可用密钥。私钥始终留在 agent 内，Cage 不导出也不上传。

### 从 Windows SSH 配置导入

Windows 主机的“SSH 主机”侧栏提供“导入 Windows SSH 配置”。点击后从当前 Windows 用户的 `%USERPROFILE%\.ssh\config` 读取可导入的主机，核对名称、地址、用户名与端口，取消不需要的条目，再点击“导入所选主机”。预览不保存内容；导入只追加本机连接记录，不发起连接、启动客机或部署服务。重复别名（不区分大小写）或相同地址、端口、用户名的已有连接会跳过，保留其认证设置与已核对的指纹。超出 100 台主机上限或保存失败时整批不导入。

导入的登录方式固定为“本机 SSH agent”。`IdentityFile` 只作为配置语法被忽略，Cage 不打开该路径，不自动加载密钥，不保存私钥路径，更不会把私钥或原始配置复制到虚拟机、远端或共享项目目录。先在 Windows 上启用并启动 OpenSSH Authentication Agent 服务，再用系统 `ssh-add` 将所需私钥加入 agent，可用 `ssh-add -l` 检查。Cage 不自动更改服务设置。配置中没有 `IdentityAgent` 时，继续使用启动 Cage 时的 `SSH_AUTH_SOCK`，未设置时使用 `\\.\pipe\openssh-ssh-agent`；也支持显式的本机 Windows 命名管道和 `IdentityAgent SSH_AUTH_SOCK`。agent 不可用或没有密钥时明确报错，不回退读取私钥文件，也不向客机或远端转发 agent。

支持 UTF-8（含 BOM）、CRLF、注释、双引号、`key=value`、一个 `Host` 下的多个别名、重复 `Host` 段、`*`/`?` 通配与 `!` 排除，按 OpenSSH 的“第一个匹配值生效”规则提取 `HostName`、`User`、`Port`、`IdentityAgent`；`HostName` 支持 `%h`。只枚举明确别名，通配符本身不会生成主机。未填写 `User` 时使用 Windows 的 `USERNAME`，默认端口 22。导入默认为 Linux 远端；Windows 远端请在导入后编辑“远程系统”。

以下限制会明确报告，现有手动添加方式保持可用：

- 含 `Include` 或 `Match` 的文件整体拒绝导入，避免漏掉条件配置后连接到错误目标。不会执行 `Match exec`、`ProxyCommand`、`LocalCommand`，也不调用 `ssh -G` 或外部解析程序。
- 依赖 `ProxyCommand`、`ProxyJump`、主机名规范化、`IdentitiesOnly yes`、`IdentityAgent none`、证书、其他认证策略或未知连接选项的主机跳过，不静默改用直连或其他密钥。跳过原因不包含命令、密钥路径等原始选项值。
- 导入只提取连接元数据；不沿用转发、环境变量、保活、日志、压缩、`known_hosts` 路径或 `StrictHostKeyChecking` 等设置。首次连接仍通过 Cage 原有的指纹核对流程，已有主机的固定指纹不变。
- 不读取系统级 SSH 配置，只读取本机磁盘上的用户配置，最多 1 MiB、每行小于 64 KiB、100 个明确别名。预览后配置发生变化会拒绝保存，需重新读取并核对。

例如，下面的 `dev` 可以导入；请先通过 Windows `ssh-add` 把对应密钥加入 agent：

```sshconfig
Host dev
    HostName dev.example.com
    User developer
    Port 2222
    IdentityFile C:/Users/you/.ssh/id_ed25519
```

配置语义参考 [OpenSSH ssh_config](https://man.openbsd.org/ssh_config)；Windows agent 的启用与密钥加载见 [Microsoft OpenSSH 密钥管理](https://learn.microsoft.com/windows-server/administration/openssh/openssh_keymanagement)。导入是一次性读取，之后修改原始 SSH 配置不会自动更新已保存连接。

<figure class="guide-shot">
  <a href="../../www/public/docs/screenshots/ssh-host-form.jpg" target="_blank" rel="noopener" aria-label="查看原图：远程主机地址、认证与私钥路径表单"><img src="../../www/public/docs/screenshots/ssh-host-form.jpg" alt="远程主机地址、认证与私钥路径表单" width="2560" height="1764" loading="lazy" decoding="async" /></a>
  <figcaption>远程主机地址、认证与私钥路径表单。截图版本：0.1.6-alpha-preview（开发构建） · macOS 原生应用 · 2026-11-07。<a href="../../www/public/docs/screenshots/ssh-host-form.jpg" target="_blank" rel="noopener">查看原图（2560 × 1764） ↗</a></figcaption>
</figure>

## 服务器项目与本机共享

远端工作台可选择服务器目录。云端账户、浏览资料、保险箱、节点与系统语言属于该服务器，本机已有配置不会自动复制过去。操作前看清当前主机名称。

需要使用本机源码时，按界面授权一个具体项目，选择所需的只读或读写方式。共享仅在授权连接期间有效。Git 目录与历史随项目共享：只读共享允许查看，读写共享允许暂存、提交和切换分支。授权目录内的隐藏文件、凭据和工具配置也会共享，目录外访问仍被拒绝。读写共享会直接修改本机源文件与仓库，先备份并审查改动。

上传由本机选择器读取你选定的文件，下载由服务器提供文件内容，再明确保存到本机。远端文件与本地副本的保存位置分别显示。

## 断开、锁定与停止

关闭本地窗口或断开 SSH 会释放客户端连接，服务器环境和任务继续运行。重新连接读取仍存活的会话。需要结束远端任务时，在对应云端工作台明确“停止环境”。

切换主机后，旧连接的文件选择、输入和请求不能继续进入新主机。连接失败时不会自动改为操作本机。对发送、保存等结果不明确的操作，重连后先核对状态再重试。

<figure class="guide-shot">
  <a href="../../www/public/docs/screenshots/ssh-install-paths.jpg" target="_blank" rel="noopener" aria-label="查看原图：自定义服务器上的程序与配置路径"><img src="../../www/public/docs/screenshots/ssh-install-paths.jpg" alt="自定义服务器上的程序与配置路径" width="2560" height="1764" loading="lazy" decoding="async" /></a>
  <figcaption>自定义服务器上的程序与配置路径。截图版本：0.1.6-alpha-preview（开发构建） · macOS 原生应用 · 2026-11-07。<a href="../../www/public/docs/screenshots/ssh-install-paths.jpg" target="_blank" rel="noopener">查看原图（2560 × 1764） ↗</a></figcaption>
</figure>

## 更新与故障恢复

客户端更新只更新本机程序，服务器维护另行执行。更新服务器前保存任务、停止相关环境与服务，完整保留配置、项目、容器文件系统及引擎数据，用原绝对路径恢复。

不要删除容器或执行清理全部引擎数据的命令来升级；容器可写文件系统也保存持久资料。出现连接、指纹或部署前提错误时，查[SSH 与云端故障](troubleshooting.md#ssh-或云端部署失败)。
