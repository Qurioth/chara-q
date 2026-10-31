"use client";

import { useState, type ReactNode } from "react";

import { NumberField } from "@/components/form/number-field";
import { SelectField } from "@/components/form/select-field";
import { TextField } from "@/components/form/text-field";
import { TextareaField } from "@/components/form/textarea-field";

const shortOptions = [
  { value: "henge", label: "へんげ" },
  { value: "kitsune", label: "きつね" },
  { value: "tanuki", label: "たぬき（選べない）", disabled: true },
];

const manyOptions = Array.from({ length: 30 }, (_, i) => ({
  value: `option-${i + 1}`,
  label: `選択肢 ${i + 1}`,
}));

const longOptions = [
  {
    value: "long",
    label: "とても長い選択肢の文言。入力欄の幅を超えるときは、省略記号で 1 行に収める",
  },
  { value: "short", label: "短い選択肢" },
];

function Group({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-4">
      <h3 className="font-bold">{title}</h3>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">{children}</div>
    </div>
  );
}

function useText(initial = "") {
  const [value, setValue] = useState(initial);
  return { value, onValueChange: setValue };
}

function useNumber(initial: number | null = null) {
  const [value, setValue] = useState<number | null>(initial);
  return { value, onValueChange: setValue };
}

function useSelect(initial: string | null = null) {
  const [value, setValue] = useState<string | null>(initial);
  return { value, onValueChange: setValue };
}

export function FormPreview() {
  return (
    <div className="flex flex-col gap-8">
      <Group title="テキスト">
        <TextField label="通常" {...useText()} />
        <TextField label="補足説明あり" description="全角 20 文字まで" {...useText()} />
        <TextField label="エラー" error="入力してください" {...useText()} />
        <TextField label="無効" disabled {...useText("変更できない値")} />
        <TextField label="必須" required {...useText()} />
        <TextField
          label="ラベルを隠す"
          hideLabel
          placeholder="ラベルは読み上げだけ"
          {...useText()}
        />
      </Group>

      <Group title="数値">
        <NumberField label="既定（0 以上の整数）" {...useNumber()} />
        <NumberField label="1〜6" min={1} max={6} description="1 以上 6 以下" {...useNumber(3)} />
        <NumberField label="負の数を許す" allowNegative {...useNumber()} />
        <NumberField label="小数を許す" allowDecimal {...useNumber()} />
        <NumberField label="エラー" error="画面側のエラー" {...useNumber(10)} />
        <NumberField label="無効" disabled {...useNumber(5)} />
        <NumberField label="必須" required {...useNumber()} />
        <NumberField
          label="大きな表示"
          inputClassName="h-14 text-center text-3xl md:text-3xl"
          {...useNumber(12)}
        />
      </Group>

      <Group title="セレクト">
        <SelectField label="通常" options={shortOptions} {...useSelect()} />
        <SelectField label="選択肢 30 件" options={manyOptions} {...useSelect()} />
        <SelectField label="長い文言" options={longOptions} {...useSelect("long")} />
        <SelectField label="エラー" error="選んでください" options={shortOptions} {...useSelect()} />
        <SelectField label="無効" disabled options={shortOptions} {...useSelect("henge")} />
        <SelectField label="必須" required options={shortOptions} {...useSelect()} />
        <SelectField label="ラベルを隠す" hideLabel options={shortOptions} {...useSelect()} />
      </Group>

      <Group title="テキストエリア">
        <TextareaField label="通常" {...useText()} />
        <TextareaField label="補足説明あり" description="改行できます" rows={5} {...useText()} />
        <TextareaField label="エラー" error="入力してください" {...useText()} />
        <TextareaField label="無効" disabled {...useText("変更できない\n複数行の値")} />
        <TextareaField label="必須" required {...useText()} />
      </Group>
    </div>
  );
}
