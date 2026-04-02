import type { FormEvent } from "react";

import { DEFAULT_STITCH_FORM } from "../app-data";
import type { StitchRequest } from "../app-types";

interface StitchFormProps {
  form: StitchRequest;
  setForm: (f: StitchRequest | ((prev: StitchRequest) => StitchRequest)) => void;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onReset: () => void;
}

export default function StitchForm({ form, setForm, isSubmitting, submitError, onSubmit, onReset }: StitchFormProps) {
  return (
    <form className="make-form compact-form" onSubmit={onSubmit}>
      <label className="field"><span>视频路径列表</span>
        <textarea
          rows={4}
          value={form.videos}
          onChange={(e) => setForm((f) => ({ ...f, videos: e.target.value }))}
          placeholder={"每行一个服务端可访问的视频路径"}
          required
        />
      </label>

      <label className="field"><span>旁白路径</span>
        <input
          type="text"
          value={form.narration}
          onChange={(e) => setForm((f) => ({ ...f, narration: e.target.value }))}
          required
        />
      </label>

      <label className="field"><span>背景音乐路径</span>
        <input
          type="text"
          value={form.music}
          onChange={(e) => setForm((f) => ({ ...f, music: e.target.value }))}
          placeholder="可选"
        />
      </label>

      {submitError ? <p className="error-banner">{submitError}</p> : null}

      <div className="form-actions">
        <button type="submit" className="primary-button" disabled={isSubmitting}>
          {isSubmitting ? "提交中..." : "合成成片"}
        </button>
        <button type="button" className="ghost-button" onClick={onReset} disabled={isSubmitting}>
          重置
        </button>
      </div>
    </form>
  );
}
