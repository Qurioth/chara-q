import { describe, expect, it } from "vitest";

import {
  defaultSheetSearch,
  hasSheetConditions,
  nextSheetSort,
  parseSheetSearch,
  sheetSearchHref,
  sheetSortDirection,
  sheetSortOptions,
  toSheetOrder,
} from "./search-params";

describe("parseSheetSearch", () => {
  it("何もなければ既定の値", () => {
    expect(parseSheetSearch({})).toEqual(defaultSheetSearch);
  });

  it("すべての値を読む", () => {
    expect(
      parseSheetSearch({ system: "beginning-idol", q: " きつね ", mine: "1", sort: "name", page: "3" }),
    ).toEqual({ system: "beginning-idol", q: "きつね", mine: true, sort: "name", page: 3 });
  });

  it.each([
    ["system", "bar", { system: null }],
    ["system", "BEGINNING_IDOL", { system: null }],
    ["sort", "foo", { sort: "updated" }],
    ["mine", "true", { mine: false }],
    ["mine", "0", { mine: false }],
    ["page", "0", { page: 1 }],
    ["page", "-2", { page: 1 }],
    ["page", "1.5", { page: 1 }],
    ["page", "abc", { page: 1 }],
    ["page", "99999999999999999999", { page: 1 }],
    ["page", "999", { page: 999 }],
  ])("%s=%s は %o として扱う", (key, value, want) => {
    expect(parseSheetSearch({ [key]: value })).toEqual({ ...defaultSheetSearch, ...want });
  });

  it("同じキーが複数あるときは最初の値", () => {
    expect(parseSheetSearch({ sort: ["created", "name"], page: ["2", "5"] })).toMatchObject({
      sort: "created",
      page: 2,
    });
  });
});

describe("sheetSearchHref", () => {
  it("既定の値はクエリに含めない", () => {
    expect(sheetSearchHref({})).toBe("/sheets");
    expect(sheetSearchHref(defaultSheetSearch)).toBe("/sheets");
  });

  it("条件をクエリにする", () => {
    expect(
      sheetSearchHref({ system: "yuyakekoyake", q: "きつね", mine: true, sort: "created", page: 2 }),
    ).toBe(`/sheets?system=yuyakekoyake&q=${encodeURIComponent("きつね")}&mine=1&sort=created&page=2`);
  });

  it("読み取った条件から作った URL を読み直すと同じ条件になる", () => {
    const search = { system: "beginning-idol", q: "a&b=c", mine: true, sort: "name", page: 4 } as const;
    const href = sheetSearchHref(search);
    const params = Object.fromEntries(new URL(href, "http://localhost").searchParams);
    expect(parseSheetSearch(params)).toEqual(search);
  });
});

it("並び順を GraphQL の値にする", () => {
  expect(toSheetOrder("updated")).toBe("UPDATED_AT_DESC");
  expect(toSheetOrder("updated-asc")).toBe("UPDATED_AT_ASC");
  expect(toSheetOrder("created")).toBe("CREATED_AT_DESC");
  expect(toSheetOrder("created-asc")).toBe("CREATED_AT_ASC");
  expect(toSheetOrder("name")).toBe("CHARACTER_NAME_ASC");
  expect(toSheetOrder("name-desc")).toBe("CHARACTER_NAME_DESC");
});

it("プルダウンは、項目ごとに最初の向きと逆の向きを並べた 6 つ", () => {
  expect(sheetSortOptions.map((o) => o.label)).toEqual([
    "更新日時の新しい順",
    "更新日時の古い順",
    "作成日時の新しい順",
    "作成日時の古い順",
    "キャラクター名の順",
    "キャラクター名の逆の順",
  ]);
});

it.each(["updated-asc", "created-asc", "name-desc"] as const)("sort=%s を読み、URL に戻せる", (sort) => {
  expect(parseSheetSearch({ sort })).toEqual({ ...defaultSheetSearch, sort });
  expect(sheetSearchHref({ sort })).toBe(`/sheets?sort=${sort}`);
});

it.each(["toString", "__proto__", "updated-desc", "name-asc"])("sort=%s は既定の並び順として扱う", (sort) => {
  expect(parseSheetSearch({ sort }).sort).toBe("updated");
});

describe("表の見出しでの並び替え", () => {
  it.each([
    // [今の並び順, 選んだ見出し, 次の並び順]
    ["updated", "updated", "updated-asc"],
    ["updated-asc", "updated", "updated"],
    ["updated", "name", "name"],
    ["name", "name", "name-desc"],
    ["name-desc", "name", "name"],
    ["name-desc", "updated", "updated"],
    ["created-asc", "name", "name"],
  ] as const)("%s で %s の見出しを選ぶと %s", (current, key, next) => {
    expect(nextSheetSort(current, key)).toBe(next);
  });

  it("並んでいる項目だけ向きを返す", () => {
    expect(sheetSortDirection("updated", "updated")).toBe("descending");
    expect(sheetSortDirection("updated-asc", "updated")).toBe("ascending");
    expect(sheetSortDirection("name", "name")).toBe("ascending");
    expect(sheetSortDirection("name-desc", "name")).toBe("descending");
    expect(sheetSortDirection("created", "updated")).toBeNull();
    expect(sheetSortDirection("updated", "name")).toBeNull();
  });
});

it("ページ以外の条件があるかを判定する", () => {
  expect(hasSheetConditions(defaultSheetSearch)).toBe(false);
  expect(hasSheetConditions({ ...defaultSheetSearch, page: 3 })).toBe(false);
  expect(hasSheetConditions({ ...defaultSheetSearch, q: "a" })).toBe(true);
  expect(hasSheetConditions({ ...defaultSheetSearch, sort: "name" })).toBe(true);
  expect(hasSheetConditions({ ...defaultSheetSearch, sort: "updated-asc" })).toBe(true);
});
