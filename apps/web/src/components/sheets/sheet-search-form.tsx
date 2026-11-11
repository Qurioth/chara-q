"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import { SelectField } from "@/components/form/select-field";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { gameSystems } from "@/lib/game-systems";
import {
  hasSheetConditions,
  sheetSearchHref,
  sheetSortOptions,
  type SheetSearch,
  type SheetSort,
} from "@/lib/sheets/search-params";
import { isGameSystemId } from "@/lib/sheets/systems";

/** システムのプルダウンの「すべて」の値。URL には含めない */
const ALL_SYSTEMS = "all";

const systemOptions = [
  { value: ALL_SYSTEMS, label: "すべて" },
  ...gameSystems.map((s) => ({ value: s.id, label: s.name })),
];

type SheetSearchFormProps = {
  search: SheetSearch;
  /** ログインしているときだけ「自分のシートだけ」を出す */
  loggedIn: boolean;
};

/**
 * 一覧の検索のフォーム（contracts/web-routes.md の「一覧の画面」）。
 * 送信すると、既定の値を除いた URL（sheetSearchHref）へ移動する。ページは含めないため、最初のページに戻る。
 * JavaScript が動かないときも、GET のフォームとして /sheets に送られる（空の値は既定として読まれる）。
 */
export function SheetSearchForm({ search, loggedIn }: SheetSearchFormProps) {
  const router = useRouter();
  const [q, setQ] = useState(search.q);
  const [system, setSystem] = useState<string>(search.system ?? ALL_SYSTEMS);
  const [sort, setSort] = useState<SheetSort>(search.sort);
  const [mine, setMine] = useState(search.mine);

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    router.push(
      sheetSearchHref({
        q,
        system: isGameSystemId(system) ? system : null,
        sort,
        mine: loggedIn && mine,
      }),
    );
  }

  return (
    <form
      method="get"
      action="/sheets"
      role="search"
      aria-label="シートの検索"
      onSubmit={onSubmit}
      className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4 text-card-foreground"
    >
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <TextField
          label="キーワード"
          description="キャラクター名・プレイヤー名"
          name="q"
          type="search"
          value={q}
          onValueChange={setQ}
        />
        <SelectField
          label="システム"
          name="system"
          value={system}
          onValueChange={(v) => setSystem(v ?? ALL_SYSTEMS)}
          options={systemOptions}
        />
        <SelectField
          label="並び順"
          name="sort"
          value={sort}
          onValueChange={(v) => setSort(sheetSortOptions.find((o) => o.value === v)?.value ?? "updated")}
          options={sheetSortOptions}
        />
      </div>
      <div className="flex flex-wrap items-center justify-between gap-4">
        {loggedIn ? (
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input
              type="checkbox"
              name="mine"
              value="1"
              checked={mine}
              onChange={(e) => setMine(e.target.checked)}
              className="size-4 accent-primary outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
            />
            自分のシートだけ
          </label>
        ) : (
          <span />
        )}
        <div className="flex items-center gap-4">
          {hasSheetConditions(search) && (
            <Link href="/sheets" className="text-sm underline underline-offset-4">
              条件を外す
            </Link>
          )}
          <Button type="submit">検索</Button>
        </div>
      </div>
    </form>
  );
}
