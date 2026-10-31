import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { mockGraphQL, renderWithUrql, requestsOf } from "@/test/graphql";
import { router } from "@/test/navigation";

import { RegisterForm } from "./register-form";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  vi.clearAllMocks();
});

function registerPayload(payload: Record<string, unknown>) {
  return { data: { register: { viewer: null, errors: [], ...payload } } };
}

async function fill(values: { code?: string; loginId?: string; password?: string; displayName?: string }) {
  const user = userEvent.setup();
  if (values.code) await user.type(screen.getByLabelText(/招待コード/), values.code);
  if (values.loginId) await user.type(screen.getByLabelText(/ログイン ID/), values.loginId);
  if (values.password) await user.type(screen.getByLabelText(/^パスワード/), values.password);
  if (values.displayName) await user.type(screen.getByLabelText(/表示名/), values.displayName);
  await user.click(screen.getByRole("button", { name: "登録する" }));
}

const valid = { code: "ABCD-EFGH", loginId: "tama", password: "password1", displayName: "たま" };

it("招待コード・ログイン ID・パスワード・表示名だけを入力する", async () => {
  const { container } = renderWithUrql(<RegisterForm />);
  const inputs = container.querySelectorAll("input");
  expect(inputs).toHaveLength(4);
  expect(screen.queryByLabelText(/メール/)).not.toBeInTheDocument();
  expect(screen.getByLabelText(/^パスワード/)).toHaveAttribute("autocomplete", "new-password");
  expect(screen.getByRole("link", { name: "ログイン" })).toHaveAttribute("href", "/login");
  await expectNoAxeViolations(container);
});

it("規則に合わない項目は、送信する前に項目の近くにエラーを出す", async () => {
  const fetchMock = mockGraphQL(() => registerPayload({}));
  renderWithUrql(<RegisterForm />);
  await fill({ loginId: "a!", password: "short" });
  expect(screen.getByText("招待コードを入力してください")).toBeInTheDocument();
  expect(screen.getByText("ログイン ID は半角英数字・_・- だけで入力してください")).toBeInTheDocument();
  expect(screen.getByText("パスワードは 8 文字以上で入力してください")).toBeInTheDocument();
  expect(screen.getByText("表示名を入力してください")).toBeInTheDocument();
  expect(screen.getByLabelText(/ログイン ID/)).toHaveAttribute("aria-invalid", "true");
  expect(fetchMock).not.toHaveBeenCalled();
});

it("使えない招待コードは、招待コードの項目に出す", async () => {
  mockGraphQL(() => registerPayload({ errors: [{ field: "INVITATION_CODE", code: "INVITATION_UNAVAILABLE" }] }));
  renderWithUrql(<RegisterForm />);
  await fill(valid);
  expect(await screen.findByText("この招待コードは使えません。管理者に確認してください")).toBeInTheDocument();
  expect(screen.getByLabelText(/招待コード/)).toHaveAttribute("aria-invalid", "true");
});

it("使われているログイン ID は、ログイン ID の項目に出す", async () => {
  mockGraphQL(() => registerPayload({ errors: [{ field: "LOGIN_ID", code: "LOGIN_ID_TAKEN" }] }));
  renderWithUrql(<RegisterForm />);
  await fill(valid);
  expect(await screen.findByText("このログイン ID は使われています。別のログイン ID にしてください")).toBeInTheDocument();
});

it("成功したらトップページへ移り、ヘッダーを描き直す", async () => {
  const fetchMock = mockGraphQL(() => registerPayload({ viewer: { id: "1" } }));
  renderWithUrql(<RegisterForm />);
  await fill(valid);
  await vi.waitFor(() => expect(router.replace).toHaveBeenCalledWith("/"));
  expect(router.refresh).toHaveBeenCalled();
  expect(requestsOf(fetchMock)[0]?.variables).toEqual({
    input: { invitationCode: "ABCD-EFGH", loginId: "tama", password: "password1", displayName: "たま" },
  });
});
