import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";

import { defaultSheetSearch } from "@/lib/sheets/search-params";
import { expectNoAxeViolations } from "@/test/axe";
import { router } from "@/test/navigation";

import { SheetSearchForm } from "./sheet-search-form";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  vi.clearAllMocks();
});

it("送信すると、条件を既定の値を除いた URL にして移動する（ページは含めない）", async () => {
  const user = userEvent.setup();
  render(<SheetSearchForm search={{ ...defaultSheetSearch, page: 3 }} loggedIn={false} />);

  await user.type(screen.getByRole("searchbox", { name: "キーワード" }), "きつね");
  await user.click(screen.getByRole("combobox", { name: "システム" }));
  await user.click(await screen.findByRole("option", { name: /ゆうやけこやけ/ }));
  await user.click(screen.getByRole("combobox", { name: "並び順" }));
  await user.click(await screen.findByRole("option", { name: "キャラクター名の順" }));
  await user.click(screen.getByRole("button", { name: "検索" }));

  expect(router.push).toHaveBeenCalledWith(
    `/sheets?system=yuyakekoyake&q=${encodeURIComponent("きつね")}&sort=name`,
  );
});

it("条件がなければ /sheets に移動する", async () => {
  const user = userEvent.setup();
  render(<SheetSearchForm search={defaultSheetSearch} loggedIn={false} />);
  await user.click(screen.getByRole("button", { name: "検索" }));
  expect(router.push).toHaveBeenCalledWith("/sheets");
});

it("システムの選択肢は「すべて」と 2 つのシステム", async () => {
  const user = userEvent.setup();
  render(<SheetSearchForm search={defaultSheetSearch} loggedIn={false} />);
  expect(screen.getByRole("combobox", { name: "システム" })).toHaveTextContent("すべて");
  await user.click(screen.getByRole("combobox", { name: "システム" }));
  const options = await screen.findAllByRole("option");
  expect(options.map((o) => o.textContent)).toEqual([
    "すべて",
    "ふしぎもののけRPG ゆうやけこやけ",
    "駆け出しアイドルRPG ビギニングアイドル",
  ]);
});

it("ログインしていなければ「自分のシートだけ」を表示しない", () => {
  render(<SheetSearchForm search={{ ...defaultSheetSearch, mine: true }} loggedIn={false} />);
  expect(screen.queryByRole("checkbox", { name: "自分のシートだけ" })).not.toBeInTheDocument();
});

it("ログインしていれば「自分のシートだけ」で絞り込める", async () => {
  const user = userEvent.setup();
  const { container } = render(<SheetSearchForm search={defaultSheetSearch} loggedIn={true} />);
  await user.click(screen.getByRole("checkbox", { name: "自分のシートだけ" }));
  await user.click(screen.getByRole("button", { name: "検索" }));
  expect(router.push).toHaveBeenCalledWith("/sheets?mine=1");
  await expectNoAxeViolations(container);
});

it("条件があるときだけ「条件を外す」を表示する", () => {
  const { rerender } = render(<SheetSearchForm search={defaultSheetSearch} loggedIn={false} />);
  expect(screen.queryByRole("link", { name: "条件を外す" })).not.toBeInTheDocument();
  rerender(<SheetSearchForm search={{ ...defaultSheetSearch, q: "a" }} loggedIn={false} />);
  expect(screen.getByRole("link", { name: "条件を外す" })).toHaveAttribute("href", "/sheets");
});
