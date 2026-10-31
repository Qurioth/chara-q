import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import { TextareaField } from "./textarea-field";

function Controlled(props: Partial<React.ComponentProps<typeof TextareaField>>) {
  const [value, setValue] = useState("");
  return (
    <TextareaField
      label="メモ"
      {...props}
      value={value}
      onValueChange={(v) => {
        setValue(v);
        props.onValueChange?.(v);
      }}
    />
  );
}

it("改行を保ったまま値を渡す", async () => {
  const user = userEvent.setup();
  const onValueChange = vi.fn();
  const { container } = render(<Controlled onValueChange={onValueChange} />);

  await user.type(screen.getByLabelText("メモ"), "一行目{Enter}二行目");
  expect(onValueChange).toHaveBeenLastCalledWith("一行目\n二行目");
  await expectNoAxeViolations(container);
});

it("rows の既定は 3", () => {
  render(<Controlled />);
  expect(screen.getByLabelText("メモ")).toHaveAttribute("rows", "3");
});
