from __future__ import annotations

import argparse
import sys
from pathlib import Path

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete, print_report
from minimax_studio.model_catalog import DEFAULT_TEXT_MODEL, PLANNING_TEXT_MODELS
from minimax_studio.client import MiniMaxClient
from minimax_studio.schemas import PlanOptions
from minimax_studio.workflows.generate import run_plan


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(add_help=add_help, description="生成分镜与旁白文案规划。")
    parser.add_argument("-t", "--theme", default="一艘小纸船在晨光湖面缓缓前行，像一个关于出发与希望的短片", help="创意主题描述。")
    parser.add_argument("-s", "--scenes", "--scene-count", dest="scene_count", type=int, default=1, help="场景数量。")
    parser.add_argument("-d", "--duration", "--scene-duration", dest="scene_duration", type=int, default=6, help="每个场景的时长，单位为秒。")
    parser.add_argument("-l", "--language", default="中文", help="旁白语言。")
    parser.add_argument(
        "-T",
        "--text-model",
        default=DEFAULT_TEXT_MODEL,
        choices=PLANNING_TEXT_MODELS,
        metavar="TEXT_MODEL",
        help="用于规划的文本模型。",
    )
    parser.add_argument("--text-max-tokens", type=int, default=4096, help="规划请求的最大 token 数。")
    parser.add_argument("-o", "--output", "--output-dir", dest="output", default="runs/plan", help="输出目录。")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    args = parser.parse_args(argv)
    options = PlanOptions(
        theme=args.theme,
        scene_count=args.scene_count,
        scene_duration=args.scene_duration,
        language=args.language,
        text_model=args.text_model,
        text_max_tokens=args.text_max_tokens,
        output_dir=Path(args.output),
    )
    try:
        run_plan(MiniMaxClient.from_env(), options, reporter=print_report)
        return 0
    except KeyboardInterrupt:
        print("已中断", file=sys.stderr)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
