---
name: version-upgrade
description: Guides GitKleen releases and the CHANGELOG check that gates them. Use when asked to release or bump a version, push a version tag, or fix a failed Release workflow. Do not use for everyday feature work.
---

# GitKleen 版本发布（version upgrade）

发布只由一个动作触发：push 一个 `v` 开头的 tag（流程见 `.github/workflows/release.yml`）。版本号取自 tag 名：`v0.2.3` → `0.2.3`，workflow 用 jq 写进 mygo.json；平时 mygo.json 的 version 保持 `0.2.0-dev`，发布时不用手动改。

## 教训（来自 v0.2.2 的失败）

`go tool mygo build` 在 Release workflow 里会校验 CHANGELOG.md 存在 `## <版本号>` 小节（精确格式：`## 0.2.3`，无 v 前缀、无后缀），没有就整个构建失败：

```
error: CHANGELOG.md has no notes for 0.2.2: add a "## 0.2.2" section
```

publish job 还会用 awk 提取该小节作为 release notes，小节为空同样失败。**先写 CHANGELOG 并提交，再打 tag。**顺序错了会浪费一整轮 CI 构建。

## 发布清单（按顺序，第 1-3 步必须在 push tag 之前完成）

1. 起草更新说明：在 CHANGELOG.md 顶部（`# Changelog` 之后、上一版小节之前）加 `## x.y.z` 小节，中文条目风格参照上一版；内容覆盖 `git log <上个tag>..HEAD` 的全部提交。
2. 提交 CHANGELOG。
3. 自检：CHANGELOG.md 中存在且仅存在一个该版本的 `## x.y.z` 标题（PowerShell 可用 `Select-String -Path CHANGELOG.md -Pattern '^## x\.y\.z$'` 验证）。
4. 打 tag：`git tag vx.y.z`，打在包含该 CHANGELOG 的提交上。
5. 推送：`git push origin vx.y.z`，触发 Release workflow（unix、windows、publish 三个 job）。

## 规则

- 版本号带 `-`（如 `v0.2.0-rc.1`）发布为 pre-release，不会成为 latest；正式版版本号不要带后缀。
- 已推送且构建失败的 tag：删除（`git tag -d vx.y.z` + `git push --delete origin vx.y.z`）并跳号发下一版（v0.2.2 失败 → 发 v0.2.3），不复用失败过的版本号。
- 发布后到 GitHub Actions 确认三个 job 全绿，Releases 页出现新版本和安装包。
- 本机构建环境：`go` 不在 PATH，用 `C:\Program Files\Go\bin\go.exe`；默认 GOPROXY 连不上时先设 `$env:GOPROXY='https://goproxy.cn,direct'`。

## 相关

- 发布流程细节：`.github/workflows/release.yml`
- 更新说明格式参照：CHANGELOG.md 顶部的小节（0.2.1、0.2.0）
