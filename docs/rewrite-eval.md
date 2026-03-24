# Rewrite Evaluation

## Conclusion

当前项目可以用 Go 或 Rust 重构，但不建议现在直接全量重写。

更合适的顺序是：

1. 继续用 Python 稳定 CLI、命名和输出约定
2. 如果目标变成“单文件分发、跨平台安装更轻”，优先做 Go 版
3. 只有在后面要引入更重的本地媒体处理、任务系统、插件系统或长期守护进程时，再考虑 Rust

核心原因很简单：

- 目前主要耗时在 MiniMax API 和 `ffmpeg` 子进程，不在 Python 运行时
- 当前代码是“API 编排 + 文件落盘 + 外部命令调用”，这类工具改成 Go 的收益明显高于改成 Rust
- Rust 能做得更硬，但当前场景下复杂度收益比不高

## Why Go First

Go 更适合这个项目的下一阶段，原因有四个：

- CLI 工具链成熟，`cobra` / `urfave/cli` 这类方案很好用
- HTTP、JSON、轮询、超时控制、并发任务都很直接
- 跨平台单二进制分发简单
- 团队后续维护成本通常低于 Rust

这个项目天然适合 Go 的部分：

- `ms clip`
- `ms make`
- `ms plan`
- `ms voice`
- `ms music`
- `ms stitch`
- `ms quota`

这些命令本质都是：

- 解析参数
- 调 MiniMax API
- 轮询任务状态
- 调 `ffmpeg`
- 写文件

这正是 Go 的强项。

## When Rust Makes Sense

Rust 不是不能做，而是要在下面这些前提成立时才更值：

- 你准备把媒体处理逻辑从“调用 ffmpeg”升级成更多本地处理
- 你准备做长期运行的后台 worker / 队列系统
- 你很在意更强的类型约束和更严格的错误边界
- 你接受更高的开发和维护成本

如果项目未来只是一个高质量 CLI，Rust 不是最优先选择。

## Suggested Go Layout

推荐 Go 版按下面的结构来：

```text
cmd/ms/main.go

internal/cli/
  root.go
  clip.go
  make.go
  plan.go
  voice.go
  music.go
  stitch.go
  models.go
  quota.go
  completion.go

internal/minimax/
  client.go
  text.go
  image.go
  video.go
  audio.go
  quota.go

internal/catalog/
  models.go

internal/schema/
  plan.go
  quota.go
  workflow.go

internal/workflow/
  clip.go
  make.go
  generate.go
  stitch.go

internal/media/
  ffmpeg.go
  probe.go

internal/fsx/
  paths.go
  write.go
```

职责映射关系：

- `cli/` 对应现在的 `src/minimax_studio/cli/`
- `minimax/` 对应现在的 `client.py`
- `catalog/` 对应现在的 `model_catalog.py`
- `schema/` 对应现在的 `schemas.py`
- `workflow/` 对应现在的 `workflows/`
- `media/` 对应现在的 `media.py`
- `fsx/` 对应现在的 `files.py`

## Migration Order

如果要重构，我建议按下面的顺序做，而不是一次性重写。

### Phase 1

先做纯 CLI 外壳，不接真实 API：

- 建 `ms clip` / `ms make` / `ms plan` 等命令树
- 保持和当前 Python 版参数兼容
- 跑通 `--help`
- 跑通配置读取和输出路径解析

### Phase 2

迁移无状态模块：

- `model_catalog.py`
- `schemas.py`
- `files.py`

这是最低风险的一步。

### Phase 3

迁移 MiniMax client：

- 统一 HTTP client
- 统一超时与错误处理
- 统一 Anthropic 文本接口、普通 API、openplatform API

这一层是 Go 重构的真正核心。

### Phase 4

迁移 workflow：

- `clip`
- `plan`
- `voice`
- `music`
- `stitch`
- 最后再迁 `make`

`make` 应该最后迁，因为它依赖前面几乎所有能力。

### Phase 5

补发行能力：

- `goreleaser`
- 多平台构建
- shell completion
- Homebrew / Scoop / Linux package

## Compatibility Target

如果要做 Go 版，建议把下面这些视为必须兼容：

- 命令名兼容：`ms clip`、`ms make` 等
- 短参数兼容：`-o`、`-m`、`-T`、`-S`、`-M`
- 输出目录兼容：`runs/`
- 产物命名兼容：`clip01.jpg`、`clip01.mp4`、`voice.txt`、`voice.mp3`、`edit.mp4`、`timed.mp4`、`final.mp4`

这样才能平滑替换。

## Recommendation

建议采用下面这个决策：

- 短期：继续维护 Python 主线
- 中期：启动 Go 分支，先做并行 PoC
- 长期：只有当 Go 版功能、参数、输出约定都稳定后，再考虑替换 Python 主入口

不建议现在直接转 Rust，也不建议现在直接停止 Python 开发去做全量重写。
