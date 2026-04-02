import { FormEvent, useEffect, useState } from "react";

import {
  DEFAULT_CLIP_FORM,
  DEFAULT_FORM,
  DEFAULT_IMAGE_FORM,
  DEFAULT_MUSIC_FORM,
  DEFAULT_PLAN_FORM,
  DEFAULT_STITCH_FORM,
  DEFAULT_VOICE_FORM,
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
  loadHistory,
  modeLabel,
  modeHint,
  normalizeQuotaEntries,
  requestJSON,
  saveHistory,
  shortJobID,
} from "./app-utils";
import { MODE_ICONS } from "./icons";

import MakeForm from "./components/MakeForm";
import PlanForm from "./components/PlanForm";
import ClipForm from "./components/ClipForm";
import ImageForm from "./components/ImageForm";
import VoiceForm from "./components/VoiceForm";
import MusicForm from "./components/MusicForm";
import StitchForm from "./components/StitchForm";
import JobPanel from "./components/JobPanel";
import QuotaPanel from "./components/QuotaPanel";

/* ══════════════════════════════════════════════
   App — thin orchestrator
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

  /* ── Derived artifacts (must be before effects that reference it) ── */

  const artifacts = (() => {
    if (!job) return null;
    const o = job.output;
    const pp = o && "plan_path" in o ? o.plan_path : undefined;
    const np = o && "narration_path" in o ? o.narration_path : undefined;
    const mp = o && "music_path" in o ? o.music_path : undefined;
    const fvp = o && "final_video_path" in o ? o.final_video_path : undefined;
    const ip = o && "image_path" in o ? o.image_path : undefined;
    const ap = o && "output_path" in o ? o.output_path : undefined;
    const fa = job.artifacts?.reduce<Record<string, string>>((a, x) => { a[x.label] = shortUrl(job.job_id, x.path); return a; }, {}) ?? {};
    return {
      plan: fa.plan ?? shortUrl(job.job_id, pp),
      narration: fa.narration ?? fa.voice ?? shortUrl(job.job_id, np ?? ap),
      music: fa.music ?? shortUrl(job.job_id, mp ?? ap),
      finalVideo: fa.final_video ?? fa.video ?? shortUrl(job.job_id, fvp),
      image: fa.image ?? shortUrl(job.job_id, ip),
    };
  })();

  /* ── Effects ── */

  useEffect(() => { void fetchJobs(); }, []);
  useEffect(() => { void fetchQuota(); }, []);

  useEffect(() => {
    if (!job || job.status !== "processing") return;
    const t = window.setInterval(() => void fetchJob(job.job_id, false), 3000);
    return () => window.clearInterval(t);
  }, [job]);

  useEffect(() => {
    if (!job || job.status !== "completed" || !artifacts?.plan) { setPlan(null); return; }
    void fetch(artifacts.plan)
      .then((r) => { if (!r.ok) throw new Error("fail"); return r.json() as Promise<PlanData>; })
      .then(setPlan).catch(() => setPlan(null));
  }, [artifacts?.plan, job]);

  useEffect(() => {
    if (!job) { setResultTab("result"); return; }
    if (job.logs?.length) setResultTab("result");
  }, [job?.job_id]);

  function shortUrl(jid: string, p?: string) {
    if (!p) return "";
    const fn = p.replaceAll("\\", "/").split("/").filter(Boolean).pop();
    return fn ? `/api/v1/output/${jid}/${encodeURIComponent(fn)}` : "";
  }

  /* ── API calls ── */

  async function fetchQuota() {
    setIsQuotaLoading(true); setQuotaError("");
    try { setQuota(await requestJSON<QuotaResult>("/api/v1/quota")); }
    catch (e) { setQuotaError(e instanceof Error ? e.message : "加载额度失败"); }
    finally { setIsQuotaLoading(false); }
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
    } catch (e) { setJobError(e instanceof Error ? e.message : "读取任务失败"); }
  }

  /* ── Submit handlers (thin wrappers) ── */

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string }>("/api/v1/make", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: form.theme.trim(), scene_count: form.scene_count, scene_duration: form.scene_duration, language: form.language, input_video: form.input_video.trim() }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handlePlanSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string }>("/api/v1/plan", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: planForm.theme.trim(), scene_count: planForm.scene_count, scene_duration: planForm.scene_duration, language: planForm.language }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "分镜任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleClipSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string }>("/api/v1/clip", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: clipForm.prompt.trim(), subject: clipForm.subject.trim(), aspect_ratio: clipForm.aspect_ratio, duration: clipForm.duration, resolution: clipForm.resolution }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "图生视频任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleVoiceSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string }>("/api/v1/voice", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: voiceForm.text.trim(), voice_id: voiceForm.voice_id, audio_format: voiceForm.audio_format }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "语音任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleImageSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string }>("/api/v1/image", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: imageForm.prompt.trim(), aspect_ratio: imageForm.aspect_ratio }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "图片任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleMusicSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    try {
      const r = await requestJSON<{ job_id: string }>("/api/v1/music", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: musicForm.prompt.trim(), audio_format: musicForm.audio_format }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "音乐任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  async function handleStitchSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); setIsSubmitting(true); setSubmitError(""); setJobError("");
    const videos = stitchForm.videos.split("\n").map((s) => s.trim()).filter(Boolean);
    try {
      const r = await requestJSON<{ job_id: string }>("/api/v1/stitch", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ videos, narration: stitchForm.narration.trim(), music: stitchForm.music.trim() }),
      });
      await fetchJob(r.job_id); await fetchJobs();
    } catch (e) { setSubmitError(e instanceof Error ? e.message : "素材合成任务提交失败"); }
    finally { setIsSubmitting(false); }
  }

  /* ── Mode switcher ── */

  const modes: WorkflowMode[] = ["make", "plan", "clip", "image", "voice", "music", "stitch"];

  /* ══════════════════════════════════════════════
     Render
     ══════════════════════════════════════════════ */

  return (
    <ErrorBoundary>
    <div className="app-shell">
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

      <main className="workspace">
        {/* Left: Form panel */}
        <section className="panel form-panel">
          <div className="panel-heading">
            <div><h2>{modeLabel(mode)}</h2></div>
            <span className="panel-note">{modeHint(mode)}</span>
          </div>

          <div className="mode-switcher">
            {modes.map((m) => (
              <button key={m} type="button" className={`mode-chip ${mode === m ? "active" : ""}`} onClick={() => setMode(m)}>
                <span className="mode-icon">{MODE_ICONS[m]}</span>
                {m.charAt(0).toUpperCase() + m.slice(1)}
              </button>
            ))}
          </div>

          {mode === "make" && <MakeForm form={form} setForm={setForm} isSubmitting={isSubmitting} submitError={submitError} onSubmit={handleSubmit} onReset={() => setForm(DEFAULT_FORM)} />}
          {mode === "plan" && <PlanForm form={planForm} setForm={setPlanForm} isSubmitting={isSubmitting} submitError={submitError} onSubmit={handlePlanSubmit} onReset={() => setPlanForm(DEFAULT_PLAN_FORM)} />}
          {mode === "clip" && <ClipForm form={clipForm} setForm={setClipForm} isSubmitting={isSubmitting} submitError={submitError} onSubmit={handleClipSubmit} onReset={() => setClipForm(DEFAULT_CLIP_FORM)} />}
          {mode === "image" && <ImageForm form={imageForm} setForm={setImageForm} isSubmitting={isSubmitting} submitError={submitError} onSubmit={handleImageSubmit} onReset={() => setImageForm(DEFAULT_IMAGE_FORM)} />}
          {mode === "voice" && <VoiceForm form={voiceForm} setForm={setVoiceForm} isSubmitting={isSubmitting} submitError={submitError} onSubmit={handleVoiceSubmit} onReset={() => setVoiceForm(DEFAULT_VOICE_FORM)} />}
          {mode === "music" && <MusicForm form={musicForm} setForm={setMusicForm} isSubmitting={isSubmitting} submitError={submitError} onSubmit={handleMusicSubmit} onReset={() => setMusicForm(DEFAULT_MUSIC_FORM)} />}
          {mode === "stitch" && <StitchForm form={stitchForm} setForm={setStitchForm} isSubmitting={isSubmitting} submitError={submitError} onSubmit={handleStitchSubmit} onReset={() => setStitchForm(DEFAULT_STITCH_FORM)} />}

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
                  <small>{item.status === "processing" ? "处理中" : item.status === "completed" ? "已完成" : item.status === "failed" ? "失败" : "等待中"}</small>
                </button>
              ))}
            </div>
          </div>
        </section>

        {/* Middle: Job status */}
        <section className="panel status-panel">
          <div className="panel-heading">
            <div><h2>任务状态</h2></div>
            {job ? <span className={`status-badge ${job.status}`}>{job.status === "processing" ? "处理中" : job.status === "completed" ? "已完成" : "失败"}</span> : null}
          </div>
          <JobPanel job={job} jobError={jobError} onRefresh={(id) => void fetchJob(id)} />
        </section>

        {/* Right: Quota */}
        <section className="panel quota-panel">
          <QuotaPanel quota={quota} isLoading={isQuotaLoading} error={quotaError} onRefresh={() => void fetchQuota()} />
        </section>
      </main>
    </div>
    </ErrorBoundary>
  );
}
