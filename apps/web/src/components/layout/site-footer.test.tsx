import { render, screen, within } from "@testing-library/react";
import { expect, it } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import { SiteFooter } from "./site-footer";

it("アプリ名と著作権表記を表示する", async () => {
  const { container } = render(<SiteFooter />);

  const footer = screen.getByRole("contentinfo");
  expect(within(footer).getByText("CharaQ")).toBeInTheDocument();
  expect(within(footer).getByText("© 2026 Qurioth")).toBeInTheDocument();
  await expectNoAxeViolations(container);
});
