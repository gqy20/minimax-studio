from __future__ import annotations

import argparse
import json

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete
from minimax_studio.client import MiniMaxClient


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(add_help=add_help, description="查看当前 MiniMax 配额使用情况与剩余额度。")
    parser.add_argument("-j", "--json", action="store_true", help="输出适合机器读取的 JSON。")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    args = parser.parse_args(argv)
    result = MiniMaxClient.from_env().get_coding_plan_remains()

    if args.json:
        print(
            json.dumps(
                {
                    "entries": [
                        {
                            "model_name": entry.model_name,
                            "current_interval_total_count": entry.current_interval_total_count,
                            "current_interval_usage_count": entry.current_interval_usage_count,
                            "current_interval_remaining_count": max(
                                0, entry.current_interval_total_count - entry.current_interval_usage_count
                            ),
                            "current_interval_start": entry.start_time.isoformat(sep=" ", timespec="seconds"),
                            "current_interval_end": entry.end_time.isoformat(sep=" ", timespec="seconds"),
                            "weekly_total_count": entry.current_weekly_total_count,
                            "weekly_usage_count": entry.current_weekly_usage_count,
                            "weekly_remaining_count": max(
                                0, entry.current_weekly_total_count - entry.current_weekly_usage_count
                            ),
                            "weekly_start": entry.weekly_start_time.isoformat(sep=" ", timespec="seconds"),
                            "weekly_end": entry.weekly_end_time.isoformat(sep=" ", timespec="seconds"),
                        }
                        for entry in result.entries
                    ]
                },
                ensure_ascii=False,
                indent=2,
            ),
            flush=True,
        )
        return 0

    print("MiniMax quota remains", flush=True)
    for entry in result.entries:
        current_remaining = max(0, entry.current_interval_total_count - entry.current_interval_usage_count)
        weekly_remaining = max(0, entry.current_weekly_total_count - entry.current_weekly_usage_count)
        current_ratio = _format_ratio(entry.current_interval_usage_count, entry.current_interval_total_count)
        weekly_ratio = _format_ratio(entry.current_weekly_usage_count, entry.current_weekly_total_count)

        print(f"\n{entry.model_name}", flush=True)
        print(
            f"  current: {entry.current_interval_usage_count}/{entry.current_interval_total_count} used"
            f" ({current_ratio}), remaining {current_remaining}",
            flush=True,
        )
        print(
            f"  window: {entry.start_time:%Y-%m-%d %H:%M:%S} -> {entry.end_time:%Y-%m-%d %H:%M:%S}",
            flush=True,
        )
        print(
            f"  weekly: {entry.current_weekly_usage_count}/{entry.current_weekly_total_count} used"
            f" ({weekly_ratio}), remaining {weekly_remaining}",
            flush=True,
        )
        print(
            f"  week: {entry.weekly_start_time:%Y-%m-%d %H:%M:%S} -> {entry.weekly_end_time:%Y-%m-%d %H:%M:%S}",
            flush=True,
        )
    return 0


def _format_ratio(used: int, total: int) -> str:
    if total <= 0:
        return "n/a"
    return f"{used / total:.0%}"


if __name__ == "__main__":
    raise SystemExit(main())
