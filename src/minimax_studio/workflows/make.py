from __future__ import annotations

import shutil
from collections.abc import Callable
from pathlib import Path

from minimax_studio.client import MiniMaxClient
from minimax_studio.files import ensure_dir, write_bytes
from minimax_studio.media import compose_final_video, concat_videos, get_duration_seconds, pad_video_to_duration
from minimax_studio.model_catalog import normalize_video_model_and_resolution
from minimax_studio.schemas import MakeOptions, MakeResult
from minimax_studio.workflows import emit_report
from minimax_studio.workflows.generate import save_plan_files


def run_make(
    client: MiniMaxClient,
    options: MakeOptions,
    *,
    reporter: Callable[[str], None] | None = None,
) -> MakeResult:
    validate_make_options(options)

    output_dir = ensure_dir(options.output_dir.resolve())
    video_model, video_resolution = normalize_video_model_and_resolution(options.video_model, options.resolution)

    emit_report(reporter, "step 1/7: planning storyboard with text model...")
    plan = client.plan_video(
        theme=options.theme,
        scene_count=options.scene_count,
        scene_duration=options.scene_duration,
        language=options.language,
        text_model=options.text_model,
        text_max_tokens=options.text_max_tokens,
    )
    plan_path = save_plan_files(output_dir, plan)
    emit_report(reporter, f"plan saved to: {plan_path}")

    scene_videos: list[Path] = []
    if options.input_video is not None:
        emit_report(reporter, f"step 2/7: reusing existing video: {options.input_video}")
        reused_video_path = output_dir / "s01.mp4"
        shutil.copy2(options.input_video, reused_video_path)
        scene_videos.append(reused_video_path)
    else:
        for index, scene in enumerate(plan.scenes, start=1):
            frame_path = output_dir / f"s{index:02d}.jpg"
            video_path = output_dir / f"s{index:02d}.mp4"

            emit_report(reporter, f"step 2/7: generating image for scene {index}...")
            image = client.generate_image(
                prompt=scene.image_prompt,
                aspect_ratio=options.aspect_ratio,
                prompt_optimizer=options.image_prompt_optimizer,
            )
            write_bytes(frame_path, image.content)
            emit_report(reporter, f"scene {index} frame saved to: {frame_path}")

            emit_report(reporter, f"step 3/7: generating video for scene {index}...")
            task_id = client.create_video_task(
                image_b64=image.base64_data,
                prompt=scene.video_prompt,
                model=video_model,
                duration=options.scene_duration,
                resolution=video_resolution,
            )
            emit_report(reporter, f"scene {index} task id: {task_id}")
            file_id = client.poll_video_task(
                task_id=task_id,
                interval_seconds=options.poll_interval,
                max_wait_seconds=options.max_wait,
                on_status=reporter,
            )
            download_url = client.fetch_download_url(file_id=file_id)
            write_bytes(video_path, client.download_file(url=download_url))
            emit_report(reporter, f"scene {index} video saved to: {video_path}")
            scene_videos.append(video_path)

    narration_path = output_dir / f"voice.{options.audio_format}"
    emit_report(reporter, "step 4/7: generating narration...")
    write_bytes(
        narration_path,
        client.generate_narration(
            text=plan.narration,
            voice_id=options.voice_id,
            tts_model=options.tts_model,
            audio_format=options.audio_format,
        ),
    )
    emit_report(reporter, f"narration saved to: {narration_path}")

    music_path: Path | None = None
    if options.music_mode != "skip":
        tentative_music_path = output_dir / f"music.{options.audio_format}"
        emit_report(reporter, "step 5/7: generating music...")
        try:
            write_bytes(
                tentative_music_path,
                client.generate_music(
                    prompt=plan.music_prompt,
                    model=options.music_model,
                    audio_format=options.audio_format,
                ),
            )
            music_path = tentative_music_path
            emit_report(reporter, f"music saved to: {music_path}")
        except Exception as exc:
            if options.music_mode == "required":
                raise
            emit_report(reporter, f"music generation unavailable, continuing without background music: {exc}")
    else:
        emit_report(reporter, "step 5/7: skipping music generation by request...")

    stitched_video_path = output_dir / "edit.mp4"
    emit_report(reporter, "step 6/7: stitching and timing video...")
    concat_videos(scene_videos, stitched_video_path)
    narration_duration = get_duration_seconds(narration_path)
    video_duration = get_duration_seconds(stitched_video_path)
    target_duration = max(video_duration, narration_duration)
    padded_video_path = output_dir / "timed.mp4"
    pad_video_to_duration(stitched_video_path, padded_video_path, target_duration)
    emit_report(
        reporter,
        f"stitched video duration={video_duration:.2f}s, narration duration={narration_duration:.2f}s, target={target_duration:.2f}s",
    )

    final_video_path = output_dir / "final.mp4"
    emit_report(reporter, "step 7/7: composing final video...")
    compose_final_video(
        video_path=padded_video_path,
        narration_path=narration_path,
        music_path=music_path,
        output_path=final_video_path,
        target_duration=target_duration,
    )
    emit_report(reporter, f"final video saved to: {final_video_path}")

    return MakeResult(
        output_dir=output_dir,
        plan_path=plan_path,
        narration_path=narration_path,
        music_path=music_path,
        final_video_path=final_video_path,
    )


def validate_make_options(options: MakeOptions) -> None:
    if options.scene_count < 1:
        raise SystemExit("--scene-count must be >= 1")

    if shutil.which("ffmpeg") is None or shutil.which("ffprobe") is None:
        raise SystemExit("ffmpeg and ffprobe are required")

    if options.input_video is not None and not options.input_video.exists():
        raise SystemExit(f"--input-video does not exist: {options.input_video}")

    if options.input_video is not None and options.scene_count != 1:
        raise SystemExit("--input-video currently requires --scene-count 1")
