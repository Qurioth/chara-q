import { render, screen, within } from "@testing-library/react";
import { expect, it } from "vitest";

import { gameSystems } from "@/lib/game-systems";
import { expectNoAxeViolations } from "@/test/axe";

import Home from "./page";

it("紹介と対応予定のゲームシステムを表示する", async () => {
  const { container } = render(<Home />);

  expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
    "TRPGのキャラクターをつくろう",
  );
  expect(
    screen.getByText(
      "キャラクターシートを作成するシステムを選択してください。",
    ),
  ).toBeInTheDocument();

  const list = screen.getByRole("list", { name: "対応予定のゲームシステム" });
  const items = within(list).getAllByRole("listitem");
  expect(items).toHaveLength(gameSystems.length);

  for (const system of gameSystems) {
    const item = items.find((el) => el.textContent?.includes(system.name));
    expect(item).toBeDefined();
    expect(within(item!).getByText("準備中")).toBeInTheDocument();
    expect(within(item!).queryByRole("link")).not.toBeInTheDocument();
  }

  await expectNoAxeViolations(container);
});

it("API の接続状態は表示しない", () => {
  render(<Home />);
  expect(screen.queryByText(/API:/)).not.toBeInTheDocument();
});
