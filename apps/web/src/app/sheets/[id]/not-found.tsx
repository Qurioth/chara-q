import Link from "next/link";

// 存在しない・削除された・形の正しくない ID のシートで、同じように表示する。
// 削除されたことがわかる文言は書かない（specs/003-character-sheet-base の FR-018）。
export default function SheetNotFound() {
  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-4 px-4 py-8">
      <h1 className="text-2xl font-bold">シートが見つかりません</h1>
      <p className="text-muted-foreground">URL が正しいか確かめてください。</p>
      <p>
        <Link href="/sheets" className="underline underline-offset-4">
          キャラクターシート一覧へ
        </Link>
      </p>
    </div>
  );
}
