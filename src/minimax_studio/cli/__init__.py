from __future__ import annotations

import argparse

__all__ = ["ChineseArgumentParser", "enable_argcomplete", "print_report"]


class ChineseArgumentParser(argparse.ArgumentParser):
    def __init__(self, *args, **kwargs) -> None:
        add_help = kwargs.pop("add_help", True)
        super().__init__(*args, add_help=False, **kwargs)
        self._positionals.title = "位置参数"
        self._optionals.title = "选项"
        if add_help:
            self.add_argument(
                "-h",
                "--help",
                action="help",
                default=argparse.SUPPRESS,
                help="显示此帮助信息并退出",
            )

    def format_help(self) -> str:
        formatter = self._get_formatter()
        formatter.add_usage(
            self.usage,
            self._actions,
            self._mutually_exclusive_groups,
            prefix="用法: ",
        )
        formatter.add_text(self.description)

        for action_group in self._action_groups:
            formatter.start_section(action_group.title)
            formatter.add_text(action_group.description)
            formatter.add_arguments(action_group._group_actions)
            formatter.end_section()

        formatter.add_text(self.epilog)
        return formatter.format_help()

    def format_usage(self) -> str:
        formatter = self._get_formatter()
        formatter.add_usage(
            self.usage,
            self._actions,
            self._mutually_exclusive_groups,
            prefix="用法: ",
        )
        return formatter.format_help()


def enable_argcomplete(parser: argparse.ArgumentParser) -> None:
    try:
        import argcomplete
    except ImportError:
        return

    argcomplete.autocomplete(parser)


def print_report(message: str) -> None:
    print(message, flush=True)
