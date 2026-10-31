import "@testing-library/jest-dom/vitest";

import { cleanup } from "@testing-library/react";
import { afterEach, vi } from "vitest";

import { mockColorScheme } from "@/test/match-media";

// next/font/google はビルド時に Next.js が変換するため、テストでは名前だけを返す偽物にする
vi.mock("next/font/google", () => {
  const font = (name: string) => () => ({
    className: `font-${name}`,
    variable: `font-variable-${name}`,
    style: { fontFamily: name },
  });
  return new Proxy({}, { get: (_, name: string) => font(name) });
});

// jsdom には matchMedia がないため、既定はライトの端末として振る舞わせる
mockColorScheme("light");

afterEach(() => {
  cleanup();
  localStorage.clear();
  document.documentElement.className = "";
  document.documentElement.removeAttribute("style");
  mockColorScheme("light");
});
