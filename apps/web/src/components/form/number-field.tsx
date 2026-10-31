"use client";

import { useState } from "react";

import { Input } from "@/components/ui/input";

import { FormField, type FieldCommonProps } from "./form-field";
import {
  isAcceptableNumberText,
  normalizeNumberText,
  parseNumberText,
  validateNumber,
  type NumberRules,
} from "./number-parse";

type NumberFieldProps = FieldCommonProps &
  NumberRules & {
    /** null は未入力（0 とは区別する） */
    value: number | null;
    /** 数値として読める入力のたびに呼ぶ。範囲外・許可外の値も渡す（部品はエラーを表示するだけ） */
    onValueChange: (value: number | null) => void;
    onBlur?: () => void;
    placeholder?: string;
  };

const format = (value: number | null) => (value === null ? "" : String(value));

export function NumberField({
  value,
  onValueChange,
  onBlur,
  min,
  max,
  allowNegative = false,
  allowDecimal = false,
  placeholder,
  error,
  inputClassName,
  ...fieldProps
}: NumberFieldProps) {
  const rules = { min, max, allowNegative, allowDecimal };
  const [text, setText] = useState(() => format(value));
  const [internalError, setInternalError] = useState<string>();

  // 外から value が変わったときは、入力欄の文字列もその値に合わせる（入力中の文字列と同じ値なら残す）
  const [prevValue, setPrevValue] = useState(value);
  if (value !== prevValue) {
    setPrevValue(value);
    if (parseNumberText(text) !== value) {
      setText(format(value));
    }
  }

  const handleChange = (raw: string) => {
    const next = normalizeNumberText(raw);
    if (!isAcceptableNumberText(next)) return;
    setText(next);

    const parsed = parseNumberText(next);
    if (parsed === undefined) return;
    // 表示中のエラーは、直した時点で消す
    if (internalError) setInternalError(validateNumber(parsed, rules));
    onValueChange(parsed);
  };

  const handleBlur = () => {
    // 入力途中の形（`-`、`1.`）のまま離れたら、値の表示に戻す
    setText(format(value));
    setInternalError(validateNumber(value, rules));
    onBlur?.();
  };

  return (
    <FormField {...fieldProps} error={error ?? internalError}>
      {(controlProps) => (
        <Input
          {...controlProps}
          type="text"
          inputMode={allowDecimal ? "decimal" : "numeric"}
          value={text}
          onChange={(e) => handleChange(e.target.value)}
          onBlur={handleBlur}
          placeholder={placeholder}
          className={inputClassName}
        />
      )}
    </FormField>
  );
}
