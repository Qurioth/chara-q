# Implementation Plan: キャラクターシートの共通基盤

**Branch**: `003-character-sheet-base` | **Date**: 2026-11-10 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/003-character-sheet-base/spec.md`

## Summary

全システム共通のキャラクターシートの型・保存・一覧・検索・閲覧と、作成・更新・複製・削除の API を
用意する。GraphQL では `interface CharacterSheet` に共通の項目を置き、`YuyakekoyakeSheet` と
`BeginningIdolSheet` がこれを実装して、システム固有の部分を `data: JSON!` で返す。DB では
`character_sheets` の表に、共通の項目・所有者・システム・日時を列で、システム固有の部分を JSONB で
持つ。所有者の表示名は dataloadgen の DataLoader でまとめて引く。

一覧と閲覧はログインなしで見られ、作成と複製はログインしたユーザー、更新と削除は所有者だけが
できる（ADR 0005）。所有者の確認と同時の更新の検出（`revision`）は、更新の SQL の 1 文で行う。
削除は `deleted_at` の論理削除で、取り消しは管理者の DB 操作で行う。

web には、一覧（検索・並び順・ページ送り、条件は URL に持つ）と閲覧画面の共通部分、ヘッダーの
一覧への入口を作る。作成・編集・複製・削除の画面は、各システムの Spec で作る。

## Technical Context

**Language/Version**: Go 1.27.3（`apps/api`）、TypeScript 5（`apps/web`）、Node.js 24.22.2、
pnpm 12.16.1（`mise.toml`。変更しない）

**Primary Dependencies**:
- api: gqlgen 0.17.98、chi 5.5.0、pgx 5.11.0、sqlc 1.31.1、goose（`tools/go.mod`）。追加は
  `github.com/vikstrous/dataloadgen` v0.0.10 だけ（research R18）。
- web: Next.js 16.4.3（App Router、`cacheComponents: true`）、React 19.3.0、urql 5.0.4、
  GraphQL Code Generator（client preset）、shadcn/ui（Base UI）。追加はない。

**Storage**: Postgres 18（ローカルは `compose.yaml` の `postgres:18.6-alpine`、本番は Neon）。
表を 1 つ追加する（`character_sheets`）。詳細は [data-model.md](data-model.md)。

**Testing**: api は `go test`（テーブル駆動、DB を使うテストは `TEST_DATABASE_URL` の実際の Postgres、
GraphQL は `httptest`、問い合わせの回数は pgx の `QueryTracer`）。web は Vitest + Testing Library +
`user-event`（既存）。手で確かめることは [quickstart.md](quickstart.md)。

**Target Platform**: api は Cloud Run（Linux コンテナ）、web は Vercel。ブラウザは最新の Chrome・
Edge・Safari・Firefox とスマートフォンの標準ブラウザ。

**Project Type**: Web アプリ（モノレポの `apps/api` と `apps/web`）

**Performance Goals**: シート 10,000 件で、一覧・検索・閲覧の api の応答が 200ms 以内（SC-002 の
1 秒に十分な余裕）。一覧の 1 回の表示で、DB への問い合わせは件数・一覧・所有者の 3 回で一定
（SC-003）。

**Constraints**:
- `cacheComponents` のため、`searchParams`・`params`・`cookies()` を読むコンポーネントは
  `<Suspense>` の中に置く（research R13）。シートのデータはキャッシュしない。
- 所有者の確認は画面ではなく api の resolver で行う（constitution「シートの公開範囲」、FR-027）。
- リクエストの本体は 1MB まで、システム固有の部分は 256KB まで（research R11）。
- キャラクター名の並び順は `ja-x-icu` の照合順序に頼る。使えない環境の扱いは research R9。

**Scale/Scope**: シートは数千〜1 万件、利用者は数十人規模。画面 2 つ（一覧、閲覧）とヘッダーの
リンク、GraphQL のクエリ 2 つとミューテーション 4 つ、表 1 つ。

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則・制約 | 判定 | 根拠 |
|---|---|---|
| I. スキーマファースト | 適合 | `schema/graphql/character_sheet.graphqls` を追加し、`auth.graphqls` の enum に値を足してから `task generate` する |
| II. 生成コードはコミットする | 適合 | gqlgen・sqlc・GraphQL Code Generator の出力を、スキーマ・SQL と同じコミットに入れる。`internal/graph/model/sheet.go` は手で書くファイルで、生成コードではない（research R3） |
| III. キャラクターシートのモデル | 適合 | `interface CharacterSheet` を 2 つのシステムの型が実装し、固有の部分は `JSON`。DB は検索・一覧の項目を列、システム固有の部分を JSONB に置く（research R1・R2） |
| IV. SQL ファースト | 適合 | 表は goose のマイグレーション、クエリは `db/queries/character_sheets.sql` を sqlc で生成する。並び順ごとに SQL を分ける（data-model） |
| V. N+1 禁止 | 適合 | `owner` は dataloadgen の `UserByID` でまとめて引く。問い合わせの回数をテストで確かめる（research R4） |
| VI. テスト | 適合 | DB を使うテストは実際の Postgres。Go のテストはテーブル駆動。`task test`・`task lint` を通す（research R17） |
| VII. Spec 駆動開発 | 適合 | specify → clarify（5 問）→ plan の順。未決定事項（WebP の変換）には関わらない |
| 技術スタック・新しい基盤技術には ADR | 適合 | 基盤技術は変えない。追加は小さなライブラリ 1 つ（dataloadgen）。ADR 0001 の「関連データは DataLoader で解決」の範囲 |
| UI: Tailwind CSS、shadcn/ui は操作の伴う部品だけ | 適合 | 既存の `SelectField` などの共通部品を使い、新しい shadcn/ui の部品は足さない。チェックボックスはネイティブの要素を Tailwind で整える |
| 認証（ADR 0006） | 適合 | 002-auth のセッションと中継を使う。Server Components はセッションの Cookie だけを付けて api を呼ぶ（research R13） |
| シートの公開範囲（ADR 0005） | 適合 | 一覧・閲覧はログイン不要、作成・複製はログイン、更新は所有者だけを resolver と SQL の条件で守る。管理者も例外にしない |
| 画像 | 該当なし | 画像は扱わない |
| シークレット・コスト | 適合 | 新しいシークレット・環境変数・クラウドのリソースはない |
| ツールとタスク | 適合 | ツールのバージョンは変えない。新しいタスクは作らない（`task db:new` でマイグレーションを作る） |
| 旧システム | 適合 | 旧リポジトリは、共通の項目と検索の決め方の参考に読んだだけ。変更しない |

違反はないため、Complexity Tracking は空のまま。

### Phase 1 後の再確認

設計（[data-model.md](data-model.md)、[contracts/](contracts/)、[quickstart.md](quickstart.md)）を
反映して再確認した。判定は上の表から変わらない。

- `data` を interface に置かず、各システムの型に置く（research R2）。各システムの Spec が型を
  切り出すときに、interface を変えずに済み、constitution III の SHOULD に沿う。
- 削除は、ADR 0005 の「後続の Spec で決めること」の 1 つ（削除を許すか、誰ができるか）を、
  「所有者だけ、データは残す」で決めた。実装の中で ADR 0005 に追記する。constitution の
  「シートの公開範囲」にも「削除は所有者だけができる」を足す。原則の節の拡張のため MINOR
  （1.3.0 → 1.4.0）とする。
- 002-auth の `FORBIDDEN` の文言（「管理者だけが行えます」）を、シートの操作にも使える
  「この操作は行えません」に変える（contracts/graphql.md）。web の文言の対応表も合わせる。

## Project Structure

### Documentation (this feature)

```text
specs/003-character-sheet-base/
├── spec.md
├── plan.md              # このファイル
├── research.md          # Phase 0: 技術の選定と調査
├── data-model.md        # Phase 1: 表、索引、状態の移り変わり、sqlc のクエリ
├── quickstart.md        # Phase 1: 動作確認の手順
├── contracts/
│   ├── graphql.md       # GraphQL のスキーマの案、ふるまい、エラーのコード
│   └── web-routes.md    # 画面、URL のクエリ、ヘッダー
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks で作る
```

### Source Code (repository root)

```text
schema/graphql/
├── character_sheet.graphqls         # CharacterSheet、User、クエリとミューテーション（新規）
└── auth.graphqls                    # InputField・InputErrorCode に値を足す

apps/api/
├── cmd/server/main.go               # sheet のサービスとローダーを組み込む
├── db/
│   ├── migrations/00003_character_sheets.sql  # 表と索引（新規、task db:new で作る）
│   └── queries/
│       ├── character_sheets.sql     # シートのクエリ（新規）
│       └── users.sql                # ListUsersByIDs を足す
├── internal/
│   ├── sheet/                       # シートの処理（新規）
│   │   ├── system.go                # システムの値の変換（DB・GraphQL）
│   │   ├── validate.go              # 共通の項目と data の規則
│   │   ├── service.go               # 一覧、閲覧、作成、更新、複製、削除
│   │   ├── search.go                # 絞り込みと並び順、ページの丸め
│   │   └── *_test.go
│   ├── loader/                      # DataLoader（新規）
│   │   ├── loader.go                # リクエストごとのローダーとミドルウェア
│   │   └── loader_test.go           # 問い合わせの回数のテスト
│   ├── db/                          # sqlc の生成コード（character_sheets.sql.go が増える）
│   ├── graph/
│   │   ├── resolver.go              # sheet のサービスを持たせる
│   │   ├── character_sheet.resolvers.go  # gqlgen が作る。実装を書く
│   │   ├── errors.go                # NOT_FOUND を足し、FORBIDDEN の文言を変える
│   │   ├── model/sheet.go           # 2 つのシステムの型（手で書く、research R3）
│   │   └── generated.go, model/models_gen.go  # gqlgen の生成コード
│   └── server/server.go             # ローダーのミドルウェア、本体の大きさの制限
├── gqlgen.yml                       # 2 つのシステムの型を model/sheet.go に結び付ける
└── go.mod                           # dataloadgen を追加

apps/web/
└── src/
    ├── app/
    │   └── sheets/
    │       ├── page.tsx             # 一覧と検索（新規）
    │       └── [id]/
    │           ├── page.tsx         # 閲覧（新規）
    │           └── not-found.tsx    # 見つからないとき（新規）
    ├── components/
    │   ├── layout/site-header.tsx   # 「シート一覧」のリンク
    │   └── sheets/                  # 検索のフォーム、一覧の行、ページ送り、閲覧の共通部分（新規）
    ├── lib/
    │   ├── api/server.ts            # queryApi を足す
    │   ├── auth/messages.ts         # 新しいエラーのコードの文言
    │   └── sheets/                  # 新規
    │       ├── search-params.ts     # URL のクエリの読み取りと URL の組み立て
    │       ├── systems.ts           # GameSystemId と GraphQL の enum の変換
    │       ├── format.ts            # 日時の表示（Asia/Tokyo）
    │       ├── get-sheet.ts         # cache() で包んだ getCharacterSheet（server-only）
    │       └── system-views.ts      # システム固有の部分と所有者の操作の部品の表
    └── gql/                         # GraphQL Code Generator の生成コード

docs/adr/0005-sheet-access-and-legacy-owner.md  # 削除の決定を追記
.specify/memory/constitution.md                  # シートの公開範囲に削除を足す（1.4.0）
```

**Structure Decision**: 既存のモノレポの構成に合わせる。api のシートの処理は `internal/sheet` に
まとめ、002-auth の `internal/auth` と同じく resolver から呼ぶ。DataLoader は、今後ほかの関連データ
（画像など）にも使えるよう `internal/loader` に分ける。web は `/sheets` 以下に画面を置き、各システムの
Spec が `system-views.ts` に部品を登録して閲覧画面に差し込み、`/sheets/new/<システム>` と
`/sheets/<ID>/edit` を足す形にする。

## 実装の順序

User Story の優先度に合わせて、次の順にする。細かいタスクは `/speckit-tasks` で分ける。

1. **基盤**: マイグレーション（`ja-x-icu` の確認を含む）、sqlc のクエリ、スキーマと `task generate`、
   `internal/sheet` の検証とシステムの変換、`internal/graph/model/sheet.go`、DataLoader と
   ミドルウェア、本体の大きさの制限、エラーのコード。
2. **US3（P1）作成**: `createCharacterSheet`。権限と入力の規則のテスト。
3. **US2（P1）閲覧**: `characterSheet`、`queryApi`、閲覧画面と見つからない表示。
4. **US1（P1）一覧と検索**: `characterSheets`（絞り込み・並び順・ページ）、DataLoader の回数の
   テスト、一覧の画面、検索のフォーム、ページ送り、ヘッダーのリンク。
5. **US4（P1）更新**: `updateCharacterSheet`（所有者とリビジョン）。
6. **US5（P2）複製**: `copyCharacterSheet`。
7. **US6（P2）削除**: `deleteCharacterSheet`、取り消しの手順の確認。
8. **仕上げ**: ADR 0005 と constitution の追記、README の更新（必要なら）、quickstart の全項目の確認。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

違反はない。
