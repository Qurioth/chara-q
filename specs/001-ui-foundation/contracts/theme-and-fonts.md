# Contract: テーマとフォント

後の機能の画面と、ゲームシステムごとの見た目が頼る CSS とブラウザ上の取り決め。

## テーマ

### 保存

| 項目 | 値 |
|---|---|
| 保存先 | `localStorage` |
| キー | `chara-q-theme` |
| 値 | `"light"` \| `"dark"`（一度も切り替えていなければキーがない） |

### `<html>` に付けるもの

| 表示中のテーマ | `class` | `style.colorScheme` |
|---|---|---|
| ライト | `light`（`dark` は付けない） | `light` |
| ダーク | `dark`（`light` は付けない） | `dark` |
| （JavaScript が無効） | どちらもなし | なし（CSS の既定 `light dark` のまま） |

- `<head>` のインラインスクリプトが、本文の描画より前に上の状態にする。React のハイドレーションと食い違う
  ため、`<html>` に `suppressHydrationWarning` を付ける。
- 保存値がないときだけ、`matchMedia("(prefers-color-scheme: dark)")` の `change` に追従する。保存値が
  できた後は追従しない。

### 色の決まり方

- Tailwind の `dark` の条件（`globals.css` の `@custom-variant dark`）は、「`dark` クラスがある」または
  「端末の設定がダークで、`light` クラスがない」。スクリプトが動かないとき（JavaScript が無効）は、
  端末の設定に合わせた色になる（spec の Edge Cases）。
- 色の変数は `:root` にライトの値を書き、同じ `:root` の中の `@variant dark { … }` にダークの値を書く。
  shadcn/ui の初期化は `:root` と `.dark` の 2 組で色を書くため、新しく取り込んだ色はこの形に書き直す。
- `:root { color-scheme: light dark; }` を既定にし、スクリプトが `style.colorScheme` を固定する
  （スクロールバーやフォーム部品など、ブラウザが描く部分の色のため）。
- CSS の `light-dark()` は使わない。Tailwind v4 が使う Lightning CSS が `prefers-color-scheme` だけを見る形に
  変換するため、スクリプトで `color-scheme` を変えても色が変わらない（実装中に確認した）。

### 画面側の書き方

- 色は Tailwind の意味の名前（`bg-background`、`text-foreground`、`text-muted-foreground`、`border-border`、
  `bg-primary`、`text-destructive`、`ring-ring` など）で書く。`dark:` で色を直接指定しない。
- 色の名前と値は `globals.css` の `:root` と、その中の `@variant dark` で定義する（shadcn/ui の初期化で入る変数名）。

## フォント

### CSS 変数

| 変数 | 既定 | 説明 |
|---|---|---|
| `--font-sans` | 端末の日本語ゴシック体の並び | 本文。`<body>` に適用 |
| `--font-heading` | `var(--font-sans)` | 見出しの役割 |
| `--font-brush` | `var(--font-sans)` | 筆文字の役割 |
| `--font-mochiy-pop-p-one` | `next/font` が定義 | 直接は使わない |
| `--font-yuji-syuku` | `next/font` が定義 | 直接は使わない |

Tailwind では `font-sans`・`font-heading`・`font-brush` として使う（`@theme inline` で対応させる）。

### システムごとの差し替え

```css
[data-game-system="yuyakekoyake"] {
  --font-heading: var(--font-mochiy-pop-p-one), var(--font-sans);
  --font-brush: var(--font-yuji-syuku), var(--font-sans);
}
[data-game-system="beginning-idol"] {
  --font-heading: var(--font-mochiy-pop-p-one), var(--font-sans);
}
```

- `next/font/google` の定義は `preload: false`、`display: "swap"`。システムの領域を表示しない画面では
  フォントをダウンロードしない。
- 代替フォントは `--font-sans` の並び。Web フォントにない文字も代替フォントで表示される（Edge Cases）。
- `src/lib/game-systems.ts` の `fonts` と、この CSS の対応が一致することをテストで確かめる。

### システムの色（後の機能）

この機能ではシステムごとの色を変えない。変えるときは、同じ `[data-game-system=…]` の中で色の変数
（`--primary` など）を上書きする。
