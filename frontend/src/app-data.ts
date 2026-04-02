import type {
  ClipRequest,
  ImageRequest,
  MakeRequest,
  MusicRequest,
  PlanRequest,
  StitchRequest,
  VoiceOption,
  VoiceRequest,
} from "./app-types";

export const API_ROOT = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");
export const HISTORY_STORAGE_KEY = "minimax-studio.job-history";

export const DEFAULT_FORM: MakeRequest = {
  theme: "一只纸船在凌晨海面漂流，最终进入发光的城市河道",
  scene_count: 1,
  scene_duration: 6,
  language: "zh",
  input_video: "",
};

export const DEFAULT_PLAN_FORM: PlanRequest = {
  theme: "清晨薄雾里的旧码头，一只纸船慢慢漂向远处灯塔",
  scene_count: 3,
  scene_duration: 6,
  language: "zh",
};

export const DEFAULT_CLIP_FORM: ClipRequest = {
  prompt: "A paper boat on reflective water at dawn, cinematic soft light",
  subject: "The paper boat drifts gently forward while the camera slowly pushes in",
  aspect_ratio: "16:9",
  duration: 5,
  resolution: "720p",
};

export const DEFAULT_IMAGE_FORM: ImageRequest = {
  prompt: "A paper boat drifting on reflective water at dawn, cinematic soft light",
  aspect_ratio: "16:9",
};

export const DEFAULT_VOICE_FORM: VoiceRequest = {
  text: "海风从旧码头吹过，纸船沿着微光漂向远方。",
  voice_id: "male-qn-qingse",
  audio_format: "mp3",
};

export const VOICE_OPTIONS: VoiceOption[] = [
  { id: "male-qn-qingse", label: "青涩青年" },
  { id: "male-qn-jingying", label: "精英青年" },
  { id: "male-qn-badao", label: "霸道青年" },
  { id: "male-qn-daxuesheng", label: "青年大学生" },
  { id: "female-shaonv", label: "少女" },
  { id: "female-yujie", label: "御姐" },
  { id: "female-chengshu", label: "成熟女性" },
  { id: "female-tianmei", label: "甜美女性" },
  { id: "clever_boy", label: "聪明男童" },
  { id: "lovely_girl", label: "萌萌女童" },
  { id: "Chinese (Mandarin)_News_Anchor", label: "新闻女声" },
  { id: "Chinese (Mandarin)_Gentleman", label: "温润男声" },
  { id: "Chinese (Mandarin)_Male_Announcer", label: "播报男声" },
  { id: "Chinese (Mandarin)_Sweet_Lady", label: "甜美女声" },
  { id: "Chinese (Mandarin)_Warm_Bestie", label: "温暖闺蜜" },
  { id: "Chinese (Mandarin)_Warm_Girl", label: "温暖少女" },
  { id: "Chinese (Mandarin)_Radio_Host", label: "电台男主播" },
  { id: "Chinese (Mandarin)_Lyrical_Voice", label: "抒情男声" },
];

export const DEFAULT_MUSIC_FORM: MusicRequest = {
  prompt: "warm cinematic piano with soft ambient texture, no vocals",
  audio_format: "mp3",
};

export const DEFAULT_STITCH_FORM: StitchRequest = {
  videos: "",
  narration: "",
  music: "",
};
