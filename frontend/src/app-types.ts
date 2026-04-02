export type MakeRequest = {
  theme: string;
  scene_count: number;
  scene_duration: number;
  language: string;
  input_video: string;
};

export type PlanRequest = {
  theme: string;
  scene_count: number;
  scene_duration: number;
  language: string;
};

export type ClipRequest = {
  prompt: string;
  subject: string;
  aspect_ratio: string;
  duration: number;
  resolution: string;
};

export type ImageRequest = {
  prompt: string;
  aspect_ratio: string;
};

export type VoiceRequest = {
  text: string;
  voice_id: string;
  audio_format: string;
};

export type MusicRequest = {
  prompt: string;
  audio_format: string;
};

export type StitchRequest = {
  videos: string;
  narration: string;
  music: string;
};

export type MakeResult = {
  output_dir: string;
  plan_path: string;
  narration_path: string;
  music_path?: string;
  final_video_path: string;
};

export type PlanResult = {
  output_dir: string;
  plan_path: string;
  narration_path: string;
};

export type ClipResult = {
  image_path: string;
  video_path: string;
};

export type ImageResult = {
  image_path: string;
};

export type VoiceResult = {
  output_path: string;
};

export type MusicResult = {
  output_path: string;
};

export type StitchResult = {
  stitched_video_path: string;
  padded_video_path: string;
  final_video_path: string;
};

export type Job = {
  job_id: string;
  status: "pending" | "processing" | "completed" | "failed";
  progress: number;
  stage: string;
  created_at?: string;
  updated_at?: string;
  output?: MakeResult | PlanResult | ClipResult | ImageResult | VoiceResult | MusicResult | StitchResult;
  error?: string;
  logs?: Array<{
    time: string;
    message: string;
  }>;
  artifacts?: Array<{
    label: string;
    kind: string;
    path: string;
  }>;
};

export type JobListResult = {
  jobs: Job[];
};

export type QuotaEntry = {
  model_name: string;
  current_interval_total_count: number;
  current_interval_usage_count: number;
  current_weekly_total_count: number;
  current_weekly_usage_count: number;
};

export type QuotaResult = QuotaEntry[] | { entries: QuotaEntry[] };

export type PlanData = {
  title: string;
  visual_style: string;
  narration: string;
  music_prompt: string;
  scenes: Array<{
    name: string;
    image_prompt: string;
    video_prompt: string;
  }>;
};

export type WorkflowMode = "make" | "plan" | "clip" | "image" | "voice" | "music" | "stitch";

export type ResultTab = "result" | "plan" | "logs";

export type VoiceOption = {
  id: string;
  label: string;
};
