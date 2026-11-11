import { describe, expect, it } from "vitest";

import { checkDisplayName, checkLoginId, checkPassword, checkRequired } from "./rules";

// api の internal/auth/validate_test.go と同じ境界値で確かめる
describe("checkLoginId", () => {
  it.each([
    ["abc", null],
    ["a".repeat(32), null],
    ["Ab_9-z", null],
    ["  admin ", null],
    ["", "REQUIRED"],
    ["   ", "REQUIRED"],
    ["ab", "TOO_SHORT"],
    ["a".repeat(33), "TOO_LONG"],
    ["ab!c", "INVALID_CHARACTERS"],
    ["ab c", "INVALID_CHARACTERS"],
    ["あいう", "INVALID_CHARACTERS"],
    ["ａｂｃ", "INVALID_CHARACTERS"],
  ])("%j → %s", (value, code) => {
    expect(checkLoginId(value)?.code ?? null).toBe(code);
  });
});

describe("checkPassword", () => {
  it.each([
    ["a".repeat(7), "TOO_SHORT"],
    ["a".repeat(8), null],
    ["a".repeat(128), null],
    ["a".repeat(129), "TOO_LONG"],
    ["", "REQUIRED"],
    ["あいうえおかきく", null],
    [" abcdef ", null],
    ["🐱".repeat(8), null],
  ])("%j → %s", (value, code) => {
    const result = checkPassword("NEW_PASSWORD", value);
    expect(result?.code ?? null).toBe(code);
    if (result) expect(result.field).toBe("NEW_PASSWORD");
  });
});

describe("checkDisplayName", () => {
  it.each([
    ["a", null],
    ["あ".repeat(32), null],
    ["あ".repeat(33), "TOO_LONG"],
    ["🐱たま", null],
    ["　たま ", null],
    ["たま ねこ", null],
    ["", "REQUIRED"],
    [" 　 ", "REQUIRED"],
    ["たま\nねこ", "INVALID_CHARACTERS"],
    ["たま\tねこ", "INVALID_CHARACTERS"],
    ["たま\u2028ねこ", "INVALID_CHARACTERS"],
  ])("%j → %s", (value, code) => {
    expect(checkDisplayName(value)?.code ?? null).toBe(code);
  });
});

describe("checkRequired", () => {
  it("空白とハイフンだけなら未入力", () => {
    expect(checkRequired("INVITATION_CODE", " - ")).toEqual({ field: "INVITATION_CODE", code: "REQUIRED" });
    expect(checkRequired("INVITATION_CODE", "ABCD-EFGH")).toBeNull();
  });
});
