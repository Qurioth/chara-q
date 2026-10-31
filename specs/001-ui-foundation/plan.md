# Implementation Plan: UI 基盤

**Branch**: `001-ui-foundation` | **Date**: 2026-10-31 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-ui-foundation/spec.md`

## Summary

web（Next.js）に Tailwind CSS v4 と shadcn/ui（土台は Base UI）を入れ、全画面共通のレイアウト
（スキップリンク・ヘッダー・本文・フッター）を作る。テーマはライト/ダークの 2 択の切り替えで、
`<head>` のインラインスクリプトが最初の描画の前に `html` の `dark` / `light` クラスと `color-scheme` を決める。
色の変数は `:root` と、その中の `@variant dark` で 1 か所に定義する。ゲーム
システムごとのフォントは `next/font/google` で読み込み、`data-game-system` 属性を付けた領域
でだけ CSS 変数を差し替える。共通入力部品は shadcn/ui の Field・Input・Select・Textarea を
土台に、ラベル・補足説明・エラーの関連付けをまとめた 4 つの部品（テキスト・数値・セレクト・
テキストエリア）を作る。トップページはアプリの紹介と対応予定のシステム一覧に作り直し、API の
接続確認は開発時だけの部品確認ページへ移す。テストは Vitest + jsdom + Testing Library で、
`task test` から Go のテストと一緒に実行し、CI でも実行する。

## Technical Context

**Language/Version**: TypeScript 5（`apps/web`）、Node.js 24.22.2、pnpm 12.16.1（`mise.toml`）

**Primary Dependencies**: Next.js 16.4.3（App Router、`cacheComponents: true`）、React 19.3.0、
Tailwind CSS 4.3.3（`@tailwindcss/postcss`）、shadcn CLI 4.23.0（Base UI、`@base-ui/react` 1.9.0）、
`lucide-react`、`class-variance-authority`、`cn`、`tw-animate-css`
（shadcn/ui の初期化で入るもの）。詳細は [research.md](research.md)。

**Storage**: ブラウザの `localStorage`（テーマの選択だけ）。DB・API の変更はない。

**Testing**: Vitest 5.0.5、jsdom 30.1.3、`@testing-library/react` 16.3.3、
`@testing-library/user-event` 14.6.8、`@testing-library/jest-dom` 7.0.1、`axe-core` 4.14.0

**Target Platform**: 最新の主要ブラウザ（Chrome、Edge、Safari、Firefox）とスマートフォンの標準ブラウザ。
web は Vercel で配信する。

**Project Type**: Web アプリ（モノレポの `apps/web`）

**Performance Goals**: トップページの主な内容をモバイル回線相当で 2.5 秒以内に表示（SC-006）。
システム用フォントは、その領域を表示しない画面ではダウンロードしない（FR-022）。

**Constraints**: 最初の描画で別のテーマを見せない（FR-014）。`cacheComponents` が有効なため、
ルートレイアウトで `cookies()` などのリクエスト時の API を使わない（使うと全画面が動的になる）。
幅 360px 以上で横スクロールなし（FR-005）。コントラスト比 4.5:1 以上（FR-016）。

**Scale/Scope**: 画面 2 つ（トップページ、開発時のみの部品確認ページ）、共通部品はレイアウト 3、
テーマ切り替え 1、入力部品 4、ゲームシステムの領域 1。

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則・制約 | 判定 | 根拠 |
|---|---|---|
| I. スキーマファースト | 該当なし | GraphQL スキーマを変えない。既存の `Health` クエリは部品確認ページへ移すだけ |
| II. 生成コードはコミットする | 適合 | GraphQL Code Generator の出力は変わらない（`task check:generated` が通る）。shadcn/ui からコピーした部品は constitution の UI 制約のとおり生成コードではなく自分のコードとして扱う |
| III. キャラクターシートのモデル | 該当なし | シートは扱わない。ゲームシステムの定義は画面表示用の定数だけ |
| IV. SQL ファースト | 該当なし | DB を変えない |
| V. N+1 禁止 | 該当なし | resolver を変えない |
| VI. テスト | 適合 | `task test` に web のテストを加え、`task lint` とともに CI で実行する。DB を触らないため Postgres の要件は関係しない |
| VII. Spec 駆動開発 | 適合 | specify → clarify（4 問）→ plan の順で進めている。この機能に関わる未決定事項（WebP）はない |
| 技術スタック（Tailwind CSS + shadcn/ui、MUI 禁止） | 適合 | Tailwind CSS v4 と shadcn/ui を使い、MUI は入れない |
| 新しい基盤技術には ADR | 適合（ADR を追加） | Vitest（web のテスト基盤）と Base UI（shadcn/ui の土台）、テーマとフォントの方式を ADR 0007 に記録する。ADR 0004 の「UI 基盤の Spec で決める」項目への回答も兼ねる |
| UI: shadcn/ui は操作の伴う部品にだけ使う | 適合 | Select と、ラベル付けの Field・Input・Textarea・Button だけを使う。レイアウトとシートの形は Tailwind で組み、`Card`・`Table` は使わない |
| ツールとタスク（mise、Taskfile だけが入口） | 適合 | ツールのバージョンは変えない。テストの入口は `task test` / `task test:web` |
| 認証（ブラウザは web とだけ通信） | 該当なし（既存の状態のまま） | 部品確認ページの API 接続確認は、現状の `NEXT_PUBLIC_API_URL` への直接接続を引き継ぐ。`rewrites` での中継は認証の機能で入れる |
| シークレット | 適合 | 新しい環境変数やシークレットはない |

違反はないため、Complexity Tracking は空のまま。

### Phase 1 後の再確認

設計（[data-model.md](data-model.md)、[contracts/](contracts/)、[quickstart.md](quickstart.md)）を
反映して再確認した。判定は上の表から変わらない。

- 追加の依存はすべて `apps/web` の `package.json` に厳密なバージョンで書き、`mise.toml` は変えない。
- 部品確認ページは本番のビルドで 404 を返すため、本番に開発用の画面や API の接続先は出ない（FR-032）。
- テーマの保存は `localStorage` だけで、サーバーにも Cookie にも送らない（FR-012）。

## Project Structure

### Documentation (this feature)

```text
specs/001-ui-foundation/
├── spec.md
├── plan.md              # このファイル
├── research.md          # Phase 0: 技術の選定
├── data-model.md        # Phase 1: テーマの選択、ゲームシステム、入力部品の共通属性
├── quickstart.md        # Phase 1: 動作確認の手順
├── contracts/
│   ├── ui-components.md # 共通部品の props とふるまい
│   └── theme-and-fonts.md # テーマの保存・html のクラス・色とフォントの CSS 変数
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks で作る
```

### Source Code (repository root)

```text
apps/web/
├── components.json                  # shadcn/ui の設定（新規）
├── postcss.config.mjs               # @tailwindcss/postcss（新規）
├── vitest.config.mts                # Vitest の設定（新規）
├── vitest.setup.ts                  # jest-dom のマッチャー、テストごとの後片付け（新規）
├── package.json                     # 依存と test スクリプトを追加
└── src/
    ├── app/
    │   ├── globals.css              # Tailwind、色の CSS 変数（ライト/ダーク）、フォントの CSS 変数
    │   ├── layout.tsx               # テーマの初期化スクリプト、フォント、共通レイアウト
    │   ├── page.tsx                 # トップページ（紹介とシステム一覧）
    │   ├── page.test.tsx
    │   └── dev/ui/
    │       ├── page.tsx             # 部品確認ページ（本番は 404）
    │       └── health-status.tsx    # 既存の API 接続確認を移動
    ├── components/
    │   ├── ui/                      # shadcn/ui からコピーした部品（button、field、input、label、select、textarea）
    │   ├── layout/
    │   │   ├── site-header.tsx
    │   │   ├── site-footer.tsx
    │   │   ├── skip-link.tsx
    │   │   └── *.test.tsx
    │   ├── theme/
    │   │   ├── theme-script.ts      # 最初の描画の前に実行するスクリプト（文字列）
    │   │   ├── theme.ts             # 保存キー・切り替えの処理（スクリプトとトグルで共有）
    │   │   ├── theme-toggle.tsx
    │   │   └── *.test.ts(x)
    │   ├── form/
    │   │   ├── form-field.tsx       # ラベル・補足説明・エラー・必須の共通の枠
    │   │   ├── text-field.tsx
    │   │   ├── number-field.tsx
    │   │   ├── number-parse.ts      # 全角→半角、許可する文字の判定、範囲の検証
    │   │   ├── select-field.tsx
    │   │   ├── textarea-field.tsx
    │   │   └── *.test.ts(x)
    │   └── game-system/
    │       ├── game-system-scope.tsx
    │       └── game-system-scope.test.tsx
    ├── lib/
    │   ├── fonts.ts                 # next/font/google の定義
    │   ├── game-systems.ts          # ゲームシステムの定義（表示名、紹介文、対応状況、フォント）
    │   ├── utils.ts                 # shadcn/ui の cn()
    │   └── urql-provider.tsx        # 既存
    └── test/
        └── axe.ts                   # axe-core で違反がないことを確かめる補助関数

Taskfile.yml                         # test に test:web を追加
.github/workflows/ci-web.yml         # テストのステップを追加
docs/adr/0007-ui-foundation.md       # Base UI、テーマ、フォント、Vitest の決定
README.md                            # テストの説明、「予定の機能」の 1 を完了に
```

**Structure Decision**: 既存のモノレポ構成のまま、変更は `apps/web` と、Taskfile・CI・ADR・README
に限る。テストは対象のファイルの隣に `*.test.ts(x)` として置く（Vitest の既定の探索範囲で見つかり、
部品と一緒に移動・削除できるため）。shadcn/ui のコピーは `src/components/ui/`、それを包む
このアプリの部品は用途ごとのディレクトリに分け、コピー元との差分を追いやすくする。

## Complexity Tracking

違反はない。
