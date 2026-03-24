from __future__ import annotations

import base64
import binascii
import json
import os
import re
import time
from typing import Any, Callable

import requests

from minimax_studio.model_catalog import DEFAULT_IMAGE_MODEL
from minimax_studio.schemas import GeneratedImage, QuotaEntry, QuotaResult, VideoPlan


API_BASE = "https://api.minimaxi.com/v1"
OPENPLATFORM_BASE = "https://www.minimaxi.com/v1/api/openplatform"
ANTHROPIC_MESSAGES_URL = "https://api.minimaxi.com/anthropic/v1/messages"


class MiniMaxClient:
    def __init__(self, api_key: str, *, session: requests.Session | None = None) -> None:
        self.api_key = api_key
        self.session = session or requests.Session()
        self.session.headers.update(
            {
                "Authorization": f"Bearer {api_key}",
                "Content-Type": "application/json",
            }
        )

    @classmethod
    def from_env(cls) -> "MiniMaxClient":
        api_key = os.environ.get("MINIMAX_API_KEY")
        if not api_key:
            raise SystemExit("MINIMAX_API_KEY is not set")
        return cls(api_key)

    def request_json(
        self,
        method: str,
        path: str,
        label: str,
        *,
        params: dict[str, Any] | None = None,
        payload: dict[str, Any] | None = None,
        timeout: int = 180,
    ) -> dict[str, Any]:
        url = f"{API_BASE}{path}"
        response = self.session.request(method, url, params=params, json=payload, timeout=timeout)
        response.raise_for_status()
        data = response.json()
        self.ensure_base_resp_ok(data, label)
        return data

    def request_openplatform_json(
        self,
        method: str,
        path: str,
        label: str,
        *,
        params: dict[str, Any] | None = None,
        payload: dict[str, Any] | None = None,
        timeout: int = 180,
    ) -> dict[str, Any]:
        url = f"{OPENPLATFORM_BASE}{path}"
        response = self.session.request(method, url, params=params, json=payload, timeout=timeout)
        response.raise_for_status()
        data = response.json()
        self.ensure_base_resp_ok(data, label)
        return data

    def request_anthropic_json(
        self,
        *,
        payload: dict[str, Any],
        timeout: int = 180,
    ) -> dict[str, Any]:
        response = self.session.request(
            "POST",
            ANTHROPIC_MESSAGES_URL,
            headers={
                "x-api-key": self.api_key,
                "anthropic-version": "2023-06-01",
                "content-type": "application/json",
            },
            json=payload,
            timeout=timeout,
        )
        response.raise_for_status()
        data = response.json()
        self.ensure_base_resp_ok(data, "anthropic_messages")
        return data

    @staticmethod
    def ensure_base_resp_ok(payload: dict[str, Any], label: str) -> None:
        base_resp = payload.get("base_resp") or {}
        status_code = base_resp.get("status_code", 0)
        if status_code != 0:
            status_msg = base_resp.get("status_msg", "unknown error")
            raise RuntimeError(f"{label} failed: {status_code} {status_msg}")

    def plan_video(
        self,
        *,
        theme: str,
        scene_count: int,
        scene_duration: int,
        language: str,
        text_model: str,
        text_max_tokens: int,
    ) -> VideoPlan:
        max_chars = max(18, scene_count * scene_duration * 5)
        system_prompt = (
            "You are a video creative planner. "
            "Return valid JSON only. Do not use markdown fences. "
            "Do not include explanations. "
            "Provide concise, production-ready prompts."
        )
        user_prompt = f"""
为主题“{theme}”生成一个短视频制作方案。

要求：
1. 输出 JSON 对象，字段必须完整。
2. scenes 数组长度必须等于 {scene_count}。
3. 每个 scene 的画面提示词和运动提示词用英文，适合 AI 图片/视频生成。
4. narration 使用 {language}，必须简短自然，总长度控制在 {max_chars} 个字符以内，适配总时长约 {scene_count * scene_duration} 秒。
5. music_prompt 用英文，描述纯音乐，不要人声。
6. 风格要统一，适合短视频成片。

JSON Schema:
{{
  "title": "string",
  "visual_style": "string",
  "narration": "string",
  "music_prompt": "string",
  "scenes": [
    {{
      "name": "string",
      "image_prompt": "string",
      "video_prompt": "string"
    }}
  ]
}}
""".strip()

        payload = {
            "model": text_model,
            "max_tokens": text_max_tokens,
            "system": system_prompt,
            "messages": [
                {
                    "role": "user",
                    "content": [{"type": "text", "text": user_prompt}],
                }
            ],
        }
        data = self.request_anthropic_json(payload=payload, timeout=180)
        content_blocks = data.get("content") or []
        text_parts = [block.get("text", "") for block in content_blocks if block.get("type") == "text"]
        content = "\n".join(part for part in text_parts if part).strip()
        if not content:
            raise RuntimeError(f"anthropic_messages returned no text content: {json.dumps(data, ensure_ascii=False)}")

        plan_payload = self.extract_json_object(content)
        return VideoPlan.from_dict(plan_payload, expected_scene_count=scene_count)

    def get_coding_plan_remains(self) -> QuotaResult:
        data = self.request_openplatform_json("GET", "/coding_plan/remains", "coding_plan_remains", timeout=60)
        entries = [QuotaEntry.from_dict(item) for item in data.get("model_remains") or []]
        return QuotaResult(entries=entries)

    def generate_image(
        self,
        *,
        prompt: str,
        aspect_ratio: str,
        prompt_optimizer: bool = False,
    ) -> GeneratedImage:
        payload = {
            "model": DEFAULT_IMAGE_MODEL,
            "prompt": prompt,
            "aspect_ratio": aspect_ratio,
            "response_format": "base64",
            "prompt_optimizer": prompt_optimizer,
        }
        data = self.request_json("POST", "/image_generation", "image_generation", payload=payload, timeout=180)
        image_list = ((data.get("data") or {}).get("image_base64")) or []
        if not image_list:
            raise RuntimeError(f"image_generation returned no images: {json.dumps(data, ensure_ascii=False)}")

        image_b64 = image_list[0]
        return GeneratedImage(base64_data=image_b64, content=base64.b64decode(image_b64))

    def create_video_task(
        self,
        *,
        image_b64: str,
        prompt: str,
        model: str,
        duration: int,
        resolution: str,
    ) -> str:
        payload = {
            "model": model,
            "prompt": prompt,
            "first_frame_image": f"data:image/jpeg;base64,{image_b64}",
            "duration": duration,
            "resolution": resolution,
        }
        data = self.request_json("POST", "/video_generation", "video_generation", payload=payload, timeout=180)
        task_id = data.get("task_id")
        if not task_id:
            raise RuntimeError(f"video_generation returned no task_id: {json.dumps(data, ensure_ascii=False)}")
        return str(task_id)

    def poll_video_task(
        self,
        *,
        task_id: str,
        interval_seconds: int,
        max_wait_seconds: int,
        on_status: Callable[[str], None] | None = None,
    ) -> str:
        deadline = time.time() + max_wait_seconds
        while True:
            data = self.request_json(
                "GET",
                "/query/video_generation",
                "query_video_generation",
                params={"task_id": task_id},
                timeout=60,
            )
            status = str(data.get("status", ""))
            if on_status is not None:
                on_status(f"video task {task_id} status: {status}")

            if status == "Success":
                file_id = data.get("file_id")
                if not file_id:
                    raise RuntimeError(f"video task succeeded without file_id: {json.dumps(data, ensure_ascii=False)}")
                return str(file_id)

            if status in {"Fail", "Failed", "Expired"}:
                raise RuntimeError(f"video task ended with status={status}: {json.dumps(data, ensure_ascii=False)}")

            if time.time() >= deadline:
                raise TimeoutError(f"video task {task_id} timed out after {max_wait_seconds} seconds")

            time.sleep(interval_seconds)

    def fetch_download_url(self, *, file_id: str) -> str:
        data = self.request_json(
            "GET",
            "/files/retrieve",
            "files_retrieve",
            params={"file_id": file_id},
            timeout=60,
        )
        download_url = ((data.get("file") or {}).get("download_url")) or ""
        if not download_url:
            raise RuntimeError(f"files_retrieve returned no download_url: {json.dumps(data, ensure_ascii=False)}")
        return download_url

    def download_file(self, *, url: str) -> bytes:
        with requests.get(url, timeout=600) as response:
            response.raise_for_status()
            return response.content

    def generate_narration(
        self,
        *,
        text: str,
        voice_id: str,
        tts_model: str,
        audio_format: str,
    ) -> bytes:
        payload = {
            "model": tts_model,
            "text": text,
            "stream": False,
            "voice_setting": {
                "voice_id": voice_id,
                "speed": 1,
                "vol": 1,
                "pitch": 0,
            },
            "language_boost": "Chinese",
            "audio_setting": {
                "sample_rate": 32000,
                "bitrate": 128000,
                "format": audio_format,
                "channel": 1,
            },
            "subtitle_enable": False,
        }
        data = self.request_json("POST", "/t2a_v2", "t2a_v2", payload=payload, timeout=180)
        audio_hex = ((data.get("data") or {}).get("audio")) or ""
        if not audio_hex:
            raise RuntimeError(f"t2a_v2 returned no audio: {json.dumps(data, ensure_ascii=False)}")
        return self.decode_hex_audio(audio_hex)

    def generate_music(
        self,
        *,
        prompt: str,
        model: str,
        audio_format: str,
    ) -> bytes:
        payload = {
            "model": model,
            "prompt": prompt,
            "stream": False,
            "output_format": "hex",
            "aigc_watermark": False,
            "audio_setting": {
                "sample_rate": 44100,
                "bitrate": 256000,
                "format": audio_format,
            },
        }
        if model == "music-2.5":
            payload["lyrics"] = "[Inst]"
            payload["lyrics_optimizer"] = False
        else:
            payload["is_instrumental"] = True
        data = self.request_json("POST", "/music_generation", "music_generation", payload=payload, timeout=600)
        audio_hex = ((data.get("data") or {}).get("audio")) or ""
        if not audio_hex:
            raise RuntimeError(f"music_generation returned no audio: {json.dumps(data, ensure_ascii=False)}")
        return self.decode_hex_audio(audio_hex)

    @staticmethod
    def decode_hex_audio(hex_text: str) -> bytes:
        try:
            return binascii.unhexlify(hex_text)
        except binascii.Error as exc:
            raise RuntimeError("invalid hex audio payload") from exc

    @staticmethod
    def strip_thinking(text: str) -> str:
        text = re.sub(r"<think>.*?</think>\s*", "", text, flags=re.DOTALL)
        text = text.strip()
        if text.startswith("```"):
            text = re.sub(r"^```[a-zA-Z0-9_-]*\n", "", text)
            text = re.sub(r"\n```$", "", text)
        return text.strip()

    @classmethod
    def extract_json_object(cls, text: str) -> dict[str, Any]:
        cleaned = cls.strip_thinking(text)
        try:
            return json.loads(cleaned)
        except json.JSONDecodeError:
            pass

        match = re.search(r"\{.*\}", cleaned, flags=re.DOTALL)
        if not match:
            raise RuntimeError(f"failed to locate JSON object in model output: {cleaned}")
        return json.loads(match.group(0))
