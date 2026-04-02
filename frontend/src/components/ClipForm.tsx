import type { FormEvent } from "react";

import { DEFAULT_CLIP_FORM } from "../app-data";
import type { ClipRequest } from "../app-types";

interface ClipFormProps {
  form: ClipRequest;
  setForm: (f: ClipRequest | ((prev: ClipRequest) => ClipRequest)) => void;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onReset: () => void;
}

export default function ClipForm({ form, setForm, isSubmitting, submitError, onSubmit, onReset }: ClipFormProps) {
  return (
    <form className="make-form compact-form" onSubmit={onSubmit}>
      <label className="field"><span>首帧提示词</span>
        <textarea
          rows={3}
          value={form.prompt}
          onChange={(e) => setForm((f) => ({ ...f, prompt: e.target.value }))}
          placeholder="用于生成首帧图像"
          required
        />
      </label>

      <label className="field"><span>视频运动提示词</span>
        <textarea
          rows={3}
          value={form.subject}
          onChange={(e) => setForm((f) => ({ ...f, subject: e.target.value }))}
          placeholder="描述镜头运动和主体行为"
          required
        />
      </label>

      <div className="field-grid">
        <label className="field"><span>画幅</span>
          <select value={form.aspect_ratio} onChange={(e) => setForm((f) => ({ ...f, aspect_ratio: e.target.value }))}>
            <option value="16:9">16:9</option>
            <option value="9:16">9:16</option>
            <option value="1:1">1:1</option>
          </select>
        </label>

        <label className="field"><span>时长</span>
          <input
            type="number"
            min={5}
            max={10}
            value={form.duration}
            onChange={(e) => setForm((f) => ({ ...f, duration: Number(e.target.value) }))}
          />
        </label>

        <label className="field"><span>分辨率</span>
          <select value={form.resolution} onChange={(e) => setForm((f) => ({ ...f, resolution: e.target.value }))}>
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
        <button type="button" className="ghost-button" onClick={onReset} disabled={isSubmitting}>
          重置
        </button>
      </div>
    </form>
  );
}
