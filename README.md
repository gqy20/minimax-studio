# minimax-studio

基于 Go 的 MiniMax CLI 工具箱，支持 6 类工作流。

## 功能

| 命令 | 说明 |
|------|------|
| `ms clip` | 图生视频 |
| `ms plan` | 分镜规划 |
| `ms voice` | 语音合成 |
| `ms music` | 背景音乐生成 |
| `ms stitch` | 素材合成 |
| `ms make` | 端到端短片生成 |
| `ms quota` | 额度查询 |
| `ms server` | HTTP API Server（供前端调用） |

## 快速开始

### 安装

```bash
# 安装依赖
make deps

# 构建
make build

# 或直接运行
make run -- clip -p "prompt" -s "subject" -o output/clip
```

### CLI 使用

```bash
# 图生视频
./bin/ms clip -p "A paper boat at sunrise" -s "The paper boat drifts slowly" -o output/clip

# 素材合成
./bin/ms stitch -v video1.mp4 -v video2.mp4 -n narration.mp3 -m music.mp3 -o output/final.mp4

# 语音合成
./bin/ms voice -t "这是一段旁白" -o output/voice.mp3

# 背景音乐
./bin/ms music -p "warm cinematic piano, no vocals" -o output/music.mp3

# 额度查询
./bin/ms quota

# JSON 输出（便于脚本解析）
./bin/ms clip -p "x" -s "y" -o z --json
```

### API Server

```bash
# 启动服务器
./bin/ms server --port 8080

# API 端点
# POST /api/v1/clip
# POST /api/v1/plan
# POST /api/v1/voice
# POST /api/v1/music
# POST /api/v1/stitch
# POST /api/v1/make
# GET  /api/v1/quota
# GET  /api/v1/jobs/:id
```

## 环境变量

```bash
export MINIMAX_API_KEY=your_api_key
export MINIMAX_GROUP_ID=your_group_id
```

## Frontend MVP

仓库内已包含一个独立前端骨架，位于 `frontend/`，当前聚焦以下能力：

- 提交 `make` 任务
- 轮询 `jobs/:id`
- 查看 quota
- 预览最终视频、旁白、音乐和 `plan.json`

### 启动方式

先启动 Go server：

```bash
export MINIMAX_API_KEY=your_api_key
./bin/ms server --port 8080
```

再启动前端：

```bash
cd frontend
npm install
npm run dev
```

默认开发地址为 `http://localhost:3000`，并通过 Vite 代理转发到 `http://localhost:8080`。

如果前端和后端不在同一地址，可设置：

```bash
export VITE_API_BASE_URL=http://localhost:8080
```

## 构建

```bash
# 当前平台
make build

# 所有平台
make build-all

# 清理
make clean
```

## 项目结构

```
minimax-studio/
├── cmd/ms/          # CLI 入口
├── internal/
│   ├── api/         # HTTP API Server
│   ├── client/      # MiniMax API 客户端
│   ├── media/       # ffmpeg 封装
│   ├── schemas/     # 数据结构
│   └── workflows/   # 业务工作流
├── Makefile
└── README.md
```

## 历史版本

Python 版本请查看 `python-version` 分支。
