import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { SheetSummary } from "@/lib/sheets/types";
import { expectNoAxeViolations } from "@/test/axe";

import { SheetList } from "./sheet-list";

const sortProps = { sort: "updated", sortHref: (sort: string) => `/sheets?sort=${sort}` } as const;

const sheets: SheetSummary[] = [
  {
    id: "id-1",
    system: "yuyakekoyake",
    characterName: "こぎつね",
    playerName: "たろう",
    updatedAt: "2026-11-09T15:30:00Z",
    owner: { displayName: "所有者 A" },
  },
  {
    id: "id-2",
    system: "beginning-idol",
    characterName: "アイドル",
    playerName: "",
    updatedAt: "2026-11-10T03:00:00Z",
    owner: { displayName: "所有者 B" },
  },
];

it("件数と、各行のキャラクター名・プレイヤー名・システム・所有者・更新日時を表示する", async () => {
  const { container } = render(<SheetList sheets={sheets} totalCount={42} hasConditions={false} {...sortProps} />);

  expect(screen.getByText("全 42 件")).toBeInTheDocument();
  const table = screen.getByRole("table");
  const rows = within(table).getAllByRole("row").slice(1);
  expect(rows).toHaveLength(2);
  const first = within(rows[0]);
  expect(first.getByRole("link", { name: "こぎつね" })).toHaveAttribute("href", "/sheets/id-1");
  expect(first.getByText("たろう")).toBeInTheDocument();
  expect(first.getByText("ふしぎもののけRPG ゆうやけこやけ")).toBeInTheDocument();
  expect(first.getByText("所有者 A")).toBeInTheDocument();
  expect(first.getByText("2026/11/10 00:30")).toBeInTheDocument();
  await expectNoAxeViolations(container);
});

it("幅の狭い画面向けのカードにも同じ行がある", () => {
  render(<SheetList sheets={sheets} totalCount={2} hasConditions={false} {...sortProps} />);
  const cards = screen.getAllByRole("listitem");
  expect(cards).toHaveLength(2);
  expect(within(cards[1]).getByRole("link", { name: "アイドル" })).toHaveAttribute("href", "/sheets/id-2");
});

it("条件ありで 0 件なら「条件に合うシートはありません」と「条件を外す」", () => {
  render(<SheetList sheets={[]} totalCount={0} hasConditions={true} {...sortProps} />);
  expect(screen.getByText("条件に合うシートはありません。")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "条件を外す" })).toHaveAttribute("href", "/sheets");
});

it("条件なしで 0 件なら「まだシートはありません」", () => {
  render(<SheetList sheets={[]} totalCount={0} hasConditions={false} {...sortProps} />);
  expect(screen.getByText("まだシートはありません。")).toBeInTheDocument();
  expect(screen.queryByRole("link", { name: "条件を外す" })).not.toBeInTheDocument();
});

describe("表の見出しでの並び替え", () => {
  function header(name: string) {
    return screen.getByRole("columnheader", { name: new RegExp(`^${name}`) });
  }

  it("既定（更新日時の新しい順）では、更新日時が降順。見出しのリンクは、更新日時は逆の向き、キャラクター名は名前の順", async () => {
    const { container } = render(<SheetList sheets={sheets} totalCount={2} hasConditions={false} {...sortProps} />);

    expect(header("更新日時")).toHaveAttribute("aria-sort", "descending");
    expect(header("キャラクター名")).not.toHaveAttribute("aria-sort");
    expect(within(header("更新日時")).getByRole("link")).toHaveAttribute("href", "/sheets?sort=updated-asc");
    expect(within(header("更新日時")).getByRole("link")).toHaveAccessibleName(
      "更新日時（更新日時の古い順に並べ替える）",
    );
    expect(within(header("キャラクター名")).getByRole("link")).toHaveAttribute("href", "/sheets?sort=name");
    // 並び替えられない列は、リンクにしない
    expect(within(screen.getByRole("columnheader", { name: "プレイヤー名" })).queryByRole("link")).toBeNull();
    await expectNoAxeViolations(container);
  });

  it.each([
    ["name", "キャラクター名", "ascending", "/sheets?sort=name-desc"],
    ["name-desc", "キャラクター名", "descending", "/sheets?sort=name"],
    ["updated-asc", "更新日時", "ascending", "/sheets?sort=updated"],
  ] as const)("sort=%s では、%s が %s。もう一度選ぶと %s", (sort, name, direction, href) => {
    render(<SheetList sheets={sheets} totalCount={2} hasConditions={false} {...sortProps} sort={sort} />);
    expect(header(name)).toHaveAttribute("aria-sort", direction);
    expect(within(header(name)).getByRole("link")).toHaveAttribute("href", href);
  });

  it("作成日時の順では、どちらの見出しにも aria-sort を付けない", () => {
    render(<SheetList sheets={sheets} totalCount={2} hasConditions={false} {...sortProps} sort="created-asc" />);
    expect(header("キャラクター名")).not.toHaveAttribute("aria-sort");
    expect(header("更新日時")).not.toHaveAttribute("aria-sort");
    expect(within(header("更新日時")).getByRole("link")).toHaveAttribute("href", "/sheets?sort=updated");
  });
});
