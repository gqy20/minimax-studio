import type { FormEvent } from "react";

import { DEFAULT_MUSIC_FORM } from "../app-data";
import type { MusicRequest } from "../app-types";

interface MusicFormProps {
  form: MusicRequest;
  setForm: (f: MusicRequest | ((prev: MusicRequest) => MusicRequest)) => void;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onReset: () => void;
}

export default function MusicForm({ form, setForm, isSubmitting, submitError, onSubmit, onReset }: MusicFormProps) {
  return (
    <form className="make-form compact-form" onSubmit={onSubmit}>
      <label className="field"><span>音乐提示词</span>
        <textarea
          rows={4}
          value={form.prompt}
          onChange={(e) => setForm((f) => ({ ...f, prompt: e.target.value }))}
          required
        />
      </label>

      <label className="field"><span>格式</span>
        <select value={form.audio_format} onChange={(e) => setForm((f) => ({ ...f, audio_format: e.target.value }))}>
          <option value="mp3">mp3</option>
          <option value="wav">wav</option>
        </select>
      </label>

      {submitError ? <p className="error-banner">{submitError}</p> : null}

      <div className="form-actions">
        <button type="submit" className="primary-button" disabled={isSubmitting}>
          {isSubmitting ? "提交中..." : "生成音乐"}
        </button>
        <button type="button" className="ghost-button" onClick={onReset} disabled={isSubmitting}>
          重置
        </button>
      </div>
    </form>
  );
}
