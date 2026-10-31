---

description: "Task list for UI 基盤"
---

# Tasks: UI 基盤

**Input**: Design documents from `specs/001-ui-foundation/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/](contracts/), [quickstart.md](quickstart.md)

**Tests**: spec の FR-037 が、入力部品・テーマ切り替え・共通レイアウトの自動テストを求めているため、各ストーリーにテストのタスクを含める。テストは部品テスト（Vitest + jsdom + Testing Library）だけで、本物のブラウザのテストは書かない（clarify の回答）。jsdom で確かめられない項目は Polish の目視確認で扱う。

**Organization**: ユーザーストーリーごとに分ける。ストーリーの番号は spec.md のもの（US1〜US5）で、フェーズは優先度の順（P1: US1・US2 → P2: US3・US5 → P3: US4）。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 並行して進められる（別のファイルで、未完了のタスクに依存しない）
- **[Story]**: どのユーザーストーリーのタスクか（US1〜US5）
- パスはリポジトリのルートからの相対パス。web のファイルはすべて `apps/web/` の下

## 全タスク共通の注意

- コードを書く前に、`apps/web/AGENTS.md` の指示に従い、`apps/web/node_modules/next/dist/docs/` の該当ガイドを読む
  （CSS: `01-app/01-getting-started/11-css.md`、フォント: `01-app/01-getting-started/13-fonts.md` と
  `01-app/03-api-reference/02-components/font.md`、テスト: `01-app/02-guides/testing/vitest.md`）。
- 依存パッケージは `pnpm --filter @chara-q/web add` で入れ、`package.json` には厳密なバージョン（`^` なし）で書く。
  バージョンは [research.md](research.md) のもの。
- 色は意味の名前（`bg-background`、`text-muted-foreground` など）だけで書き、`dark:` で色を書かない
  （[contracts/theme-and-fonts.md](contracts/theme-and-fonts.md)）。
- コミットは CLAUDE.md の形式（`Refs: specs/001-ui-foundation`）で、論理的なまとまりごとに分ける。

---

## Phase 1: Setup（共通の土台）

**Purpose**: Tailwind CSS・shadcn/ui・Vitest を入れる

- [X] T001 Tailwind CSS 4.3.3 と `@tailwindcss/postcss` 4.3.3 を `apps/web` の devDependencies に入れ、`apps/web/postcss.config.mjs` を作る（`plugins: { "@tailwindcss/postcss": {} }`）
- [X] T002 `apps/web` で `pnpm dlx shadcn@4.23.0 init` を実行し、既定の Base UI と既定のスタイルで初期化する。`apps/web/components.json` の別名が `@/components`・`@/components/ui`・`@/lib/utils`・`@/lib`・`@/hooks` になり、`css` が `src/app/globals.css` を指すことを確かめる。生成された `apps/web/src/lib/utils.ts`（`cn()`）を残し、追加された依存（`@base-ui/react`、`class-variance-authority`、`clsx`、`tailwind-merge`、`lucide-react`、`tw-animate-css` など）を `apps/web/package.json` で厳密なバージョンに直す
- [X] T003 `apps/web` で `pnpm dlx shadcn@4.23.0 add button field input label select textarea` を実行し、`apps/web/src/components/ui/` に部品をコピーする（T002 の後）
- [X] T004 [P] Vitest 5.0.5、`@vitejs/plugin-react` 6.1.2、`vite-tsconfig-paths` 6.1.1、jsdom 30.1.3、`@testing-library/react` 16.3.3、`@testing-library/dom` 10.4.2、`@testing-library/user-event` 14.6.8、`@testing-library/jest-dom` 7.0.1、`axe-core` 4.14.0 を `apps/web` の devDependencies に入れる。`apps/web/vitest.config.mts`（`plugins: [tsconfigPaths(), react()]`、`environment: "jsdom"`、`setupFiles: ["./vitest.setup.ts"]`、`include: ["src/**/*.test.{ts,tsx}"]`）と `apps/web/vitest.setup.ts`（`@testing-library/jest-dom/vitest` の読み込み、`afterEach(cleanup)`、`localStorage.clear()`、`next/font/google` を `vi.mock` して各フォント関数が `{ className, variable, style }` を返すようにする、jsdom にない `window.matchMedia` の既定の偽物）を作る。`apps/web/package.json` に `"test": "vitest run"` と `"test:watch": "vitest"` を足す
- [X] T005 [P] `apps/web/src/test/axe.ts` に、要素を受け取り `axe-core` を実行して違反が 0 件であることを `expect` する補助関数 `expectNoAxeViolations(container)` を作る。jsdom では判定できない `color-contrast` の規則は無効にする（T004 の依存の導入後）
- [X] T006 [P] `apps/web/src/test/match-media.ts` に、テストから端末の設定（ライト/ダーク）を切り替えて `change` イベントを発火できる `matchMedia` の偽物 `mockColorScheme(initial)` を作る（戻り値で `setScheme("dark" | "light")` を呼べる）

**Checkpoint**: `pnpm --filter @chara-q/web test` が「テストなし」で正常終了し、`pnpm --filter @chara-q/web build` が通る

---

## Phase 2: Foundational（全ストーリーの前提）

**Purpose**: 色・フォントの CSS 変数と、ゲームシステムの定義

**⚠️ CRITICAL**: このフェーズが終わるまで、ストーリーの作業を始めない

- [X] T007 `apps/web/src/app/globals.css` を書き直す。既存の素の CSS（`:root` の `--background`/`--foreground`、`body` の `font-family: Arial`、`* { margin: 0 }`、`html, body { overflow-x: hidden }` など）を消し、次の形にする: `@import "tailwindcss";`、`@import "tw-animate-css";`、`@custom-variant dark (&:where(.dark, .dark *));`、`:root { color-scheme: light dark; }`、shadcn/ui の初期化が書いた色の変数（`:root` と `.dark` の 2 組）を `:root` 1 か所の `light-dark(ライトの値, ダークの値)` に書き直す（`.dark` の色のブロックは消す）、`@theme inline` で色の変数を Tailwind の色名に対応させる、`--font-sans: system-ui, -apple-system, "Hiragino Sans", "Hiragino Kaku Gothic ProN", "Yu Gothic UI", "Meiryo", "Noto Sans JP", sans-serif;` と `--font-heading: var(--font-sans);`・`--font-brush: var(--font-sans);` を定義して `font-sans`・`font-heading`・`font-brush` を使えるようにする、`@layer base` で `body` に `bg-background text-foreground font-sans antialiased` を当てる
- [X] T008 [P] `apps/web/src/lib/game-systems.ts` に、[data-model.md](data-model.md) の GameSystem を定義する。型 `GameSystemId = "yuyakekoyake" | "beginning-idol"`、`FontRole = "heading" | "brush"`、`FontId = "mochiy-pop-p-one" | "yuji-syuku"`、`status: "available" | "coming-soon"`、`href` は `status` が `"available"` のときだけ持つ型にする。値は `yuyakekoyake`（name「ふしぎもののけRPG ゆうやけこやけ」、fonts `{ heading: "mochiy-pop-p-one", brush: "yuji-syuku" }`、status `"coming-soon"`）と `beginning-idol`（name「駆け出しアイドルRPG ビギニングアイドル」、fonts `{ heading: "mochiy-pop-p-one" }`、status `"coming-soon"`）。紹介文は旧システムのトップページ（`E:\develop\chara-q-old\chara-q\src\app\page.tsx`、参照のみ）を元に 1 文で書く。一覧の順序を持つ配列 `gameSystems` と、`id` から引く関数を公開する
- [X] T009 [P] `apps/web/src/lib/game-systems.test.ts` に、`id` が小文字の英字とハイフンだけであること、`status` が `"coming-soon"` のシステムは `href` を持たないこと、`id` が重複しないことのテストを書く

**Checkpoint**: `pnpm --filter @chara-q/web test` と `build` が通る。画面の見た目はまだ変わらなくてよい

---

## Phase 3: User Story 1 - 共通レイアウトでトップページを見る (Priority: P1) 🎯 MVP

**Goal**: 全画面共通のスキップリンク・ヘッダー・本文・フッターと、新しいトップページ

**Independent Test**: トップページを開き、ヘッダー（アプリ名）・紹介・システム一覧（準備中）・フッターが表示され、アプリ名が `/` へのリンクで、axe の違反がないことを部品テストで確かめる

### Tests for User Story 1

> 先に書き、実装前に失敗することを確かめる

- [X] T010 [P] [US1] `apps/web/src/components/layout/skip-link.test.tsx`: 「本文へスキップ」のリンクが `#main-content` を指すこと、最初の Tab でフォーカスされること（`user-event` の `tab()`）
- [X] T011 [P] [US1] `apps/web/src/components/layout/site-header.test.tsx`: `banner` のランドマークがあり、アプリ名「chara-q」が `href="/"` のリンクであること、`actions` に渡した要素が右端の操作の領域に描画されること、`expectNoAxeViolations` が通ること
- [X] T012 [P] [US1] `apps/web/src/components/layout/site-footer.test.tsx`: `contentinfo` のランドマークに「chara-q」と「© 2026 Qurioth」があること、`expectNoAxeViolations` が通ること
- [X] T013 [P] [US1] `apps/web/src/app/page.test.tsx`: 見出しと紹介文があること、`gameSystems` の 2 つの表示名が出ること、それぞれに「準備中」が表示され、システム名がリンク（`role="link"`）になっていないこと、API の接続状態（「API:」）が表示されないこと、`expectNoAxeViolations` が通ること

### Implementation for User Story 1

- [X] T014 [P] [US1] `apps/web/src/components/layout/skip-link.tsx`: `href="#main-content"` の「本文へスキップ」。普段は `sr-only`、フォーカス時に `focus:not-sr-only` で画面左上に見える
- [X] T015 [P] [US1] `apps/web/src/components/layout/site-header.tsx`: Server Component。`<header>` の中にアプリ名「chara-q」（`next/link` で `/`、`font-bold`、幅が足りないときは `truncate` で省略）と、右端の操作の領域（props の `actions?: ReactNode` を `shrink-0` で描画）。下端に `border-border` の線。幅 360px で要素が重ならないよう `flex items-center justify-between gap-2 px-4` を基本にする
- [X] T016 [P] [US1] `apps/web/src/components/layout/site-footer.tsx`: Server Component。`<footer>` に「chara-q」と「© 2026 Qurioth」を `text-sm text-muted-foreground` で表示
- [X] T017 [US1] `apps/web/src/app/layout.tsx` を書き直す: Geist と Geist Mono の読み込みを消す。`metadata` を `title: { template: "%s | chara-q", default: "chara-q" }`・`description: "TRPG キャラクターシート管理"` にする。`<html lang="ja">`、`<body className="flex min-h-dvh flex-col">` の中に `SkipLink`、`SiteHeader`、`<main id="main-content" tabIndex={-1} className="flex-1 outline-none">`（中に `UrqlProvider` と `children`）、`SiteFooter` を並べる（T014〜T016 の後）
- [X] T018 [P] [US1] 既存の `apps/web/src/app/health-status.tsx` を `apps/web/src/app/dev/ui/health-status.tsx` へ `git mv` で移す（中身は変えない）
- [X] T019 [US1] `apps/web/src/app/dev/ui/page.tsx` を作る: 先頭で `process.env.NODE_ENV === "production"` なら `notFound()` を呼ぶ。`metadata.title` は「部品の確認」。この時点では見出しと「API の接続状態」の節（`HealthStatus`）だけを置く。後のストーリーで節を足せるよう、節ごとに `<section aria-labelledby>` で区切る（T018 の後）
- [X] T020 [US1] `apps/web/src/app/page.tsx` を書き直す: `HealthStatus` を外し、`<h1>`（アプリ名と短い説明）、紹介文、「対応予定のゲームシステム」の見出しと `gameSystems` の一覧（`<ul>`、各項目に表示名・紹介文・`status` が `"coming-soon"` なら「準備中」のバッジ、`"available"` なら `href` へのリンク）を置く。幅は `mx-auto w-full max-w-5xl px-4 py-8`、一覧は狭い幅で 1 列、`sm:` 以上で 2 列のグリッド。色は意味の名前だけを使う

**Checkpoint**: T010〜T013 が通る。`task dev:web` でトップページが新しいレイアウトで表示される

---

## Phase 4: User Story 2 - ライト/ダークモードを切り替える (Priority: P1)

**Goal**: ヘッダーのボタン 1 つでライト/ダークを切り替え、保存し、最初の描画から反映する。未選択の間は端末の設定に追従する

**Independent Test**: テーマの判定・スクリプト・トグルの部品テストで、保存値・端末の設定・不正な保存値・保存先が使えないときの挙動と、キーボード操作を確かめる

### Tests for User Story 2

- [X] T021 [P] [US2] `apps/web/src/components/theme/theme.test.ts`: [data-model.md](data-model.md) の ThemePreference の決まりを確かめる。保存値 `"dark"`/`"light"` ならそれを使う、保存値がなければ端末の設定、「保存値が `"light"`・`"dark"` 以外（手で書き換えられたなど）のときは、保存値がないものとして扱う」、「保存先が使えない（読み書きで例外が出る）ときは、保存値がないものとして扱い、エラーは表示しない」（`localStorage.getItem`/`setItem` を例外にする）、`applyTheme` が `<html>` の `dark` クラスと `style.colorScheme` を [contracts/theme-and-fonts.md](contracts/theme-and-fonts.md) の表どおりに設定すること、`toggleTheme` が表示中の反対を保存して適用すること
- [X] T022 [P] [US2] `apps/web/src/components/theme/theme-script.test.ts`: `themeScript` の文字列を jsdom で実行し（`new Function(themeScript)()`）、保存値と `mockColorScheme` の組み合わせごとに `<html>` の状態が決まること、保存値がないときは `setScheme` で端末の設定を変えると追従し、保存値を書いた後は追従しないこと、保存先が例外を投げても例外が外に出ないこと
- [X] T023 [P] [US2] `apps/web/src/components/theme/theme-toggle.test.tsx`: ボタンを押すと `<html>` の `dark` クラスと `localStorage` の `chara-q-theme` が切り替わること、読み上げ用の名前が「ダークモードに切り替える」/「ライトモードに切り替える」で、押した後に入れ替わること、`user-event` でフォーカスして Enter と Space で切り替わること、未選択で端末の設定が変わると読み上げ用の名前も追従すること、`expectNoAxeViolations` が通ること

### Implementation for User Story 2

- [X] T024 [US2] `apps/web/src/components/theme/theme.ts`: `THEME_STORAGE_KEY = "chara-q-theme"`、型 `Theme = "light" | "dark"`、`readStoredTheme()`（不正な値と例外は `null`）、`getSystemTheme()`、`resolveTheme()`、`applyTheme(theme)`（`documentElement.classList.toggle("dark", …)` と `style.colorScheme`）、`toggleTheme()`（表示中の反対を保存して適用し、保存の例外は無視）、表示中のテーマの変化を購読する `subscribeTheme(callback)`（`<html>` の `class` を `MutationObserver` で監視）
- [X] T025 [US2] `apps/web/src/components/theme/theme-script.ts`: 最初の描画の前に実行する即時関数の文字列 `themeScript` を作る。T024 の `THEME_STORAGE_KEY` を埋め込み、保存値（`"light"`/`"dark"` 以外は無視）か `prefers-color-scheme` で `<html>` の `dark` クラスと `style.colorScheme` を設定し、保存値がないときだけ `matchMedia` の `change` を購読して追従する（購読の中でも毎回保存値を読み、保存値ができていたら何もしない）。全体を `try { … } catch {}` で包む（T024 の後）
- [X] T026 [US2] `apps/web/src/components/theme/theme-toggle.tsx`: `"use client"`。`useSyncExternalStore(subscribeTheme, resolveTheme, () => null)` で表示中のテーマを読み、`<button type="button">`（shadcn/ui の `Button` の `variant="ghost" size="icon"`）を描画する。読み上げ用の名前は [contracts/ui-components.md](contracts/ui-components.md) の ThemeToggle のとおり（ハイドレーション前は「テーマを切り替える」）。アイコンは `lucide-react` の `Moon`（`dark:hidden`）と `Sun`（`hidden dark:block`）を `aria-hidden` で出し分ける（T024 の後）
- [X] T027 [US2] `apps/web/src/app/layout.tsx` を更新する: `<html>` に `suppressHydrationWarning` を付け、`<head>` に `<script dangerouslySetInnerHTML={{ __html: themeScript }} />` を置き、`SiteHeader` の `actions` に `ThemeToggle` を渡す（T025・T026 の後）

**Checkpoint**: T021〜T023 が通る。`task dev:web` でトグルが動き、再読み込み後も選択が残る

---

## Phase 5: User Story 3 - 共通の入力部品でフォームを組む (Priority: P2)

**Goal**: ラベル・補足説明・エラー・必須・無効の形がそろった 4 種類の入力部品と、部品確認ページ

**Independent Test**: 4 つの部品の部品テストで、[contracts/ui-components.md](contracts/ui-components.md) の属性と操作を確かめる

### Tests for User Story 3

- [X] T028 [P] [US3] `apps/web/src/components/form/number-parse.test.ts`: [data-model.md](data-model.md) の「入力の決まり」をテーブル駆動で確かめる。「全角の数字・マイナス・ピリオド（`０-９`、`－`、`ー`、`．`）は半角に置き換える」、「数字・`-`・`.` 以外の文字を含む入力は受け付けず」、途中の形（`-`、`1.`）は値なし・空欄は `null`、blur 時の検証の文言「0 以上の数を入力してください」「整数を入力してください」「{min} 以上 {max} 以下で入力してください」（片方だけのときは「{min} 以上で入力してください」「{max} 以下で入力してください」）
- [X] T029 [P] [US3] `apps/web/src/components/form/form-field.test.tsx`: ラベルの `for` と入力欄の `id` が一致し、`getByLabelText` で入力欄が取れること、`aria-describedby` が「補足説明の `id` と、エラーがあればエラーの `id`（この順、空白区切り）」になること、エラーで `aria-invalid="true"` と `role="alert"`、`required` でラベルに「必須」と `aria-required="true"`、`hideLabel` でもラベル名で取れること、`id` を渡したときはそれを使うこと
- [X] T030 [P] [US3] `apps/web/src/components/form/text-field.test.tsx`: 入力で `onValueChange` が呼ばれる、`disabled` で入力できない、`className` と `inputClassName` がそれぞれ外枠と入力欄に付く、`expectNoAxeViolations`
- [X] T031 [P] [US3] `apps/web/src/components/form/number-field.test.tsx`: spec の US3 の受け入れシナリオ 2〜4 を確かめる。「数字以外の文字を入力・貼り付けする」と入力欄が直前のまま（`user-event` の `type` と `paste`）、全角「１２」が `12` になる、空欄で `onValueChange(null)`、既定（`allowNegative`・`allowDecimal` なし）で `-3` と `1.5` は blur 後にエラー、`min`/`max` の範囲外で blur 後にエラー、範囲外でも `onValueChange` には値を渡す、画面側の `error` が内部のエラーより優先、`inputMode` が `numeric`（`allowDecimal` で `decimal`）、外から `value` を変えると入力欄が追従
- [X] T032 [P] [US3] `apps/web/src/components/form/select-field.test.tsx`: 未選択でプレースホルダー「選択してください」、キーボード（フォーカス → Enter で開く → ArrowDown → Enter）で選ぶと `onValueChange` が呼ばれフォーカスがトリガーに戻る、Escape で閉じる、`disabled` の選択肢は選べない、ラベル名でトリガーが取れる、`expectNoAxeViolations`。Base UI の部品が jsdom で必要とする API（`scrollIntoView`、`PointerEvent`、`ResizeObserver` など）が足りなければ `apps/web/vitest.setup.ts` に最小の偽物を足す
- [X] T033 [P] [US3] `apps/web/src/components/form/textarea-field.test.tsx`: 複数行（`{Enter}` を含む）の入力で改行 `\n` が `onValueChange` の値に残る、`rows` の既定が 3、`expectNoAxeViolations`

### Implementation for User Story 3

- [X] T034 [P] [US3] `apps/web/src/components/form/number-parse.ts`: 純粋関数で、`normalizeNumberText(text)`（全角→半角）、`isAcceptableNumberText(text)`（数字・`-`・`.` だけで、`-` は先頭に 1 つ、`.` は 1 つまで）、`parseNumberText(text)`（空欄 → `null`、途中の形 → `undefined`（値を変えない）、それ以外 → `number`）、`validateNumber(value, { min, max, allowNegative, allowDecimal })`（エラー文言か `undefined`）を実装する
- [X] T035 [US3] `apps/web/src/components/form/form-field.tsx`: [contracts/ui-components.md](contracts/ui-components.md) の「入力部品の共通の枠」。shadcn/ui の `Field`・`FieldLabel`・`FieldDescription`・`FieldError` を使い、`useId()` で `id`・説明の `id`・エラーの `id` を作る。入力欄の描画は render prop（`children: (controlProps) => ReactNode`）で受け取り、`controlProps` に `id`・`name`・`aria-describedby`・`aria-invalid`・`aria-required`・`disabled` を入れて渡す。`required` のラベルの後ろに「必須」、`hideLabel` は `sr-only`、エラーは `role="alert"` とエラーのアイコン（`lucide-react` の `CircleAlert`、`aria-hidden`）付きで `text-destructive`。外枠に `className`、`data-invalid` をエラー時に付ける
- [X] T036 [P] [US3] `apps/web/src/components/form/text-field.tsx`: `FormField` と shadcn/ui の `Input`。props は contracts の TextField（`type` は `"text" | "password" | "search"`）。`inputClassName` を `Input` に `cn()` で足す（T035 の後）
- [X] T037 [P] [US3] `apps/web/src/components/form/number-field.tsx`: `"use client"`。`FormField` と `Input`（`type="text"`、`inputMode` は `allowDecimal ? "decimal" : "numeric"`）。入力欄の文字列を内部の状態に持ち、`onChange` で `normalizeNumberText` → `isAcceptableNumberText` が偽なら状態を変えない → `parseNumberText` の結果が `undefined` 以外なら `onValueChange` を呼ぶ。`onBlur` で `validateNumber` を実行して内部のエラーを持ち、表示するエラーは `error ?? 内部のエラー`。外から `value` が変わり、入力中の文字列を読んだ値と違うときは、文字列を `value` に合わせる（T034・T035 の後）
- [X] T038 [P] [US3] `apps/web/src/components/form/select-field.tsx`: `"use client"`。`FormField` と shadcn/ui（Base UI）の `Select`。props は contracts の SelectField。`null` を未選択として扱い、プレースホルダーの既定は「選択してください」。選択肢の一覧に `max-h` を付けてスクロールさせ、トリガーの表示は `truncate` で 1 行にする。`controlProps` はトリガーに渡す（T035 の後）
- [X] T039 [P] [US3] `apps/web/src/components/form/textarea-field.tsx`: `FormField` と shadcn/ui の `Textarea`。props は contracts の TextareaField（`rows` の既定 3）（T035 の後）
- [X] T040 [US3] `apps/web/src/app/dev/ui/form-preview.tsx`（`"use client"`、各部品の値を `useState` で持つ）を作り、`apps/web/src/app/dev/ui/page.tsx` に「入力部品」の節として置く。4 つの部品それぞれについて、通常・補足説明あり・エラー・無効・必須・ラベルを隠した状態を並べ、数値入力は既定・`min`/`max` あり・`allowNegative`・`allowDecimal` の例を、セレクトは選択肢 30 件の例と長い文言の例を含める（T036〜T039 の後）

**Checkpoint**: T028〜T033 が通る。`/dev/ui` で 4 つの部品を操作できる

---

## Phase 6: User Story 5 - web の自動テストを実行する (Priority: P2)

**Goal**: web のテストを既存の入口（`task test`）と CI から実行する

**Independent Test**: API と DB を止めた状態で `task test:web` が成功し、わざと壊したテストで終了コードが 0 以外になる

### Implementation for User Story 5

- [X] T041 [P] [US5] `Taskfile.yml` に `test:web`（`desc: web のテスト`、`cmds: [pnpm --filter {{.WEB_PKG}} test]`）を足し、`test` の `cmds` を `task: test:api` と `task: test:web` の 2 つにする。`test` の `desc` を「テストを実行する（api と web。api の DB を使うテストには `task db:up` が必要）」にする
- [X] T042 [P] [US5] `.github/workflows/ci-web.yml` の「型チェック」と「build」の間に `- name: テスト` / `run: pnpm --filter @chara-q/web test` を足す
- [X] T043 [US5] API と DB を止めた状態で `task test:web` を実行し、すべて成功して 1 分以内に終わることを確かめる（SC-008）。続けて任意のテストの期待値を一時的に壊して実行し、失敗したテスト名が表示され終了コードが 0 以外になることを確かめてから元に戻す（壊した変更はコミットしない）（T041 の後）

**Checkpoint**: `task test` で api と web の両方のテストが走る

---

## Phase 7: User Story 4 - ゲームシステムごとのフォントで表示する (Priority: P3)

**Goal**: `data-game-system` を付けた領域で、見出し・筆文字の役割のフォントがシステムのものに変わる

**Independent Test**: `GameSystemScope` の属性と、`game-systems.ts` の `fonts` と `globals.css` の差し替えの一致を部品テストで確かめる。見た目は `/dev/ui` で目視

### Tests for User Story 4

- [X] T044 [P] [US4] `apps/web/src/components/game-system/game-system-scope.test.tsx`: `system` の値が `data-game-system` になる、`as` で要素が変わる（既定 `div`）、入れ子にすると内側の要素に内側の値が付く、`className` が付く
- [X] T045 [P] [US4] `apps/web/src/lib/game-system-fonts.test.ts`: `apps/web/src/app/globals.css` をテキストとして読み（`node:fs`）、`gameSystems` の各システムの `fonts` の役割ごとに、`[data-game-system="<id>"]` のブロックに `--font-<role>: var(--font-<fontId>)` があること、`fonts` にない役割は上書きしていないことを確かめる

### Implementation for User Story 4

- [X] T046 [P] [US4] `apps/web/src/lib/fonts.ts`: `next/font/google` の `Mochiy_Pop_P_One`（`weight: "400"`、`variable: "--font-mochiy-pop-p-one"`）と `Yuji_Syuku`（`weight: "400"`、`variable: "--font-yuji-syuku"`）を、どちらも `preload: false`・`display: "swap"`・`fallback` に `--font-sans` と同じ並び（`["Hiragino Sans", "Yu Gothic UI", "Meiryo", "sans-serif"]` など）で定義し、`<html>` に付けるクラスをまとめた `fontVariables` を公開する。`FontId` から CSS 変数名を引く対応表も置く
- [X] T047 [US4] `apps/web/src/app/globals.css` に、[contracts/theme-and-fonts.md](contracts/theme-and-fonts.md) の「システムごとの差し替え」の 2 つのブロック（`[data-game-system="yuyakekoyake"]` で `--font-heading` と `--font-brush`、`[data-game-system="beginning-idol"]` で `--font-heading`）を足す。値は `var(--font-…), var(--font-sans)` の形（T007 の後）
- [X] T048 [P] [US4] `apps/web/src/components/game-system/game-system-scope.tsx`: [contracts/ui-components.md](contracts/ui-components.md) の GameSystemScope。`<{as} data-game-system={system} className={className}>` だけを描画する Server Component
- [X] T049 [US4] `apps/web/src/app/layout.tsx` の `<html>` の `className` に `fontVariables` を付ける（T046 の後）
- [X] T050 [US4] `apps/web/src/app/dev/ui/page.tsx` に「ゲームシステムのフォント」の節を足す: 領域の外、`GameSystemScope system="yuyakekoyake"`、`GameSystemScope system="beginning-idol"` の 3 つを並べ、それぞれに `font-heading` の見出しと `font-brush` の文（旧字体や記号を含む文も 1 つ）を置く（T048・T049 の後）

**Checkpoint**: T044・T045 が通る。`/dev/ui` でフォントの違いが見える

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: 決定の記録、ドキュメント、全体の確認

- [X] T051 [P] `docs/adr/0007-ui-foundation.md` を書く（ADR 0004 と同じ形式: ステータス・日付・背景・決定・検討した代案・結果）。内容は [research.md](research.md) の R2（Base UI）、R5・R6（テーマと `light-dark()`）、R7（フォント）、R8（Vitest）、R4（フォームのライブラリは後で決める）。Base UI は MUI のチームが保守するが、スタイルを持たない部品だけを使うため constitution の「MUI は使わない」に当たらないことも書く
- [X] T052 [P] `docs/adr/0004-ui-library.md` の「次は UI 基盤の Spec で決める。」の節の後に、「決定は ADR 0007 を参照」の 1 行とリンクを足す
- [X] T053 [P] `README.md` を更新する: タスク一覧に `task test:web` を足し、`task test` の説明を api と web の両方に直す。「Spec Kit で進める予定の機能」の 1（UI 基盤）を完了の印にし、`specs/001-ui-foundation` へリンクする。`CLAUDE.md` の「よく使うコマンド」の `task test` の説明も同じに直す
- [X] T054 `task lint`、`task test`（`task db:up` の後）、`task check:generated`、`task build:web` をすべて実行し、通ることを確かめる。`apps/web/AGENTS.md` に `next dev` が差分を出していたら、そのままコミットに含める
- [ ] T055 [quickstart.md](quickstart.md) の「2. 本物のブラウザでの目視」をすべて行う（画面幅、スキップリンク、テーマの 5 項目、部品確認ページの 6 項目、本番ビルドで `/dev/ui` が 404 と Lighthouse の LCP）。満たさない項目があれば直し、結果（コントラストの最小値、LCP の値）を PR の説明に書く

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 依存なし。T002 → T003 の順。T004〜T006 は T001〜T003 と並行できる（T005・T006 は T004 の後）
- **Foundational (Phase 2)**: Setup の後。T007 は T002 の後（shadcn/ui の初期化が書いた色を書き直すため）
- **US1 (Phase 3)**: Foundational の後
- **US2 (Phase 4)**: Foundational の後。T027 だけは US1 の T017（layout.tsx）の後
- **US3 (Phase 5)**: Foundational の後。T040 だけは US1 の T019（部品確認ページ）の後
- **US5 (Phase 6)**: Setup の後ならいつでもよい。T043 は、テストが 1 件以上ある状態（どれかのストーリーの後）で行う
- **US4 (Phase 7)**: Foundational の後。T049 は US1 の T017 の後、T050 は US1 の T019 の後
- **Polish (Phase 8)**: すべてのストーリーの後

### User Story Dependencies

- **US1 (P1)**: 他のストーリーに依存しない
- **US2 (P1)**: 部品とテストは独立。ヘッダーへの組み込み（T027）だけ US1 に依存
- **US3 (P2)**: 部品とテストは独立。部品確認ページへの配置（T040）だけ US1 に依存
- **US5 (P2)**: 独立（Taskfile と CI だけ）
- **US4 (P3)**: 部品とテストは独立。layout と部品確認ページへの組み込みだけ US1 に依存

### Within Each User Story

- テストを先に書き、失敗することを確かめてから実装する
- 純粋関数（`theme.ts`、`number-parse.ts`）→ 部品 → 画面への組み込みの順
- `apps/web/src/app/layout.tsx` は T017・T027・T049、`apps/web/src/app/dev/ui/page.tsx` は T019・T040・T050、`apps/web/src/app/globals.css` は T007・T047 で変更するため、同じファイルのタスクは並行しない

### Parallel Opportunities

- Setup: T004 と T001〜T003 は並行できる。T005・T006 は互いに並行できる
- Foundational: T008・T009 は T007 と並行できる
- 各ストーリーのテスト（[P]）はすべて並行して書ける
- US3 の T036〜T039 は T035 の後に並行できる
- Foundational の後、US1・US2・US3・US4・US5 の部品とテストは並行できる（layout.tsx・page.tsx への組み込みだけ順に行う）

---

## Parallel Example: User Story 3

```text
# テストをまとめて書く:
Task: "number-parse のテストを apps/web/src/components/form/number-parse.test.ts に書く"
Task: "FormField のテストを apps/web/src/components/form/form-field.test.tsx に書く"
Task: "NumberField のテストを apps/web/src/components/form/number-field.test.tsx に書く"
Task: "SelectField のテストを apps/web/src/components/form/select-field.test.tsx に書く"

# FormField（T035）の後、4 つの部品をまとめて作る:
Task: "TextField を apps/web/src/components/form/text-field.tsx に作る"
Task: "NumberField を apps/web/src/components/form/number-field.tsx に作る"
Task: "SelectField を apps/web/src/components/form/select-field.tsx に作る"
Task: "TextareaField を apps/web/src/components/form/textarea-field.tsx に作る"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: Setup
2. Phase 2: Foundational
3. Phase 3: US1（共通レイアウトとトップページ）
4. **STOP and VALIDATE**: US1 のテストと、トップページの目視（画面幅、スキップリンク）

### Incremental Delivery

1. Setup + Foundational → Tailwind・shadcn/ui・Vitest が動く
2. US1 → 新しいレイアウトのトップページ（MVP）
3. US2 → テーマの切り替え
4. US3 → 入力部品と部品確認ページ
5. US5 → `task test` と CI で web のテスト
6. US4 → ゲームシステムのフォント
7. Polish → ADR 0007、README、全体の目視確認

コミットの区切りの目安（CLAUDE.md の形式、`Refs: specs/001-ui-foundation`）:
Setup + Foundational（`build(web)` または `feat(web)`）→ US1 → US2 → US3 → US5（`ci` と Taskfile は別コミットでもよい）→ US4 → ADR とドキュメント（`docs`）。

---

## Notes

- 実装時に計画から変えた点（詳細は [research.md](research.md) の「実装時の変更」「実装時の記録」）:
  - T002: shadcn/ui の初期化は `clsx`・`tailwind-merge` ではなく `cn` パッケージを入れた。`shadcn` と `tw-animate-css` は devDependencies に置いた
  - T004: `vite-tsconfig-paths` は入れず、Vite 8 の `resolve.tsconfigPaths` を使った
  - T007・T024・T025: `light-dark()` は Lightning CSS の変換でスクリプトから切り替えられないため、`@custom-variant dark` のブロック形式と `@variant dark` に変え、スクリプトは `dark` / `light` のクラスを付ける

- [P] のタスクは別のファイルで、未完了のタスクに依存しない
- [Story] はタスクとユーザーストーリーの対応を示す
- shadcn/ui からコピーした `apps/web/src/components/ui/` は自分のコードとして変更してよい（constitution）。変えたときは、コピー元との違いがわかるよう短いコメントを残す
- 旧リポジトリ（`E:\develop\chara-q-old`）は参照だけで、変更しない
