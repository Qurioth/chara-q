# ADR 0007: UI 基盤の詳細（Base UI、テーマ、フォント、web のテスト）

- ステータス: 採用
- 日付: 2026-10-31

## 背景

ADR 0004 で Tailwind CSS（v4）と shadcn/ui を採用し、次の事項は UI 基盤の Spec で決めることにした。

- 導入する Tailwind CSS と shadcn/ui のバージョン、shadcn/ui の土台（Radix UI か Base UI か）
- 最初に入れる部品と、共通の入力部品の形
- フォームのライブラリ
- テーマの CSS 変数とダークモードの切り替え方

また、web には自動テストの仕組みがなかった。

UI 基盤の Spec（`specs/001-ui-foundation/`）と clarify で、次のことが決まっている。

- テーマは、ボタン 1 つでライトとダークを切り替える。一度も切り替えていなければ端末の設定に従う。
- web のテストは部品テストだけにし、本物のブラウザのテストは入れない。

## 決定

### shadcn/ui の土台は Base UI

- **shadcn/ui は Base UI（`@base-ui/react`）を土台にして初期化する。** スタイルは `base-nova`、色は `neutral`。
  - shadcn/ui は 2026 年 7 月に、新しいプロジェクトの既定を Base UI に変えた。既定に合わせると、
    ドキュメントと例をそのまま使える。
  - Base UI は MUI のチームが保守している。ただし使うのはスタイルを持たない部品だけで、
    MUI のスタイル層（Material UI、Emotion）は入れない。constitution の「MUI は使わない」には当たらない。
- shadcn/ui の初期化が入れる `cn` パッケージ（shadcn が公開している、clsx と tailwind-merge の代わり）と
  `shadcn/tailwind.css` をそのまま使う。`shadcn` と `tw-animate-css` は CSS の読み込みにだけ使うため、
  devDependencies に置く。

### 共通の入力部品

- **shadcn/ui の Field・Input・Select・Textarea を包んだ 4 つの部品（`TextField`、`NumberField`、
  `SelectField`、`TextareaField`）を、画面から使う。** 共通の枠 `FormField` が、ラベル・補足説明・エラーと
  入力欄を `for`・`aria-describedby`・`aria-invalid`・`aria-required` で関連付ける。
  shadcn/ui の Field は `aria-describedby` を付けないため、部品ごとに書くと漏れる。
- **数値入力は `type="text"` と `inputMode` で作る。** 全角数字を受け付け、空欄を `null`（未入力）にし、
  既定では 0 以上の整数だけを受け付ける。`type="number"` は全角数字を受け付けず、空欄と不正な値を
  区別できない。
- **フォームのライブラリはまだ入れない。** 部品は `value` と `onValueChange` で値を受け渡す形にしておき、
  送信するフォームが出てくる機能（認証）の Spec で選ぶ。

### テーマ

- **テーマの切り替えは自前で書く（next-themes は使わない）。**
  - 選択は `localStorage` の `chara-q-theme` に `light` か `dark` で保存する。サーバーには送らない。
  - `<head>` のインラインスクリプトが、最初の描画の前に `<html>` に `dark` か `light` のクラスと
    `color-scheme` を付ける。保存値がない間は、端末の設定の変化に追従する。
  - next-themes は 2025 年 3 月から更新がなく、3 択（端末の設定を含む）を前提にしている。
    2 択の切り替えなら数十行で書ける。
  - Cookie でテーマを持ってサーバーで描き分けると、`cacheComponents` のもとでルートレイアウトが
    リクエストごとの描画になり、全画面の静的な配信ができなくなる。
- **色は CSS 変数で 1 か所に定義し、ダークの値は `@variant dark` の中に書く。**
  - Tailwind の `dark` の条件は、「`dark` クラスがある」または「端末の設定がダークで、`light` クラスがない」
    にする。JavaScript が無効なときも、端末の設定に合わせた色で表示される。
  - CSS の `light-dark()` は使わない。Tailwind v4 が使う Lightning CSS は、対象ブラウザに合わせて
    `light-dark()` を `prefers-color-scheme` だけを見る形に変換するため、スクリプトで `color-scheme` を
    変えても色が変わらない。
- **配色は「和紙」にする。** ライトは生成りの地に焦げ茶の文字、ダークは焦げ茶寄りの炭色に生成りの文字で、
  真っ白・真っ黒を使わない。紙のキャラクターシートに近い暖かさを出すため。変数名は shadcn/ui のものを使い、
  値だけを差し替える。本文・補足・エラーの文字は、両テーマで背景とのコントラスト比 4.5:1 以上を保つ。
- 画面は色を意味の名前（`bg-background`、`text-muted-foreground` など）で書き、`dark:` で色を直接書かない。

### フォント

- **本文は端末の日本語ゴシック体を使い、Web フォントを読み込まない。** 日本語の Web フォントは大きく、
  最初の表示が遅くなる。
- **ゲームシステムのフォントは `next/font/google` で `preload: false` にして読み込み、`[data-game-system]`
  の中で役割（`--font-heading`、`--font-brush`）に割り当てる。** 画面は `font-heading` と書くだけで、
  どのシステムかを意識しない。ブラウザは使われるまでフォントをダウンロードしないため、システムの
  領域がない画面は遅くならない。
  - ゆうやけこやけ: 見出しに Mochiy Pop P One、筆文字に Yuji Syuku
  - ビギニングアイドル: 見出しに Mochiy Pop P One
- **アプリ名の表示は `CharaQ` にし、ロゴの書体は Hina Mincho にする。** 細めでやわらかい明朝が、
  和紙の配色と紙のシートの雰囲気に合うため。全画面のヘッダーに出るので、英字の範囲（latin）だけを先読みし、
  Tailwind の `font-logo` で使う。リポジトリ名やパッケージ名（`chara-q`、`@chara-q/web`）は変えない。
- **ロゴは朱文の丸印（判子）にする。** 二重の輪の中に Hina Mincho の Q を彫った形で、色は朱（`--seal`）。
  旧システムのロゴ（輪の Q と立方体のサイコロ）から、手書き風の明朝と和紙の配色に合わせて作り直した。
  - Q は文字ではなくアウトライン（パス）にして埋め込む。閲覧者の環境にフォントがなくても同じ形になる。
    Hina Mincho は SIL Open Font License で、字形をロゴに使うことは許されている。
  - ヘッダーのロゴは `LogoMark`（`src/components/logo/logo-mark.tsx`）。色は `currentColor` で、テーマに合わせて変わる。
  - ファビコン（`src/app/icon.svg`、`favicon.ico`、`apple-icon.png`）は、16px でもつぶれないよう内側の細い輪を省き、
    線を太くした版にする。`icon.svg` は端末の設定がダークなら明るい朱に切り替わる。`apple-icon.png` は、
    iOS が透過部分を黒で埋めるため和紙の地を敷く。PNG と ICO は SVG から書き出したもので、ロゴを変えたら作り直す。

### web のテスト

- **web のテストは Vitest（jsdom）と Testing Library で書く。** キーボード操作は `user-event` で再現し、
  アクセシビリティは `axe-core` で確かめる（色のコントラストは jsdom で判定できないため除く）。
  - Next.js のドキュメントが案内している組み合わせで、API と DB なしで動く。
  - `@/` の別名は Vite 8 の `resolve.tsconfigPaths` で解決する（`vite-tsconfig-paths` は入れない）。
- テストは対象のファイルの隣に `*.test.ts(x)` で置く。`task test` が api と web の両方を実行し、
  `task test:web` で web だけを実行できる。CI（`ci-web.yml`）でも実行する。
- 最初の描画のテーマ、画面幅ごとの配置、色のコントラスト、フォントの見た目は、本物のブラウザで目視する
  （`specs/001-ui-foundation/quickstart.md`）。画面をまたぐ操作が増えたら、E2E テストの導入を検討する。

## 検討した代案

- **Radix UI を土台にする**: 保守は続いているが、shadcn/ui の既定ではなくなった。既存のコードがない今のうちに
  既定に合わせる。
- **next-themes**: 上記のとおり、更新が止まっており、2 択には過剰。
- **`light-dark()` と `color-scheme` だけで色を切り替える**: 上記の Lightning CSS の変換のため、
  スクリプトからの切り替えが効かない。
- **本文に Noto Sans JP の Web フォントを使う**: ダウンロードが数百 KB 増える。
- **Jest**: ESM と TypeScript の設定が Vitest より多い。
- **Playwright などの本物のブラウザでのテスト**: clarify で、この機能では入れないと決めた。

## 結果

- 画面は、共通レイアウトの中に中身だけを書けばよい。入力部品のアクセシビリティの属性は部品がまとめて付ける。
- テーマの切り替えは依存を増やさずに済み、全画面を静的に配信できる。
- shadcn/ui の部品は `src/components/ui/` にコピーされ、自分のコードとして変更できる。元のリポジトリで
  直った不具合は自動では入らない。
- Lightning CSS の変換に左右されないよう、ダークの色は `@variant dark` で書く必要がある。shadcn/ui から
  新しく取り込んだ部品の色の定義（`.dark { … }`）は、この形に書き直す。
