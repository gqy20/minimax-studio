import { FormEvent, useEffect, useMemo, useState } from "react";

type MakeRequest = {
  theme: string;
  scene_count: number;
  scene_duration: number;
  language: string;
  input_video: string;
};

type PlanRequest = {
  theme: string;
  scene_count: number;
  scene_duration: number;
  language: string;
};

type ClipRequest = {
  prompt: string;
  subject: string;
  aspect_ratio: string;
  duration: number;
  resolution: string;
};

type MakeResult = {
  output_dir: string;
  plan_path: string;
  narration_path: string;
  music_path?: string;
  final_video_path: string;
};

type PlanResult = {
  output_dir: string;
  plan_path: string;
  narration_path: string;
};

type ClipResult = {
  image_path: string;
  video_path: string;
};

type Job = {
  job_id: string;
  status: "pending" | "processing" | "completed" | "failed";
  progress: number;
  stage: string;
  created_at?: string;
  updated_at?: string;
  output?: MakeResult | PlanResult | ClipResult;
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

type JobListResult = {
  jobs: Job[];
};

type QuotaEntry = {
  model_name: string;
  current_interval_total_count: number;
  current_interval_usage_count: number;
  current_weekly_total_count: number;
  current_weekly_usage_count: number;
};

type QuotaResult = {
  entries: QuotaEntry[];
};

type PlanData = {
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

const API_ROOT = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");
const HISTORY_STORAGE_KEY = "minimax-studio.job-history";

const DEFAULT_FORM: MakeRequest = {
  theme: "一只纸船在凌晨海面漂流，最终进入发光的城市河道",
  scene_count: 1,
  scene_duration: 6,
  language: "zh",
  input_video: "",
};

const DEFAULT_PLAN_FORM: PlanRequest = {
  theme: "清晨薄雾里的旧码头，一只纸船慢慢漂向远处灯塔",
  scene_count: 3,
  scene_duration: 6,
  language: "zh",
};

const DEFAULT_CLIP_FORM: ClipRequest = {
  prompt: "A paper boat on reflective water at dawn, cinematic soft light",
  subject: "The paper boat drifts gently forward while the camera slowly pushes in",
  aspect_ratio: "16:9",
  duration: 5,
  resolution: "720p",
};

type WorkflowMode = "make" | "plan" | "clip";

function apiUrl(path: string) {
  if (!API_ROOT) {
    return path;
  }

  if (API_ROOT.endsWith("/api/v1") && path.startsWith("/api/v1")) {
    return `${API_ROOT}${path.slice("/api/v1".length)}`;
  }

  return `${API_ROOT}${path}`;
}

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(apiUrl(path), init);
  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    const message =
      typeof data?.error === "string"
        ? data.error
        : `Request failed with status ${response.status}`;
    throw new Error(message);
  }

  return data as T;
}

function loadHistory() {
  const raw = localStorage.getItem(HISTORY_STORAGE_KEY);
  if (!raw) {
    return [] as string[];
  }

  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter((item) => typeof item === "string") : [];
  } catch {
    return [];
  }
}

function saveHistory(items: string[]) {
  localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(items.slice(0, 8)));
}

function toAssetUrl(jobID: string, filePath?: string) {
  if (!filePath) {
    return "";
  }

  const normalized = filePath.replaceAll("\\", "/");
  const fileName = normalized.split("/").filter(Boolean).pop();
  if (!fileName) {
    return "";
  }

  return apiUrl(`/api/v1/output/${jobID}/${encodeURIComponent(fileName)}`);
}

function usagePercent(used: number, total: number) {
  if (!total) {
    return 0;
  }
  return Math.min(100, Math.round((used / total) * 100));
}

function formatStatus(status: Job["status"]) {
  switch (status) {
    case "completed":
      return "已完成";
    case "failed":
      return "失败";
    case "pending":
      return "等待中";
    default:
      return "处理中";
  }
}

export default function App() {
  const [mode, setMode] = useState<WorkflowMode>("make");
  const [form, setForm] = useState<MakeRequest>(DEFAULT_FORM);
  const [planForm, setPlanForm] = useState<PlanRequest>(DEFAULT_PLAN_FORM);
  const [clipForm, setClipForm] = useState<ClipRequest>(DEFAULT_CLIP_FORM);
  const [job, setJob] = useState<Job | null>(null);
  const [recentJobs, setRecentJobs] = useState<Job[]>([]);
  const [quota, setQuota] = useState<QuotaResult | null>(null);
  const [plan, setPlan] = useState<PlanData | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isQuotaLoading, setIsQuotaLoading] = useState(false);
  const [quotaError, setQuotaError] = useState("");
  const [jobError, setJobError] = useState("");
  const [submitError, setSubmitError] = useState("");

  const artifacts = useMemo(() => {
    if (!job) {
      return null;
    }

    const output = job.output;
    const planPath = output && "plan_path" in output ? output.plan_path : undefined;
    const narrationPath =
      output && "narration_path" in output ? output.narration_path : undefined;
    const musicPath = output && "music_path" in output ? output.music_path : undefined;
    const finalVideoPath =
      output && "final_video_path" in output ? output.final_video_path : undefined;
    const imagePath = output && "image_path" in output ? output.image_path : undefined;

    const fromArtifacts =
      job.artifacts?.reduce<Record<string, string>>((accumulator, artifact) => {
        accumulator[artifact.label] = toAssetUrl(job.job_id, artifact.path);
        return accumulator;
      }, {}) ?? {};

    return {
      plan: fromArtifacts.plan ?? toAssetUrl(job.job_id, planPath),
      narration:
        fromArtifacts.narration ??
        fromArtifacts.voice ??
        toAssetUrl(job.job_id, narrationPath),
      music: fromArtifacts.music ?? toAssetUrl(job.job_id, musicPath),
      finalVideo: fromArtifacts.final_video ?? fromArtifacts.video ?? toAssetUrl(job.job_id, finalVideoPath),
      image: fromArtifacts.image ?? toAssetUrl(job.job_id, imagePath),
    };
  }, [job]);

  useEffect(() => {
    void fetchJobs();
  }, []);

  useEffect(() => {
    void fetchQuota();
  }, []);

  useEffect(() => {
    if (!job || job.status !== "processing") {
      return;
    }

    const timer = window.setInterval(() => {
      void fetchJob(job.job_id, false);
    }, 3000);

    return () => {
      window.clearInterval(timer);
    };
  }, [job]);

  useEffect(() => {
    if (!job || job.status !== "completed" || !artifacts?.plan) {
      setPlan(null);
      return;
    }

    void fetch(artifacts.plan)
      .then((response) => {
        if (!response.ok) {
          throw new Error("failed to fetch plan");
        }
        return response.json() as Promise<PlanData>;
      })
      .then(setPlan)
      .catch(() => {
        setPlan(null);
      });
  }, [artifacts?.plan, job]);

  async function fetchQuota() {
    setIsQuotaLoading(true);
    setQuotaError("");

    try {
      const nextQuota = await requestJSON<QuotaResult>("/api/v1/quota");
      setQuota(nextQuota);
    } catch (error) {
      setQuotaError(error instanceof Error ? error.message : "加载额度失败");
    } finally {
      setIsQuotaLoading(false);
    }
  }

  async function fetchJobs() {
    try {
      const result = await requestJSON<JobListResult>("/api/v1/jobs");
      setRecentJobs(result.jobs.slice(0, 8));

      const persistedHistory = loadHistory();
      if (!job && !result.jobs.length && persistedHistory.length > 0) {
        setRecentJobs(
          persistedHistory.map((jobID) => ({
            job_id: jobID,
            status: "pending",
            progress: 0,
            stage: "unknown",
          })),
        );
      }
    } catch {
      const persistedHistory = loadHistory();
      setRecentJobs(
        persistedHistory.map((jobID) => ({
          job_id: jobID,
          status: "pending",
          progress: 0,
          stage: "unknown",
        })),
      );
    }
  }

  async function fetchJob(jobID: string, focusJob = true) {
    setJobError("");

    try {
      const nextJob = await requestJSON<Job>(`/api/v1/jobs/${jobID}`);
      if (focusJob) {
        setJob(nextJob);
      } else {
        setJob((current) => (current?.job_id === jobID ? nextJob : current));
      }

      setRecentJobs((current) => {
        const nextJobs = [nextJob, ...current.filter((item) => item.job_id !== jobID)].slice(0, 8);
        const nextHistory = nextJobs.map((item) => item.job_id);
        saveHistory(nextHistory);
        return nextJobs;
      });
    } catch (error) {
      setJobError(error instanceof Error ? error.message : "读取任务失败");
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setSubmitError("");
    setJobError("");

    try {
      const response = await requestJSON<{ job_id: string; status: string }>("/api/v1/make", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          theme: form.theme.trim(),
          scene_count: form.scene_count,
          scene_duration: form.scene_duration,
          language: form.language,
          input_video: form.input_video.trim(),
        }),
      });

      await fetchJob(response.job_id);
      await fetchJobs();
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : "任务提交失败");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handlePlanSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setSubmitError("");
    setJobError("");

    try {
      const response = await requestJSON<{ job_id: string; status: string }>("/api/v1/plan", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          theme: planForm.theme.trim(),
          scene_count: planForm.scene_count,
          scene_duration: planForm.scene_duration,
          language: planForm.language,
        }),
      });

      await fetchJob(response.job_id);
      await fetchJobs();
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : "分镜任务提交失败");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleClipSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setSubmitError("");
    setJobError("");

    try {
      const response = await requestJSON<{ job_id: string; status: string }>("/api/v1/clip", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          prompt: clipForm.prompt.trim(),
          subject: clipForm.subject.trim(),
          aspect_ratio: clipForm.aspect_ratio,
          duration: clipForm.duration,
          resolution: clipForm.resolution,
        }),
      });

      await fetchJob(response.job_id);
      await fetchJobs();
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : "图生视频任务提交失败");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="app-shell">
      <header className="hero">
        <div className="hero-copy">
          <p className="eyebrow">MiniMax Studio</p>
          <h1>把 Go 工作流包装成真正可用的视频工作台</h1>
          <p className="hero-text">
            当前前端先聚焦最高频路径：输入主题，提交整片生成任务，持续查看状态，并直接预览最终视频与中间资产。
          </p>
        </div>
        <div className="hero-metrics">
          <div className="metric-card">
            <span>接口入口</span>
            <strong>`/api/v1/make`</strong>
          </div>
          <div className="metric-card">
            <span>任务查询</span>
            <strong>`/api/v1/jobs/:id`</strong>
          </div>
          <div className="metric-card">
            <span>产物访问</span>
            <strong>`/api/v1/output/*`</strong>
          </div>
        </div>
      </header>

      <main className="workspace">
        <section className="panel form-panel">
          <div className="panel-heading">
            <div>
              <p className="panel-kicker">Create</p>
              <h2>Workflow Launcher</h2>
            </div>
            <span className="panel-note">多入口工作流</span>
          </div>

          <div className="mode-switcher">
            <button
              type="button"
              className={`mode-chip ${mode === "make" ? "active" : ""}`}
              onClick={() => setMode("make")}
            >
              Make
            </button>
            <button
              type="button"
              className={`mode-chip ${mode === "plan" ? "active" : ""}`}
              onClick={() => setMode("plan")}
            >
              Plan
            </button>
            <button
              type="button"
              className={`mode-chip ${mode === "clip" ? "active" : ""}`}
              onClick={() => setMode("clip")}
            >
              Clip
            </button>
          </div>

          {mode === "make" ? (
            <form className="make-form" onSubmit={handleSubmit}>
              <label className="field">
                <span>主题</span>
                <textarea
                  rows={5}
                  value={form.theme}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, theme: event.target.value }))
                  }
                  placeholder="输入一个适合生成短片的故事主题"
                  required
                />
              </label>

              <div className="field-grid">
                <label className="field">
                  <span>镜头数</span>
                  <input
                    type="number"
                    min={1}
                    max={8}
                    value={form.scene_count}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        scene_count: Number(event.target.value),
                      }))
                    }
                  />
                </label>

                <label className="field">
                  <span>单镜头时长</span>
                  <input
                    type="number"
                    min={3}
                    max={10}
                    value={form.scene_duration}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        scene_duration: Number(event.target.value),
                      }))
                    }
                  />
                </label>

                <label className="field">
                  <span>语言</span>
                  <select
                    value={form.language}
                    onChange={(event) =>
                      setForm((current) => ({ ...current, language: event.target.value }))
                    }
                  >
                    <option value="zh">中文</option>
                    <option value="en">English</option>
                  </select>
                </label>
              </div>

              <label className="field">
                <span>复用已有视频</span>
                <input
                  type="text"
                  value={form.input_video}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, input_video: event.target.value }))
                  }
                  placeholder="可选，本地服务端可访问的路径"
                />
              </label>

              {submitError ? <p className="error-banner">{submitError}</p> : null}

              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>
                  {isSubmitting ? "提交中..." : "生成短片"}
                </button>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => setForm(DEFAULT_FORM)}
                  disabled={isSubmitting}
                >
                  重置
                </button>
              </div>
            </form>
          ) : null}

          {mode === "plan" ? (
            <form className="make-form" onSubmit={handlePlanSubmit}>
              <label className="field">
                <span>主题</span>
                <textarea
                  rows={5}
                  value={planForm.theme}
                  onChange={(event) =>
                    setPlanForm((current) => ({ ...current, theme: event.target.value }))
                  }
                  placeholder="先用 plan 快速验证故事结构和旁白"
                  required
                />
              </label>

              <div className="field-grid">
                <label className="field">
                  <span>镜头数</span>
                  <input
                    type="number"
                    min={1}
                    max={8}
                    value={planForm.scene_count}
                    onChange={(event) =>
                      setPlanForm((current) => ({
                        ...current,
                        scene_count: Number(event.target.value),
                      }))
                    }
                  />
                </label>

                <label className="field">
                  <span>单镜头时长</span>
                  <input
                    type="number"
                    min={3}
                    max={10}
                    value={planForm.scene_duration}
                    onChange={(event) =>
                      setPlanForm((current) => ({
                        ...current,
                        scene_duration: Number(event.target.value),
                      }))
                    }
                  />
                </label>

                <label className="field">
                  <span>语言</span>
                  <select
                    value={planForm.language}
                    onChange={(event) =>
                      setPlanForm((current) => ({ ...current, language: event.target.value }))
                    }
                  >
                    <option value="zh">中文</option>
                    <option value="en">English</option>
                  </select>
                </label>
              </div>

              {submitError ? <p className="error-banner">{submitError}</p> : null}

              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>
                  {isSubmitting ? "提交中..." : "生成分镜规划"}
                </button>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => setPlanForm(DEFAULT_PLAN_FORM)}
                  disabled={isSubmitting}
                >
                  重置
                </button>
              </div>
            </form>
          ) : null}

          {mode === "clip" ? (
            <form className="make-form" onSubmit={handleClipSubmit}>
              <label className="field">
                <span>首帧提示词</span>
                <textarea
                  rows={4}
                  value={clipForm.prompt}
                  onChange={(event) =>
                    setClipForm((current) => ({ ...current, prompt: event.target.value }))
                  }
                  placeholder="用于生成首帧图像"
                  required
                />
              </label>

              <label className="field">
                <span>视频运动提示词</span>
                <textarea
                  rows={4}
                  value={clipForm.subject}
                  onChange={(event) =>
                    setClipForm((current) => ({ ...current, subject: event.target.value }))
                  }
                  placeholder="描述镜头运动和主体行为"
                  required
                />
              </label>

              <div className="field-grid">
                <label className="field">
                  <span>画幅</span>
                  <select
                    value={clipForm.aspect_ratio}
                    onChange={(event) =>
                      setClipForm((current) => ({ ...current, aspect_ratio: event.target.value }))
                    }
                  >
                    <option value="16:9">16:9</option>
                    <option value="9:16">9:16</option>
                    <option value="1:1">1:1</option>
                  </select>
                </label>

                <label className="field">
                  <span>时长</span>
                  <input
                    type="number"
                    min={5}
                    max={10}
                    value={clipForm.duration}
                    onChange={(event) =>
                      setClipForm((current) => ({
                        ...current,
                        duration: Number(event.target.value),
                      }))
                    }
                  />
                </label>

                <label className="field">
                  <span>分辨率</span>
                  <select
                    value={clipForm.resolution}
                    onChange={(event) =>
                      setClipForm((current) => ({ ...current, resolution: event.target.value }))
                    }
                  >
                    <option value="720p">720p</option>
                    <option value="1080p">1080p</option>
                  </select>
                </label>
              </div>

              {submitError ? <p className="error-banner">{submitError}</p> : null}

              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>
                  {isSubmitting ? "提交中..." : "生成单镜头"}
                </button>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => setClipForm(DEFAULT_CLIP_FORM)}
                  disabled={isSubmitting}
                >
                  重置
                </button>
              </div>
            </form>
          ) : null}

          <div className="history-block">
            <div className="history-header">
              <h3>最近任务</h3>
              <button type="button" className="text-link history-refresh" onClick={() => void fetchJobs()}>
                刷新列表
              </button>
            </div>
            <div className="history-list">
              {recentJobs.length === 0 ? (
                <p className="muted-text">还没有任务记录。</p>
              ) : (
                recentJobs.map((item) => (
                  <button
                    key={item.job_id}
                    className="history-chip"
                    type="button"
                    onClick={() => void fetchJob(item.job_id)}
                  >
                    <span>{item.job_id}</span>
                    <small>{formatStatus(item.status)}</small>
                  </button>
                ))
              )}
            </div>
          </div>
        </section>

        <section className="panel status-panel">
          <div className="panel-heading">
            <div>
              <p className="panel-kicker">Monitor</p>
              <h2>Job Console</h2>
            </div>
            {job ? <span className={`status-badge ${job.status}`}>{formatStatus(job.status)}</span> : null}
          </div>

          {job ? (
            <div className="job-stack">
              <div className="job-meta">
                <div>
                  <span className="meta-label">Job ID</span>
                  <strong>{job.job_id}</strong>
                </div>
                <div>
                  <span className="meta-label">Stage</span>
                  <strong>{job.stage || "make"}</strong>
                </div>
                <div>
                  <span className="meta-label">Progress</span>
                  <strong>{Math.round((job.progress || 0) * 100)}%</strong>
                </div>
              </div>

              <div className="progress-rail" aria-hidden="true">
                <div
                  className="progress-fill"
                  style={{ width: `${Math.max(8, Math.round((job.progress || 0) * 100))}%` }}
                />
              </div>

              <p className="status-copy">
                {job.status === "processing"
                  ? "当前后端只暴露任务级状态，前端会持续轮询直到完成或失败。"
                  : job.status === "completed"
                    ? "任务已完成，可以直接预览视频、旁白和分镜规划。"
                    : "任务执行失败，优先检查 API Key、MiniMax 配额和输入路径。"}
              </p>

              {job.error ? <p className="error-banner">{job.error}</p> : null}
              {jobError ? <p className="error-banner">{jobError}</p> : null}

              <div className="console-actions">
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => void fetchJob(job.job_id)}
                >
                  手动刷新
                </button>
                {artifacts?.finalVideo ? (
                  <a className="text-link" href={artifacts.finalVideo} target="_blank" rel="noreferrer">
                    打开最终视频
                  </a>
                ) : null}
              </div>

              {job.logs?.length ? (
                <div className="log-panel">
                  <div className="artifact-header">
                    <h3>Recent Logs</h3>
                    <span>{job.logs.length} entries</span>
                  </div>
                  <div className="log-list">
                    {job.logs.slice().reverse().map((entry, index) => (
                      <div className="log-row" key={`${entry.time}-${index}`}>
                        <span>{new Date(entry.time).toLocaleString()}</span>
                        <p>{entry.message}</p>
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}

              {artifacts ? (
                <div className="artifact-grid">
                  {artifacts.image ? (
                    <article className="artifact-card">
                      <div className="artifact-header">
                        <h3>Key Frame</h3>
                        <a href={artifacts.image} target="_blank" rel="noreferrer">
                          查看原图
                        </a>
                      </div>
                      <img className="image-frame" src={artifacts.image} alt="Generated key frame" />
                    </article>
                  ) : null}

                  <article className="artifact-card feature-card">
                    <div className="artifact-header">
                      <h3>{artifacts.image && !artifacts.plan ? "Clip Video" : "Final Video"}</h3>
                      {artifacts.finalVideo ? (
                        <a href={artifacts.finalVideo} target="_blank" rel="noreferrer">
                          新窗口打开
                        </a>
                      ) : null}
                    </div>
                    {artifacts.finalVideo ? (
                      <video controls className="media-frame" src={artifacts.finalVideo} />
                    ) : (
                      <p className="muted-text">等待产物生成。</p>
                    )}
                  </article>

                  <article className="artifact-card">
                    <div className="artifact-header">
                      <h3>Narration</h3>
                      {artifacts.narration ? (
                        <a href={artifacts.narration} target="_blank" rel="noreferrer">
                          下载
                        </a>
                      ) : null}
                    </div>
                    {artifacts.narration ? (
                      <audio controls className="audio-frame" src={artifacts.narration} />
                    ) : (
                      <p className="muted-text">暂无旁白文件。</p>
                    )}
                  </article>

                  <article className="artifact-card">
                    <div className="artifact-header">
                      <h3>Music</h3>
                      {artifacts.music ? (
                        <a href={artifacts.music} target="_blank" rel="noreferrer">
                          下载
                        </a>
                      ) : null}
                    </div>
                    {artifacts.music ? (
                      <audio controls className="audio-frame" src={artifacts.music} />
                    ) : (
                      <p className="muted-text">当前任务没有背景音乐文件，或后端以 optional 模式跳过了生成。</p>
                    )}
                  </article>

                  <article className="artifact-card plan-card">
                    <div className="artifact-header">
                      <h3>Storyboard Plan</h3>
                      {artifacts.plan ? (
                        <a href={artifacts.plan} target="_blank" rel="noreferrer">
                          查看 JSON
                        </a>
                      ) : null}
                    </div>
                    {plan ? (
                      <div className="plan-stack">
                        <div className="plan-summary">
                          <strong>{plan.title || "未命名计划"}</strong>
                          <p>{plan.visual_style}</p>
                        </div>
                        <div className="scene-list">
                          {plan.scenes.map((scene, index) => (
                            <div className="scene-card" key={`${scene.name}-${index}`}>
                              <span className="scene-index">{String(index + 1).padStart(2, "0")}</span>
                              <div>
                                <strong>{scene.name}</strong>
                                <p>{scene.video_prompt}</p>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    ) : (
                      <p className="muted-text">完成后会在这里加载 `plan.json` 的结构化内容。</p>
                    )}
                  </article>
                </div>
              ) : (
                <div className="empty-state">
                  <p>提交一个 `make` 任务后，这里会显示状态、日志说明和最终产物。</p>
                </div>
              )}
            </div>
          ) : (
            <div className="empty-state">
              <p>当前没有选中任务。左侧提交新任务，或者从最近任务列表恢复一个 Job。</p>
            </div>
          )}
        </section>

        <section className="panel quota-panel">
          <div className="panel-heading">
            <div>
              <p className="panel-kicker">Capacity</p>
              <h2>Quota Snapshot</h2>
            </div>
            <button type="button" className="ghost-button" onClick={() => void fetchQuota()}>
              {isQuotaLoading ? "刷新中..." : "刷新额度"}
            </button>
          </div>

          {quotaError ? <p className="error-banner">{quotaError}</p> : null}

          <div className="quota-list">
            {quota?.entries?.length ? (
              quota.entries.map((entry) => {
                const intervalUsage = usagePercent(
                  entry.current_interval_usage_count,
                  entry.current_interval_total_count,
                );
                const weeklyUsage = usagePercent(
                  entry.current_weekly_usage_count,
                  entry.current_weekly_total_count,
                );

                return (
                  <article className="quota-card" key={entry.model_name}>
                    <div className="quota-title">
                      <strong>{entry.model_name}</strong>
                    </div>
                    <div className="quota-row">
                      <span>当前周期</span>
                      <strong>
                        {entry.current_interval_usage_count}/{entry.current_interval_total_count}
                      </strong>
                    </div>
                    <div className="mini-bar">
                      <div style={{ width: `${intervalUsage}%` }} />
                    </div>
                    <div className="quota-row">
                      <span>周配额</span>
                      <strong>
                        {entry.current_weekly_usage_count}/{entry.current_weekly_total_count}
                      </strong>
                    </div>
                    <div className="mini-bar warm">
                      <div style={{ width: `${weeklyUsage}%` }} />
                    </div>
                  </article>
                );
              })
            ) : (
              <p className="muted-text">还没有额度数据。启动 Go server 并配置好 `MINIMAX_API_KEY` 后再刷新。</p>
            )}
          </div>
        </section>
      </main>
    </div>
  );
}
