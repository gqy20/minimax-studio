from __future__ import annotations

import argparse
import sys
from dataclasses import dataclass

from minimax_studio.cli import ChineseArgumentParser, completion, enable_argcomplete, models, music, plan, quota, stitch
from minimax_studio.cli import clip, make, voice


@dataclass(frozen=True)
class CommandSpec:
    module: object
    aliases: tuple[str, ...]
    help: str
    description: str


COMMANDS: dict[str, CommandSpec] = {
    "clip": CommandSpec(
        module=clip,
        aliases=("i2v", "image", "i"),
        help="根据提示词生成短视频片段。",
        description="先生成关键帧图片，再将其转换为短视频片段。",
    ),
    "make": CommandSpec(
        module=make,
        aliases=("full", "video", "f"),
        help="生成完整视频流程。",
        description="从规划到最终合成，执行完整的视频生成流程。",
    ),
    "plan": CommandSpec(
        module=plan,
        aliases=("story", "p"),
        help="仅生成分镜规划。",
        description="只生成分镜与旁白文案规划。",
    ),
    "voice": CommandSpec(
        module=voice,
        aliases=("tts", "t"),
        help="仅生成旁白音频。",
        description="只生成旁白音频。",
    ),
    "music": CommandSpec(
        module=music,
        aliases=("score", "m"),
        help="仅生成背景音乐。",
        description="只生成背景音乐。",
    ),
    "stitch": CommandSpec(
        module=stitch,
        aliases=("compose", "render", "s"),
        help="合成已有素材。",
        description="将已有的视频与音频素材合成为最终视频。",
    ),
    "models": CommandSpec(
        module=models,
        aliases=("ls",),
        help="列出项目使用的模型分组。",
        description="列出本项目记录的 MiniMax 模型分组及默认值。",
    ),
    "quota": CommandSpec(
        module=quota,
        aliases=("usage", "q"),
        help="查看配额使用情况。",
        description="查看当前 MiniMax 配额使用情况与剩余额度。",
    ),
    "completion": CommandSpec(
        module=completion,
        aliases=("comp",),
        help="输出 shell 补全脚本。",
        description="输出适用于 ms 的 shell 补全脚本。",
    ),
}

ALIASES = {
    alias: command
    for command, spec in COMMANDS.items()
    for alias in spec.aliases
}


def build_parser() -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(prog="ms", description="MiniMax Studio 命令行工具。")
    subparsers = parser.add_subparsers(dest="command", title="子命令")

    for command, spec in COMMANDS.items():
        subparsers.add_parser(
            command,
            aliases=list(spec.aliases),
            parents=[spec.module.build_parser(add_help=False)],
            help=spec.help,
            description=spec.description,
        )
    return parser


def main(argv: list[str] | None = None) -> int:
    args = list(sys.argv[1:] if argv is None else argv)
    parser = build_parser()
    enable_argcomplete(parser)

    if not args:
        parser.print_help()
        return 0
    if args[0] in {"-h", "--help"}:
        parser.print_help()
        return 0

    command, rest = args[0], args[1:]
    target = command if command in COMMANDS else ALIASES.get(command)
    if target is not None:
        return COMMANDS[target].module.main(rest)

    parser.error(f"未知命令: {command}")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
