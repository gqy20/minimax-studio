import { useMemo } from "react";

import { isTextWindowQuota, normalizeQuotaEntries, remainingCount, usagePercent } from "../app-utils";
import type { QuotaResult } from "../app-types";

interface QuotaSummary {
  window: number;
  daily: number;
}

interface QuotaPanelProps {
  quota: QuotaResult | null;
  isLoading: boolean;
  error: string;
  onRefresh: () => void;
}

export default function QuotaPanel({ quota, isLoading, error, onRefresh }: QuotaPanelProps) {
  const summary = useMemo<QuotaSummary>(() => {
    const entries = normalizeQuotaEntries(quota);
    return entries.reduce(
      (acc, entry) => {
        if (isTextWindowQuota(entry.model_name)) {
          acc.window += remainingCount(entry.current_interval_total_count, entry.current_interval_usage_count);
        } else {
          acc.daily += remainingCount(entry.current_interval_total_count, entry.current_interval_usage_count);
        }
        return acc;
      },
      { window: 0, daily: 0 },
    );
  }, [quota]);

  return (
    <>
      <div className="panel-heading">
        <div><h2>API 额度</h2></div>
        <div className="quota-actions">
          <div className="quota-total">
            <span>剩余</span>
            <strong>{summary.window + summary.daily}</strong>
            <small>总计</small>
          </div>
          <button type="button" className="ghost-button" style={{ padding: "6px 11px", fontSize: "0.78rem" }} onClick={onRefresh}>
            {isLoading ? "..." : "刷新"}
          </button>
        </div>
      </div>

      {error ? <p className="error-banner">{error}</p> : null}

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
        }) : (
          <div className="empty-state">
            <span className="empty-icon">📊</span>
            <p>无数据</p>
            <small style={{ color: "var(--text-muted)", opacity: 0.6 }}>点击刷新获取额度信息</small>
          </div>
        )}
      </div>
    </>
  );
}
