// テーマ（ライト/ダーク）の判定と切り替え。最初の描画の前に動くスクリプト（theme-script.ts）と
// 同じ決まりで動く。決まりは specs/001-ui-foundation/data-model.md の ThemePreference。

export type Theme = "light" | "dark";

export const THEME_STORAGE_KEY = "chara-q-theme";

const DARK_QUERY = "(prefers-color-scheme: dark)";

function isTheme(value: unknown): value is Theme {
  return value === "light" || value === "dark";
}

/** 保存値。不正な値と、保存先が使えないときは null（保存値がないものとして扱う） */
export function readStoredTheme(): Theme | null {
  try {
    const value = localStorage.getItem(THEME_STORAGE_KEY);
    return isTheme(value) ? value : null;
  } catch {
    return null;
  }
}

export function getSystemTheme(): Theme {
  return window.matchMedia(DARK_QUERY).matches ? "dark" : "light";
}

/** 使うべきテーマ。保存値があればそれ、なければ端末の設定 */
export function resolveTheme(): Theme {
  return readStoredTheme() ?? getSystemTheme();
}

/** 表示中のテーマ（html の dark クラス） */
export function getDisplayedTheme(): Theme {
  return document.documentElement.classList.contains("dark") ? "dark" : "light";
}

/**
 * html に dark か light のクラスと color-scheme を付ける。どちらのクラスもないときは、
 * CSS が端末の設定に従う（globals.css の dark の条件）。
 */
export function applyTheme(theme: Theme) {
  const html = document.documentElement;
  html.classList.toggle("dark", theme === "dark");
  html.classList.toggle("light", theme === "light");
  html.style.colorScheme = theme;
}

/** 表示中の反対のテーマを保存して適用し、適用したテーマを返す。保存できなくても画面は切り替える */
export function toggleTheme(): Theme {
  const next: Theme = getDisplayedTheme() === "dark" ? "light" : "dark";
  try {
    localStorage.setItem(THEME_STORAGE_KEY, next);
  } catch {
    // 保存先が使えないときは、この画面の中だけ切り替える
  }
  applyTheme(next);
  return next;
}

/** 表示中のテーマの変化（html の class の変化）を購読する */
export function subscribeTheme(callback: () => void): () => void {
  const observer = new MutationObserver(callback);
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ["class"] });
  return () => observer.disconnect();
}
