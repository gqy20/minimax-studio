import type { FormEvent } from "react";

import { DEFAULT_VOICE_FORM, VOICE_OPTIONS } from "../app-data";
import type { VoiceRequest } from "../app-types";

interface VoiceFormProps {
  form: VoiceRequest;
  setForm: (f: VoiceRequest | ((prev: VoiceRequest) => VoiceRequest)) => void;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onReset: () => void;
}

export default function VoiceForm({ form, setForm, isSubmitting, submitError, onSubmit, onReset }: VoiceFormProps) {
  return (
    <form className="make-form compact-form" onSubmit={onSubmit}>
      <label className="field"><span>旁白文本</span>
        <textarea
          rows={4}
          value={form.text}
          onChange={(e) => setForm((f) => ({ ...f, text: e.target.value }))}
          required
        />
      </label>

      <div className="field-grid">
        <label className="field"><span>音色</span>
          <select value={form.voice_id} onChange={(e) => setForm((f) => ({ ...f, voice_id: e.target.value }))}>
            {VOICE_OPTIONS.map((o) => (
              <option key={o.id} value={o.id}>{o.label}</option>
            ))}
          </select>
        </label>

        <label className="field"><span>格式</span>
          <select value={form.audio_format} onChange={(e) => setForm((f) => ({ ...f, audio_format: e.target.value }))}>
            <option value="mp3">mp3</option>
            <option value="wav">wav</option>
          </select>
        </label>
      </div>

      {submitError ? <p className="error-banner">{submitError}</p> : null}

      <div className="form-actions">
        <button type="submit" className="primary-button" disabled={isSubmitting}>
          {isSubmitting ? "提交中..." : "生成语音"}
        </button>
        <button type="button" className="ghost-button" onClick={onReset} disabled={isSubmitting}>
          重置
        </button>
      </div>
    </form>
  );
}
