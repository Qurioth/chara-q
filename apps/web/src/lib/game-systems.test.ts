import { describe, expect, it } from "vitest";

import { gameSystems, getGameSystem } from "./game-systems";

describe("gameSystems", () => {
  it.each(gameSystems.map((s) => [s.id, s] as const))(
    "%s の id は小文字の英字とハイフンだけ",
    (id) => {
      expect(id).toMatch(/^[a-z]+(-[a-z]+)*$/);
    },
  );

  it.each(gameSystems.map((s) => [s.id, s] as const))(
    "%s は準備中なら入口の href を持たない",
    (_, system) => {
      if (system.status === "coming-soon") {
        expect(system.href).toBeUndefined();
      } else {
        expect(system.href).toMatch(/^\//);
      }
    },
  );

  it("id が重複しない", () => {
    const ids = gameSystems.map((s) => s.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it("getGameSystem は id から定義を引く", () => {
    expect(getGameSystem("yuyakekoyake").name).toContain("ゆうやけこやけ");
  });
});
