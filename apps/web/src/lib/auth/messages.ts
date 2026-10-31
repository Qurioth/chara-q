import type { InputErrorCode, InputField } from "@/gql/graphql";

import { DISPLAY_NAME_MAX, LOGIN_ID_MAX, LOGIN_ID_MIN, PASSWORD_MAX, PASSWORD_MIN } from "./rules";

/** 項目の名前 */
export const fieldLabels: Record<InputField, string> = {
  INVITATION_CODE: "招待コード",
  LOGIN_ID: "ログイン ID",
  PASSWORD: "パスワード",
  CURRENT_PASSWORD: "現在のパスワード",
  NEW_PASSWORD: "新しいパスワード",
  DISPLAY_NAME: "表示名",
  RESET_CODE: "再設定コード",
};

const minLength: Partial<Record<InputField, number>> = {
  LOGIN_ID: LOGIN_ID_MIN,
  PASSWORD: PASSWORD_MIN,
  NEW_PASSWORD: PASSWORD_MIN,
};

const maxLength: Partial<Record<InputField, number>> = {
  LOGIN_ID: LOGIN_ID_MAX,
  PASSWORD: PASSWORD_MAX,
  NEW_PASSWORD: PASSWORD_MAX,
  DISPLAY_NAME: DISPLAY_NAME_MAX,
};

const pad = (n: number) => String(n).padStart(2, "0");

/** ロックが解ける時刻の目安（HH:mm、利用者の端末の時刻） */
export function formatLockTime(iso: string) {
  const d = new Date(iso);
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** ロック中の文言。purpose はログインかパスワードの再設定か */
export function lockedMessage(purpose: "login" | "reset", lockedUntil: string | null | undefined) {
  const [action, verb] = purpose === "login" ? ["ログイン", "ログイン"] : ["パスワードの再設定", "再設定"];
  const until = lockedUntil ? `${formatLockTime(lockedUntil)} ごろまで` : "しばらく";
  return `${action}に続けて失敗したため、${until}${verb}できません`;
}

/** 入力の失敗のコードから、利用者に見せる文言を返す（エラーの文言はこのファイルだけに書く） */
export function inputErrorMessage(field: InputField | null | undefined, code: InputErrorCode): string {
  // 「ログイン ID」のように英数字で終わる名前は、助詞との間に空白を入れる
  const label = field ? fieldLabels[field] : "入力";
  const subject = /[A-Za-z0-9]$/.test(label) ? `${label} ` : label;
  const min = field ? minLength[field] : undefined;
  const max = field ? maxLength[field] : undefined;
  switch (code) {
    case "REQUIRED":
      return field ? `${subject}を入力してください` : "入力してください";
    case "TOO_SHORT":
      return min ? `${subject}は ${min} 文字以上で入力してください` : `${subject}が短すぎます`;
    case "TOO_LONG":
      return max ? `${subject}は ${max} 文字以内で入力してください` : `${subject}が長すぎます`;
    case "INVALID_CHARACTERS":
      return field === "LOGIN_ID"
        ? "ログイン ID は半角英数字・_・- だけで入力してください"
        : `${subject}に改行などの制御文字は使えません`;
    case "INVITATION_UNAVAILABLE":
      return "この招待コードは使えません。管理者に確認してください";
    case "LOGIN_ID_TAKEN":
      return "このログイン ID は使われています。別のログイン ID にしてください";
    case "INVALID_CREDENTIALS":
      return "ログイン ID またはパスワードが違います";
    case "LOCKED":
      return lockedMessage("login", null);
    case "WRONG_CURRENT_PASSWORD":
      return "現在のパスワードが違います";
    case "RESET_CODE_UNAVAILABLE":
      return "この再設定コードは使えません。管理者に確認してください";
  }
}

/** 通信の失敗など、入力では直せない失敗の文言 */
export const networkErrorMessage = "通信に失敗しました。時間をおいてもう一度お試しください";

/** 管理者だけの操作を、管理者でないユーザーが行おうとしたときの文言 */
export const forbiddenMessage = "この操作は管理者だけが行えます";

/** 日時を yyyy/MM/dd HH:mm で表す（利用者の端末の時刻） */
export function formatDateTime(iso: string) {
  const d = new Date(iso);
  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
