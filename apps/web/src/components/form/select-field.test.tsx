import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import { SelectField } from "./select-field";

const options = [
  { value: "henge", label: "へんげ" },
  { value: "kitsune", label: "きつね" },
  { value: "tanuki", label: "たぬき", disabled: true },
];

function Controlled(props: Partial<React.ComponentProps<typeof SelectField>>) {
  const [value, setValue] = useState<string | null>(null);
  return (
    <SelectField
      label="種族"
      options={options}
      {...props}
      value={value}
      onValueChange={(v) => {
        setValue(v);
        props.onValueChange?.(v);
      }}
    />
  );
}

const trigger = () => screen.getByRole("combobox", { name: "種族" });

it("未選択ではプレースホルダーを表示し、ラベル名でトリガーを取れる", async () => {
  const { container } = render(<Controlled />);
  expect(trigger()).toHaveTextContent("選択してください");
  await expectNoAxeViolations(container);
});

it("キーボードで開いて選ぶと値が変わり、フォーカスがトリガーに戻る", async () => {
  const user = userEvent.setup();
  const onValueChange = vi.fn();
  render(<Controlled onValueChange={onValueChange} />);

  await user.tab();
  expect(trigger()).toHaveFocus();

  await user.keyboard("{Enter}");
  expect(await screen.findByRole("listbox")).toBeInTheDocument();

  await user.keyboard("{ArrowDown}");
  await user.keyboard("{Enter}");

  const selected = onValueChange.mock.lastCall?.[0];
  const option = options.find((o) => o.value === selected);
  expect(option).toBeDefined();
  expect(trigger()).toHaveTextContent(option!.label);
  expect(trigger()).toHaveFocus();
});

it("Escape で閉じる", async () => {
  const user = userEvent.setup();
  render(<Controlled />);

  await user.click(trigger());
  expect(await screen.findByRole("listbox")).toBeInTheDocument();
  await user.keyboard("{Escape}");
  expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
});

it("無効な選択肢は選べない", async () => {
  const user = userEvent.setup();
  const onValueChange = vi.fn();
  render(<Controlled onValueChange={onValueChange} />);

  await user.click(trigger());
  const disabled = await screen.findByRole("option", { name: "たぬき" });
  expect(disabled).toHaveAttribute("aria-disabled", "true");
  await user.click(disabled);
  expect(onValueChange).not.toHaveBeenCalledWith("tanuki");
});

it("エラーを示す", () => {
  render(<Controlled error="選んでください" />);
  expect(trigger()).toHaveAttribute("aria-invalid", "true");
  expect(screen.getByRole("alert")).toHaveTextContent("選んでください");
});
