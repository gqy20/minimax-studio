# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.1.0] - 2026-04-02

### Added

- 首次公开发布 `ms` 命令行工具。
- CLI 子命令：`clip`（图生视频）、`plan`（分镜规划）、`voice`（语音合成）、`music`（背景音乐）、`stitch`（素材合成）、`make`（端到端短片生成）、`quota`（额度查询）。
- HTTP API Server 模式（`ms server`），支持前端集成。
- 前端 MVP 面板，支持提交任务、轮询进度、预览产物。
- 单二进制打包：通过 `go:embed` 将前端嵌入 Go 二进制，`make build-release` 一键构建。
- ffmpeg 启动检测：server 启动时检查 ffmpeg/ffprobe，缺失时给出友好提示。
- 自动打开浏览器：server 启动后自动打开 `http://localhost:<port>`，支持 `--open=false` 关闭。
- GitHub Actions CI/CD：PR/push 触发验证 + 跨平台构建（linux/darwin/windows，含 arm64）。
- GitHub Actions Release：推送 `v*` tag 自动构建压缩包、生成 checksums 并发布 GitHub Release。

[0.1.0]: https://github.com/gqy20/minimax-studio/releases/tag/v0.1.0
