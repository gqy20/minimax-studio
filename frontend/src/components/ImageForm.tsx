import type { FormEvent } from "react";

import { DEFAULT_IMAGE_FORM } from "../app-data";
import type { ImageRequest } from "../app-types";

interface ImageFormProps {
  form: ImageRequest;
  setForm: (f: ImageRequest | ((prev: ImageRequest) => ImageRequest)) => void;
  isSubmitting: boolean;
  submitError: string;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onReset: () => void;
}

export default function ImageForm({ form, setForm, isSubmitting, submitError, onSubmit, onReset }: ImageFormProps) {
  return (
    <form className="make-form compact-form" onSubmit={onSubmit}>
      <label className="field"><span>图片提示词</span>
        <textarea
          rows={4}
          value={form.prompt}
          onChange={(e) => setForm((f) => ({ ...f, prompt: e.target.value }))}
          placeholder="输入单图生成提示词"
          required
        />
      </label>

      <label className="field"><span>画幅</span>
        <select value={form.aspect_ratio} onChange={(e) => setForm((f) => ({ ...f, aspect_ratio: e.target.value }))}>
          <option value="16:9">16:9</option>
          <option value="9:16">9:16</option>
          <option value="1:1">1:1</option>
        </select>
      </label>

      {submitError ? <p className="error-banner">{submitError}</p> : null}

      <div className="form-actions">
        <button type="submit" className="primary-button" disabled={isSubmitting}>
          {isSubmitting ? "提交中..." : "生成图片"}
        </button>
        <button type="button" className="ghost-button" onClick={onReset} disabled={isSubmitting}>
          重置
        </button>
      </div>
    </form>
  );
}
