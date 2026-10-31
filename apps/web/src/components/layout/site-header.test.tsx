import { render, screen, within } from "@testing-library/react";
import { expect, it } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import { SiteHeader } from "./site-header";

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
