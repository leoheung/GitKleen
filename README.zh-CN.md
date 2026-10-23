<p align="center">
  <img src="docs/images/logo.png" width="128" height="128" alt="GitKleen">
</p>

<h1 align="center">GitKleen</h1>

<p align="center">
  <b>轻快的原生 Git 客户端，远程仓库也能用</b><br>
  JetBrains 那样的分支图和历史，VS Code 那样的暂存，安装包不到 7 MB，秒开；本地和 SSH 远程的仓库都能用。<br>
  也为 agent 写代码而做：终端里 <code>gitkleen .</code> 打开，agent 改到哪你看到哪，没问题的暂存，其余的交回给它。
</p>

<p align="center">
  <a href="https://github.com/leoheung/GitKleen/actions/workflows/ci.yml"><img src="https://github.com/leoheung/GitKleen/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <img src="https://img.shields.io/badge/macOS-Apple%20芯片-111?logo=apple" alt="macOS">
  <img src="https://img.shields.io/badge/Windows-x64-0078D4?logo=windows" alt="Windows">
  <img src="https://img.shields.io/badge/Linux-x64%20%C2%B7%20arm64-FCC624?logo=linux&logoColor=111" alt="Linux">
  <img src="https://img.shields.io/badge/built%20with-Go%20%2B%20MyGo-00ADD8?logo=go&logoColor=white" alt="Go">
</p>

<p align="center">
  <a href="README.md">English</a> · <b>简体中文</b>
</p>

## 为什么是 GitKleen

代码交给 agent 写以后，用 git 的大部分时间都花在看改了什么上。GitKleen 让这件事变快：

1. **看**：文件一改，这里马上就能看到，不用手动刷新。每个文件并排对比、带语法高亮。锁文件、压缩文件和生成的文件自动折叠，图片改动显示前后对比。
2. **留**：直接在 diff 里修一行，把没问题的文件、改动块或行暂存起来，然后提交。
3. **退**：在要返工的行上写一句，复制为 Markdown（每条带文件、行号和代码），粘给 Claude Code、Codex 或同事。

GitKleen 读的是仓库，不是某个 agent 的会话，所以换哪个 agent、同时用几个、还是你自己改的代码，用起来都一样。

| 现在看 agent 改动的办法 | 换成 GitKleen |
|---|---|
| agent 自带的 diff 视图 | 哪个 agent 都能用，还能逐行暂存，提交前直接改代码 |
| 打开 IDE | 秒开，内存只要零头 |
| 终端里跑 `git diff` | 并排对比，生成文件自动折叠，文件一改马上能看到 |

### 轻量、快速、现代

用起来就是你熟悉的 IDE：JetBrains 的分支图和历史操作，VS Code 的改动列表和暂存方式，⌘K 打开命令栏，⌘P 转到文件，深浅色都有。但它不是 IDE：GitKleen 是一个原生二进制文件，由 `https://github.com/egoist/mygo` 用 GPU 绘制界面，没有 webview，也没有 JVM。打开一个 2 万个提交的仓库后的内存占用，应用的所有进程都计算在内（Apple M4）：

| 应用 | 内存 |
|---|---|
| **GitKleen** | **109 MB** |
| 精简版 IntelliJ 客户端 | 750 MB |
| VS Code，不带插件 | 1000 MB |
| VS Code，167 个插件 | 1400 MB |
| IntelliJ IDEA 2026.2 | 2300 MB |

在同一个仓库上，GitKleen 读完全部历史并布局分支图只要 **0.13 秒**，打开一个 20 MB 补丁的提交（5677 个文件）只要 **0.24 秒**。

GitKleen 读取靠解析 `git` 的输出，所有改动都交给 `git` 命令本身完成，所以仓库在 GitKleen 里的表现和在终端里完全一致。macOS、Windows 和 Linux 上是同一个应用，功能完全相同。

## Fork 自 MyGit

GitKleen fork 自 `https://github.com/Binaryify/MyGit`（作者 Binaryify）——它本身已经是一个轻快、原生的 git 客户端，本地和 SSH 远程仓库都能用，MIT 许可证。这个底子值得保留，GitKleen 也确实保留了：解析 `git` 的输出、由 `git` 本身完成所有改动的思路不变，功能不变，还是那个单一的小二进制文件。

有两件事，在 fork 里做比在上游做合适：

- **自己的更新。** 装好的应用只会去它出生时定好的地方找更新——MyGit 的 releases、用 MyGit 的密钥签名——所以自己构建的版本都是死路：你分发出去的副本永远更新不到它。GitKleen 用自己的密钥给更新签名，从自己的 releases 提供，装的是什么就一直是什么。
- **自己的方向。** 界面在这里继续打磨——分支列、diff 抽屉、标签和 stash 面板，以及下面其余的新东西——这些在自己 fork 里做总比在别人的项目里容易。名字和图标也是：命令行上是 `gitkleen .`，SSH 连过去的电脑上是 `~/.gitkleen`。

许可证仍是 MIT，MyGit 的原始版权保留在 [LICENSE](LICENSE) 里。

## GitKleen 的新内容

MyGit 有的都有，再加上 GitKleen 的 [0.2.0](CHANGELOG.md) 在 MyGit 0.3.5（上游最后一个版本）之上的新东西：

**标签和 stash**

- 每个标签的名字下面显示其提交的哈希，右边显示日期；点一行跳到历史里的那个提交并选中
- stash 按时间排序，可按关键词筛选；点一行直接打开这个 stash 的 diff，单条改动也能从它自己的行 stash
- 在标签或 stash 里选了新的一行，之前选中的行——那边的或历史里的——就取消选中，一次只选一样，不会一边高亮着一边显示另一个的 diff

**窗口**

- 分支从标题栏下拉挪到工作区旁边的一列，远程分支按文件夹分组，列上方有搜索栏，可搜分支和仓库
- 分支、stash 和标签是这一列上方的三个标签页
- diff 在窗口右侧的抽屉里打开
- 标题栏更高、石板色：左边是分支块，操作居中，原来放箭头的地方显示拉取和推送的计数
- 更改面板：两个列表互换位置，提交框移到它们下面，stash 入口在列表头部

**SSH 远程**

- GitKleen 先用 git 远程操作配置的密钥去试，默认那把排在前面，都试过了才会询问

**还有名字**

- GitKleen 自己的名字、图标和命令行（`gitkleen .`）；SSH 连过去的电脑上，辅助程序放在 `~/.gitkleen`；更新用自己的密钥签名，从自己的 releases 提供

## 功能

- **看改动**：生成的文件（锁文件、压缩文件、`linguist-generated`）默认折叠；可以把文件标记为已看；图片改动显示前后对比；可以评论某几行，并把评论复制为 Markdown 交给 AI 或同事
- **暂存**：按文件、改动块或选中的行暂存、取消暂存、丢弃；在 diff 里编辑并保存工作区文件；改动成千上万时按路径筛选
- **提交**：完整说明、改动文件、带语法高亮的并排 diff；合并提交可以选对比哪个父提交；比较两个提交（⌘ 点选）、分支或工作区；在菜单里修改上次提交；确实要跳过时可以不跑钩子（`--no-verify`）
- **历史**：所有分支、标签、远程，带分支图，边读边显示；按说明、作者、哈希搜索；按分支或路径筛选（文件历史）
- **历史操作**：新建分支和标签、检出、cherry-pick、revert、reset（soft / mixed / hard）、交互式变基（重排、改写说明、停下修改、squash、fixup、丢弃）
- **拉取请求**：打开 GitHub 拉取请求，在对应行下阅读和回复讨论、解决讨论，并提交审阅：评论、批准或请求修改
- **冲突**：三栏合并视图；合并、变基、cherry-pick、revert 中途可以继续、跳过或中止
- **Stash、Blame 和远程**：stash（可包括未跟踪文件）、逐行追溯、远程管理、拉取（快进 / 变基 / 合并）、推送前预览、force-with-lease
- **登录**：git 需要用户名、令牌或口令时，在窗口里询问
- **工作区**：打开包含多个仓库的文件夹（比如 agent 同时改几个仓库时的父目录），侧边栏像 VS Code 源代码管理那样列出每个仓库，各自有提交框、同步和改动列表
- **SSH 远程**：打开另一台电脑上的仓库（`host:/path/to/repo`，Linux、macOS、Windows 都行），沿用本机的 ssh 配置和密钥；GitKleen 会在对方的 `~/.gitkleen` 里放一个小的辅助程序，对方只要有 git。Windows 上 WSL 的发行版也能这样打开，不用配 sshd（`wsl+Ubuntu:~/repo`）
- **随处可用**：每个仓库一个窗口、拖入文件夹打开、终端里 `gitkleen .`、自定义标题栏、深浅色（跟随系统，或在设置里指定）、English 和简体中文、签名的自动更新

## 安装

下载最新版本：

| 平台 | 下载 |
|---|---|
| macOS（Apple 芯片） | `https://github.com/leoheung/GitKleen/releases/latest/download/GitKleen-macos-arm64.dmg` |
| Windows（x64） | `https://github.com/leoheung/GitKleen/releases/latest/download/GitKleen-windows-amd64-setup.exe` · `https://github.com/leoheung/GitKleen/releases/latest/download/GitKleen-windows-amd64.zip` |
| Linux（x64） | `https://github.com/leoheung/GitKleen/releases/latest/download/GitKleen-linux-amd64.deb`，或用下面的脚本 |
| Linux（arm64） | `https://github.com/leoheung/GitKleen/releases/latest/download/GitKleen-linux-arm64.deb`，或用下面的脚本 |

Linux 上这条命令把 GitKleen 装到 `~/.local`，不需要 root：

```bash
curl -fsSL https://github.com/leoheung/GitKleen/releases/latest/download/install.sh | sh
```

GitKleen 暂未经过 Apple 公证。如果 macOS 提示无法打开，先把它拖进「应用程序」，再运行：

```bash
xattr -dr com.apple.quarantine /Applications/GitKleen.app
```

Windows 安装程序暂时也没有代码签名。如果 Windows 提示「已保护你的电脑」，点「更多信息」，再点「仍要运行」。

GitKleen 需要 `PATH` 里有 `git`，安装后会自动更新。不支持 Intel Mac。所有版本见 `https://github.com/leoheung/GitKleen/releases`。

想从源码构建，见[开发](#开发)。

## 快捷键

| 按键 | 作用 |
|---|---|
| ⌘↩ / Ctrl+Enter | 提交 |
| ⌘F / Ctrl+F | 搜索历史 |
| ⌘R / Ctrl+R | 刷新 |
| ⌘P / Ctrl+P | 转到提交或更改里的文件 |
| ⌘K / Ctrl+K | 命令栏 |
| ⌘S / Ctrl+S | 保存在 diff 里编辑的文件 |
| F7 / ⇧F7 | 跳到 diff 里的下一处 / 上一处改动；到头后再按一次进入下一个 / 上一个文件 |
| 空格 | 暂存或取消暂存选中的改动；把提交里的文件标记为已看 |
| Delete | 丢弃选中的改动 |
| ⌘ 点选 | 比较两个提交 |

## 开发

```bash
git clone https://github.com/leoheung/GitKleen.git && cd GitKleen   # 需要 Go 1.27 或更新版本
go tool mygo dev      # 运行，改动后自动重新加载
go tool mygo build -platform darwin/arm64,windows/amd64,linux/amd64,linux/arm64
```

界面文字在 [`locales/`](locales) 里：`en.json` 是源语言，`zh-Hans.json` 是简体中文，改完运行 `go tool mygo generate`。欢迎贡献其他语言的翻译：新增 `locales/<语言标签>.json` 即可。

发布：在 [`CHANGELOG.md`](CHANGELOG.md) 写好这个版本的说明（会成为 Release 的说明），再推送 `v0.2.0` 这样的标签。CI 会用仓库 Secret `MYGO_UPDATER_PRIVATE_KEY` 给更新包签名，打包 macOS 磁盘镜像、Windows 安装程序和 zip、Linux 安装包，并创建 Release。签名私钥（`mygo keygen` 生成的 `mygo-update.key` 的内容）保存在该 Secret 里，丢失后已安装的应用就收不到更新了。

SSH 远程：发布时 CI 会同时编译、签名 6 个平台的辅助程序（`gitkleen-helper-<系统>-<架构>.gz`）放进 Release，应用第一次连某种系统时下载并验签。开发时用 `go run ./internal/tools/helpers` 编到 `build/helpers`，再用 `GITKLEEN_HELPER_DIR=$PWD/build/helpers` 启动应用。

## 致谢

基于 `https://github.com/Binaryify/MyGit`（作者 Binaryify），使用 MIT 许可证。

## 许可证

[MIT](LICENSE)
