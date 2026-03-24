from __future__ import annotations

import argparse
import sys
from pathlib import Path

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete, print_report
from minimax_studio.model_catalog import DEFAULT_VIDEO_MODEL, VIDEO_MODEL_CHOICES, VIDEO_RESOLUTIONS
from minimax_studio.client import MiniMaxClient
from minimax_studio.schemas import ClipOptions
from minimax_studio.workflows.clip import run_clip


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(
        add_help=add_help,
        description="使用 MiniMax 生成关键帧图片，再将其转换为短视频。",
    )
    parser.add_argument(
        "-i",
        "--image",
        "--image-prompt",
        dest="image_prompt",
        default="A cinematic close-up of a small paper boat floating on a calm tea-colored lake at sunrise, soft golden light, detailed ripples, photorealistic.",
        help="图片生成阶段使用的提示词。",
    )
    parser.add_argument(
        "-p",
        "--prompt",
        "--video-prompt",
        dest="video_prompt",
        default="The paper boat drifts slowly forward on gentle ripples while the camera makes a subtle push-in. Soft morning light shimmers on the water, cinematic and natural motion.",
        help="图生视频阶段使用的提示词。",
    )
    parser.add_argument(
        "-a",
        "--aspect",
        "--aspect-ratio",
        dest="aspect_ratio",
        default="16:9",
        help="生成图片的宽高比。",
    )
    parser.add_argument(
        "-m",
        "--model",
        "--video-model",
        dest="video_model",
        default=DEFAULT_VIDEO_MODEL,
        choices=VIDEO_MODEL_CHOICES,
        metavar="VIDEO_MODEL",
        help="使用的视频模型。",
    )
    parser.add_argument(
        "-d",
        "--duration",
        type=int,
        default=6,
        help="视频时长，单位为秒。",
    )
    parser.add_argument(
        "-r",
        "--resolution",
        default="768P",
        choices=VIDEO_RESOLUTIONS,
        metavar="RESOLUTION",
        help="视频分辨率。",
    )
    parser.add_argument(
        "-n",
        "--interval",
        "--poll-interval",
        dest="poll_interval",
        type=int,
        default=10,
        help="轮询视频任务状态的时间间隔，单位为秒。",
    )
    parser.add_argument(
        "-w",
        "--wait",
        "--max-wait",
        dest="max_wait",
        type=int,
        default=1800,
        help="等待视频生成完成的最长时间，单位为秒。",
    )
    parser.add_argument(
        "-o",
        "--output",
        "--output-prefix",
        dest="output",
        default="runs/clip",
        help="输出文件前缀，最终会写出 `.jpg` 和 `.mp4` 文件。",
    )
    parser.add_argument(
        "-x",
        "--optimize-image",
        "--image-prompt-optimizer",
        dest="image_prompt_optimizer",
        action="store_true",
        help="为图片生成阶段启用提示词优化器。",
    )
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    args = parser.parse_args(argv)
    options = ClipOptions(
        image_prompt=args.image_prompt,
        video_prompt=args.video_prompt,
        aspect_ratio=args.aspect_ratio,
        video_model=args.video_model,
        duration=args.duration,
        resolution=args.resolution,
        poll_interval=args.poll_interval,
        max_wait=args.max_wait,
        output_prefix=Path(args.output),
        image_prompt_optimizer=args.image_prompt_optimizer,
    )
    try:
        run_clip(MiniMaxClient.from_env(), options, reporter=print_report)
        return 0
    except KeyboardInterrupt:
        print("已中断", file=sys.stderr)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
