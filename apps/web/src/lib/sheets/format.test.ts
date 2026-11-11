import { expect, it } from "vitest";

import { formatDateTime } from "./format";

it.each([
  ["2026-11-09T15:30:00Z", "2026/11/10 00:30"],
  ["2026-01-01T00:05:00+09:00", "2026/01/01 00:05"],
  ["2026-12-31T14:59:59.999Z", "2026/12/31 23:59"],
])("%s は %s と表示する（Asia/Tokyo）", (value, want) => {
  expect(formatDateTime(value)).toBe(want);
});
