from __future__ import annotations

AUDIO_FORMATS = (
    "mp3",
    "wav",
    "flac",
)

MUSIC_MODES = (
    "auto",
    "required",
    "skip",
)

TEXT_MODELS = (
    "MiniMax-M2.7-highspeed",
    "MiniMax-M2.7",
    "MiniMax-M2.5",
    "MiniMax-M2.5-highspeed",
    "MiniMax-M2.1",
    "MiniMax-M2.1-highspeed",
    "MiniMax-M2",
)

PLANNING_TEXT_MODELS = (
    "MiniMax-M2.7-highspeed",
    "MiniMax-M2.7",
    "MiniMax-M2.5-highspeed",
    "MiniMax-M2.5",
    "MiniMax-M2.1-highspeed",
    "MiniMax-M2.1",
    "MiniMax-M2",
)

CODING_PLAN_TEXT_MODELS = (
    "MiniMax-M2.5",
    "MiniMax-M2.1",
    "MiniMax-M2",
)

CODING_PLAN_HIGHSPEED_MODELS = (
    "MiniMax-M2.7-highspeed",
    "MiniMax-M2.5-highspeed",
)

VIDEO_MODELS = (
    "MiniMax-Hailuo-2.3",
    "MiniMax-Hailuo-2.3-Fast",
    "MiniMax-Hailuo-02",
)

VIDEO_MODEL_ALIASES = (
    "Hailuo-2.3-768P",
    "MiniMax-Hailuo-2.3-768P",
    "Hailuo-2.3-1080P",
    "MiniMax-Hailuo-2.3-1080P",
)

VIDEO_MODEL_CHOICES = VIDEO_MODELS + VIDEO_MODEL_ALIASES

VIDEO_RESOLUTIONS = (
    "768P",
    "1080P",
)


def normalize_video_model_and_resolution(model: str, resolution: str) -> tuple[str, str]:
    alias_map = {
        "Hailuo-2.3-768P": ("MiniMax-Hailuo-2.3", "768P"),
        "MiniMax-Hailuo-2.3-768P": ("MiniMax-Hailuo-2.3", "768P"),
        "Hailuo-2.3-1080P": ("MiniMax-Hailuo-2.3", "1080P"),
        "MiniMax-Hailuo-2.3-1080P": ("MiniMax-Hailuo-2.3", "1080P"),
    }
    return alias_map.get(model, (model, resolution))

IMAGE_MODELS = ("image-01",)

SPEECH_MODELS = (
    "speech-2.8-hd",
    "speech-2.8-turbo",
    "speech-2.6-hd",
    "speech-2.6-turbo",
    "speech-02-hd",
    "speech-02-turbo",
)

MUSIC_MODELS = (
    "music-2.5",
    "music-2.0",
)

DEFAULT_TEXT_MODEL = "MiniMax-M2.7-highspeed"
DEFAULT_VIDEO_MODEL = "MiniMax-Hailuo-2.3-Fast"
DEFAULT_TTS_MODEL = "speech-2.8-hd"
DEFAULT_MUSIC_MODEL = "music-2.5"
DEFAULT_IMAGE_MODEL = "image-01"
DEFAULT_VOICE_ID = "male-qn-qingse"
