// シートの一覧の URL のクエリ（specs/003-character-sheet-base の contracts/web-routes.md）。
// 知らない値・形の正しくない値は、エラーにせず既定の値にする（spec の Edge Cases）。

import type { CharacterSheetOrder } from "@/gql/graphql";
import type { GameSystemId } from "@/lib/game-systems";

import { isGameSystemId } from "./systems";

/** 並び替えの項目。値は、その項目の最初の向きの並び順と同じ */
export type SheetSortKey = "updated" | "created" | "name";

/**
 * 一覧の並び順（URL の `sort`）。項目の最初の向き（日時は新しい順、名前は名前の順）は接尾辞なし、
 * 逆の向きは `-asc`・`-desc` を付ける（research R14）。
 */
export type SheetSort = SheetSortKey | "updated-asc" | "created-asc" | "name-desc";

/** 表の見出しの `aria-sort` の値 */
export type SheetSortDirection = "ascending" | "descending";

export type SheetSearch = {
  system: GameSystemId | null;
  q: string;
  mine: boolean;
  sort: SheetSort;
  page: number;
};

export const defaultSheetSearch: SheetSearch = {
  system: null,
  q: "",
  mine: false,
  sort: "updated",
  page: 1,
};

type SortSpec = {
  label: string;
  key: SheetSortKey;
  direction: SheetSortDirection;
  reverse: SheetSort;
  order: CharacterSheetOrder;
};

// 並べた順が、検索のフォームのプルダウンの順になる
const sorts = {
  updated: {
    label: "更新日時の新しい順",
    key: "updated",
    direction: "descending",
    reverse: "updated-asc",
    order: "UPDATED_AT_DESC",
  },
  "updated-asc": {
    label: "更新日時の古い順",
    key: "updated",
    direction: "ascending",
    reverse: "updated",
    order: "UPDATED_AT_ASC",
  },
  created: {
    label: "作成日時の新しい順",
    key: "created",
    direction: "descending",
    reverse: "created-asc",
    order: "CREATED_AT_DESC",
  },
  "created-asc": {
    label: "作成日時の古い順",
    key: "created",
    direction: "ascending",
    reverse: "created",
    order: "CREATED_AT_ASC",
  },
  name: {
    label: "キャラクター名の順",
    key: "name",
    direction: "ascending",
    reverse: "name-desc",
    order: "CHARACTER_NAME_ASC",
  },
  "name-desc": {
    label: "キャラクター名の逆の順",
    key: "name",
    direction: "descending",
    reverse: "name",
    order: "CHARACTER_NAME_DESC",
  },
} as const satisfies Record<SheetSort, SortSpec>;

export const sheetSortOptions: readonly { value: SheetSort; label: string }[] = (
  Object.keys(sorts) as SheetSort[]
).map((value) => ({ value, label: sorts[value].label }));

export function toSheetOrder(sort: SheetSort): CharacterSheetOrder {
  return sorts[sort].order;
}

export function sheetSortLabel(sort: SheetSort): string {
  return sorts[sort].label;
}

/** その項目で並んでいれば向き、並んでいなければ null（表の見出しの `aria-sort`） */
export function sheetSortDirection(sort: SheetSort, key: SheetSortKey): SheetSortDirection | null {
  return sorts[sort].key === key ? sorts[sort].direction : null;
}

/** 表の見出しを選んだときの並び順。並んでいない項目は最初の向き、並んでいる項目は逆の向き */
export function nextSheetSort(current: SheetSort, key: SheetSortKey): SheetSort {
  return sorts[current].key === key ? sorts[current].reverse : key;
}

type RawSearchParams = Record<string, string | string[] | undefined>;

function first(value: string | string[] | undefined): string | undefined {
  return Array.isArray(value) ? value[0] : value;
}

function isSort(value: unknown): value is SheetSort {
  return typeof value === "string" && Object.hasOwn(sorts, value);
}

/** Next.js の searchParams から一覧の条件を読む。同じキーが複数あるときは最初の値を使う */
export function parseSheetSearch(params: RawSearchParams): SheetSearch {
  const system = first(params.system);
  const sort = first(params.sort);
  const page = first(params.page);
  const pageNumber = page !== undefined && /^\d+$/.test(page) ? Number(page) : NaN;
  return {
    system: isGameSystemId(system) ? system : null,
    q: (first(params.q) ?? "").trim(),
    mine: first(params.mine) === "1",
    sort: isSort(sort) ? sort : "updated",
    page: Number.isSafeInteger(pageNumber) && pageNumber >= 1 ? pageNumber : 1,
  };
}

/** 一覧の条件から URL を作る。既定の値はクエリに含めない */
export function sheetSearchHref(search: Partial<SheetSearch>): string {
  const s = { ...defaultSheetSearch, ...search };
  const query = new URLSearchParams();
  if (s.system) query.set("system", s.system);
  if (s.q.trim()) query.set("q", s.q.trim());
  if (s.mine) query.set("mine", "1");
  if (s.sort !== "updated") query.set("sort", s.sort);
  if (s.page > 1) query.set("page", String(s.page));
  const qs = query.toString();
  return qs ? `/sheets?${qs}` : "/sheets";
}

/** ページ以外の条件が既定の値から変わっているか（「条件を外す」を表示するか） */
export function hasSheetConditions(search: SheetSearch): boolean {
  return search.system !== null || search.q !== "" || search.mine || search.sort !== "updated";
}
