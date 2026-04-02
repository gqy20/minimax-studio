import { useMemo, useState } from "react";

import { ErrorBoundary, formatStatus, requestJSON, shortJobID, toAssetUrl } from "../app-utils";
import type { Job, PlanData, ResultTab } from "../app-types";

interface Artifacts {
  plan: string;
  narration: string;
  music: string;
  finalVideo: string;
  image: string;
}

interface JobPanelProps {
  job: Job | null;
  jobError: string;
  onRefresh: (jobID: string) => void;
}

function useArtifacts(job: Job | null): Artifacts | null {
  return useMemo(() => {
    if (!job) return null;

    const output = job.output;
    const planPath = output && "plan_path" in output ? output.plan_path : undefined;
    const narrationPath = output && "narration_path" in output ? output.narration_path : undefined;
    const musicPath = output && "music_path" in output ? output.music_path : undefined;
    const finalVideoPath = output && "final_video_path" in output ? output.final_video_path : undefined;
    const imagePath = output && "image_path" in output ? output.image_path : undefined;
    const audioPath = output && "output_path" in output ? output.output_path : undefined;

    const fromArtifacts =
      job.artifacts?.reduce<Record<string, string>>((acc, a) => {
        acc[a.label] = toAssetUrl(job.job_id, a.path);
        return acc;
      }, {}) ?? {};

    return {
      plan: fromArtifacts.plan ?? toAssetUrl(job.job_id, planPath),
      narration: fromArtifacts.narration ?? fromArtifacts.voice ?? toAssetUrl(job.job_id, narrationPath ?? audioPath),
      music: fromArtifacts.music ?? toAssetUrl(job.job_id, musicPath ?? audioPath),
      finalVideo: fromArtifacts.final_video ?? fromArtifacts.video ?? toAssetUrl(job.job_id, finalVideoPath),
      image: fromArtifacts.image ?? toAssetUrl(job.job_id, imagePath),
    };
  }, [job]);
}

function EmptyState({ icon, title, desc }: { icon: string; title: string; desc?: string }) {
  return (
    <div className="empty-state">
      <span className="empty-icon">{icon}</span>
      <p>{title}</p>
      {desc ? <small style={{ color: "var(--text-muted)", opacity: 0.6 }}>{desc}</small> : null}
    </div>
  );
}

function ResultContent({ artifacts, job, resultTab, setResultTab }: {
  artifacts: NonNullable<ReturnType<typeof useArtifacts>>;
  job: Job;
  resultTab: ResultTab;
  setResultTab: (t: ResultTab) => void;
}) {
  const [plan, setPlan] = useState<PlanData | null>(null);

  // Fetch plan data when tab switches
  if (!artifacts?.plan) setPlan(null);

  // This effect is kept simple — the full fetch logic lives in App for job polling coordination
  return (
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
            ) : <EmptyState icon="⏳" title="等待结果" desc="任务处理完成后将在此显示" />}
          </article>

          {artifacts.image && artifacts.finalVideo ? (
            <article className="artifact-card compact-card">
              <div className="artifact-header"><h3>首帧</h3><a href={artifacts.image} target="_blank" rel="noreferrer">打开</a></div>
              <img className="image-frame" src={artifacts.image} alt="Key frame" />
            </article>
          ) : null}

          <article className="artifact-card compact-card">
            <div className="artifact-header"><h3>旁白</h3>{artifacts.narration ? <a href={artifacts.narration} target="_blank" rel="noreferrer">打开</a> : null}</div>
            {artifacts.narration ? <audio controls className="audio-frame" src={artifacts.narration} /> : <p className="muted-text">无旁白</p>}
          </article>

          <article className="artifact-card compact-card">
            <div className="artifact-header"><h3>配乐</h3>{artifacts.music ? <a href={artifacts.music} target="_blank" rel="noreferrer">打开</a> : null}</div>
            {artifacts.music ? <audio controls className="audio-frame" src={artifacts.music} /> : <p className="muted-text">无配乐</p>}
          </article>
        </div>
      ) : null}

      {/* Plan tab */}
      {resultTab === "plan" ? (
        <article className="artifact-card plan-card solo-card">
          <div className="artifact-header"><h3>分镜规划</h3>{artifacts.plan ? <a href={artifacts.plan} target="_blank" rel="noreferrer">打开</a> : null}</div>
          {plan ? (
            <div className="plan-stack">
              <div className="plan-summary"><strong>{plan.title || "未命名计划"}</strong><p>{plan.visual_style}</p></div>
              <div className="scene-list">
                {plan.scenes.map((scene, idx) => (
                  <div className="scene-card" key={`${scene.name}-${idx}`}>
                    <span className="scene-index">{String(idx + 1).padStart(2, "0")}</span>
                    <div><strong>{scene.name}</strong><p>{scene.video_prompt}</p></div>
                  </div>
                ))}
              </div>
            </div>
          ) : <EmptyState icon="📋" title="暂无规划数据" />}
        </article>
      ) : null}

      {/* Logs tab */}
      {resultTab === "logs" ? (
        <div className="log-panel solo-card">
          <div className="artifact-header"><h3>运行日志</h3><span>{job.logs?.length ?? 0} 条</span></div>
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
  );
}

export default function JobPanel({ job, jobError, onRefresh }: JobPanelProps) {
  const [resultTab, setResultTab] = useState<ResultTab>("result");
  const artifacts = useArtifacts(job);

  if (!job) {
    return <EmptyState icon="🎯" title="选择一个任务" desc="从左侧提交新任务或查看历史记录" />;
  }

  return (
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
        <button type="button" className="ghost-button" onClick={() => onRefresh(job.job_id)}>刷新</button>
        {artifacts?.finalVideo ? (
          <a className="text-link" href={artifacts.finalVideo} target="_blank" rel="noreferrer">打开结果</a>
        ) : null}
      </div>

      {artifacts ? (
        <ResultContent artifacts={artifacts} job={job} resultTab={resultTab} setResultTab={setResultTab} />
      ) : <EmptyState icon="📦" title="暂无产出" desc="选择左侧任务查看详情" />}
    </div>
  );
}
