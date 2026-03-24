from __future__ import annotations

import argparse
import sys
from pathlib import Path

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete, print_report
from minimax_studio.model_catalog import (
    AUDIO_FORMATS,
    DEFAULT_MUSIC_MODEL,
    DEFAULT_TEXT_MODEL,
    DEFAULT_TTS_MODEL,
    DEFAULT_VIDEO_MODEL,
    DEFAULT_VOICE_ID,
    MUSIC_MODES,
    MUSIC_MODELS,
    PLANNING_TEXT_MODELS,
    SPEECH_MODELS,
    VIDEO_MODEL_CHOICES,
    VIDEO_RESOLUTIONS,
)
from minimax_studio.client import MiniMaxClient
from minimax_studio.schemas import MakeOptions
from minimax_studio.workflows.make import run_make


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(
        add_help=add_help,
        description="使用 MiniMax 的文本、图片、视频、语音与音乐模型生成完整短视频。",
    )
    parser.add_argument(
        "-t",
        "--theme",
        default="一艘小纸船在晨光湖面缓缓前行，像一个关于出发与希望的短片",
        help="最终视频的高层创意描述。",
    )
    parser.add_argument(
        "-s",
        "--scenes",
        "--scene-count",
        dest="scene_count",
        type=int,
        default=1,
        help="要生成的视频场景数。每个场景会消耗一次视频生成任务。",
    )
    parser.add_argument(
        "-d",
        "--duration",
        "--scene-duration",
        dest="scene_duration",
        type=int,
        default=6,
        help="每个生成视频场景的时长，单位为秒。",
    )
    parser.add_argument(
        "-a",
        "--aspect",
        "--aspect-ratio",
        dest="aspect_ratio",
        default="16:9",
        help="每个关键帧图片的宽高比。",
    )
    parser.add_argument(
        "-r",
        "--resolution",
        choices=VIDEO_RESOLUTIONS,
        metavar="RESOLUTION",
        default="768P",
        help="视频生成分辨率。",
    )
    parser.add_argument(
        "-T",
        "--text-model",
        default=DEFAULT_TEXT_MODEL,
        choices=PLANNING_TEXT_MODELS,
        metavar="TEXT_MODEL",
        help="用于生成创意规划的 MiniMax 文本模型。",
    )
    parser.add_argument(
        "--text-max-tokens",
        type=int,
        default=4096,
        help="兼容 Anthropic 接口的文本规划请求最大 token 数。",
    )
    parser.add_argument(
        "-m",
        "--video-model",
        default=DEFAULT_VIDEO_MODEL,
        choices=VIDEO_MODEL_CHOICES,
        metavar="VIDEO_MODEL",
        help="用于图生视频的 MiniMax 视频模型，支持 `Hailuo-2.3-768P` 这类别名。",
    )
    parser.add_argument(
        "-S",
        "--speech-model",
        "--tts-model",
        default=DEFAULT_TTS_MODEL,
        choices=SPEECH_MODELS,
        metavar="TTS_MODEL",
        help="用于生成旁白的 MiniMax 语音模型。",
    )
    parser.add_argument(
        "-M",
        "--music-model",
        default=DEFAULT_MUSIC_MODEL,
        choices=MUSIC_MODELS,
        metavar="MUSIC_MODEL",
        help="用于生成背景音乐的 MiniMax 音乐模型。",
    )
    parser.add_argument(
        "--music-mode",
        default="auto",
        choices=MUSIC_MODES,
        help="是否强制生成 MiniMax 音乐。若为 `auto`，当当前 API 套餐无音乐权限时会继续执行但跳过音乐生成。",
    )
    parser.add_argument(
        "--voice-id",
        default=DEFAULT_VOICE_ID,
        help="旁白使用的 Voice ID。",
    )
    parser.add_argument(
        "--audio-format",
        default="mp3",
        choices=AUDIO_FORMATS,
        help="旁白与音乐素材的音频格式。",
    )
    parser.add_argument(
        "-n",
        "--interval",
        "--poll-interval",
        dest="poll_interval",
        type=int,
        default=10,
        help="轮询视频生成任务状态的时间间隔，单位为秒。",
    )
    parser.add_argument(
        "-w",
        "--wait",
        "--max-wait",
        dest="max_wait",
        type=int,
        default=1800,
        help="每个视频生成任务的最长等待时间，单位为秒。",
    )
    parser.add_argument(
        "-l",
        "--language",
        default="中文",
        help="规划提示词中使用的旁白语言标签。",
    )
    parser.add_argument(
        "-o",
        "--output",
        "--output-dir",
        dest="output",
        default="runs/make",
        help="输出目录。",
    )
    parser.add_argument(
        "-i",
        "--input",
        "--input-video",
        dest="input_video",
        default="",
        help="可选的已有视频文件。提供后将复用该视频而不是调用 MiniMax 视频生成，建议配合 `--scene-count 1` 使用。",
    )
    parser.add_argument(
        "-x",
        "--optimize-image",
        "--image-prompt-optimizer",
        dest="image_prompt_optimizer",
        action="store_true",
        help="为图片生成启用提示词优化器。",
    )
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    args = parser.parse_args(argv)
    options = MakeOptions(
        theme=args.theme,
        scene_count=args.scene_count,
        scene_duration=args.scene_duration,
        aspect_ratio=args.aspect_ratio,
        resolution=args.resolution,
        text_model=args.text_model,
        text_max_tokens=args.text_max_tokens,
        video_model=args.video_model,
        tts_model=args.tts_model,
        music_model=args.music_model,
        music_mode=args.music_mode,
        voice_id=args.voice_id,
        audio_format=args.audio_format,
        poll_interval=args.poll_interval,
        max_wait=args.max_wait,
        language=args.language,
        output_dir=Path(args.output),
        input_video=Path(args.input_video).resolve() if args.input_video else None,
        image_prompt_optimizer=args.image_prompt_optimizer,
    )
    try:
        run_make(MiniMaxClient.from_env(), options, reporter=print_report)
        return 0
    except KeyboardInterrupt:
        print("已中断", file=sys.stderr)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
