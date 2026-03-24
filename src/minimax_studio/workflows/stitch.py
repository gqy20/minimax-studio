from __future__ import annotations

import shutil
from collections.abc import Callable
from pathlib import Path

from minimax_studio.files import ensure_dir
from minimax_studio.media import compose_final_video, concat_videos, get_duration_seconds, pad_video_to_duration
from minimax_studio.schemas import StitchOptions, StitchResult
from minimax_studio.workflows import emit_report


def run_stitch(
    options: StitchOptions,
    *,
    reporter: Callable[[str], None] | None = None,
) -> StitchResult:
    validate_stitch_options(options)

    output_dir = ensure_dir(options.output_path.resolve().parent)

    stitched_video_path = output_dir / "edit.mp4"
    padded_video_path = output_dir / "timed.mp4"
    final_video_path = options.output_path.resolve()

    emit_report(reporter, "step 1/3: stitching source video...")
    concat_videos(options.video_paths, stitched_video_path)

    narration_duration = get_duration_seconds(options.narration_path)
    video_duration = get_duration_seconds(stitched_video_path)
    target_duration = max(video_duration, narration_duration)

    emit_report(reporter, "step 2/3: padding video to target duration...")
    pad_video_to_duration(stitched_video_path, padded_video_path, target_duration)
    emit_report(
        reporter,
        f"stitched video duration={video_duration:.2f}s, narration duration={narration_duration:.2f}s, target={target_duration:.2f}s",
    )

    emit_report(reporter, "step 3/3: composing final video...")
    compose_final_video(
        video_path=padded_video_path,
        narration_path=options.narration_path,
        music_path=options.music_path,
        output_path=final_video_path,
        target_duration=target_duration,
    )
    emit_report(reporter, f"final video saved to: {final_video_path}")

    return StitchResult(
        stitched_video_path=stitched_video_path,
        padded_video_path=padded_video_path,
        final_video_path=final_video_path,
    )


def validate_stitch_options(options: StitchOptions) -> None:
    if shutil.which("ffmpeg") is None or shutil.which("ffprobe") is None:
        raise SystemExit("ffmpeg and ffprobe are required")

    if not options.video_paths:
        raise SystemExit("at least one input video is required")

    missing_video = [path for path in options.video_paths if not path.exists()]
    if missing_video:
        raise SystemExit(f"input video does not exist: {missing_video[0]}")

    if not options.narration_path.exists():
        raise SystemExit(f"narration file does not exist: {options.narration_path}")

    if options.music_path is not None and not options.music_path.exists():
        raise SystemExit(f"music file does not exist: {options.music_path}")
