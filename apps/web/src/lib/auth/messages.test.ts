import { describe, expect, it } from "vitest";

import type { InputErrorCode, InputField } from "@/gql/graphql";

import { formatLockTime, inputErrorMessage, lockedMessage } from "./messages";

const codes: InputErrorCode[] = [
  "REQUIRED",
  "TOO_SHORT",
  "TOO_LONG",
  "INVALID_CHARACTERS",
  "INVITATION_UNAVAILABLE",
  "LOGIN_ID_TAKEN",
  "INVALID_CREDENTIALS",
  "LOCKED",
  "WRONG_CURRENT_PASSWORD",
  "RESET_CODE_UNAVAILABLE",
];

const fields: InputField[] = [
  "INVITATION_CODE",
  "LOGIN_ID",
  "PASSWORD",
  "CURRENT_PASSWORD",
  "NEW_PASSWORD",
  "DISPLAY_NAME",
  "RESET_CODE",
];

describe("inputErrorMessage", () => {
  it("すべてのコードに文言がある", () => {
    for (const code of codes) {
      expect(inputErrorMessage(null, code)).not.toBe("");
      for (const field of fields) {
        expect(inputErrorMessage(field, code)).not.toMatch(/undefined/);
      }
    }
  });

  it.each([
    ["LOGIN_ID", "TOO_SHORT", "ログイン ID は 3 文字以上で入力してください"],
    ["LOGIN_ID", "TOO_LONG", "ログイン ID は 32 文字以内で入力してください"],
    ["LOGIN_ID", "INVALID_CHARACTERS", "ログイン ID は半角英数字・_・- だけで入力してください"],
    ["NEW_PASSWORD", "TOO_SHORT", "新しいパスワードは 8 文字以上で入力してください"],
    ["DISPLAY_NAME", "REQUIRED", "表示名を入力してください"],
    [null, "INVALID_CREDENTIALS", "ログイン ID またはパスワードが違います"],
  ] as const)("%s %s", (field, code, want) => {
    expect(inputErrorMessage(field, code)).toBe(want);
  });
});

describe("lockedMessage", () => {
  it("解ける時刻の目安を HH:mm で出す", () => {
    const until = new Date(2026, 10, 1, 9, 5).toISOString();
    expect(formatLockTime(until)).toBe("09:05");
    expect(lockedMessage("login", until)).toBe("ログインに続けて失敗したため、09:05 ごろまでログインできません");
    expect(lockedMessage("reset", until)).toBe(
      "パスワードの再設定に続けて失敗したため、09:05 ごろまで再設定できません",
    );
  });
});
