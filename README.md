# minimax-studio

`minimax-studio` 是一个基于 `uv` 的 MiniMax CLI 工具箱，聚焦 6 类工作流：

- 文生分镜规划
- 图生短视频
- 旁白语音合成
- 背景音乐生成
- 现有音视频素材合成
- 端到端短片生成

统一入口是 `ms`，同时保留一组便于直用的短脚本：

- `ms-clip`
- `ms-make`
- `ms-plan`
- `ms-voice`
- `ms-music`
- `ms-stitch`
- `ms-models`
- `ms-quota`
- `ms-completion`

## 先看效果

GitHub 仓库里展示视频时，通常使用“缩略图 + 相对路径视频链接”的方式，避免 README 依赖外部附件。下面这几段测试视频已经整理到 `docs/assets/`，可以直接点开查看。

[![单镜头图生视频 Demo](docs/assets/images/clip-demo-cover.jpeg)](docs/assets/videos/clip-demo.mp4)
[![完整流程 Demo](docs/assets/images/full-pipeline-demo-cover.jpeg)](docs/assets/videos/full-pipeline-demo.mp4)

- 单镜头图生视频：[`docs/assets/videos/clip-demo.mp4`](docs/assets/videos/clip-demo.mp4)
- 完整流程成片：[`docs/assets/videos/full-pipeline-demo.mp4`](docs/assets/videos/full-pipeline-demo.mp4)
- 复用已有视频再合成：[`docs/assets/videos/reuse-input-video-demo.mp4`](docs/assets/videos/reuse-input-video-demo.mp4)

这些演示素材来自仓库内已有测试产物：

- `runs/minimax_demo_video.mp4`
- `runs/minimax_full_video_demo/final_video.mp4`
- `runs/minimax_full_video_scene1_reuse/final_video.mp4`

## 这个仓库能做什么

| 入口 | 等价短脚本 | 作用 | 典型输出 |
| --- | --- | --- | --- |
| `uv run ms clip` | `uv run ms-clip` | 先生成关键帧，再生成单条短视频 | `xxx.jpg`、`xxx.mp4` |
| `uv run ms plan` | `uv run ms-plan` | 只做分镜和旁白文案规划 | `plan.json`、`voice.txt` |
| `uv run ms voice` | `uv run ms-voice` | 只做旁白语音合成 | `voice.mp3` / `voice.wav` / `voice.flac` |
| `uv run ms music` | `uv run ms-music` | 只做背景音乐生成 | `music.mp3` / `music.wav` / `music.flac` |
| `uv run ms stitch` | `uv run ms-stitch` | 把已有视频、旁白、音乐合成最终视频 | `edit.mp4`、`timed.mp4`、最终成片 |
| `uv run ms make` | `uv run ms-make` | 从规划到成片的一条龙流程 | `plan.json`、`voice.txt`、`s01.jpg`、`s01.mp4`、`voice.mp3`、`music.mp3`、`final.mp4` |
| `uv run ms models` | `uv run ms-models` | 列出项目内记录的模型组与默认值 | 控制台输出 |
| `uv run ms quota` | `uv run ms-quota` | 查询当前 key 的额度信息 | 控制台输出 / JSON |
| `uv run ms completion` | `uv run ms-completion` | 生成 shell 自动补全脚本 | 控制台输出 |

如果你只记一个入口，记住 `ms` 就够了；短脚本只是为了少打字。

## 快速开始

### 依赖

- Python `>= 3.12`
- `uv`
- `ffmpeg` 和 `ffprobe`
  说明：`ms make` 和 `ms stitch` 在合成视频时会直接调用它们
- MiniMax API Key

### 安装

```bash
uv sync
export MINIMAX_API_KEY=your_key_here
uv run ms -h
```

如果你还没装 `ffmpeg`：

```bash
# macOS
brew install ffmpeg

# Ubuntu / Debian
sudo apt-get update
sudo apt-get install -y ffmpeg
```

### 建议先跑这几个命令确认环境

```bash
uv run ms -h
uv run ms models
uv run ms quota
uv run ms completion -s zsh > ~/.zfunc/_ms
```

## 常用工作流

### 1. 只生成一个图生视频片段

```bash
uv run ms clip \
  -i "A paper boat at sunrise, photorealistic, cinematic lighting" \
  -p "The paper boat drifts slowly forward with gentle ripples" \
  -o runs/clip01
```

输出：

- `runs/clip01.jpg`
- `runs/clip01.mp4`

### 2. 只做分镜和旁白文案规划

```bash
uv run ms plan \
  -t "一艘小纸船在晨光湖面缓缓前行，像一个关于出发与希望的短片" \
  -s 1 \
  -d 6 \
  -o runs/plan01
```

输出：

- `runs/plan01/plan.json`
- `runs/plan01/voice.txt`

### 3. 只生成旁白音频

```bash
uv run ms voice "这是一段旁白" -o runs/voice.mp3
```

### 4. 只生成背景音乐

```bash
uv run ms music "warm cinematic piano, no vocals" -o runs/music.mp3
```

### 5. 合成已有素材

```bash
uv run ms stitch \
  runs/clip01.mp4 \
  -n runs/voice.mp3 \
  -m runs/music.mp3 \
  -o runs/stitch-demo/final.mp4
```

同目录下还会生成：

- `runs/stitch-demo/edit.mp4`
- `runs/stitch-demo/timed.mp4`

### 6. 一条命令跑完整流程

```bash
uv run ms make \
  -t "纸船晨光短片" \
  -s 1 \
  -d 6 \
  -o runs/make01
```

典型输出：

- `runs/make01/plan.json`
- `runs/make01/voice.txt`
- `runs/make01/s01.jpg`
- `runs/make01/s01.mp4`
- `runs/make01/voice.mp3`
- `runs/make01/music.mp3`
- `runs/make01/edit.mp4`
- `runs/make01/timed.mp4`
- `runs/make01/final.mp4`

### 7. 复用已有视频，跳过视频生成

当你已经有一条视频，只想让这个工具补上规划、旁白、音乐和最终合成时，可以这样跑：

```bash
uv run ms make \
  -t "纸船晨光短片" \
  --scene-count 1 \
  --input-video runs/minimax_demo_video.mp4 \
  -o runs/reuse-demo
```

说明：

- 当前 `--input-video` 需要配合 `--scene-count 1`
- 这是源码里明确校验过的行为，不是 README 的额外约定

## 输出目录结构

下面这个结构更接近当前脚本真实行为：

```text
runs/
├── clip01.jpg
├── clip01.mp4
├── plan01/
│   ├── plan.json
│   └── voice.txt
├── voice.mp3
├── music.mp3
├── stitch-demo/
│   ├── edit.mp4
│   ├── timed.mp4
│   └── final.mp4
└── make01/
    ├── plan.json
    ├── voice.txt
    ├── s01.jpg
    ├── s01.mp4
    ├── voice.mp3
    ├── music.mp3
    ├── edit.mp4
    ├── timed.mp4
    └── final.mp4
```

历史测试产物统一放在 `runs/` 下，README 展示素材整理在 `docs/assets/` 下。

## 模型与默认值

当前代码里的默认模型如下：

- 文本规划模型：`MiniMax-M2.7-highspeed`
- 视频模型：`MiniMax-Hailuo-2.3-Fast`
- 图片模型：`image-01`
- 语音模型：`speech-2.8-hd`
- 音乐模型：`music-2.5`
- 默认语音 ID：`male-qn-qingse`

补充说明：

- `ms plan` 和 `ms make` 的规划阶段走的是 Anthropic 兼容消息接口
- `ms quota` 当前查询的是源码中 `/coding_plan/remains` 这个接口，返回内容更适合理解为 Coding Plan 相关额度

## MiniMax 官方验证链接

下面这些链接是为了方便你对照 README 和源码自己核验能力边界、模型和接口。

- MiniMax 开放平台首页：https://platform.minimaxi.com/
- API 文档首页：https://platform.minimaxi.com/docs
- API Overview：https://platform.minimaxi.com/docs/api-reference/api-overview
- Pricing Overview：https://platform.minimaxi.com/docs/pricing/overview
- 按量计费说明：https://platform.minimaxi.com/docs/pricing/pay-as-you-go
- Coding Plan 说明：https://platform.minimaxi.com/docs/pricing/coding-plan
- 文本生成 / Anthropic 兼容接口：https://platform.minimaxi.com/docs/api-reference/text-anthropic-api
- 图像生成：https://platform.minimaxi.com/docs/api-reference/image-generation-intro
- 视频生成：https://platform.minimaxi.com/docs/api-reference/video-generation-intro
- 同步语音合成（T2A HTTP）：https://platform.minimaxi.com/docs/api-reference/speech-t2a-http
- 音乐生成：https://platform.minimaxi.com/docs/api-reference/music-generation
- API Key 与常见问题：https://platform.minimaxi.com/docs/faq/about-apis
- Coding Plan FAQ（含额度相关说明）：https://platform.minimaxi.com/docs/coding-plan/faq

如果这些官方页面后续调整了路径，请以 MiniMax 开放平台导航为准。

## 代码结构

```text
src/minimax_studio/
├── cli/          # 命令行入口
├── client.py     # MiniMax API 调用封装
├── files.py      # 文件写入与目录处理
├── media.py      # ffmpeg / ffprobe 媒体处理
├── model_catalog.py
├── schemas.py
└── workflows/    # clip / plan / voice / music / stitch / make 的具体流程
```

## 其他文档

- Go / Rust 重构评估：[docs/rewrite-eval.md](docs/rewrite-eval.md)
