"use client";

import { Textarea } from "@/components/ui/textarea";

import { FormField, type FieldCommonProps } from "./form-field";

type TextareaFieldProps = FieldCommonProps & {
  /** 改行（\n）を含む */
  value: string;
  onValueChange: (value: string) => void;
  onBlur?: () => void;
  rows?: number;
  maxLength?: number;
  placeholder?: string;
};

export function TextareaField({
  value,
  onValueChange,
  onBlur,
  rows = 3,
  maxLength,
  placeholder,
  inputClassName,
  ...fieldProps
}: TextareaFieldProps) {
  return (
    <FormField {...fieldProps}>
      {(controlProps) => (
        <Textarea
          {...controlProps}
          value={value}
          onChange={(e) => onValueChange(e.target.value)}
          onBlur={onBlur}
          rows={rows}
          maxLength={maxLength}
          placeholder={placeholder}
          className={inputClassName}
        />
      )}
    </FormField>
  );
}
