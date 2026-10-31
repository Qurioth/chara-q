"use client";

import { CircleAlertIcon } from "lucide-react";
import { useId, type ReactNode } from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { cn } from "@/lib/utils";

/** 4 種類の入力部品が共通して受け取る props（specs/001-ui-foundation/contracts/ui-components.md） */
export type FieldCommonProps = {
  label: ReactNode;
  /** ラベルを見た目だけ隠す（シートの表のマス目など）。支援技術には読ませる */
  hideLabel?: boolean;
  description?: ReactNode;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  /** 入力欄の id。省略時は自動で採番する */
  id?: string;
  name?: string;
  /** 外枠（ラベル・入力欄・説明・エラー）の見た目 */
  className?: string;
  /** 入力欄そのものの見た目 */
  inputClassName?: string;
};

/** 入力欄に付ける属性。FormField が作り、各部品が入力欄に渡す */
export type ControlProps = {
  id: string;
  name?: string;
  disabled?: boolean;
  "aria-describedby"?: string;
  "aria-invalid"?: true;
  "aria-required"?: true;
};

type FormFieldProps = Omit<FieldCommonProps, "inputClassName"> & {
  children: (controlProps: ControlProps) => ReactNode;
};

/** ラベル・補足説明・エラーを入力欄と関連付ける共通の枠 */
export function FormField({
  label,
  hideLabel = false,
  description,
  error,
  required = false,
  disabled = false,
  id,
  name,
  className,
  children,
}: FormFieldProps) {
  const generatedId = useId();
  const controlId = id ?? generatedId;
  const descriptionId = `${controlId}-description`;
  const errorId = `${controlId}-error`;

  const describedBy = [description ? descriptionId : null, error ? errorId : null]
    .filter(Boolean)
    .join(" ");

  return (
    <Field
      data-invalid={error ? true : undefined}
      data-disabled={disabled ? true : undefined}
      className={className}
    >
      <FieldLabel htmlFor={controlId} className={cn(hideLabel && "sr-only")}>
        {label}
        {required && (
          <span className="rounded-sm bg-muted px-1 text-xs font-normal text-foreground">
            必須
          </span>
        )}
      </FieldLabel>
      {children({
        id: controlId,
        name,
        disabled: disabled || undefined,
        "aria-describedby": describedBy || undefined,
        "aria-invalid": error ? true : undefined,
        "aria-required": required ? true : undefined,
      })}
      {description && <FieldDescription id={descriptionId}>{description}</FieldDescription>}
      {error && (
        <FieldError id={errorId} className="flex items-start gap-1">
          <CircleAlertIcon aria-hidden className="mt-0.5 size-4 shrink-0" />
          <span>{error}</span>
        </FieldError>
      )}
    </Field>
  );
}
