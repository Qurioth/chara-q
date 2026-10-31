import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { FormField } from "./form-field";

function renderField(props: Partial<React.ComponentProps<typeof FormField>> = {}) {
  return render(
    <FormField label="名前" {...props}>
      {(controlProps) => <input {...controlProps} />}
    </FormField>,
  );
}

describe("FormField", () => {
  it("ラベルと入力欄を関連付ける", () => {
    renderField();
    expect(screen.getByLabelText("名前")).toBeInstanceOf(HTMLInputElement);
  });

  it("渡した id を入力欄に使う", () => {
    renderField({ id: "character-name" });
    expect(screen.getByLabelText("名前")).toHaveAttribute("id", "character-name");
  });

  it("補足説明とエラーを、この順で aria-describedby に入れる", () => {
    renderField({ description: "全角 20 文字まで", error: "入力してください" });
    const input = screen.getByLabelText("名前");

    const ids = input.getAttribute("aria-describedby")?.split(" ") ?? [];
    expect(ids).toHaveLength(2);
    expect(document.getElementById(ids[0])).toHaveTextContent("全角 20 文字まで");
    expect(document.getElementById(ids[1])).toHaveTextContent("入力してください");
    expect(input).toHaveAccessibleDescription("全角 20 文字まで 入力してください");
  });

  it("エラーがあると aria-invalid を付け、エラーを alert で示す", () => {
    renderField({ error: "入力してください" });
    expect(screen.getByLabelText("名前")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("alert")).toHaveTextContent("入力してください");
  });

  it("エラーがなければ aria-invalid と aria-describedby を付けない", () => {
    renderField();
    const input = screen.getByLabelText("名前");
    expect(input).not.toHaveAttribute("aria-invalid");
    expect(input).not.toHaveAttribute("aria-describedby");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("必須ならラベルに「必須」を付け、aria-required を付ける", () => {
    renderField({ required: true });
    const input = screen.getByLabelText(/名前/);
    expect(input).toHaveAttribute("aria-required", "true");
    expect(screen.getByText("必須")).toBeInTheDocument();
  });

  it("ラベルを隠しても、ラベル名で入力欄を取れる", () => {
    renderField({ hideLabel: true });
    expect(screen.getByLabelText("名前")).toBeInTheDocument();
    expect(screen.getByText("名前").closest("label")).toHaveClass("sr-only");
  });

  it("無効なら入力欄を disabled にする", () => {
    renderField({ disabled: true });
    expect(screen.getByLabelText("名前")).toBeDisabled();
  });
});
