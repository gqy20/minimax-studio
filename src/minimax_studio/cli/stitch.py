from __future__ import annotations

import argparse
import sys
from pathlib import Path

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete, print_report
from minimax_studio.schemas import StitchOptions
from minimax_studio.workflows.stitch import run_stitch


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(add_help=add_help, description="将生成好的素材合成为最终视频。")
    parser.add_argument("videos", nargs="+", help="输入视频文件。")
    parser.add_argument("-n", "--narration", required=True, help="旁白音频文件。")
    parser.add_argument("-m", "--music", default="", help="可选的背景音乐文件。")
    parser.add_argument("-o", "--output", default="runs/final.mp4", help="最终视频输出路径。")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    args = parser.parse_args(argv)
    options = StitchOptions(
        video_paths=[Path(item).resolve() for item in args.videos],
        narration_path=Path(args.narration).resolve(),
        music_path=Path(args.music).resolve() if args.music else None,
        output_path=Path(args.output),
    )
    try:
        run_stitch(options, reporter=print_report)
        return 0
    except KeyboardInterrupt:
        print("已中断", file=sys.stderr)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
