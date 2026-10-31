import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import { TextField } from "./text-field";

function Controlled(props: Partial<React.ComponentProps<typeof TextField>>) {
  const [value, setValue] = useState("");
  return (
    <TextField
      label="キャラクター名"
      {...props}
      value={value}
      onValueChange={(v) => {
        setValue(v);
        props.onValueChange?.(v);
      }}
    />
  );
}

it("入力すると onValueChange に値を渡す", async () => {
  const user = userEvent.setup();
  const onValueChange = vi.fn();
  const { container } = render(<Controlled onValueChange={onValueChange} />);

  await user.type(screen.getByLabelText("キャラクター名"), "あかね");
  expect(screen.getByLabelText("キャラクター名")).toHaveValue("あかね");
  expect(onValueChange).toHaveBeenLastCalledWith("あかね");
  await expectNoAxeViolations(container);
});

it("無効なら入力できない", async () => {
  const user = userEvent.setup();
  render(<Controlled disabled />);

  await user.type(screen.getByLabelText("キャラクター名"), "x");
  expect(screen.getByLabelText("キャラクター名")).toHaveValue("");
});

it("className は外枠、inputClassName は入力欄に付く", () => {
  render(<Controlled className="outer-class" inputClassName="input-class" />);

  const input = screen.getByLabelText("キャラクター名");
  expect(input).toHaveClass("input-class");
  expect(input).not.toHaveClass("outer-class");
  expect(input.closest(".outer-class")).not.toBeNull();
});
