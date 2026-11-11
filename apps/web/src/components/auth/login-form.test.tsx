import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { mockGraphQL, renderWithUrql, requestsOf } from "@/test/graphql";
import { router } from "@/test/navigation";

import { LoginForm } from "./login-form";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  vi.clearAllMocks();
});

function loginPayload(payload: Record<string, unknown>) {
  return { data: { login: { viewer: null, errors: [], lockedUntil: null, ...payload } } };
}

async function fillAndSubmit(loginId: string, password: string) {
  const user = userEvent.setup();
  if (loginId) await user.type(screen.getByLabelText(/ログイン ID/), loginId);
  if (password) await user.type(screen.getByLabelText(/^パスワード/), password);
  await user.click(screen.getByRole("button", { name: "ログイン" }));
}

it("入力欄はパスワード管理ソフトで使える形にする", async () => {
  const { container } = renderWithUrql(<LoginForm next="/" />);
  expect(screen.getByLabelText(/ログイン ID/)).toHaveAttribute("autocomplete", "username");
  const password = screen.getByLabelText(/^パスワード/);
  expect(password).toHaveAttribute("type", "password");
  expect(password).toHaveAttribute("autocomplete", "current-password");
  expect(screen.getByRole("link", { name: "ユーザー登録" })).toHaveAttribute("href", "/register");
  expect(screen.getByRole("link", { name: "パスワードの再設定" })).toHaveAttribute("href", "/reset-password");
  await expectNoAxeViolations(container);
});

it("空のまま送ると項目のエラーを出し、送信しない", async () => {
  const fetchMock = mockGraphQL(() => loginPayload({}));
  renderWithUrql(<LoginForm next="/" />);
  await fillAndSubmit("", "");
  expect(screen.getByText("ログイン ID を入力してください")).toBeInTheDocument();
  expect(screen.getByText("パスワードを入力してください")).toBeInTheDocument();
  expect(fetchMock).not.toHaveBeenCalled();
});

it("ログイン ID かパスワードが違えば、フォームの上に文言を出す", async () => {
  mockGraphQL(() => loginPayload({ errors: [{ field: null, code: "INVALID_CREDENTIALS" }] }));
  renderWithUrql(<LoginForm next="/" />);
  await fillAndSubmit("alice", "wrong-password");
  expect(await screen.findByRole("alert")).toHaveTextContent("ログイン ID またはパスワードが違います");
  expect(screen.getByLabelText(/^パスワード/)).toHaveValue("");
  expect(router.replace).not.toHaveBeenCalled();
});

it("ロック中なら、解ける時刻の目安を出す", async () => {
  const until = new Date(2026, 10, 1, 9, 5).toISOString();
  mockGraphQL(() => loginPayload({ errors: [{ field: null, code: "LOCKED" }], lockedUntil: until }));
  renderWithUrql(<LoginForm next="/" />);
  await fillAndSubmit("alice", "wrong-password");
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "ログインに続けて失敗したため、09:05 ごろまでログインできません",
  );
});

it("成功したら戻り先へ移り、ヘッダーを描き直す", async () => {
  const fetchMock = mockGraphQL(() => loginPayload({ viewer: { id: "1" } }));
  renderWithUrql(<LoginForm next="/account" />);
  await fillAndSubmit("alice", "password1");
  await vi.waitFor(() => expect(router.replace).toHaveBeenCalledWith("/account"));
  expect(router.refresh).toHaveBeenCalled();
  expect(requestsOf(fetchMock)[0]?.variables).toEqual({ input: { loginId: "alice", password: "password1" } });
});

it("パスワードを再設定した直後は、そのことを知らせる", () => {
  renderWithUrql(<LoginForm next="/" reset />);
  expect(screen.getByRole("status")).toHaveTextContent("パスワードを再設定しました");
});
