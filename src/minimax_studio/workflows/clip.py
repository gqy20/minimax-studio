from __future__ import annotations

from collections.abc import Callable

from minimax_studio.client import MiniMaxClient
from minimax_studio.files import write_bytes
from minimax_studio.schemas import ClipOptions, ClipResult
from minimax_studio.workflows import emit_report


def run_clip(
    client: MiniMaxClient,
    options: ClipOptions,
    *,
    reporter: Callable[[str], None] | None = None,
) -> ClipResult:
    image_path = options.output_prefix.with_suffix(".jpg")
    video_path = options.output_prefix.with_suffix(".mp4")

    emit_report(reporter, "step 1/4: generating image...")
    image = client.generate_image(
        prompt=options.image_prompt,
        aspect_ratio=options.aspect_ratio,
        prompt_optimizer=options.image_prompt_optimizer,
    )
    write_bytes(image_path, image.content)
    emit_report(reporter, f"image saved to: {image_path}")

    emit_report(reporter, "step 2/4: creating video task...")
    task_id = client.create_video_task(
        image_b64=image.base64_data,
        prompt=options.video_prompt,
        model=options.video_model,
        duration=options.duration,
        resolution=options.resolution,
    )
    emit_report(reporter, f"video task id: {task_id}")

    emit_report(reporter, "step 3/4: polling video task...")
    file_id = client.poll_video_task(
        task_id=task_id,
        interval_seconds=options.poll_interval,
        max_wait_seconds=options.max_wait,
        on_status=reporter,
    )
    emit_report(reporter, f"video file id: {file_id}")

    emit_report(reporter, "step 4/4: downloading video...")
    download_url = client.fetch_download_url(file_id=file_id)
    write_bytes(video_path, client.download_file(url=download_url))
    emit_report(reporter, f"video saved to: {video_path}")
    emit_report(reporter, f"download url: {download_url}")

    return ClipResult(
        image_path=image_path,
        video_path=video_path,
        task_id=task_id,
        file_id=file_id,
        download_url=download_url,
    )
