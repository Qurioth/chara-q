import { expect, it } from "vitest";

import { safeNext } from "./safe-next";

it.each([
  ["/account", "/account"],
  ["/admin/invitations", "/admin/invitations"],
  ["/sheets?q=a#top", "/sheets?q=a#top"],
  ["https://example.com", "/"],
  ["//example.com", "/"],
  ["/\\example.com", "/"],
  ["javascript:alert(1)", "/"],
  ["", "/"],
  [null, "/"],
  [undefined, "/"],
  ["/login?next=/x", "/"],
  ["/register", "/"],
  ["/reset-password", "/"],
  ["/loginx", "/loginx"],
])("%j → %j", (next, want) => {
  expect(safeNext(next)).toBe(want);
});
