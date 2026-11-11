import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import { Pagination, pageItems } from "./pagination";

const hrefFor = (page: number) => `/sheets?q=a&page=${page}`;

it("1 ページだけなら何も表示しない", () => {
  const { container } = render(<Pagination page={1} totalPages={1} hrefFor={hrefFor} />);
  expect(container).toBeEmptyDOMElement();
});

it("現在のページに aria-current を付け、各リンクは条件を保つ", async () => {
  const { container } = render(<Pagination page={2} totalPages={3} hrefFor={hrefFor} />);
  expect(screen.getByRole("link", { name: "2 ページ目" })).toHaveAttribute("aria-current", "page");
  expect(screen.getByRole("link", { name: "1 ページ目" })).not.toHaveAttribute("aria-current");
  expect(screen.getByRole("link", { name: "前へ" })).toHaveAttribute("href", "/sheets?q=a&page=1");
  expect(screen.getByRole("link", { name: "次へ" })).toHaveAttribute("href", "/sheets?q=a&page=3");
  await expectNoAxeViolations(container);
});

it("最初のページに「前へ」、最後のページに「次へ」はない", () => {
  const { rerender } = render(<Pagination page={1} totalPages={3} hrefFor={hrefFor} />);
  expect(screen.queryByRole("link", { name: "前へ" })).not.toBeInTheDocument();
  rerender(<Pagination page={3} totalPages={3} hrefFor={hrefFor} />);
  expect(screen.queryByRole("link", { name: "次へ" })).not.toBeInTheDocument();
});

describe("pageItems", () => {
  it.each([
    [1, 5, [1, 2, 3, 4, 5]],
    [1, 10, [1, 2, 3, "gap", 10]],
    [5, 10, [1, 2, 3, 4, 5, 6, 7, "gap", 10]],
    [6, 12, [1, "gap", 4, 5, 6, 7, 8, "gap", 12]],
    [12, 12, [1, "gap", 10, 11, 12]],
  ])("page=%i, totalPages=%i は %o", (page, total, want) => {
    expect(pageItems(page, total)).toEqual(want);
  });
});
