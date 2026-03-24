from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class ScenePlan:
    name: str
    image_prompt: str
    video_prompt: str

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "ScenePlan":
        return cls(
            name=str(payload.get("name", "")),
            image_prompt=str(payload.get("image_prompt", "")),
            video_prompt=str(payload.get("video_prompt", "")),
        )

    def to_dict(self) -> dict[str, str]:
        return {
            "name": self.name,
            "image_prompt": self.image_prompt,
            "video_prompt": self.video_prompt,
        }


@dataclass(frozen=True)
class VideoPlan:
    title: str
    visual_style: str
    narration: str
    music_prompt: str
    scenes: list[ScenePlan]

    @classmethod
    def from_dict(cls, payload: dict[str, Any], *, expected_scene_count: int | None = None) -> "VideoPlan":
        scenes_raw = payload.get("scenes")
        if not isinstance(scenes_raw, list):
            raise RuntimeError("video plan is missing a valid scenes array")

        scenes = [ScenePlan.from_dict(scene) for scene in scenes_raw]
        if expected_scene_count is not None and len(scenes) != expected_scene_count:
            raise RuntimeError(f"video plan returned {len(scenes)} scenes, expected {expected_scene_count}")

        return cls(
            title=str(payload.get("title", "")),
            visual_style=str(payload.get("visual_style", "")),
            narration=str(payload.get("narration", "")),
            music_prompt=str(payload.get("music_prompt", "")),
            scenes=scenes,
        )

    def to_dict(self) -> dict[str, Any]:
        return {
            "title": self.title,
            "visual_style": self.visual_style,
            "narration": self.narration,
            "music_prompt": self.music_prompt,
            "scenes": [scene.to_dict() for scene in self.scenes],
        }


@dataclass(frozen=True)
class GeneratedImage:
    base64_data: str
    content: bytes


@dataclass(frozen=True)
class ClipOptions:
    image_prompt: str
    video_prompt: str
    aspect_ratio: str
    video_model: str
    duration: int
    resolution: str
    poll_interval: int
    max_wait: int
    output_prefix: Path
    image_prompt_optimizer: bool = False


@dataclass(frozen=True)
class ClipResult:
    image_path: Path
    video_path: Path
    task_id: str
    file_id: str
    download_url: str


@dataclass(frozen=True)
class MakeOptions:
    theme: str
    scene_count: int
    scene_duration: int
    aspect_ratio: str
    resolution: str
    text_model: str
    text_max_tokens: int
    video_model: str
    tts_model: str
    music_model: str
    music_mode: str
    voice_id: str
    audio_format: str
    poll_interval: int
    max_wait: int
    language: str
    output_dir: Path
    input_video: Path | None
    image_prompt_optimizer: bool = False


@dataclass(frozen=True)
class MakeResult:
    output_dir: Path
    plan_path: Path
    narration_path: Path
    music_path: Path | None
    final_video_path: Path


@dataclass(frozen=True)
class PlanOptions:
    theme: str
    scene_count: int
    scene_duration: int
    language: str
    text_model: str
    text_max_tokens: int
    output_dir: Path


@dataclass(frozen=True)
class PlanResult:
    output_dir: Path
    plan_path: Path
    narration_path: Path


@dataclass(frozen=True)
class VoiceOptions:
    text: str
    output_path: Path
    voice_id: str
    tts_model: str
    audio_format: str


@dataclass(frozen=True)
class VoiceResult:
    output_path: Path


@dataclass(frozen=True)
class MusicOptions:
    prompt: str
    output_path: Path
    model: str
    audio_format: str


@dataclass(frozen=True)
class MusicResult:
    output_path: Path


@dataclass(frozen=True)
class StitchOptions:
    video_paths: list[Path]
    narration_path: Path
    output_path: Path
    music_path: Path | None = None


@dataclass(frozen=True)
class StitchResult:
    stitched_video_path: Path
    padded_video_path: Path
    final_video_path: Path


@dataclass(frozen=True)
class QuotaEntry:
    model_name: str
    start_time: datetime
    end_time: datetime
    remains_time_ms: int
    current_interval_total_count: int
    current_interval_usage_count: int
    weekly_start_time: datetime
    weekly_end_time: datetime
    current_weekly_total_count: int
    current_weekly_usage_count: int
    weekly_remains_time_ms: int

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "QuotaEntry":
        return cls(
            model_name=str(payload.get("model_name", "")),
            start_time=datetime.fromtimestamp(int(payload.get("start_time", 0)) / 1000),
            end_time=datetime.fromtimestamp(int(payload.get("end_time", 0)) / 1000),
            remains_time_ms=int(payload.get("remains_time", 0)),
            current_interval_total_count=int(payload.get("current_interval_total_count", 0)),
            current_interval_usage_count=int(payload.get("current_interval_usage_count", 0)),
            weekly_start_time=datetime.fromtimestamp(int(payload.get("weekly_start_time", 0)) / 1000),
            weekly_end_time=datetime.fromtimestamp(int(payload.get("weekly_end_time", 0)) / 1000),
            current_weekly_total_count=int(payload.get("current_weekly_total_count", 0)),
            current_weekly_usage_count=int(payload.get("current_weekly_usage_count", 0)),
            weekly_remains_time_ms=int(payload.get("weekly_remains_time", 0)),
        )


@dataclass(frozen=True)
class QuotaResult:
    entries: list[QuotaEntry]
