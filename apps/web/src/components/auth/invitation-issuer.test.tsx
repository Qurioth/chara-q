import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Activity } from "react";
import { beforeEach, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";
import { UrqlProvider } from "@/lib/urql-provider";
import { mockGraphQL, renderWithUrql } from "@/test/graphql";

import { InvitationIssuer } from "./invitation-issuer";

vi.mock("next/navigation", () => import("@/test/navigation"));

beforeEach(() => {
  vi.clearAllMocks();
});

it("発行するとコードと有効期限を表示し、コピーできる", async () => {
  const user = userEvent.setup();
  const writeText = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
  const expiresAt = new Date(2026, 10, 8, 12, 0).toISOString();
  mockGraphQL(() => ({ data: { issueInvitation: { code: "ABCD-EFGH-IJKL", expiresAt } } }));
  const { container } = renderWithUrql(<InvitationIssuer />);

  await user.click(screen.getByRole("button", { name: "招待コードを発行する" }));
  expect(await screen.findByText("ABCD-EFGH-IJKL")).toBeInTheDocument();
  expect(screen.getByText("有効期限: 2026/11/08 12:00")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "コピー" }));
  expect(writeText).toHaveBeenCalledWith("ABCD-EFGH-IJKL");
  expect(screen.getByRole("status")).toHaveTextContent("コピーしました");
  await expectNoAxeViolations(container);
});

it("管理者でなければ文言を出す", async () => {
  const user = userEvent.setup();
  mockGraphQL(() => ({
    data: null,
    errors: [{ message: "管理者だけが行えます", extensions: { code: "FORBIDDEN" } }],
  }));
  renderWithUrql(<InvitationIssuer />);
  await user.click(screen.getByRole("button", { name: "招待コードを発行する" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("この操作は管理者だけが行えます");
});

it("画面を離れたら（アンマウント・Activity で隠れたら）発行したコードを消す", async () => {
  const user = userEvent.setup();
  mockGraphQL(() => ({ data: { issueInvitation: { code: "ABCD-EFGH", expiresAt: new Date().toISOString() } } }));
  const { rerender } = renderWithUrql(<InvitationIssuer />);
  await user.click(screen.getByRole("button", { name: "招待コードを発行する" }));
  expect(await screen.findByText("ABCD-EFGH")).toBeInTheDocument();

  // React の Activity で隠してから表示し直す
  rerender(
    <UrqlProvider>
      <Activity mode="hidden">
        <InvitationIssuer />
      </Activity>
    </UrqlProvider>,
  );
  rerender(
    <UrqlProvider>
      <Activity mode="visible">
        <InvitationIssuer />
      </Activity>
    </UrqlProvider>,
  );
  expect(screen.queryByText("ABCD-EFGH")).not.toBeInTheDocument();
});
