import type { InputErrorCode, InputField } from "@/gql/graphql";

/**
 * 入力の規則（specs/002-auth の Clarifications Q1）。api の internal/auth/validate.go と同じにする。
 * api が正であり、ここでは送信する前に項目の近くへ表示するために確かめる。
 */
export const LOGIN_ID_MIN = 3;
export const LOGIN_ID_MAX = 32;
export const PASSWORD_MIN = 8;
export const PASSWORD_MAX = 128;
export const DISPLAY_NAME_MAX = 32;

export type RuleError = { field: InputField; code: InputErrorCode };

/** 文字数はコードポイントで数える（絵文字 1 つを 1 文字とする） */
function length(s: string) {
  return [...s].length;
}

export function checkLoginId(value: string): RuleError | null {
  const v = value.trim();
  const fail = (code: InputErrorCode): RuleError => ({ field: "LOGIN_ID", code });
  if (v === "") return fail("REQUIRED");
  if (!/^[A-Za-z0-9_-]+$/.test(v)) return fail("INVALID_CHARACTERS");
  if (v.length < LOGIN_ID_MIN) return fail("TOO_SHORT");
  if (v.length > LOGIN_ID_MAX) return fail("TOO_LONG");
  return null;
}

/** パスワードの前後の空白は取り除かない */
export function checkPassword(field: InputField, value: string): RuleError | null {
  const n = length(value);
  if (value === "") return { field, code: "REQUIRED" };
  if (n < PASSWORD_MIN) return { field, code: "TOO_SHORT" };
  if (n > PASSWORD_MAX) return { field, code: "TOO_LONG" };
  return null;
}

export function checkDisplayName(value: string): RuleError | null {
  const v = value.trim();
  const fail = (code: InputErrorCode): RuleError => ({ field: "DISPLAY_NAME", code });
  if (v === "") return fail("REQUIRED");
  if (/[\p{Cc}\u2028\u2029]/u.test(v)) return fail("INVALID_CHARACTERS");
  if (length(v) > DISPLAY_NAME_MAX) return fail("TOO_LONG");
  return null;
}

/** 招待コード・再設定コードなど、空でないことだけを確かめる項目 */
export function checkRequired(field: InputField, value: string): RuleError | null {
  return value.replace(/[\s-]/g, "") === "" ? { field, code: "REQUIRED" } : null;
}
