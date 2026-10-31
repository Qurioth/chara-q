import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import { gameSystems, type FontRole } from "./game-systems";

// game-systems.ts の fonts と、globals.css のフォントの差し替えが食い違わないことを確かめる
// テストは apps/web をルートに実行する（vitest.config.mts）
const css = readFileSync(resolve(process.cwd(), "src/app/globals.css"), "utf8");

const roles: FontRole[] = ["heading", "brush"];

function blockOf(id: string): string {
  const match = css.match(new RegExp(`\\[data-game-system="${id}"\\]\\s*\\{([^}]*)\\}`));
  return match?.[1] ?? "";
}

describe.each(gameSystems.map((s) => [s.id, s] as const))("%s", (id, system) => {
  it.each(roles)("役割 %s の差し替えが定義と一致する", (role) => {
    const declaration = new RegExp(`--font-${role}\\s*:\\s*([^;]+);`).exec(blockOf(id));
    const fontId = system.fonts[role];

    if (fontId) {
      expect(declaration?.[1]).toBe(`var(--font-${fontId}), var(--font-sans)`);
    } else {
      expect(declaration).toBeNull();
    }
  });
});
