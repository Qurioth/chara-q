import { afterEach, describe, expect, it, vi } from "vitest";

import { mockColorScheme } from "@/test/match-media";

import {
  THEME_STORAGE_KEY,
  applyTheme,
  getDisplayedTheme,
  readStoredTheme,
  resolveTheme,
  toggleTheme,
} from "./theme";

const html = document.documentElement;

afterEach(() => {
  vi.restoreAllMocks();
});

describe("resolveTheme", () => {
  it.each([
    { stored: "dark", system: "light", want: "dark" },
    { stored: "light", system: "dark", want: "light" },
    { stored: null, system: "dark", want: "dark" },
    { stored: null, system: "light", want: "light" },
    // 手で書き換えられた不正な値は、保存値がないものとして扱う
    { stored: "blue", system: "dark", want: "dark" },
    { stored: "", system: "light", want: "light" },
  ] as const)("保存値 $stored, 端末 $system → $want", ({ stored, system, want }) => {
    if (stored !== null) localStorage.setItem(THEME_STORAGE_KEY, stored);
    mockColorScheme(system);
    expect(resolveTheme()).toBe(want);
  });

  it("保存先が例外を投げるときは端末の設定に合わせる", () => {
    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("SecurityError");
    });
    mockColorScheme("dark");
    expect(readStoredTheme()).toBeNull();
    expect(resolveTheme()).toBe("dark");
  });
});

describe("applyTheme", () => {
  it("ダークは dark クラスと color-scheme: dark", () => {
    html.classList.add("light");
    applyTheme("dark");
    expect(html).toHaveClass("dark");
    expect(html).not.toHaveClass("light");
    expect(html.style.colorScheme).toBe("dark");
    expect(getDisplayedTheme()).toBe("dark");
  });

  it("ライトは light クラスと color-scheme: light", () => {
    html.classList.add("dark");
    applyTheme("light");
    expect(html).toHaveClass("light");
    expect(html).not.toHaveClass("dark");
    expect(html.style.colorScheme).toBe("light");
    expect(getDisplayedTheme()).toBe("light");
  });
});

describe("toggleTheme", () => {
  it("表示中の反対を保存して適用する", () => {
    applyTheme("light");
    expect(toggleTheme()).toBe("dark");
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe("dark");
    expect(html).toHaveClass("dark");

    expect(toggleTheme()).toBe("light");
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe("light");
    expect(html).not.toHaveClass("dark");
  });

  it("保存先が例外を投げても、画面の切り替えは働く", () => {
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("QuotaExceededError");
    });
    applyTheme("light");
    expect(() => toggleTheme()).not.toThrow();
    expect(html).toHaveClass("dark");
  });
});
