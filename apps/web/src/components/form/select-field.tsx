"use client";

import type { ReactNode } from "react";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

import { FormField, type FieldCommonProps } from "./form-field";

export type SelectOption = {
  value: string;
  label: ReactNode;
  disabled?: boolean;
};

type SelectFieldProps = FieldCommonProps & {
  /** null は未選択 */
  value: string | null;
  onValueChange: (value: string | null) => void;
  options: readonly SelectOption[];
  placeholder?: string;
};

export function SelectField({
  value,
  onValueChange,
  options,
  placeholder = "選択してください",
  inputClassName,
  ...fieldProps
}: SelectFieldProps) {
  return (
    <FormField {...fieldProps}>
      {({ id, name, disabled, ...ariaProps }) => (
        <Select
          id={id}
          name={name}
          disabled={disabled}
          items={options}
          value={value}
          onValueChange={(next) => onValueChange(next)}
        >
          <SelectTrigger {...ariaProps} className={cn("w-full", inputClassName)}>
            <SelectValue placeholder={placeholder} className="min-w-0 truncate" />
          </SelectTrigger>
          <SelectContent className="max-h-80">
            {options.map((option) => (
              <SelectItem key={option.value} value={option.value} disabled={option.disabled}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </FormField>
  );
}
