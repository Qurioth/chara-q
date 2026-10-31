import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";

import { LogoMark } from "./logo-mark";

it("title がなければ装飾として支援技術から隠す", () => {
  const { container } = render(<LogoMark className="size-7" />);
  const svg = container.querySelector("svg")!;
  expect(svg).toHaveAttribute("aria-hidden", "true");
  expect(svg).toHaveClass("size-7");
});

it("title があれば画像として名前を付ける", () => {
  render(<LogoMark title="CharaQ のロゴ" />);
  expect(screen.getByRole("img", { name: "CharaQ のロゴ" })).toBeInTheDocument();
});

it("色は currentColor で、使う側の文字色に従う", () => {
  const { container } = render(<LogoMark />);
  const painted = container.querySelectorAll("[stroke], [fill]");
  for (const el of painted) {
    for (const attr of ["stroke", "fill"]) {
      const value = el.getAttribute(attr);
      if (value) expect(["currentColor", "none"]).toContain(value);
    }
  }
});
