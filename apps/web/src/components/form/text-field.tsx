"use client";

import { Input } from "@/components/ui/input";

import { FormField, type FieldCommonProps } from "./form-field";

type TextFieldProps = FieldCommonProps & {
  value: string;
  onValueChange: (value: string) => void;
  onBlur?: () => void;
  type?: "text" | "password" | "search";
  maxLength?: number;
  placeholder?: string;
  autoComplete?: string;
};

export function TextField({
  value,
  onValueChange,
  onBlur,
  type = "text",
  maxLength,
  placeholder,
  autoComplete,
  inputClassName,
  ...fieldProps
}: TextFieldProps) {
  return (
    <FormField {...fieldProps}>
      {(controlProps) => (
        <Input
          {...controlProps}
          type={type}
          value={value}
          onChange={(e) => onValueChange(e.target.value)}
          onBlur={onBlur}
          maxLength={maxLength}
          placeholder={placeholder}
          autoComplete={autoComplete}
          className={inputClassName}
        />
      )}
    </FormField>
  );
}
