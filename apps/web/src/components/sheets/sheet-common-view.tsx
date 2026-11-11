import type { ReactNode } from "react";

import { GameSystemScope } from "@/components/game-system/game-system-scope";
import { getGameSystem } from "@/lib/game-systems";
import { formatDateTime } from "@/lib/sheets/format";
import { sheetSystemViews } from "@/lib/sheets/system-views";
import type { SheetDetail } from "@/lib/sheets/types";

type SheetCommonViewProps = {
  sheet: SheetDetail;
  /** 見ている人がシートの所有者か。所有者の操作を描くかどうかに使う */
  isOwner: boolean;
};

/**
 * 閲覧画面の共通部分（contracts/web-routes.md の「閲覧の画面」）。
 * システム固有の部分と所有者の操作は、lib/sheets/system-views.tsx の部品を差し込む。
 */
export function SheetCommonView({ sheet, isOwner }: SheetCommonViewProps) {
  const system = getGameSystem(sheet.system);
  const view = sheetSystemViews[sheet.system];
  const Body = view.body;
  const OwnerActions = view.ownerActions;

  return (
    <GameSystemScope system={sheet.system} as="article" className="flex flex-col gap-8">
      <header className="flex flex-col gap-3">
        <p className="text-sm text-muted-foreground">{system.name}</p>
        <h1 className="font-heading text-2xl font-bold break-words">{sheet.characterName}</h1>
        <dl className="flex flex-wrap gap-x-6 gap-y-1 text-sm text-muted-foreground">
          <div className="flex gap-2">
            <dt>所有者</dt>
            <dd className="min-w-0 break-words text-foreground">{sheet.owner.displayName}</dd>
          </div>
          <div className="flex gap-2">
            <dt>作成日時</dt>
            <dd>
              <time dateTime={sheet.createdAt}>{formatDateTime(sheet.createdAt)}</time>
            </dd>
          </div>
          <div className="flex gap-2">
            <dt>更新日時</dt>
            <dd>
              <time dateTime={sheet.updatedAt}>{formatDateTime(sheet.updatedAt)}</time>
            </dd>
          </div>
        </dl>
        {isOwner && <OwnerActions sheet={sheet} />}
      </header>

      <section aria-labelledby="sheet-common-heading" className="flex flex-col gap-4">
        <h2 id="sheet-common-heading" className="font-heading text-xl font-bold">
          基本情報
        </h2>
        <dl className="grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-[8rem_1fr]">
          <Item label="プレイヤー名">{sheet.playerName}</Item>
          <Item label="年齢">{sheet.age}</Item>
          <Item label="性別">{sheet.sex}</Item>
          <Item label="メモ" multiline>
            {sheet.memo}
          </Item>
        </dl>
      </section>

      <section aria-labelledby="sheet-system-heading" className="flex flex-col gap-4">
        <h2 id="sheet-system-heading" className="font-heading text-xl font-bold">
          {system.name}の項目
        </h2>
        <Body sheet={sheet} />
      </section>
    </GameSystemScope>
  );
}

function Item({ label, multiline = false, children }: { label: string; multiline?: boolean; children: ReactNode }) {
  const empty = children === "";
  return (
    <div className="contents">
      <dt className="text-sm font-bold text-muted-foreground sm:pt-0.5">{label}</dt>
      <dd
        className={
          empty
            ? "text-muted-foreground"
            : `min-w-0 break-words ${multiline ? "whitespace-pre-wrap" : ""}`
        }
      >
        {empty ? "未入力" : children}
      </dd>
    </div>
  );
}
