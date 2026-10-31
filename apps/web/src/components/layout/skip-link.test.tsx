import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, it } from "vitest";

import { SkipLink } from "./skip-link";

it("本文へのスキップリンクが最初の Tab でフォーカスされる", async () => {
  const user = userEvent.setup();
  render(
    <>
      <SkipLink />
      <button type="button">次の要素</button>
    </>,
  );

  const link = screen.getByRole("link", { name: "本文へスキップ" });
  expect(link).toHaveAttribute("href", "#main-content");

  await user.tab();
  expect(link).toHaveFocus();
});
