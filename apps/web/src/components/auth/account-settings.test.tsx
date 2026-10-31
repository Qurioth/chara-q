import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { mockGraphQL, renderWithUrql, type GraphQLRequest } from "@/test/graphql";
import { router } from "@/test/navigation";

import { AccountSettings } from "./account-settings";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  vi.clearAllMocks();
});

const unauthenticated = {
  data: null,
  errors: [{ message: "ログインが必要です", extensions: { code: "UNAUTHENTICATED" } }],
};

function render(handler: (req: GraphQLRequest) => unknown = () => ({})) {
  mockGraphQL(handler);
  return renderWithUrql(<AccountSettings loginId="alice" displayName="ありす" />);
}

it("ログイン ID は表示だけで、変えられない", async () => {
  const { container } = render();
  expect(screen.getByText("alice")).toBeInTheDocument();
  expect(screen.queryByDisplayValue("alice")).not.toBeInTheDocument();
  expect(screen.getByRole("textbox", { name: /^表示名/ })).toHaveValue("ありす");
  await expectNoAxeViolations(container);
});

it("表示名を保存したら成功を知らせ、ヘッダーを描き直す", async () => {
  const user = userEvent.setup();
  render(() => ({
    data: { updateDisplayName: { viewer: { id: "1", displayName: "アリス" }, errors: [] } },
  }));
  const input = screen.getByRole("textbox", { name: /^表示名/ });
  await user.clear(input);
  await user.type(input, "アリス");
  await user.click(screen.getByRole("button", { name: "表示名を保存する" }));

  expect(await screen.findByRole("status")).toHaveTextContent("表示名を変えました");
  expect(router.refresh).toHaveBeenCalled();
});

it("現在のパスワードが違えば、その項目に出す", async () => {
  const user = userEvent.setup();
  render(() => ({
    data: { changePassword: { viewer: null, errors: [{ field: "CURRENT_PASSWORD", code: "WRONG_CURRENT_PASSWORD" }] } },
  }));
  await user.type(screen.getByLabelText(/現在のパスワード/), "wrong-password");
  await user.type(screen.getByLabelText(/新しいパスワード/), "password2");
  await user.click(screen.getByRole("button", { name: "パスワードを変える" }));
  expect(await screen.findByText("現在のパスワードが違います")).toBeInTheDocument();
});

it("パスワードを変えたら入力欄を空にして成功を知らせる", async () => {
  const user = userEvent.setup();
  render(() => ({ data: { changePassword: { viewer: { id: "1" }, errors: [] } } }));
  await user.type(screen.getByLabelText(/現在のパスワード/), "password1");
  await user.type(screen.getByLabelText(/新しいパスワード/), "password2");
  await user.click(screen.getByRole("button", { name: "パスワードを変える" }));

  expect(await screen.findByRole("status")).toHaveTextContent("パスワードを変えました");
  expect(screen.getByLabelText(/現在のパスワード/)).toHaveValue("");
  expect(screen.getByLabelText(/新しいパスワード/)).toHaveValue("");
});

it("ログイン状態が切れていたら、ログイン画面へ移る", async () => {
  const user = userEvent.setup();
  render(() => unauthenticated);
  await user.click(screen.getByRole("button", { name: "表示名を保存する" }));
  await vi.waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login?next=%2Faccount"));
});
