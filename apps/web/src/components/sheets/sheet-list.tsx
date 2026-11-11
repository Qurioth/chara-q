import { ArrowDownIcon, ArrowUpDownIcon, ArrowUpIcon } from "lucide-react";
import Link from "next/link";

import { getGameSystem } from "@/lib/game-systems";
import { formatDateTime } from "@/lib/sheets/format";
import {
  nextSheetSort,
  sheetSortDirection,
  sheetSortLabel,
  type SheetSort,
  type SheetSortKey,
} from "@/lib/sheets/search-params";
import type { SheetSummary } from "@/lib/sheets/types";
import { cn } from "@/lib/utils";

type SheetListProps = {
  sheets: readonly SheetSummary[];
  totalCount: number;
  /** 絞り込みの条件があるか（0 件のときの文言を変える） */
  hasConditions: boolean;
  /** 今の並び順（表の見出しの向きの表示に使う） */
  sort: SheetSort;
  /** 並び順を変えた一覧の URL（ほかの条件は保ち、最初のページに戻す） */
  sortHref: (sort: SheetSort) => string;
};

/**
 * 一覧の行（contracts/web-routes.md の「一覧の画面」）。
 * sm 以上は表、sm 未満は縦に積んだカードの形にし、幅 360px でも横スクロールを出さない。
 * 表のキャラクター名と更新日時の見出しは、並び替えのリンクにする（FR-011、research R14）。
 */
export function SheetList({ sheets, totalCount, hasConditions, sort, sortHref }: SheetListProps) {
  if (sheets.length === 0) {
    return (
      <div className="flex flex-col items-start gap-2 py-8">
        <p>{hasConditions ? "条件に合うシートはありません。" : "まだシートはありません。"}</p>
        {hasConditions && (
          <Link href="/sheets" className="text-sm underline underline-offset-4">
            条件を外す
          </Link>
        )}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <p className="text-sm text-muted-foreground">全 {totalCount} 件</p>

      {/* 幅の狭い画面: カード */}
      <ul className="flex flex-col gap-3 sm:hidden">
        {sheets.map((s) => (
          <li key={s.id} className="flex flex-col gap-1 rounded-lg border border-border bg-card p-3 text-card-foreground">
            <SheetLink sheet={s} />
            <p className="text-xs text-muted-foreground">{getGameSystem(s.system).name}</p>
            <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5 text-sm">
              <dt className="text-muted-foreground">プレイヤー名</dt>
              <dd className="min-w-0 break-words">{s.playerName || "—"}</dd>
              <dt className="text-muted-foreground">所有者</dt>
              <dd className="min-w-0 break-words">{s.owner.displayName}</dd>
              <dt className="text-muted-foreground">更新日時</dt>
              <dd>
                <time dateTime={s.updatedAt}>{formatDateTime(s.updatedAt)}</time>
              </dd>
            </dl>
          </li>
        ))}
      </ul>

      {/* 幅の広い画面: 表 */}
      <table className="hidden w-full table-fixed border-collapse text-sm sm:table">
        <thead>
          <tr className="border-b border-border text-left text-muted-foreground">
            <SortableHeader sortKey="name" sort={sort} sortHref={sortHref} className="w-[30%] pr-3">
              キャラクター名
            </SortableHeader>
            <th scope="col" className="w-[18%] py-2 pr-3 font-bold">
              プレイヤー名
            </th>
            <th scope="col" className="w-[22%] py-2 pr-3 font-bold">
              システム
            </th>
            <th scope="col" className="w-[15%] py-2 pr-3 font-bold">
              所有者
            </th>
            <SortableHeader sortKey="updated" sort={sort} sortHref={sortHref} className="w-[15%]">
              更新日時
            </SortableHeader>
          </tr>
        </thead>
        <tbody>
          {sheets.map((s) => (
            <tr key={s.id} className="border-b border-border align-top">
              <td className="py-2 pr-3 break-words">
                <SheetLink sheet={s} />
              </td>
              <td className="py-2 pr-3 break-words">{s.playerName || "—"}</td>
              <td className="py-2 pr-3 break-words">{getGameSystem(s.system).name}</td>
              <td className="py-2 pr-3 break-words">{s.owner.displayName}</td>
              <td className="py-2">
                <time dateTime={s.updatedAt}>{formatDateTime(s.updatedAt)}</time>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

type SortableHeaderProps = Pick<SheetListProps, "sort" | "sortHref"> & {
  sortKey: SheetSortKey;
  className?: string;
  children: string;
};

/**
 * 並び替えのリンクを持つ見出し。並んでいれば `aria-sort` と向きの矢印、並んでいなければ薄い
 * 上下の矢印を表示する。リンクには、選んだときの並び順を視覚的に隠した文で添える。
 */
function SortableHeader({ sortKey, sort, sortHref, className, children }: SortableHeaderProps) {
  const direction = sheetSortDirection(sort, sortKey);
  const next = nextSheetSort(sort, sortKey);
  const Icon = direction === "ascending" ? ArrowUpIcon : direction === "descending" ? ArrowDownIcon : ArrowUpDownIcon;
  return (
    <th
      scope="col"
      aria-sort={direction ?? undefined}
      className={cn("py-2 font-bold", direction && "text-foreground", className)}
    >
      <Link
        href={sortHref(next)}
        className="inline-flex items-center gap-1 rounded-sm outline-none hover:text-foreground hover:underline underline-offset-4 focus-visible:ring-3 focus-visible:ring-ring/50"
      >
        {children}
        <Icon aria-hidden="true" className={cn("size-3.5 shrink-0", !direction && "opacity-50")} />
        <span className="sr-only">（{sheetSortLabel(next)}に並べ替える）</span>
      </Link>
    </th>
  );
}

function SheetLink({ sheet }: { sheet: SheetSummary }) {
  return (
    <Link href={`/sheets/${sheet.id}`} className="font-bold break-words underline-offset-4 hover:underline">
      {sheet.characterName}
    </Link>
  );
}
