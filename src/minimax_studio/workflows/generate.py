from __future__ import annotations

import json
from collections.abc import Callable

from minimax_studio.client import MiniMaxClient
from minimax_studio.files import ensure_dir, write_bytes, write_text
from minimax_studio.schemas import MusicOptions, MusicResult, PlanOptions, PlanResult, VideoPlan, VoiceOptions, VoiceResult
from minimax_studio.workflows import emit_report


def run_plan(
    client: MiniMaxClient,
    options: PlanOptions,
    *,
    reporter: Callable[[str], None] | None = None,
) -> PlanResult:
    output_dir = ensure_dir(options.output_dir.resolve())

    emit_report(reporter, "step 1/1: planning storyboard with text model...")
    plan = client.plan_video(
        theme=options.theme,
        scene_count=options.scene_count,
        scene_duration=options.scene_duration,
        language=options.language,
        text_model=options.text_model,
        text_max_tokens=options.text_max_tokens,
    )
    plan_path = save_plan_files(output_dir, plan)
    narration_path = output_dir / "voice.txt"
    emit_report(reporter, f"plan saved to: {plan_path}")
    return PlanResult(output_dir=output_dir, plan_path=plan_path, narration_path=narration_path)


def run_voice(
    client: MiniMaxClient,
    options: VoiceOptions,
    *,
    reporter: Callable[[str], None] | None = None,
) -> VoiceResult:
    emit_report(reporter, "step 1/1: generating narration...")
    write_bytes(
        options.output_path,
        client.generate_narration(
            text=options.text,
            voice_id=options.voice_id,
            tts_model=options.tts_model,
            audio_format=options.audio_format,
        ),
    )
    emit_report(reporter, f"narration saved to: {options.output_path}")
    return VoiceResult(output_path=options.output_path)


def run_music(
    client: MiniMaxClient,
    options: MusicOptions,
    *,
    reporter: Callable[[str], None] | None = None,
) -> MusicResult:
    emit_report(reporter, "step 1/1: generating music...")
    write_bytes(
        options.output_path,
        client.generate_music(
            prompt=options.prompt,
            model=options.model,
            audio_format=options.audio_format,
        ),
    )
    emit_report(reporter, f"music saved to: {options.output_path}")
    return MusicResult(output_path=options.output_path)


def save_plan_files(output_dir, plan: VideoPlan):
    plan_path = output_dir / "plan.json"
    write_text(plan_path, json.dumps(plan.to_dict(), ensure_ascii=False, indent=2))
    write_text(output_dir / "voice.txt", plan.narration)
    return plan_path
