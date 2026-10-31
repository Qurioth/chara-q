// 数値入力の文字列の扱い。決まりは specs/001-ui-foundation/data-model.md の「NumberField の値」。

const FULL_WIDTH_DIGITS = /[０-９]/g;
const FULL_WIDTH_MINUS = /[－ー−]/g;

/** 全角の数字・マイナス・ピリオドを半角に置き換える */
export function normalizeNumberText(text: string): string {
  return text
    .replace(FULL_WIDTH_DIGITS, (c) => String.fromCharCode(c.charCodeAt(0) - 0xfee0))
    .replace(FULL_WIDTH_MINUS, "-")
    .replace(/．/g, ".");
}

/** 入力欄に置いてよい文字列か（数字・先頭の - 1 つ・. 1 つだけ。入力途中の形を含む） */
export function isAcceptableNumberText(text: string): boolean {
  return /^-?\d*\.?\d*$/.test(text);
}

/**
 * 入力欄の文字列を値に変える。空欄は null（未入力）、入力途中の形（`-`、`1.` など）は
 * undefined（値を変えない）、それ以外は数値。
 */
export function parseNumberText(text: string): number | null | undefined {
  if (text === "") return null;
  if (!/^-?(\d+|\d*\.\d+)$/.test(text)) return undefined;
  return Number(text);
}

export type NumberRules = {
  min?: number;
  max?: number;
  allowNegative?: boolean;
  allowDecimal?: boolean;
};

/** 部品単位の検証。問題があればエラー文言を返す */
export function validateNumber(
  value: number | null,
  { min, max, allowNegative = false, allowDecimal = false }: NumberRules,
): string | undefined {
  if (value === null) return undefined;
  if (!allowNegative && value < 0) return "0 以上の数を入力してください";
  if (!allowDecimal && !Number.isInteger(value)) return "整数を入力してください";

  const belowMin = min !== undefined && value < min;
  const aboveMax = max !== undefined && value > max;
  if (!belowMin && !aboveMax) return undefined;
  if (min !== undefined && max !== undefined) return `${min} 以上 ${max} 以下で入力してください`;
  if (min !== undefined) return `${min} 以上で入力してください`;
  return `${max} 以下で入力してください`;
}
