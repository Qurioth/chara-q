/** キーボードの利用者がヘッダーを飛ばして本文へ移るためのリンク。フォーカスしたときだけ見える */
export function SkipLink() {
  return (
    <a
      href="#main-content"
      className="sr-only rounded-md bg-background px-3 py-2 text-sm font-medium text-foreground ring-2 ring-ring focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-50"
    >
      本文へスキップ
    </a>
  );
}
