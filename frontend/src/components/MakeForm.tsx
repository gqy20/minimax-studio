import type { FormEvent } from "react";

import { DEFAULT_FORM } from "../app-data";
import type { MakeRequest } from "../app-types";

interface MakeFormProps {
  form: MakeRequest;
  setForm: (f: MakeRequest | ((prev: MakeRequest) => MakeRequest)) => void;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onReset: () => void;
}

export default function MakeForm({ form, setForm, isSubmitting, submitError, onSubmit, onReset }: MakeFormProps) {
  return (
    <form className="make-form compact-form" onSubmit={onSubmit}>
      <label className="field"><span>主题</span>
        <textarea
          rows={4}
          value={form.theme}
          onChange={(e) => setForm((f) => ({ ...f, theme: e.target.value }))}
          placeholder="输入一个适合生成短片的故事主题"
          required
        />
      </label>

      <div className="field-grid">
        <label className="field"><span>镜头数</span>
          <input
            type="number"
            min={1}
            max={8}
            value={form.scene_count}
            onChange={(e) => setForm((f) => ({ ...f, scene_count: Number(e.target.value) }))}
          />
        </label>

        <label className="field"><span>单镜头时长</span>
          <input
            type="number"
            min={3}
            max={10}
            value={form.scene_duration}
            onChange={(e) => setForm((f) => ({ ...f, scene_duration: Number(e.target.value) }))}
          />
        </label>

        <label className="field"><span>语言</span>
          <select
            value={form.language}
            onChange={(e) => setForm((f) => ({ ...f, language: e.target.value }))}
          >
            <option value="zh">中文</option>
            <option value="en">English</option>
          </select>
        </label>
      </div>

      <label className="field"><span>复用已有视频</span>
        <input
          type="text"
          value={form.input_video}
          onChange={(e) => setForm((f) => ({ ...f, input_video: e.target.value }))}
          placeholder="可选，本地服务端可访问的路径"
        />
      </label>

      {submitError ? <p className="error-banner">{submitError}</p> : null}

      <div className="form-actions">
        <button type="submit" className="primary-button" disabled={isSubmitting}>
          {isSubmitting ? "提交中..." : "生成短片"}
        </button>
        <button type="button" className="ghost-button" onClick={onReset} disabled={isSubmitting}>
          重置
        </button>
      </div>
    </form>
  );
}
