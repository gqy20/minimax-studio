from __future__ import annotations

import shutil
import subprocess
import tempfile
from pathlib import Path

from minimax_studio.files import write_text

def run_command(command: list[str]) -> None:
    subprocess.run(command, check=True)


def get_duration_seconds(path: Path) -> float:
    result = subprocess.run(
        [
            "ffprobe",
            "-v",
            "error",
            "-show_entries",
            "format=duration",
            "-of",
            "default=noprint_wrappers=1:nokey=1",
            str(path),
        ],
        check=True,
        capture_output=True,
        text=True,
    )
    return float(result.stdout.strip())


def normalize_single_video(input_path: Path, output_path: Path) -> None:
    run_command(
        [
            "ffmpeg",
            "-y",
            "-i",
            str(input_path),
            "-an",
            "-c:v",
            "libx264",
            "-pix_fmt",
            "yuv420p",
            "-movflags",
            "+faststart",
            str(output_path),
        ]
    )


def concat_videos(video_paths: list[Path], output_path: Path) -> None:
    if not video_paths:
        raise RuntimeError("no video segments to concatenate")

    if len(video_paths) == 1:
        normalize_single_video(video_paths[0], output_path)
        return

    with tempfile.TemporaryDirectory() as temp_dir:
        concat_file = Path(temp_dir) / "concat.txt"
        lines = [f"file '{path.resolve().as_posix()}'" for path in video_paths]
        write_text(concat_file, "\n".join(lines) + "\n")
        run_command(
            [
                "ffmpeg",
                "-y",
                "-f",
                "concat",
                "-safe",
                "0",
                "-i",
                str(concat_file),
                "-an",
                "-c:v",
                "libx264",
                "-pix_fmt",
                "yuv420p",
                "-movflags",
                "+faststart",
                str(output_path),
            ]
        )


def pad_video_to_duration(input_path: Path, output_path: Path, target_duration: float) -> None:
    current_duration = get_duration_seconds(input_path)
    extra_duration = max(0.0, target_duration - current_duration)
    if extra_duration < 0.05:
        shutil.copy2(input_path, output_path)
        return

    run_command(
        [
            "ffmpeg",
            "-y",
            "-i",
            str(input_path),
            "-vf",
            f"tpad=stop_mode=clone:stop_duration={extra_duration:.3f},format=yuv420p",
            "-an",
            "-c:v",
            "libx264",
            "-pix_fmt",
            "yuv420p",
            "-movflags",
            "+faststart",
            str(output_path),
        ]
    )


def compose_final_video(
    video_path: Path,
    narration_path: Path,
    music_path: Path | None,
    output_path: Path,
    target_duration: float,
) -> None:
    if music_path is None:
        run_command(
            [
                "ffmpeg",
                "-y",
                "-i",
                str(video_path),
                "-i",
                str(narration_path),
                "-filter_complex",
                f"[1:a]volume=1.0,atrim=0:{target_duration:.3f},asetpts=N/SR/TB[aout]",
                "-map",
                "0:v",
                "-map",
                "[aout]",
                "-c:v",
                "copy",
                "-c:a",
                "aac",
                "-b:a",
                "192k",
                "-shortest",
                str(output_path),
            ]
        )
        return

    run_command(
        [
            "ffmpeg",
            "-y",
            "-i",
            str(video_path),
            "-stream_loop",
            "-1",
            "-i",
            str(music_path),
            "-i",
            str(narration_path),
            "-filter_complex",
            (
                f"[1:a]volume=0.16,atrim=0:{target_duration:.3f},asetpts=N/SR/TB[bgm];"
                f"[2:a]volume=1.0,atrim=0:{target_duration:.3f},asetpts=N/SR/TB[narr];"
                "[bgm][narr]amix=inputs=2:duration=longest:dropout_transition=2[aout]"
            ),
            "-map",
            "0:v",
            "-map",
            "[aout]",
            "-c:v",
            "copy",
            "-c:a",
            "aac",
            "-b:a",
            "192k",
            "-shortest",
            str(output_path),
        ]
    )
