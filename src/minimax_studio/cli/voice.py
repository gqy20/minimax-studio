from __future__ import annotations

import argparse
import sys
from pathlib import Path

from minimax_studio.cli import ChineseArgumentParser, enable_argcomplete, print_report
from minimax_studio.model_catalog import AUDIO_FORMATS, DEFAULT_TTS_MODEL, DEFAULT_VOICE_ID, SPEECH_MODELS
from minimax_studio.client import MiniMaxClient
from minimax_studio.schemas import VoiceOptions
from minimax_studio.workflows.generate import run_voice


def build_parser(*, add_help: bool = True) -> argparse.ArgumentParser:
    parser = ChineseArgumentParser(add_help=add_help, description="生成旁白音频。")
    parser.add_argument("text", help="要合成的文本。")
    parser.add_argument("-o", "--output", default="runs/voice.mp3", help="输出音频路径。")
    parser.add_argument("-v", "--voice", "--voice-id", dest="voice_id", default=DEFAULT_VOICE_ID, help="旁白语音 ID。")
    parser.add_argument(
        "-m",
        "--model",
        "--tts-model",
        dest="tts_model",
        default=DEFAULT_TTS_MODEL,
        choices=SPEECH_MODELS,
        metavar="TTS_MODEL",
        help="TTS 模型。",
    )
    parser.add_argument("-f", "--format", "--audio-format", dest="audio_format", choices=AUDIO_FORMATS, default="mp3", help="音频格式。")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    enable_argcomplete(parser)
    args = parser.parse_args(argv)
    options = VoiceOptions(
        text=args.text,
        output_path=Path(args.output),
        voice_id=args.voice_id,
        tts_model=args.tts_model,
        audio_format=args.audio_format,
    )
    try:
        run_voice(MiniMaxClient.from_env(), options, reporter=print_report)
        return 0
    except KeyboardInterrupt:
        print("已中断", file=sys.stderr)
        raise


if __name__ == "__main__":
    raise SystemExit(main())
