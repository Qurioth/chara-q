import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import SheetNotFound from "./not-found";

it("見つからないことと、一覧へのリンクを表示する", async () => {
  const { container } = render(<SheetNotFound />);

  expect(screen.getByRole("heading", { level: 1, name: "シートが見つかりません" })).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "キャラクターシート一覧へ" })).toHaveAttribute("href", "/sheets");
  expect(container).not.toHaveTextContent("削除");
  await expectNoAxeViolations(container);
});
