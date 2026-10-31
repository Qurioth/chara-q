import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";

import { GameSystemScope } from "./game-system-scope";

it("system の値を data-game-system に付ける", () => {
  render(
    <GameSystemScope system="yuyakekoyake" className="scope-class">
      <p>中身</p>
    </GameSystemScope>,
  );

  const scope = screen.getByText("中身").parentElement!;
  expect(scope.tagName).toBe("DIV");
  expect(scope).toHaveAttribute("data-game-system", "yuyakekoyake");
  expect(scope).toHaveClass("scope-class");
});

it("as で要素を変えられる", () => {
  render(
    <GameSystemScope system="beginning-idol" as="section" aria-label="シート">
      <p>中身</p>
    </GameSystemScope>,
  );

  expect(screen.getByRole("region", { name: "シート" })).toHaveAttribute(
    "data-game-system",
    "beginning-idol",
  );
});

it("入れ子では内側の指定が近い祖先になる", () => {
  render(
    <GameSystemScope system="yuyakekoyake">
      <GameSystemScope system="beginning-idol">
        <p>内側</p>
      </GameSystemScope>
    </GameSystemScope>,
  );

  expect(screen.getByText("内側").closest("[data-game-system]")).toHaveAttribute(
    "data-game-system",
    "beginning-idol",
  );
});
