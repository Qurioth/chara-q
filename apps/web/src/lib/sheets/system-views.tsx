// システムごとに、閲覧画面へ差し込む部品の表（specs/003-character-sheet-base の research R15）。
// 各システムの Spec（ゆうやけこやけ・ビギニングアイドル）は、ここに body と ownerActions を登録して、
// システム固有の部分と所有者の操作（編集・複製・削除）を閲覧画面に加える。
import type { ComponentType } from "react";

import type { GameSystemId } from "@/lib/game-systems";

import type { SheetDetail } from "./types";

export type SheetSystemView = {
  /** システム固有の部分 */
  body: ComponentType<{ sheet: SheetDetail }>;
  /** 所有者の操作。所有者が見ているときだけ描く */
  ownerActions: ComponentType<{ sheet: SheetDetail }>;
};

function ComingSoonBody() {
  return <p className="text-muted-foreground">このシステムの項目は準備中です。</p>;
}

function NoOwnerActions() {
  return null;
}

export const sheetSystemViews: Record<GameSystemId, SheetSystemView> = {
  yuyakekoyake: { body: ComingSoonBody, ownerActions: NoOwnerActions },
  "beginning-idol": { body: ComingSoonBody, ownerActions: NoOwnerActions },
};
