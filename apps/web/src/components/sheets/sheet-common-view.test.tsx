import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";

import type { SheetDetail } from "@/lib/sheets/types";
import { expectNoAxeViolations } from "@/test/axe";

import { SheetCommonView } from "./sheet-common-view";

const sheet: SheetDetail = {
  id: "0190a000-0000-7000-8000-000000000001",
  system: "yuyakekoyake",
  characterName: "こぎつね",
  playerName: "たろう",
  age: "",
  sex: "",
  memo: "1 行目\n2 行目",
  revision: 1,
  createdAt: "2026-11-09T15:30:00Z",
  updatedAt: "2026-11-10T03:00:00Z",
  owner: { id: "u1", displayName: "所有者さん" },
  data: {},
};

it("共通の項目・システム・所有者・日時を表示する", async () => {
  const { container } = render(<SheetCommonView sheet={sheet} isOwner={false} />);

  expect(screen.getByRole("heading", { level: 1, name: "こぎつね" })).toBeInTheDocument();
  expect(screen.getByText("ふしぎもののけRPG ゆうやけこやけ")).toBeInTheDocument();
  expect(screen.getByText("所有者さん")).toBeInTheDocument();
  expect(screen.getByText("2026/11/10 00:30")).toBeInTheDocument();
  expect(screen.getByText("2026/11/10 12:00")).toBeInTheDocument();
  expect(screen.getByText("たろう")).toBeInTheDocument();
  await expectNoAxeViolations(container);
});

it("空の項目は「未入力」と表示する", () => {
  render(<SheetCommonView sheet={sheet} isOwner={false} />);
  expect(screen.getAllByText("未入力")).toHaveLength(2);
});

it("メモの改行を保つ", () => {
  render(<SheetCommonView sheet={sheet} isOwner={false} />);
  const memo = screen.getByText((_, el) => el?.tagName === "DD" && el.textContent === "1 行目\n2 行目");
  expect(memo).toHaveClass("whitespace-pre-wrap");
});

it("HTML の文字列は文字として表示する", () => {
  const { container } = render(
    <SheetCommonView sheet={{ ...sheet, characterName: "<b>太字</b>" }} isOwner={false} />,
  );
  expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("<b>太字</b>");
  expect(container.querySelector("b")).toBeNull();
});

it("システム固有の部分の場所に「準備中」を表示する", () => {
  render(<SheetCommonView sheet={sheet} isOwner={true} />);
  expect(screen.getByText("このシステムの項目は準備中です。")).toBeInTheDocument();
});

it("システムの領域として描く（システムのフォントを使う）", () => {
  const { container } = render(<SheetCommonView sheet={{ ...sheet, system: "beginning-idol" }} isOwner={false} />);
  expect(container.querySelector("[data-game-system='beginning-idol']")).not.toBeNull();
});
