from __future__ import annotations

import argparse
import sys

from argcomplete.shell_integration import shellcode
from minimax_studio.cli import ChineseArgumentParser


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(add_help=add_help, description="输出适用于 ms 的 shell 补全脚本。")
    parser.add_argument("-s", "--shell", choices=["bash", "zsh"], default="zsh", help="目标 shell。")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    sys.stdout.write(shellcode(["ms"], shell=args.shell))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
