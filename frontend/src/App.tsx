import { FormEvent, useEffect, useMemo, useState } from "react";

type MakeRequest = {
  theme: string;
  scene_count: number;
  scene_duration: number;
  language: string;
  input_video: string;
};

type MakeResult = {
  output_dir: string;
  plan_path: string;
  narration_path: string;
  music_path?: string;
  final_video_path: string;
};

type Job = {
  job_id: string;
  status: "pending" | "processing" | "completed" | "failed";
  progress: number;
  stage: string;
  output?: MakeResult;
  error?: string;
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
  const [form, setForm] = useState<MakeRequest>(DEFAULT_FORM);
  const [job, setJob] = useState<Job | null>(null);
  const [jobHistory, setJobHistory] = useState<string[]>([]);
  const [quota, setQuota] = useState<QuotaResult | null>(null);
  const [plan, setPlan] = useState<PlanData | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isQuotaLoading, setIsQuotaLoading] = useState(false);
  const [quotaError, setQuotaError] = useState("");
  const [jobError, setJobError] = useState("");
  const [submitError, setSubmitError] = useState("");

  const artifacts = useMemo(() => {
    if (!job?.output) {
      return null;
    }

    return {
      plan: toAssetUrl(job.job_id, job.output.plan_path),
      narration: toAssetUrl(job.job_id, job.output.narration_path),
      music: toAssetUrl(job.job_id, job.output.music_path),
      finalVideo: toAssetUrl(job.job_id, job.output.final_video_path),
    };
  }, [job]);

  useEffect(() => {
    setJobHistory(loadHistory());
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

  async function fetchJob(jobID: string, focusJob = true) {
    setJobError("");

    try {
      const nextJob = await requestJSON<Job>(`/api/v1/jobs/${jobID}`);
      if (focusJob) {
        setJob(nextJob);
      } else {
        setJob((current) => (current?.job_id === jobID ? nextJob : current));
      }

      setJobHistory((current) => {
        const nextHistory = [jobID, ...current.filter((item) => item !== jobID)];
        saveHistory(nextHistory);
        return nextHistory;
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
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : "任务提交失败");
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
              <h2>Make Workflow</h2>
            </div>
            <span className="panel-note">端到端短片生成</span>
          </div>

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

          <div className="history-block">
            <div className="history-header">
              <h3>最近任务</h3>
              <span>保存在浏览器本地</span>
            </div>
            <div className="history-list">
              {jobHistory.length === 0 ? (
                <p className="muted-text">还没有任务记录。</p>
              ) : (
                jobHistory.map((item) => (
                  <button
                    key={item}
                    className="history-chip"
                    type="button"
                    onClick={() => void fetchJob(item)}
                  >
                    {item}
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

              {artifacts ? (
                <div className="artifact-grid">
                  <article className="artifact-card feature-card">
                    <div className="artifact-header">
                      <h3>Final Video</h3>
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
