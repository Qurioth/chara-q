import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { mockGraphQL, renderWithUrql, requestsOf } from "@/test/graphql";
import { navigation, router } from "@/test/navigation";

import { AccountMenu } from "./account-menu";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  vi.clearAllMocks();
  navigation.pathname = "/sheets";
});

it("ログインしていなければ、今の画面に戻るログインのリンクを出す", async () => {
  const { container } = renderWithUrql(<AccountMenu viewer={null} />);
  const link = screen.getByRole("link", { name: "ログイン" });
  expect(link).toHaveAttribute("href", "/login?next=%2Fsheets");
  await expectNoAxeViolations(container);
});

it("ログインしていれば表示名のボタンを出し、メニューにアカウントの設定とログアウトがある", async () => {
  const user = userEvent.setup();
  const { container } = renderWithUrql(<AccountMenu viewer={{ displayName: "たま", isAdmin: false }} />);

  const trigger = screen.getByRole("button", { name: "アカウントのメニュー（たま）" });
  expect(trigger).toHaveTextContent("たま");
  expect(screen.queryByRole("link", { name: "ログイン" })).not.toBeInTheDocument();
  await expectNoAxeViolations(container);

  await user.click(trigger);
  expect(await screen.findByRole("menuitem", { name: "アカウントの設定" })).toHaveAttribute("href", "/account");
  expect(screen.getByRole("menuitem", { name: "ログアウト" })).toBeInTheDocument();
  expect(screen.queryByRole("menuitem", { name: "招待コードの発行" })).not.toBeInTheDocument();
});

it("管理者には招待コードの発行の入口を出す", async () => {
  const user = userEvent.setup();
  renderWithUrql(<AccountMenu viewer={{ displayName: "管理者", isAdmin: true }} />);
  await user.click(screen.getByRole("button", { name: "アカウントのメニュー（管理者）" }));
  expect(await screen.findByRole("menuitem", { name: "招待コードの発行" })).toHaveAttribute(
    "href",
    "/admin/invitations",
  );
});

it("長い表示名は省略して表示する", () => {
  renderWithUrql(<AccountMenu viewer={{ displayName: "とても長い表示名".repeat(4), isAdmin: false }} />);
  const trigger = screen.getByRole("button", { name: /アカウントのメニュー/ });
  expect(trigger).toHaveClass("max-w-32");
  expect(trigger.querySelector("span")).toHaveClass("truncate");
});

it("キーボードでメニューを開き、ログアウトを選ぶとログアウトして描き直す", async () => {
  const user = userEvent.setup();
  const fetchMock = mockGraphQL(() => ({ data: { logout: true } }));
  renderWithUrql(<AccountMenu viewer={{ displayName: "たま", isAdmin: false }} />);

  await user.tab();
  expect(screen.getByRole("button", { name: "アカウントのメニュー（たま）" })).toHaveFocus();
  await user.keyboard("{Enter}");
  const logout = await screen.findByRole("menuitem", { name: "ログアウト" });
  await user.click(logout);

  await vi.waitFor(() => expect(router.refresh).toHaveBeenCalled());
  expect(requestsOf(fetchMock)[0]?.query).toContain("logout");
});
