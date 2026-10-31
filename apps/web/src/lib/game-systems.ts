// ゲームシステムの定義（トップページの一覧と、システムごとのフォントの差し替えに使う）。
// DB には保存しない。詳細は specs/001-ui-foundation/data-model.md。

export type GameSystemId = "yuyakekoyake" | "beginning-idol";

/** フォントの役割。globals.css の --font-<role> に対応する */
export type FontRole = "heading" | "brush";

/** システム用のフォント。src/lib/fonts.ts の --font-<id> に対応する */
export type FontId = "mochiy-pop-p-one" | "yuji-syuku";

type GameSystemBase = {
  id: GameSystemId;
  name: string;
  description: string;
  /** 役割ごとのフォント。持たない役割はアプリ共通のフォントを使う */
  fonts: Partial<Record<FontRole, FontId>>;
};

export type GameSystem = GameSystemBase &
  (
    | { status: "available"; href: string }
    | { status: "coming-soon"; href?: never }
  );

/** トップページに並べる順 */
export const gameSystems: readonly GameSystem[] = [
  {
    id: "yuyakekoyake",
    name: "ふしぎもののけRPG ゆうやけこやけ",
    description:
      "小さな町を舞台に、もののけたちが人間や仲間との交流を通じてやさしい物語を作る TRPG。",
    fonts: { heading: "mochiy-pop-p-one", brush: "yuji-syuku" },
    status: "coming-soon",
  },
  {
    id: "beginning-idol",
    name: "駆け出しアイドルRPG ビギニングアイドル",
    description:
      "新人アイドルとして仲間とユニットを組み、ライブや成長を描いていくアイドル活動の TRPG。",
    fonts: { heading: "mochiy-pop-p-one" },
    status: "coming-soon",
  },
];

export function getGameSystem(id: GameSystemId): GameSystem {
  const system = gameSystems.find((s) => s.id === id);
  if (!system) {
    throw new Error(`unknown game system: ${id}`);
  }
  return system;
}
