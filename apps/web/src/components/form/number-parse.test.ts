import { describe, expect, it } from "vitest";

import {
  isAcceptableNumberText,
  normalizeNumberText,
  parseNumberText,
  validateNumber,
} from "./number-parse";

describe("normalizeNumberText", () => {
  it.each([
    ["１２３", "123"],
    ["－５", "-5"],
    ["ー５", "-5"],
    ["１．５", "1.5"],
    ["12", "12"],
  ])("%s → %s", (input, want) => {
    expect(normalizeNumberText(input)).toBe(want);
  });
});

describe("isAcceptableNumberText", () => {
  it.each([
    ["", true],
    ["12", true],
    ["-", true],
    ["-3", true],
    ["1.", true],
    ["1.5", true],
    ["abc", false],
    ["1a", false],
    ["1e3", false],
    ["1-2", false],
    ["--1", false],
    ["1.2.3", false],
    [" 1", false],
  ])("%j → %s", (input, want) => {
    expect(isAcceptableNumberText(input)).toBe(want);
  });
});

describe("parseNumberText", () => {
  it.each([
    ["", null],
    ["12", 12],
    ["-3", -3],
    ["1.5", 1.5],
    ["0", 0],
    // 入力途中の形は値を変えない
    ["-", undefined],
    ["1.", undefined],
    [".", undefined],
    ["-.", undefined],
  ])("%j → %s", (input, want) => {
    expect(parseNumberText(input)).toBe(want);
  });
});

describe("validateNumber", () => {
  it.each([
    { value: null, options: {}, want: undefined },
    { value: 3, options: {}, want: undefined },
    { value: 0, options: {}, want: undefined },
    { value: -3, options: {}, want: "0 以上の数を入力してください" },
    { value: -3, options: { allowNegative: true }, want: undefined },
    { value: 1.5, options: {}, want: "整数を入力してください" },
    { value: 1.5, options: { allowDecimal: true }, want: undefined },
    { value: 0, options: { min: 1, max: 6 }, want: "1 以上 6 以下で入力してください" },
    { value: 7, options: { min: 1, max: 6 }, want: "1 以上 6 以下で入力してください" },
    { value: 6, options: { min: 1, max: 6 }, want: undefined },
    { value: 0, options: { min: 1 }, want: "1 以上で入力してください" },
    { value: 100, options: { max: 99 }, want: "99 以下で入力してください" },
    { value: -5, options: { allowNegative: true, min: -3 }, want: "-3 以上で入力してください" },
  ])("$value $options → $want", ({ value, options, want }) => {
    expect(validateNumber(value, options)).toBe(want);
  });
});
