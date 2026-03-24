from __future__ import annotations

from collections.abc import Callable

__all__ = ["emit_report"]


def emit_report(reporter: Callable[[str], None] | None, message: str) -> None:
    if reporter is not None:
        reporter(message)
