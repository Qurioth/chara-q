// シートのシステムの値の変換（specs/003-character-sheet-base の data-model）。
// URL と web の内部は GameSystemId（kebab-case）、GraphQL は GameSystem（大文字の enum）。

import type { GameSystem } from "@/gql/graphql";
import { gameSystems, type GameSystemId } from "@/lib/game-systems";

const toGraphQL = {
  yuyakekoyake: "YUYAKEKOYAKE",
  "beginning-idol": "BEGINNING_IDOL",
} as const satisfies Record<GameSystemId, GameSystem>;

const fromGraphQL = {
  YUYAKEKOYAKE: "yuyakekoyake",
  BEGINNING_IDOL: "beginning-idol",
} as const satisfies Record<GameSystem, GameSystemId>;

export function toGameSystem(id: GameSystemId): GameSystem {
  return toGraphQL[id];
}

export function toGameSystemId(system: GameSystem): GameSystemId {
  return fromGraphQL[system];
}

/** URL の値などが、知っているシステムの id かを確かめる */
export function isGameSystemId(value: unknown): value is GameSystemId {
  return typeof value === "string" && gameSystems.some((s) => s.id === value);
}
