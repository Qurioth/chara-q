// シートの日時の表示（specs/003-character-sheet-base の contracts/web-routes.md）。
// サーバーとブラウザで同じ表示になるよう、タイムゾーンは Asia/Tokyo に固定する。

const formatter = new Intl.DateTimeFormat("ja-JP", {
  timeZone: "Asia/Tokyo",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  hourCycle: "h23",
});

/** RFC 3339 の日時を YYYY/MM/DD HH:mm（Asia/Tokyo）にする */
export function formatDateTime(value: string): string {
  const parts = Object.fromEntries(formatter.formatToParts(new Date(value)).map((p) => [p.type, p.value]));
  return `${parts.year}/${parts.month}/${parts.day} ${parts.hour}:${parts.minute}`;
}
