import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { mockGraphQL, renderWithUrql } from "@/test/graphql";
import { router } from "@/test/navigation";

import { ResetPasswordForm } from "./reset-password-form";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  vi.clearAllMocks();
});

function resetPayload(payload: Record<string, unknown>) {
  return { data: { resetPassword: { ok: false, errors: [], lockedUntil: null, ...payload } } };
}

async function fill(values: { loginId?: string; code?: string; password?: string }) {
  const user = userEvent.setup();
  if (values.loginId) await user.type(screen.getByLabelText(/ログイン ID/), values.loginId);
  if (values.code) await user.type(screen.getByLabelText(/再設定コード/), values.code);
  if (values.password) await user.type(screen.getByLabelText(/新しいパスワード/), values.password);
  await user.click(screen.getByRole("button", { name: "パスワードを再設定する" }));
}

const valid = { loginId: "alice", code: "ABCD-EFGH", password: "password2" };

it("3 つの項目の規則のエラーを出し、送信しない", async () => {
  const fetchMock = mockGraphQL(() => resetPayload({}));
  const { container } = renderWithUrql(<ResetPasswordForm />);
  expect(screen.getByLabelText(/新しいパスワード/)).toHaveAttribute("autocomplete", "new-password");
  await expectNoAxeViolations(container);

  await fill({ password: "short" });
  expect(screen.getByText("ログイン ID を入力してください")).toBeInTheDocument();
  expect(screen.getByText("再設定コードを入力してください")).toBeInTheDocument();
  expect(screen.getByText("新しいパスワードは 8 文字以上で入力してください")).toBeInTheDocument();
  expect(fetchMock).not.toHaveBeenCalled();
});

it("使えない再設定コードは、再設定コードの項目に出す", async () => {
  mockGraphQL(() => resetPayload({ errors: [{ field: "RESET_CODE", code: "RESET_CODE_UNAVAILABLE" }] }));
  renderWithUrql(<ResetPasswordForm />);
  await fill(valid);
  expect(await screen.findByText("この再設定コードは使えません。管理者に確認してください")).toBeInTheDocument();
  expect(router.replace).not.toHaveBeenCalled();
});

it("ロック中の文言を出す", async () => {
  const until = new Date(2026, 10, 1, 21, 30).toISOString();
  mockGraphQL(() => resetPayload({ errors: [{ field: null, code: "LOCKED" }], lockedUntil: until }));
  renderWithUrql(<ResetPasswordForm />);
  await fill(valid);
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "パスワードの再設定に続けて失敗したため、21:30 ごろまで再設定できません",
  );
});

it("成功したらログイン画面へ移る", async () => {
  mockGraphQL(() => resetPayload({ ok: true }));
  renderWithUrql(<ResetPasswordForm />);
  await fill(valid);
  await vi.waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login?reset=1"));
});
