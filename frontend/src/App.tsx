import { FormEvent, useEffect, useMemo, useState } from "react";

import {
  DEFAULT_CLIP_FORM,
  DEFAULT_FORM,
  DEFAULT_IMAGE_FORM,
  DEFAULT_MUSIC_FORM,
  DEFAULT_PLAN_FORM,
  DEFAULT_STITCH_FORM,
  DEFAULT_VOICE_FORM,
  VOICE_OPTIONS,
} from "./app-data";
import type {
  ClipRequest,
  ImageRequest,
  Job,
  JobListResult,
  MakeRequest,
  MusicRequest,
  PlanData,
  PlanRequest,
  QuotaResult,
  ResultTab,
  StitchRequest,
  VoiceRequest,
  WorkflowMode,
} from "./app-types";
import {
  ErrorBoundary,
  formatStatus,
  isTextWindowQuota,
  loadHistory,
  modeHint,
  modeLabel,
  normalizeQuotaEntries,
  remainingCount,
  requestJSON,
  saveHistory,
  shortJobID,
  toAssetUrl,
  usagePercent,
} from "./app-utils";

/* ══════════════════════════════════════════════
   Inline SVG Icons (zero-dependency)
   ══════════════════════════════════════════════ */

function IconMake() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
    </svg>
  );
}

function IconPlan() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
      <polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/>
    </svg>
  );
}

function IconClip() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
    </svg>
  );
}

function IconImage() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/>
    </svg>
  );
}

function IconVoice() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"/>
    </svg>
  );
}

function IconMusic() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
    </svg>
  );
}

function IconStitch() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="4 17 10 11 16 17"/><line x1="4" y1="7" x2="20" y2="7"/><line x1="4" y1="12" x2="16" y2="12"/><line x1="4" y1="17" x2="10" y2="17"/>
    </svg>
  );
}

const MODE_ICONS: Record<WorkflowMode, React.ReactNode> = {
  make: <IconMake />,
  plan: <IconPlan />,
  clip: <IconClip />,
  image: <IconImage />,
  voice: <IconVoice />,
  music: <IconMusic />,
  stitch: <IconStitch />,
};

/* ══════════════════════════════════════════════
   App Component
   ══════════════════════════════════════════════ */

export default function App() {
  const [mode, setMode] = useState<WorkflowMode>("make");
  const [resultTab, setResultTab] = useState<ResultTab>("result");
  const [form, setForm] = useState<MakeRequest>(DEFAULT_FORM);
  const [planForm, setPlanForm] = useState<PlanRequest>(DEFAULT_PLAN_FORM);
  const [clipForm, setClipForm] = useState<ClipRequest>(DEFAULT_CLIP_FORM);
  const [imageForm, setImageForm] = useState<ImageRequest>(DEFAULT_IMAGE_FORM);
  const [voiceForm, setVoiceForm] = useState<VoiceRequest>(DEFAULT_VOICE_FORM);
  const [musicForm, setMusicForm] = useState<MusicRequest>(DEFAULT_MUSIC_FORM);
  const [stitchForm, setStitchForm] = useState<StitchRequest>(DEFAULT_STITCH_FORM);
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
    if (!job) return null;

    const output = job.output;
    const planPath = output && "plan_path" in output ? output.plan_path : undefined;
    const narrationPath =
      output && "narration_path" in output ? output.narration_path : undefined;
    const musicPath = output && "music_path" in output ? output.music_path : undefined;
    const finalVideoPath =
      output && "final_video_path" in output ? output.final_video_path : undefined;
    const imagePath = output && "image_path" in output ? output.image_path : undefined;
    const audioPath = output && "output_path" in output ? output.output_path : undefined;

    const fromArtifacts =
      job.artifacts?.reduce<Record<string, string>>((acc, a) => {
        acc[a.label] = toAssetUrl(job.job_id, a.path);
        return acc;
      }, {}) ?? {};

    return {
      plan: fromArtifacts.plan ?? toAssetUrl(job.job_id, planPath),
      narration:
        fromArtifacts.narration ??
        fromArtifacts.voice ??
        toAssetUrl(job.job_id, narrationPath ?? audioPath),
      music: fromArtifacts.music ?? toAssetUrl(job.job_id, musicPath ?? audioPath),
      finalVideo: fromArtifacts.final_video ?? fromArtifacts.video ?? toAssetUrl(job.job_id, finalVideoPath),
      image: fromArtifacts.image ?? toAssetUrl(job.job_id, imagePath),
    };
  }, [job]);

  const quotaSummary = useMemo(() => {
    const entries = normalizeQuotaEntries(quota);
    return entries.reduce(
      (acc, entry) => {
        if (isTextWindowQuota(entry.model_name)) {
          acc.window += remainingCount(
            entry.current_interval_total_count,
            entry.current_interval_usage_count,
          );
        } else {
          acc.daily += remainingCount(
            entry.current_interval_total_count,
            entry.current_interval_usage_count,
          );
        }
        return acc;
      },
      { window: 0, daily: 0 },
    );
  }, [quota]);

  /* ── Effects ── */

  useEffect(() => { void fetchJobs(); }, []);
  useEffect(() => { void fetchQuota(); }, []);

  useEffect(() => {
    if (!job || job.status !== "processing") return;
    const timer = window.setInterval(() => void fetchJob(job.job_id, false), 3000);
    return () => window.clearInterval(timer);
  }, [job]);

  useEffect(() => {
    if (!job || job.status !== "completed" || !artifacts?.plan) { setPlan(null); return; }
    void fetch(artifacts.plan)
      .then((r) => { if (!r.ok) throw new Error("fail"); return r.json() as Promise<PlanData>; })
      .then(setPlan)
      .catch(() => setPlan(null));
  }, [artifacts?.plan, job]);

  useEffect(() => {
    if (!job) { setResultTab("result"); return; }
    if (job.logs?.length) setResultTab("result");
  }, [job?.job_id]);

  /* ── API calls ── */

  async function fetchQuota() {
    setIsQuotaLoading(true);
    setQuotaError("");
    try {
      const q = await requestJSON<QuotaResult>("/api/v1/quota");
      setQuota(q);
    } catch (e) {
      setQuotaError(e instanceof Error ? e.message : "加载额度失败");
    } finally {
      setIsQuotaLoading(false);
    }
  }

  async function fetchJobs() {
    try {
      const r = await requestJSON<JobListResult>("/api/v1/jobs");
      setRecentJobs(r.jobs.slice(0, 8));
      const h = loadHistory();
      if (!job && !r.jobs.length && h.length > 0) {
        setRecentJobs(h.map((id) => ({ job_id: id, status: "pending", progress: 0, stage: "unknown" })));
      }
    } catch {
      const h = loadHistory();
      setRecentJobs(h.map((id) => ({ job_id: id, status: "pending", progress: 0, stage: "unknown" })));
    }
  }

  async function fetchJob(jobID: string, focusJob = true) {
    setJobError("");
    try {
      const j = await requestJSON<Job>(`/api/v1/jobs/${jobID}`);
      if (focusJob) setJob(j);
      else setJob((c) => (c?.job_id === jobID ? j : c));
      setRecentJobs((prev) => {
        const next = [j, ...prev.filter((i) => i.job_id !== jobID)].slice(0, 8);
        saveHistory(next.map((i) => i.job_id));
        return next;
      });
    } catch (e) {
      setJobError(e instanceof Error ? e.message : "读取任务失败");
    }
  }

  /* ── Submit handlers ── */

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string; status: string }>("/api/v1/make", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: form.theme.trim(), scene_count: form.scene_count, scene_duration: form.scene_duration, language: form.language, input_video: form.input_video.trim() }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handlePlanSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string; status: string }>("/api/v1/plan", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: planForm.theme.trim(), scene_count: planForm.scene_count, scene_duration: planForm.scene_duration, language: planForm.language }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "分镜任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleClipSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string; status: string }>("/api/v1/clip", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: clipForm.prompt.trim(), subject: clipForm.subject.trim(), aspect_ratio: clipForm.aspect_ratio, duration: clipForm.duration, resolution: clipForm.resolution }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "图生视频任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleVoiceSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string; status: string }>("/api/v1/voice", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: voiceForm.text.trim(), voice_id: voiceForm.voice_id, audio_format: voiceForm.audio_format }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "语音任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleImageSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string; status: string }>("/api/v1/image", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: imageForm.prompt.trim(), aspect_ratio: imageForm.aspect_ratio }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "图片任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleMusicSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string; status: string }>("/api/v1/music", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: musicForm.prompt.trim(), audio_format: musicForm.audio_format }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "音乐任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleStitchSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    const videos = stitchForm.videos.split("\n").map((s) => s.trim()).filter(Boolean);
    try {
      const r = await requestJSON<{ job_id: string; status: string }>("/api/v1/stitch", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ videos, narration: stitchForm.narration.trim(), music: stitchForm.music.trim() }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "素材合成任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  /* ── Render helpers ── */

  function renderModeSwitcher() {
    const modes: WorkflowMode[] = ["make", "plan", "clip", "image", "voice", "music", "stitch"];
    return (
      <div className="mode-switcher">
        {modes.map((m) => (
          <button key={m} type="button" className={`mode-chip ${mode === m ? "active" : ""}`} onClick={() => setMode(m)}>
            <span className="mode-icon">{MODE_ICONS[m]}</span>
            {m.charAt(0).toUpperCase() + m.slice(1)}
          </button>
        ))}
      </div>
    );
  }

  function renderEmptyState(icon: string, title: string, desc?: string) {
    return (
      <div className="empty-state">
        <span className="empty-icon">{icon}</span>
        <p>{title}</p>
        {desc ? <small style={{ color: "var(--text-muted)", opacity: 0.6 }}>{desc}</small> : null}
      </div>
    );
  }

  /* ══════════════════════════════════════════════
     JSX
     ══════════════════════════════════════════════ */

  return (
    <ErrorBoundary>
    <div className="app-shell">
      {/* ── Top bar ── */}
      <header className="topbar">
        <div className="topbar-brand">
          <h1><span className="brand-icon">M</span> MiniMax Studio</h1>
          <span>AI Creative Workbench</span>
        </div>
        <div className="topbar-actions">
          <button type="button" className="ghost-button" style={{ fontSize: "0.78rem", padding: "6px 12px" }} onClick={() => void fetchJobs()}>
            刷新
          </button>
        </div>
      </header>

      {/* ── Workspace grid ── */}
      <main className="workspace">

        {/* ═══ Left: Form panel ═══ */}
        <section className="panel form-panel">
          <div className="panel-heading">
            <div><h2>{modeLabel(mode)}</h2></div>
            <span className="panel-note">{modeHint(mode)}</span>
          </div>

          {renderModeSwitcher()}

          {/* Make form */}
          {mode === "make" ? (
            <form className="make-form compact-form" onSubmit={handleSubmit}>
              <label className="field"><span>主题</span>
                <textarea rows={4} value={form.theme} onChange={(e) => setForm((f) => ({ ...f, theme: e.target.value }))} placeholder="输入一个适合生成短片的故事主题" required />
              </label>
              <div className="field-grid">
                <label className="field"><span>镜头数</span>
                  <input type="number" min={1} max={8} value={form.scene_count} onChange={(e) => setForm((f) => ({ ...f, scene_count: Number(e.target.value) }))} />
                </label>
                <label className="field"><span>单镜头时长</span>
                  <input type="number" min={3} max={10} value={form.scene_duration} onChange={(e) => setForm((f) => ({ ...f, scene_duration: Number(e.target.value) }))} />
                </label>
                <label className="field"><span>语言</span>
                  <select value={form.language} onChange={(e) => setForm((f) => ({ ...f, language: e.target.value }))}>
                    <option value="zh">中文</option><option value="en">English</option>
                  </select>
                </label>
              </div>
              <label className="field"><span>复用已有视频</span>
                <input type="text" value={form.input_video} onChange={(e) => setForm((f) => ({ ...f, input_video: e.target.value }))} placeholder="可选，本地服务端可访问的路径" />
              </label>
              {submitError ? <p className="error-banner">{submitError}</p> : null}
              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>
                  {isSubmitting ? "提交中..." : "生成短片"}
                </button>
                <button type="button" className="ghost-button" onClick={() => setForm(DEFAULT_FORM)} disabled={isSubmitting}>重置</button>
              </div>
            </form>
          ) : null}

          {/* Plan form */}
          {mode === "plan" ? (
            <form className="make-form compact-form" onSubmit={handlePlanSubmit}>
              <label className="field"><span>主题</span>
                <textarea rows={4} value={planForm.theme} onChange={(e) => setPlanForm((f) => ({ ...f, theme: e.target.value }))} placeholder="先用 plan 快速验证故事结构和旁白" required />
              </label>
              <div className="field-grid">
                <label className="field"><span>镜头数</span>
                  <input type="number" min={1} max={8} value={planForm.scene_count} onChange={(e) => setPlanForm((f) => ({ ...f, scene_count: Number(e.target.value) }))} />
                </label>
                <label className="field"><span>单镜头时长</span>
                  <input type="number" min={3} max={10} value={planForm.scene_duration} onChange={(e) => setPlanForm((f) => ({ ...f, scene_duration: Number(e.target.value) }))} />
                </label>
                <label className="field"><span>语言</span>
                  <select value={planForm.language} onChange={(e) => setPlanForm((f) => ({ ...f, language: e.target.value }))}>
                    <option value="zh">中文</option><option value="en">English</option>
                  </select>
                </label>
              </div>
              {submitError ? <p className="error-banner">{submitError}</p> : null}
              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>{isSubmitting ? "提交中..." : "生成分镜规划"}</button>
                <button type="button" className="ghost-button" onClick={() => setPlanForm(DEFAULT_PLAN_FORM)} disabled={isSubmitting}>重置</button>
              </div>
            </form>
          ) : null}

          {/* Clip form */}
          {mode === "clip" ? (
            <form className="make-form compact-form" onSubmit={handleClipSubmit}>
              <label className="field"><span>首帧提示词</span>
                <textarea rows={3} value={clipForm.prompt} onChange={(e) => setClipForm((f) => ({ ...f, prompt: e.target.value }))} placeholder="用于生成首帧图像" required />
              </label>
              <label className="field"><span>视频运动提示词</span>
                <textarea rows={3} value={clipForm.subject} onChange={(e) => setClipForm((f) => ({ ...f, subject: e.target.value }))} placeholder="描述镜头运动和主体行为" required />
              </label>
              <div className="field-grid">
                <label className="field"><span>画幅</span>
                  <select value={clipForm.aspect_ratio} onChange={(e) => setClipForm((f) => ({ ...f, aspect_ratio: e.target.value }))}>
                    <option value="16:9">16:9</option><option value="9:16">9:16</option><option value="1:1">1:1</option>
                  </select>
                </label>
                <label className="field"><span>时长</span>
                  <input type="number" min={5} max={10} value={clipForm.duration} onChange={(e) => setClipForm((f) => ({ ...f, duration: Number(e.target.value) }))} />
                </label>
                <label className="field"><span>分辨率</span>
                  <select value={clipForm.resolution} onChange={(e) => setClipForm((f) => ({ ...f, resolution: e.target.value }))}>
                    <option value="720p">720p</option><option value="1080p">1080p</option>
                  </select>
                </label>
              </div>
              {submitError ? <p className="error-banner">{submitError}</p> : null}
              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>{isSubmitting ? "提交中..." : "生成单镜头"}</button>
                <button type="button" className="ghost-button" onClick={() => setClipForm(DEFAULT_CLIP_FORM)} disabled={isSubmitting}>重置</button>
              </div>
            </form>
          ) : null}

          {/* Image form */}
          {mode === "image" ? (
            <form className="make-form compact-form" onSubmit={handleImageSubmit}>
              <label className="field"><span>图片提示词</span>
                <textarea rows={4} value={imageForm.prompt} onChange={(e) => setImageForm((f) => ({ ...f, prompt: e.target.value }))} placeholder="输入单图生成提示词" required />
              </label>
              <label className="field"><span>画幅</span>
                <select value={imageForm.aspect_ratio} onChange={(e) => setImageForm((f) => ({ ...f, aspect_ratio: e.target.value }))}>
                  <option value="16:9">16:9</option><option value="9:16">9:16</option><option value="1:1">1:1</option>
                </select>
              </label>
              {submitError ? <p className="error-banner">{submitError}</p> : null}
              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>{isSubmitting ? "提交中..." : "生成图片"}</button>
                <button type="button" className="ghost-button" onClick={() => setImageForm(DEFAULT_IMAGE_FORM)} disabled={isSubmitting}>重置</button>
              </div>
            </form>
          ) : null}

          {/* Voice form */}
          {mode === "voice" ? (
            <form className="make-form compact-form" onSubmit={handleVoiceSubmit}>
              <label className="field"><span>旁白文本</span>
                <textarea rows={4} value={voiceForm.text} onChange={(e) => setVoiceForm((f) => ({ ...f, text: e.target.value }))} required />
              </label>
              <div className="field-grid">
                <label className="field"><span>音色</span>
                  <select value={voiceForm.voice_id} onChange={(e) => setVoiceForm((f) => ({ ...f, voice_id: e.target.value }))}>
                    {VOICE_OPTIONS.map((o) => <option key={o.id} value={o.id}>{o.label}</option>)}
                  </select>
                </label>
                <label className="field"><span>格式</span>
                  <select value={voiceForm.audio_format} onChange={(e) => setVoiceForm((f) => ({ ...f, audio_format: e.target.value }))}>
                    <option value="mp3">mp3</option><option value="wav">wav</option>
                  </select>
                </label>
              </div>
              {submitError ? <p className="error-banner">{submitError}</p> : null}
              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>{isSubmitting ? "提交中..." : "生成语音"}</button>
                <button type="button" className="ghost-button" onClick={() => setVoiceForm(DEFAULT_VOICE_FORM)} disabled={isSubmitting}>重置</button>
              </div>
            </form>
          ) : null}

          {/* Music form */}
          {mode === "music" ? (
            <form className="make-form compact-form" onSubmit={handleMusicSubmit}>
              <label className="field"><span>音乐提示词</span>
                <textarea rows={4} value={musicForm.prompt} onChange={(e) => setMusicForm((f) => ({ ...f, prompt: e.target.value }))} required />
              </label>
              <label className="field"><span>格式</span>
                <select value={musicForm.audio_format} onChange={(e) => setMusicForm((f) => ({ ...f, audio_format: e.target.value }))}>
                  <option value="mp3">mp3</option><option value="wav">wav</option>
                </select>
              </label>
              {submitError ? <p className="error-banner">{submitError}</p> : null}
              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>{isSubmitting ? "提交中..." : "生成音乐"}</button>
                <button type="button" className="ghost-button" onClick={() => setMusicForm(DEFAULT_MUSIC_FORM)} disabled={isSubmitting}>重置</button>
              </div>
            </form>
          ) : null}

          {/* Stitch form */}
          {mode === "stitch" ? (
            <form className="make-form compact-form" onSubmit={handleStitchSubmit}>
              <label className="field"><span>视频路径列表</span>
                <textarea rows={4} value={stitchForm.videos} onChange={(e) => setStitchForm((f) => ({ ...f, videos: e.target.value }))} placeholder={"每行一个服务端可访问的视频路径"} required />
              </label>
              <label className="field"><span>旁白路径</span>
                <input type="text" value={stitchForm.narration} onChange={(e) => setStitchForm((f) => ({ ...f, narration: e.target.value }))} required />
              </label>
              <label className="field"><span>背景音乐路径</span>
                <input type="text" value={stitchForm.music} onChange={(e) => setStitchForm((f) => ({ ...f, music: e.target.value }))} placeholder="可选" />
              </label>
              {submitError ? <p className="error-banner">{submitError}</p> : null}
              <div className="form-actions">
                <button type="submit" className="primary-button" disabled={isSubmitting}>{isSubmitting ? "提交中..." : "合成成片"}</button>
                <button type="button" className="ghost-button" onClick={() => setStitchForm(DEFAULT_STITCH_FORM)} disabled={isSubmitting}>重置</button>
              </div>
            </form>
          ) : null}

          {/* History */}
          <div className="history-block">
            <div className="history-header">
              <h3>最近任务</h3>
              <button type="button" className="text-link history-refresh" onClick={() => void fetchJobs()}>刷新</button>
            </div>
            <div className="history-list">
              {recentJobs.length === 0 ? (
                <p className="muted-text" style={{ textAlign: "center", padding: "12px 0", fontSize: "0.82rem" }}>暂无任务记录</p>
              ) : recentJobs.map((item) => (
                <button key={item.job_id} className={`history-chip ${job?.job_id === item.job_id ? "active" : ""}`} type="button" onClick={() => void fetchJob(item.job_id)}>
                  <span>{shortJobID(item.job_id)}</span>
                  <small>{formatStatus(item.status)}</small>
                </button>
              ))}
            </div>
          </div>
        </section>

        {/* ═══ Middle: Job status panel ═══ */}
        <section className="panel status-panel">
          <div className="panel-heading">
            <div><h2>任务状态</h2></div>
            {job ? <span className={`status-badge ${job.status}`}>{formatStatus(job.status)}</span> : null}
          </div>

          {job ? (
            <div className="job-stack">
              <div className="job-strip">
                <span className="job-pill">{shortJobID(job.job_id)}</span>
                <span className="job-pill">{job.stage || "job"}</span>
                <span className="job-pill">{Math.round((job.progress || 0) * 100)}%</span>
              </div>

              <div className="progress-rail" aria-hidden="true">
                <div className="progress-fill" style={{ width: `${Math.max(6, Math.round((job.progress || 0) * 100))}%` }} />
              </div>

              <p className="status-copy">
                {job.status === "processing" ? "正在处理..." :
                 job.status === "completed" ? "已完成" : "失败"}
              </p>

              {job.error ? <p className="error-banner">{job.error}</p> : null}
              {jobError ? <p className="error-banner">{jobError}</p> : null}

              <div className="console-actions">
                <button type="button" className="ghost-button" onClick={() => void fetchJob(job.job_id)}>刷新</button>
                {artifacts?.finalVideo ? (
                  <a className="text-link" href={artifacts.finalVideo} target="_blank" rel="noreferrer">打开结果</a>
                ) : null}
              </div>

              {artifacts ? (
                <div className="result-shell">
                  <div className="result-tabs">
                    <button type="button" className={`result-tab ${resultTab === "result" ? "active" : ""}`} onClick={() => setResultTab("result")}>结果</button>
                    <button type="button" className={`result-tab ${resultTab === "plan" ? "active" : ""}`} onClick={() => setResultTab("plan")} disabled={!artifacts.plan}>规划</button>
                    <button type="button" className={`result-tab ${resultTab === "logs" ? "active" : ""}`} onClick={() => setResultTab("logs")} disabled={!job.logs?.length}>日志</button>
                  </div>

                  {/* Result tab */}
                  {resultTab === "result" ? (
                    <div className="result-grid">
                      <article className="artifact-card feature-card">
                        <div className="artifact-header">
                          <h3>{artifacts.finalVideo ? (artifacts.image && !artifacts.plan ? "视频" : "最终成片") : "图片"}</h3>
                          {artifacts.finalVideo ? (
                            <a href={artifacts.finalVideo} target="_blank" rel="noreferrer">打开</a>
                          ) : artifacts.image ? (
                            <a href={artifacts.image} target="_blank" rel="noreferrer">打开</a>
                          ) : null}
                        </div>
                        {artifacts.finalVideo ? (
                          <video controls className="media-frame" src={artifacts.finalVideo} />
                        ) : artifacts.image ? (
                          <img className="image-frame hero-image" src={artifacts.image} alt="Generated result" />
                        ) : renderEmptyState("⏳", "等待结果", "任务处理完成后将在此显示")}
                      </article>

                      {artifacts.image && artifacts.finalVideo ? (
                        <article className="artifact-card compact-card">
                          <div className="artifact-header"><h3>首帧</h3><a href={artifacts.image} target="_blank" rel="noreferrer">打开</a></div>
                          <img className="image-frame" src={artifacts.image} alt="Key frame" />
                        </article>
                      ) : null}

                      <article className="artifact-card compact-card">
                        <div className="artifact-header">
                          <h3>旁白</h3>
                          {artifacts.narration ? <a href={artifacts.narration} target="_blank" rel="noreferrer">打开</a> : null}
                        </div>
                        {artifacts.narration ? (
                          <audio controls className="audio-frame" src={artifacts.narration} />
                        ) : <p className="muted-text">无旁白</p>}
                      </article>

                      <article className="artifact-card compact-card">
                        <div className="artifact-header">
                          <h3>配乐</h3>
                          {artifacts.music ? <a href={artifacts.music} target="_blank" rel="noreferrer">打开</a> : null}
                        </div>
                        {artifacts.music ? (
                          <audio controls className="audio-frame" src={artifacts.music} />
                        ) : <p className="muted-text">无配乐</p>}
                      </article>
                    </div>
                  ) : null}

                  {/* Plan tab */}
                  {resultTab === "plan" ? (
                    <article className="artifact-card plan-card solo-card">
                      <div className="artifact-header">
                        <h3>分镜规划</h3>
                        {artifacts.plan ? <a href={artifacts.plan} target="_blank" rel="noreferrer">打开</a> : null}
                      </div>
                      {plan ? (
                        <div className="plan-stack">
                          <div className="plan-summary">
                            <strong>{plan.title || "未命名计划"}</strong>
                            <p>{plan.visual_style}</p>
                          </div>
                          <div className="scene-list">
                            {plan.scenes.map((scene, idx) => (
                              <div className="scene-card" key={`${scene.name}-${idx}`}>
                                <span className="scene-index">{String(idx + 1).padStart(2, "0")}</span>
                                <div>
                                  <strong>{scene.name}</strong>
                                  <p>{scene.video_prompt}</p>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      ) : renderEmptyState("📋", "暂无规划数据")}
                    </article>
                  ) : null}

                  {/* Logs tab */}
                  {resultTab === "logs" ? (
                    <div className="log-panel solo-card">
                      <div className="artifact-header">
                        <h3>运行日志</h3>
                        <span>{job.logs?.length ?? 0} 条</span>
                      </div>
                      <div className="log-list">
                        {job.logs?.slice().reverse().map((entry, idx) => (
                          <div className="log-row" key={`${entry.time}-${idx}`}>
                            <span>{new Date(entry.time).toLocaleString()}</span>
                            <p>{entry.message}</p>
                          </div>
                        ))}
                      </div>
                    </div>
                  ) : null}
                </div>
              ) : renderEmptyState("📦", "暂无产出", "选择左侧任务查看详情")}
            </div>
          ) : renderEmptyState("🎯", "选择一个任务", "从左侧提交新任务或查看历史记录")}

          {/* Error fallback */}
          {!job && jobError ? <p className="error-banner" style={{ marginTop: 12 }}>{jobError}</p> : null}
        </section>

        {/* ═══ Right: Quota panel ═══ */}
        <section className="panel quota-panel">
          <div className="panel-heading">
            <div><h2>API 额度</h2></div>
            <div className="quota-actions">
              <div className="quota-total">
                <span>剩余</span>
                <strong>{quotaSummary.window + quotaSummary.daily}</strong>
                <small>总计</small>
              </div>
              <button type="button" className="ghost-button" style={{ padding: "6px 11px", fontSize: "0.78rem" }} onClick={() => void fetchQuota()}>
                {isQuotaLoading ? "..." : "刷新"}
              </button>
            </div>
          </div>

          {quotaError ? <p className="error-banner">{quotaError}</p> : null}

          <div className="quota-list">
            {normalizeQuotaEntries(quota).length ? normalizeQuotaEntries(quota).map((entry) => {
              const intervalUsage = usagePercent(entry.current_interval_usage_count, entry.current_interval_total_count);
              const weeklyUsage = usagePercent(entry.current_weekly_usage_count, entry.current_weekly_total_count);

              return (
                <article className="quota-card" key={entry.model_name}>
                  <div className="quota-title"><strong>{entry.model_name}</strong></div>
                  <p className="quota-kind">{isTextWindowQuota(entry.model_name) ? "5 小时窗口" : "每日配额"}</p>
                  <div className="quota-remain-grid">
                    <div className="quota-remain-card">
                      <span>{isTextWindowQuota(entry.model_name) ? "窗口剩余" : "今日剩余"}</span>
                      <strong>{remainingCount(entry.current_interval_total_count, entry.current_interval_usage_count)}</strong>
                    </div>
                    {isTextWindowQuota(entry.model_name) ? (
                      <div className="quota-remain-card warm">
                        <span>周剩余</span>
                        <strong>{remainingCount(entry.current_weekly_total_count, entry.current_weekly_usage_count)}</strong>
                      </div>
                    ) : null}
                  </div>
                  <div className="quota-row">
                    <span>{isTextWindowQuota(entry.model_name) ? "窗口" : "今日"}</span>
                    <strong>{entry.current_interval_usage_count}/{entry.current_interval_total_count}</strong>
                  </div>
                  <div className="mini-bar"><div style={{ width: `${intervalUsage}%` }} /></div>
                  {isTextWindowQuota(entry.model_name) ? (
                    <>
                      <div className="quota-row"><span>本周</span><strong>{entry.current_weekly_usage_count}/{entry.current_weekly_total_count}</strong></div>
                      <div className="mini-bar warm"><div style={{ width: `${weeklyUsage}%` }} /></div>
                    </>
                  ) : null}
                </article>
              );
            }) : renderEmptyState("📊", "无数据", "点击刷新获取额度信息")}
          </div>
        </section>

      </main>
    </div>
    </ErrorBoundary>
  );
}
