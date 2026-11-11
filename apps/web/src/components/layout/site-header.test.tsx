import { render, screen, within } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { navigation } from "@/test/navigation";

import { SiteHeader } from "./site-header";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  navigation.pathname = "/";
});

it("シートの一覧への入口がある（幅の狭い画面では文字を隠し、aria-label で読ませる）", () => {
  render(<SiteHeader />);
  const link = screen.getByRole("link", { name: "シート一覧" });
  expect(link).toHaveAttribute("href", "/sheets");
  expect(link).not.toHaveAttribute("aria-current");
  expect(link.querySelector("span")).toHaveClass("hidden", "sm:inline");
});

it("右側の操作の領域は「補助の操作・シートの一覧・アカウント」の順に並べる（ロゴの隣には置かない）", () => {
  render(
    <SiteHeader
      utilities={<button type="button">テーマ</button>}
      actions={<button type="button">アカウント</button>}
    />,
  );
  const theme = screen.getByRole("button", { name: "テーマ" });
  const link = screen.getByRole("link", { name: "シート一覧" });
  const account = screen.getByRole("button", { name: "アカウント" });
  const logo = screen.getByRole("link", { name: "CharaQ" });

  const region = theme.parentElement!;
  // ロゴのすぐ隣は右側の操作の領域で、間にほかの要素を挟まない
  expect(logo.nextElementSibling).toBe(region);
  const children = [...region.children];
  expect(children).toHaveLength(3);
  expect(children[0]).toBe(theme);
  expect(children[1]).toContainElement(link);
  expect(children[2]).toBe(account);
});

it.each(["/sheets", "/sheets/0190a000-0000-7000-8000-000000000001"])(
  "%s を表示しているときは、シートの一覧の入口に aria-current を付ける",
  (pathname) => {
    navigation.pathname = pathname;
    render(<SiteHeader />);
    expect(screen.getByRole("link", { name: "シート一覧" })).toHaveAttribute("aria-current", "page");
  },
);

it("アプリ名がトップページへのリンクになっている", async () => {
  const { container } = render(<SiteHeader />);

  const header = screen.getByRole("banner");
  expect(within(header).getByRole("link", { name: "CharaQ" })).toHaveAttribute("href", "/");
  await expectNoAxeViolations(container);
});

it("actions に渡した要素を操作の領域に描画する", () => {
  render(<SiteHeader actions={<button type="button">操作</button>} />);

  expect(
    within(screen.getByRole("banner")).getByRole("button", { name: "操作" }),
  ).toBeInTheDocument();
});

it("アプリ名はロゴの書体（font-logo）で表示する", () => {
  render(<SiteHeader />);
  expect(screen.getByRole("link", { name: "CharaQ" })).toHaveClass("font-logo");
});

it("アプリ名の前に朱のロゴを置き、ロゴは読み上げない", () => {
  render(<SiteHeader />);
  const link = screen.getByRole("link", { name: "CharaQ" });
  const logo = link.querySelector("svg");
  expect(logo).toHaveAttribute("aria-hidden", "true");
  expect(logo).toHaveClass("text-seal");
});
