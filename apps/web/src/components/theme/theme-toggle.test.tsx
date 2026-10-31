import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, it } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { mockColorScheme } from "@/test/match-media";

import { THEME_STORAGE_KEY } from "./theme";
import { themeScript } from "./theme-script";
import { ThemeToggle } from "./theme-toggle";

const html = document.documentElement;

it("押すとテーマを切り替えて保存し、読み上げ用の名前も入れ替わる", async () => {
  const user = userEvent.setup();
  const { container } = render(<ThemeToggle />);

  await user.click(screen.getByRole("button", { name: "ダークモードに切り替える" }));
  expect(html).toHaveClass("dark");
  expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe("dark");

  await user.click(screen.getByRole("button", { name: "ライトモードに切り替える" }));
  expect(html).not.toHaveClass("dark");
  expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe("light");

  await expectNoAxeViolations(container);
});

it("キーボードの Enter と Space で切り替えられる", async () => {
  const user = userEvent.setup();
  render(<ThemeToggle />);

  await user.tab();
  expect(screen.getByRole("button")).toHaveFocus();

  await user.keyboard("{Enter}");
  expect(html).toHaveClass("dark");

  await user.keyboard(" ");
  expect(html).not.toHaveClass("dark");
});

it("未選択の間は端末の設定の変化に読み上げ用の名前も追従する", async () => {
  const media = mockColorScheme("light");
  new Function(themeScript)();
  render(<ThemeToggle />);
  expect(screen.getByRole("button", { name: "ダークモードに切り替える" })).toBeInTheDocument();

  await act(async () => media.setScheme("dark"));
  expect(
    await screen.findByRole("button", { name: "ライトモードに切り替える" }),
  ).toBeInTheDocument();
});
