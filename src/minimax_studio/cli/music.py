from __future__ import annotations

import argparse
import sys
from pathlib import Path

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete, print_report
from minimax_studio.model_catalog import AUDIO_FORMATS, DEFAULT_MUSIC_MODEL, MUSIC_MODELS
from minimax_studio.client import MiniMaxClient
from minimax_studio.schemas import MusicOptions
from minimax_studio.workflows.generate import run_music


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(add_help=add_help, description="生成背景音乐。")
    parser.add_argument("prompt", help="音乐提示词。")
    parser.add_argument("-o", "--output", default="runs/music.mp3", help="输出音频路径。")
    parser.add_argument(
        "-m",
        "--model",
        default=DEFAULT_MUSIC_MODEL,
        choices=MUSIC_MODELS,
        metavar="MUSIC_MODEL",
        help="音乐模型。",
    )
    parser.add_argument("-f", "--format", "--audio-format", dest="audio_format", choices=AUDIO_FORMATS, default="mp3", help="音频格式。")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    args = parser.parse_args(argv)
    options = MusicOptions(
        prompt=args.prompt,
        output_path=Path(args.output),
        model=args.model,
        audio_format=args.audio_format,
    )
    try:
        run_music(MiniMaxClient.from_env(), options, reporter=print_report)
        return 0
    except KeyboardInterrupt:
        print("已中断", file=sys.stderr)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
