import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { expectNoAxeViolations } from "@/test/axe";

import { NumberField } from "./number-field";

type Props = Partial<React.ComponentProps<typeof NumberField>> & { initial?: number | null };

function Controlled({ initial = null, ...props }: Props) {
  const [value, setValue] = useState<number | null>(initial);
  return (
    <>
      <NumberField
        label="年齢"
        {...props}
        value={value}
        onValueChange={(v) => {
          setValue(v);
          props.onValueChange?.(v);
        }}
      />
      <button type="button" onClick={() => setValue(42)}>
        外から変える
      </button>
      <output data-testid="value">{String(value)}</output>
    </>
  );
}

const input = () => screen.getByLabelText("年齢");
const value = () => screen.getByTestId("value").textContent;

describe("NumberField", () => {
  it("数字を入力すると数値を渡す", async () => {
    const user = userEvent.setup();
    const { container } = render(<Controlled />);

    await user.type(input(), "12");
    expect(input()).toHaveValue("12");
    expect(value()).toBe("12");
    await expectNoAxeViolations(container);
  });

  it("数字以外の文字は受け付けず、直前のまま", async () => {
    const user = userEvent.setup();
    render(<Controlled initial={12} />);

    await user.type(input(), "a");
    expect(input()).toHaveValue("12");

    await user.click(input());
    await user.paste("abc");
    expect(input()).toHaveValue("12");
    expect(value()).toBe("12");
  });

  it("全角数字は半角の数値として受け付ける", async () => {
    const user = userEvent.setup();
    render(<Controlled />);

    await user.type(input(), "１２");
    expect(input()).toHaveValue("12");
    expect(value()).toBe("12");
  });

  it("空欄は 0 ではなく未入力（null）", async () => {
    const user = userEvent.setup();
    render(<Controlled initial={5} />);

    await user.clear(input());
    expect(value()).toBe("null");
  });

  it("既定では負の数と小数は、入力欄を離れたときにエラー", async () => {
    const user = userEvent.setup();
    render(<Controlled />);

    await user.type(input(), "-3");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    await user.tab();
    expect(screen.getByRole("alert")).toHaveTextContent("0 以上の数を入力してください");
    expect(input()).toHaveAttribute("aria-invalid", "true");
    // 値は捨てずに渡す
    expect(value()).toBe("-3");

    await user.clear(input());
    await user.type(input(), "1.5");
    await user.tab();
    expect(screen.getByRole("alert")).toHaveTextContent("整数を入力してください");
  });

  it("許可すれば負の数と小数を受け付ける", async () => {
    const user = userEvent.setup();
    render(<Controlled allowNegative allowDecimal />);

    await user.type(input(), "-1.5");
    await user.tab();
    expect(value()).toBe("-1.5");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("範囲外は入力欄を離れたときにエラー、直すと消える", async () => {
    const user = userEvent.setup();
    render(<Controlled min={1} max={6} />);

    await user.type(input(), "7");
    await user.tab();
    expect(screen.getByRole("alert")).toHaveTextContent("1 以上 6 以下で入力してください");

    await user.clear(input());
    await user.type(input(), "6");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("画面側のエラーを、部品のエラーより優先する", async () => {
    const user = userEvent.setup();
    render(<Controlled min={1} error="サーバーのエラー" />);

    await user.type(input(), "0");
    await user.tab();
    expect(screen.getByRole("alert")).toHaveTextContent("サーバーのエラー");
  });

  it("inputMode は整数なら numeric、小数を許すと decimal", () => {
    const { unmount } = render(<Controlled />);
    expect(input()).toHaveAttribute("inputmode", "numeric");
    unmount();
    render(<Controlled allowDecimal />);
    expect(input()).toHaveAttribute("inputmode", "decimal");
  });

  it("外から値が変わると入力欄も変わる", async () => {
    const user = userEvent.setup();
    render(<Controlled initial={1} />);

    await user.click(screen.getByRole("button", { name: "外から変える" }));
    expect(input()).toHaveValue("42");
  });

  it("入力途中の形は値を変えず、離れると値の表示に戻す", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    render(<Controlled initial={3} allowDecimal onValueChange={onValueChange} />);

    await user.type(input(), ".");
    expect(input()).toHaveValue("3.");
    expect(onValueChange).not.toHaveBeenCalled();
    await user.tab();
    expect(input()).toHaveValue("3");
  });
});
