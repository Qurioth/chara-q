# Research: UI 基盤

ADR 0004 で「UI 基盤の Spec で決める」とした項目と、spec の要件を満たすための技術の選定を記録する。
バージョンは 2026-10-31 時点の npm の最新版で、`package.json` には厳密なバージョンで書く。

## R1. Tailwind CSS のバージョンと導入方法

- **Decision**: Tailwind CSS 4.3.3 と `@tailwindcss/postcss` 4.3.3。`postcss.config.mjs` にプラグインを書き、
  `globals.css` の先頭で `@import "tailwindcss";` する。
- **Rationale**: ADR 0004 で v4 に決めている。同梱の Next.js 16.4 のドキュメント
  （`node_modules/next/dist/docs/01-app/01-getting-started/11-css.md`）の手順がこの形である。v4 は設定を
  CSS に書くため、`tailwind.config` が要らない。
- **Alternatives considered**: Tailwind CSS v3（`tailwind-v3-css.md` の手順）は、shadcn/ui の現行の部品が
  v4 前提のため採らない。

## R2. shadcn/ui の土台（Radix UI か Base UI か）

- **Decision**: Base UI（`@base-ui/react` 1.9.0）。shadcn CLI 4.23.0 の `init` を `apps/web` で実行し、
  既定（Base UI）のまま初期化する。スタイルは初期化の既定のものを使い、`components.json` に記録する。
- **Rationale**:
  - shadcn/ui は 2026 年 7 月に新規プロジェクトの既定を Base UI に変えた。部品の更新と新しい部品は両方に
    出るが、Base UI にしかない部品もある。これから部品を足していくこのプロジェクトでは、既定に乗る方が
    ドキュメントと例をそのまま使える。
  - Base UI は 1.x で安定しており、MUI のチームが保守している（MUI のスタイル層は含まないため、
    constitution の「MUI は使わない」には当たらない。使うのはスタイルを持たない部品だけ）。
  - ADR 0004 にある「大きく表示するセレクト」のように、部品の見た目を画面側から変えたい。どちらの
    土台でも部品のソースは手元にコピーされるため、差はない。
- **Alternatives considered**: Radix UI（`radix-ui` 1.8.1）。保守は続いており、既存の情報も多い。ただし
  既定でなくなったため、ドキュメントの例が Base UI 側で先に出る。今は既存のコードがないので、
  移行の手間がかからない今のうちに既定へ合わせる。
- **実装時の記録**: `init -b base -p nova` で初期化し、スタイルは `base-nova`、色は `neutral` になった。
  初期化は `clsx`・`tailwind-merge` ではなく、shadcn が公開している `cn` パッケージ（0.5.0）と、CSS 用の
  `shadcn/tailwind.css` を入れる。CSS の読み込みにだけ使う `shadcn` と `tw-animate-css` は devDependencies に移した。

## R3. 共通入力部品の形

- **Decision**: shadcn/ui の `field`・`input`・`select`・`textarea`・`label`・`button` を追加する。これを
  包んで、このアプリ用の 4 つの部品（`TextField`、`NumberField`、`SelectField`、`TextareaField`）を
  `src/components/form/` に作る。共通の枠 `FormField` が、`id` の採番、`FieldLabel` の `htmlFor`、
  補足説明とエラーの `id`、入力欄の `aria-describedby`・`aria-invalid`・`aria-required` をまとめて付ける。
  詳細は [contracts/ui-components.md](contracts/ui-components.md)。
- **Rationale**:
  - shadcn/ui の Field は `aria-describedby` を自動では付けない（ドキュメントの例にもない）。FR-025 の
    「補足説明とエラーを入力欄と関連付けて読み上げる」を部品ごとに書くと漏れるため、枠でまとめる。
  - 数値入力は、全角数字の変換、空欄を `null` で返すこと、0 以上の整数を既定にすること（clarify の
    回答）など、このアプリ固有の決まりが多い。Base UI の NumberField（増減ボタン付き）は使わず、
    `Input` を `type="text"`・`inputMode="numeric"`（小数を許すときは `"decimal"`）で包む。
    `type="number"` は全角数字を受け付けず、ブラウザごとに挙動が違い、空欄と不正な値を区別できないため使わない。
  - シートの形に合わせて大きさを変えられるよう（FR-031）、4 つの部品は外枠と入力欄それぞれに
    `className` を受け取る。
- **Alternatives considered**: shadcn/ui の部品を画面で直接使う（ラベルとエラーの関連付けを毎回書く
  ことになり、漏れる）。Base UI の NumberField（全角数字や空欄の扱いを合わせるのに手間がかかる。
  シートでは増減ボタンを出さない形が多い）。

## R4. フォームのライブラリ

- **Decision**: この機能では入れない。4 つの部品は `value` と `onValueChange` で値を受け渡す制御
  コンポーネントにし、`name`・`onBlur`・`ref` も受け取れるようにして、後からフォームのライブラリと
  つなげられる形にしておく。
- **Rationale**: この機能には送信するフォームがなく、検証も部品単位（数値の範囲など）だけである。
  フォーム全体の検証と送信は、認証とシートの機能で扱う（spec の Assumptions）。必要になった機能の
  Spec で、react-hook-form（7.91.0）などを比較して決める。
- **Alternatives considered**: 今のうちに react-hook-form を入れる（使う画面がまだなく、部品の形を
  ライブラリに合わせて決めると、後で別のものを選びにくくなる）。

## R5. テーマの切り替え方

- **Decision**: 自前で実装する（依存を足さない）。
  - `localStorage` のキー `chara-q-theme` に `"light"` か `"dark"` を保存する。一度も切り替えていなければ
    値はない。
  - ルートレイアウトの `<head>` に、最初の描画の前に実行するインラインスクリプトを置く。保存値があれば
    それを、なければ `prefers-color-scheme` を使い、`<html>` の `dark` クラスと `color-scheme` を設定する。
    保存値がないときは `matchMedia` の `change` を購読し、端末の設定に追従する。
  - トグルボタン（クライアントコンポーネント）は、押すと保存値を書き、`dark` クラスを付け替える。
    表示中のテーマは `<html>` のクラスから `useSyncExternalStore` で読む（サーバーでは不明として描画し、
    ハイドレーション後に確定する。アイコンは CSS の `dark:` で出し分けるため、ちらつかない）。
  - Tailwind v4 の `dark:` は `@custom-variant dark (&:where(.dark, .dark *));` でクラスに合わせる
    （shadcn/ui の初期化が追加する）。
  - 詳細は [contracts/theme-and-fonts.md](contracts/theme-and-fonts.md)。
- **Rationale**:
  - clarify でテーマは 2 択（端末の設定に戻す手段なし）に決まった。必要な処理は「保存値か OS の設定かを
    決めて html のクラスを付ける」だけで、数十行で書ける。
  - shadcn/ui のドキュメントが勧める next-themes は、最新版（0.4.6）が 2025 年 3 月で止まっている。
    3 択（system を含む）を前提とした API で、2 択にすると使わない部分が多い。
  - `cacheComponents` が有効なので、Cookie でテーマを持ってサーバー側で描き分けると、ルートレイアウトが
    リクエストごとの描画になり、全画面の静的な配信ができなくなる。`localStorage` とインラインスクリプトなら
    静的なまま最初の描画に間に合う（FR-014）。
  - 保存先の読み書きは `try` で包み、使えないときは端末の設定に合わせる（spec の Edge Cases）。
  - スクリプトは `<html>` に `dark` か `light` のクラスを付け、`style.colorScheme` も固定する。
    色の切り替えはクラスで行う（R6）。
- **Alternatives considered**: next-themes（上記）。Cookie とサーバー側の描き分け（上記の
  `cacheComponents` の理由）。CSS の `prefers-color-scheme` だけ（切り替えができない）。

## R6. 色の定義

- **Decision**: shadcn/ui の初期化が `globals.css` に書く色の CSS 変数（`--background`、`--foreground`、
  `--primary`、`--muted`、`--accent`、`--destructive`、`--border`、`--input`、`--ring` など）の名前と値を
  土台にし、`@theme inline` で Tailwind の色名（`bg-background` など）に対応させる。部品はこの名前だけを
  使い、`gray-900` のような直接の色を使わない（FR-017）。
  - 初期化が書く `:root` と `.dark` の 2 組は、`:root` のライトの値と、その中の `@variant dark { … }` の
    ダークの値に書き直す。Tailwind の `dark` の条件は「`dark` クラスがある」または「端末の設定がダークで、
    `light` クラスがない」にする（`@custom-variant dark` のブロック形式）。JavaScript が無効なときも端末の
    設定に合わせた色になり（spec の Edge Cases）、色の定義が 1 か所で済む。
  - **実装時の変更**: 当初は CSS の `light-dark()` と `color-scheme` で切り替える計画だった。Tailwind v4 が
    使う Lightning CSS は、対象ブラウザ（Chrome 111、Safari 16.4 など）に合わせて `light-dark()` を
    `prefers-color-scheme` のメディアクエリだけを見る変数に変換するため、スクリプトで `color-scheme` を
    変えても色が変わらなかった。そのため上のクラスの方式に変えた。
  コントラストは部品確認ページを本物のブラウザで開き、開発者ツールのコントラストの表示で確かめる
  （clarify の回答: 部品テストだけ）。
  - **実装後の変更**: 色の値は、初期化の無彩色（白と黒）から「和紙」の配色に差し替えた（ユーザーの選択）。
    ライトは生成り `oklch(0.975 0.01 85)` の地に焦げ茶 `oklch(0.24 0.02 55)`、ダークは `oklch(0.2 0.012 60)` の地に
    生成り `oklch(0.93 0.015 85)`。ブラウザで測ったコントラスト比の最小は、ライトの補足の文字と `muted` の背景の 4.99:1。
- **Rationale**: shadcn/ui の部品がこの変数名を前提にしている。後でゲームシステムごとに色を変えるときは、
  `[data-game-system=…]` の中で同じ変数を上書きすればよい。
- **Alternatives considered**: 独自の変数名を作る（shadcn/ui の部品をすべて書き換えることになる）。

## R7. フォント

- **Decision**:
  - アプリ共通のフォントは、端末の日本語ゴシック体を使う（Web フォントを読み込まない）。
    `--font-sans` を `system-ui, -apple-system, "Hiragino Sans", "Hiragino Kaku Gothic ProN", "Yu Gothic UI",
    "Meiryo", "Noto Sans JP", sans-serif` にする。既存の Geist（ラテン文字だけ）は外す。
  - システム用のフォントは `next/font/google` で `Mochiy_Pop_P_One`（weight 400）と `Yuji_Syuku`（weight 400）を
    定義する。`preload: false`、`display: "swap"`、`variable`（`--font-mochiy-pop-p-one`、`--font-yuji-syuku`）、
    代替フォントを指定する。変数のクラスは `<html>` に付ける。
  - 役割ごとの CSS 変数 `--font-heading` と `--font-brush` を用意し、通常は `--font-sans` を指す。
    `[data-game-system="yuyakekoyake"]` の中では見出しを Mochiy Pop P One、筆文字を Yuji Syuku に、
    `[data-game-system="beginning-idol"]` の中では見出しを Mochiy Pop P One に差し替える。
    Tailwind の `font-heading`・`font-brush` で使う。
- **Rationale**:
  - 日本語の Web フォントは大きい。本文に使うと最初の表示が遅くなる（SC-006）。端末のゴシック体なら
    ダウンロードがなく、どの環境でも読みやすい（FR-018）。
  - ブラウザは `@font-face` を宣言しただけではフォントをダウンロードせず、その書体で描く文字が出たときに
    取りに行く。`preload: false` にしておけば、システムの領域を表示しない画面ではダウンロードが起きない
    （FR-022）。Google Fonts の日本語フォントは文字の範囲ごとに分割されているため、表示した文字の分だけを読む。
  - `display: "swap"` で、読み込み中や失敗時は代替フォントで文字を表示する（FR-021、SC-007）。
  - 役割の変数にしておけば、シートの画面は `font-heading` と書くだけで、どのシステムかを意識しない（FR-019）。
- **Alternatives considered**: 本文に Noto Sans JP の Web フォントを使う（数百 KB のダウンロードが増える）。
  システムごとに `className` を直接付ける（旧システムの方式。シート側がフォント名を知る必要があり、
  システムの追加時に変更箇所が散らばる）。

## R8. web のテスト基盤

- **Decision**: Vitest 5.0.5、`@vitejs/plugin-react` 6.1.2、jsdom 30.1.3、`@testing-library/react` 16.3.3、`@testing-library/dom` 10.4.2、
  `@testing-library/user-event` 14.6.8、`@testing-library/jest-dom` 7.0.1。アクセシビリティの自動チェックは
  `axe-core` 4.14.0 を直接呼ぶ補助関数（`src/test/axe.ts`）で行い、jsdom では判定できない
  `color-contrast` の規則は外す。
  - `package.json` に `"test": "vitest run"` と `"test:watch": "vitest"` を足す。
  - Taskfile に `test:web`（`pnpm --filter @chara-q/web test`）を足し、`test` から `test:api` と並べて呼ぶ。
  - `ci-web.yml` の lint・型チェックの後に、テストのステップを足す。
  - `@/` の別名は、Vitest 5 が使う Vite 8 の `resolve.tsconfigPaths: true` で解決する。
    **実装時の変更**: 当初は Next.js のドキュメントにある `vite-tsconfig-paths` を入れる計画だったが、Vite 8 が
    同じ機能を持ち、`vite-tsconfig-paths` は非推奨の依存（`tsconfck`）を含むため入れなかった。
  - `next/font/google` はビルド時に Next.js が変換するため、`vitest.setup.ts` で偽物に差し替える。
- **Rationale**:
  - 同梱の Next.js 16.4 のドキュメント（`02-guides/testing/vitest.md`）がこの組み合わせを案内している。
    Vite を使うため TypeScript と JSX の変換の設定がほとんど要らず、jsdom で API と DB なしに動く（FR-034）。
  - `user-event` は、キーボードとクリックを実際の操作に近い順序のイベントで再現する（FR-036）。
  - 非同期の Server Component は Vitest で描画できない。この機能のページとレイアウトは同期の
    コンポーネントにするため、問題にならない。
  - jsdom は配置と色を計算しないため、最初の描画のテーマ・画面幅・コントラスト・フォントの見た目は
    目視で確かめる（clarify の回答）。手順は [quickstart.md](quickstart.md)。
  - `vitest-axe` は 2025 年 1 月から更新がなく、`jest-axe` は Jest 用のため、`axe-core` を直接使う。
- **Alternatives considered**: Jest（ESM と TypeScript の設定が増える。Next.js の SWC 変換を使う設定が
  あるが、Vitest の方が設定が少ない）。happy-dom（速いが、Base UI の部品が使う DOM の API の再現が
  jsdom より少なく、差分の調査に時間を取られやすい）。Playwright などの本物のブラウザでのテスト
  （clarify で今回は入れないと決めた）。

## R9. 部品確認ページを本番で公開しない方法

- **Decision**: `src/app/dev/ui/page.tsx` の先頭で、`process.env.NODE_ENV === "production"` なら `notFound()` を
  呼ぶ。
- **Rationale**: `NODE_ENV` はビルド時に定数として埋め込まれる。`next build` の結果ではこのページは常に 404 に
  なり、ページの中身と API の接続先は本番の HTML に含まれない。設定ファイルやミドルウェアを増やさずに済む。
- **Alternatives considered**: 環境変数のフラグで切り替える（設定を間違えると本番で公開される）。ルートを
  ビルドから外す（Next.js に公式の方法がない）。

## R10. 判断を記録する ADR

- **Decision**: ADR 0007「UI 基盤の詳細（Base UI、テーマ、フォント、web のテスト）」を追加する。
  ADR 0004 の「次は UI 基盤の Spec で決める」に、ADR 0007 へのリンクを足す。
- **Rationale**: constitution の「新しい基盤技術の追加には ADR を必須とする」に従う。Vitest と Base UI は
  新しい基盤技術に当たる。
