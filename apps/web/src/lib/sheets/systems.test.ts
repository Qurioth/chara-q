import { describe, expect, it } from "vitest";

import { gameSystems } from "@/lib/game-systems";

import { isGameSystemId, toGameSystem, toGameSystemId } from "./systems";

describe("システムの値の変換", () => {
  it.each(gameSystems.map((s) => [s.id] as const))("%s は GraphQL の値と往復できる", (id) => {
    expect(toGameSystemId(toGameSystem(id))).toBe(id);
  });

  it("GraphQL の enum に変える", () => {
    expect(toGameSystem("yuyakekoyake")).toBe("YUYAKEKOYAKE");
    expect(toGameSystem("beginning-idol")).toBe("BEGINNING_IDOL");
  });

  it.each([
    ["yuyakekoyake", true],
    ["beginning-idol", true],
    ["YUYAKEKOYAKE", false],
    ["beginning_idol", false],
    ["", false],
    [undefined, false],
  ])("isGameSystemId(%s) は %s", (value, want) => {
    expect(isGameSystemId(value)).toBe(want);
  });
});
