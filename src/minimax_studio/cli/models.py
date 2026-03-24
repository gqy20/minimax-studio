from __future__ import annotations

import argparse

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete
from minimax_studio import model_catalog


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    return ChineseArgumentParser(add_help=add_help, description="列出本项目使用的 MiniMax 模型分组。")


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    parser.parse_args(argv)

    print("Text models:", flush=True)
    for name in model_catalog.TEXT_MODELS:
        print(f"  - {name}", flush=True)

    print("\nPlanning text models:", flush=True)
    for name in model_catalog.PLANNING_TEXT_MODELS:
        print(f"  - {name}", flush=True)

    print("\nVideo models:", flush=True)
    for name in model_catalog.VIDEO_MODELS:
        print(f"  - {name}", flush=True)

    print("\nSpeech models:", flush=True)
    for name in model_catalog.SPEECH_MODELS:
        print(f"  - {name}", flush=True)

    print("\nMusic models:", flush=True)
    for name in model_catalog.MUSIC_MODELS:
        print(f"  - {name}", flush=True)

    print("\nDefaults in this project:", flush=True)
    print(f"  - text: {model_catalog.DEFAULT_TEXT_MODEL}", flush=True)
    print(f"  - video: {model_catalog.DEFAULT_VIDEO_MODEL}", flush=True)
    print(f"  - speech: {model_catalog.DEFAULT_TTS_MODEL}", flush=True)
    print(f"  - music: {model_catalog.DEFAULT_MUSIC_MODEL}", flush=True)
    print("\nNote: Planning uses text models only. Video, speech, image, and music require pay-as-you-go API keys.", flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
