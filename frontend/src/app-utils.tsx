import { Component } from "react";
import type { ReactNode } from "react";

import { API_ROOT, HISTORY_STORAGE_KEY } from "./app-data";
import type {
  Job,
  QuotaEntry,
  QuotaResult,
  WorkflowMode,
} from "./app-types";

function apiUrl(path: string) {
  if (!API_ROOT) {
    return path;
  }

  if (API_ROOT.endsWith("/api/v1") && path.startsWith("/api/v1")) {
    return `${API_ROOT}${path.slice("/api/v1".length)}`;
  }

  return `${API_ROOT}${path}`;
}

export async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
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

export function loadHistory() {
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

export function saveHistory(items: string[]) {
  localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(items.slice(0, 8)));
}

export function toAssetUrl(jobID: string, filePath?: string) {
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

export function usagePercent(used: number, total: number) {
  if (!total) {
    return 0;
  }
  return Math.min(100, Math.round((used / total) * 100));
}

export function remainingCount(total: number, used: number) {
  return Math.max(0, total - used);
}

export function isTextWindowQuota(modelName: string) {
  return /m2/i.test(modelName);
}

export function normalizeQuotaEntries(input: QuotaResult | null) {
  if (!input) {
    return [] as QuotaEntry[];
  }

  if (Array.isArray(input)) {
    return input;
  }

  return Array.isArray(input.entries) ? input.entries : [];
}

export function formatStatus(status: Job["status"]) {
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

export function shortJobID(jobID: string) {
  return jobID.length > 8 ? jobID.slice(0, 8) : jobID;
}

export function modeLabel(mode: WorkflowMode) {
  switch (mode) {
    case "make":
      return "整片";
    case "plan":
      return "分镜";
    case "clip":
      return "镜头";
    case "image":
      return "图片";
    case "voice":
      return "语音";
    case "music":
      return "音乐";
    case "stitch":
      return "合成";
  }
}

export function modeHint(mode: WorkflowMode) {
  switch (mode) {
    case "make":
      return "一步出片";
    case "plan":
      return "先看结构";
    case "clip":
      return "单镜头试跑";
    case "image":
      return "单图生成";
    case "voice":
      return "旁白输出";
    case "music":
      return "配乐输出";
    case "stitch":
      return "素材合成";
  }
}

type ErrorBoundaryProps = {
  children: ReactNode;
};

type ErrorBoundaryState = {
  hasError: boolean;
};

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(error: unknown) {
    console.error("MiniMax Studio render error", error);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="app-shell">
          <div className="panel">
            <p className="error-banner">页面渲染失败，请刷新或重启前端。</p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
