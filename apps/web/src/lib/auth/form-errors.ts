import type { CombinedError } from "urql";

import type { InputErrorCode, InputField } from "@/gql/graphql";

import { inputErrorMessage } from "./messages";
import type { RuleError } from "./rules";

/** フォームに表示する失敗。fields は項目ごと、form はフォーム全体の文言 */
export type FormErrors = {
  fields: Partial<Record<InputField, string>>;
  form?: string;
};

export const noFormErrors: FormErrors = { fields: {} };

/** payload の errors を、フォームに表示する形にする */
export function toFormErrors(
  errors: readonly { field?: InputField | null; code: InputErrorCode }[],
): FormErrors {
  const result: FormErrors = { fields: {} };
  for (const e of errors) {
    const message = inputErrorMessage(e.field, e.code);
    if (e.field) {
      result.fields[e.field] ??= message;
    } else {
      result.form ??= message;
    }
  }
  return result;
}

/** 送信する前の検証の結果。失敗がなければ null */
export function ruleErrors(checks: (RuleError | null)[]): FormErrors | null {
  const errors = checks.filter((e): e is RuleError => e !== null);
  return errors.length ? toFormErrors(errors) : null;
}

/** ログインが必要な操作に、ログインせずにアクセスしたエラーか */
export function isUnauthenticated(error: CombinedError | undefined) {
  return error?.graphQLErrors.some((e) => e.extensions?.code === "UNAUTHENTICATED") ?? false;
}

/** 管理者だけの操作に、管理者でないユーザーがアクセスしたエラーか */
export function isForbidden(error: CombinedError | undefined) {
  return error?.graphQLErrors.some((e) => e.extensions?.code === "FORBIDDEN") ?? false;
}
