import { afterEach, expect, it, vi } from "vitest";

import { mockColorScheme } from "@/test/match-media";

import { THEME_STORAGE_KEY } from "./theme";
import { themeScript } from "./theme-script";

const html = document.documentElement;
const runScript = () => new Function(themeScript)();

afterEach(() => {
  vi.restoreAllMocks();
});

it.each([
  { stored: "dark", system: "light", dark: true },
  { stored: "light", system: "dark", dark: false },
  { stored: null, system: "dark", dark: true },
  { stored: null, system: "light", dark: false },
  { stored: "blue", system: "dark", dark: true },
] as const)("保存値 $stored, 端末 $system → dark: $dark", ({ stored, system, dark }) => {
  if (stored !== null) localStorage.setItem(THEME_STORAGE_KEY, stored);
  mockColorScheme(system);

  runScript();

  expect(html.classList.contains("dark")).toBe(dark);
  expect(html.classList.contains("light")).toBe(!dark);
  expect(html.style.colorScheme).toBe(dark ? "dark" : "light");
});

it("保存値がない間は端末の設定に追従し、保存値ができたら追従しない", () => {
  const media = mockColorScheme("light");
  runScript();
  expect(html).not.toHaveClass("dark");

  media.setScheme("dark");
  expect(html).toHaveClass("dark");

  localStorage.setItem(THEME_STORAGE_KEY, "dark");
  media.setScheme("light");
  expect(html).toHaveClass("dark");
});

it("保存値があるときは端末の設定の変化を反映しない", () => {
  localStorage.setItem(THEME_STORAGE_KEY, "light");
  const media = mockColorScheme("light");
  runScript();

  media.setScheme("dark");
  expect(html).not.toHaveClass("dark");
});

it("保存先が例外を投げても、例外を外に出さず端末の設定に合わせる", () => {
  vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
    throw new Error("SecurityError");
  });
  mockColorScheme("dark");

  expect(runScript).not.toThrow();
  expect(html).toHaveClass("dark");
});
