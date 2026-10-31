# Implementation Plan: 認証（招待コードでの登録、ログイン・ログアウト）

**Branch**: `002-auth` | **Date**: 2026-10-31 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/002-auth/spec.md`

## Summary

管理者が発行した招待コードでユーザー登録し、ログイン ID とパスワードでログイン・ログアウト
できるようにする。ブラウザは web とだけ通信し、`/graphql` は Next.js の `rewrites` で api に中継する。
api（Go + gqlgen）は DB のセッションでログイン状態を持ち、`__Host-` 接頭辞の Cookie（`HttpOnly`・
`Secure`・`SameSite=Lax`）にランダムなセッション ID だけを入れる。パスワードは Argon2id、招待
コード・再設定コード・セッション ID は SHA-256 のハッシュだけを保存する。ログインの失敗は
Postgres に記録し、5 回続けて失敗したら 15 分ロックする。

web はヘッダーの右端にログイン状態を出し、登録・ログイン・アカウントの設定・パスワードの再設定・
招待コードの発行（管理者だけ）の画面を持つ。Server Components は、セッションの Cookie だけを
付けて api を直接呼ぶ。最初の管理者の作成、再設定コードの発行、ユーザーの一覧・無効化・有効化は、
Taskfile から呼ぶ Go のコマンドで行う。ブラウザが api に直接つながなくなるため、api の CORS の許可を
外す。

実装の最初に、中継を通しても `Set-Cookie` がブラウザに届き、Cookie が api に届くことを確かめる
（FR-004）。

## Technical Context

**Language/Version**: Go 1.27.3（`apps/api`）、TypeScript 5（`apps/web`）、Node.js 24.22.2、
pnpm 12.16.1（`mise.toml`。変更しない）

**Primary Dependencies**:
- api: gqlgen 0.17.98、chi 5.5.0、pgx 5.11.0、sqlc 1.31.1、goose（`tools/go.mod`）。追加は
  `golang.org/x/crypto`（Argon2id）、`golang.org/x/term`。`github.com/go-chi/cors` を外す。
- web: Next.js 16.4.3（App Router、`cacheComponents: true`）、React 19.3.0、urql 5.0.4、
  GraphQL Code Generator（client preset）、shadcn/ui（Base UI）。追加は `server-only` と
  shadcn/ui の `dropdown-menu`。
- 詳細は [research.md](research.md) の R16。

**Storage**: Postgres 18（ローカルは `compose.yaml`、本番は Neon）。表を 5 つ追加する
（`users`、`sessions`、`invitations`、`password_reset_codes`、`auth_failures`）。
詳細は [data-model.md](data-model.md)。

**Testing**: api は `go test`（テーブル駆動、DB を使うテストは `TEST_DATABASE_URL` の実際の Postgres、
HTTP は `httptest`）。web は Vitest + Testing Library + `user-event`（既存）。中継を通した Cookie と
ブラウザごとの動作は手で確かめる（[quickstart.md](quickstart.md)）。

**Target Platform**: api は Cloud Run（Linux コンテナ）、web は Vercel。ブラウザは最新の Chrome・
Edge・Safari・Firefox とスマートフォンの標準ブラウザ。

**Project Type**: Web アプリ（モノレポの `apps/api` と `apps/web`）

**Performance Goals**: ログイン・登録の応答は Argon2id の照合を含めて 1 秒以内（SC-002 の 30 秒に
十分な余裕）。ログインしている画面でも、ヘッダーのログイン状態は外枠と同じ応答の中で表示する
（SC-009）。

**Constraints**:
- `cacheComponents` のため、`cookies()` を読むコンポーネントは `<Suspense>` の中に置き、ルート
  レイアウトの一番上では読まない（research R3）。
- Cookie の属性は本番とローカルで変えない（research R2）。
- ミューテーションは `POST` かつ `application/json` だけ（research R4）。
- 常時起動の仕組みは置かない。期限切れの行の掃除はログインのときに行う（research R9）。

**Scale/Scope**: 利用者は数十人規模。画面 5 つ（ログイン、登録、パスワードの再設定、アカウントの
設定、招待コードの発行）とヘッダーのアカウントのメニュー、GraphQL のクエリ 1 つとミューテーション 7 つ、
管理者のコマンド 5 つ。

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則・制約 | 判定 | 根拠 |
|---|---|---|
| I. スキーマファースト | 適合 | `schema/graphql/auth.graphqls` を追加してから `task generate` する。web と api はどちらもこのファイルを直接読む |
| II. 生成コードはコミットする | 適合 | gqlgen・sqlc・GraphQL Code Generator の出力をスキーマ・SQL と同じコミットに入れる |
| III. キャラクターシートのモデル | 該当なし | シートは扱わない。所有者になるユーザーの表を用意するだけ |
| IV. SQL ファースト | 適合 | 表は goose のマイグレーション、クエリは `db/queries/*.sql` を sqlc で生成する。ORM は使わない |
| V. N+1 禁止 | 適合 | Viewer はリクエストごとに 1 回、結合した 1 つのクエリで読む。一覧の中で関連データを引く resolver はない（research R13） |
| VI. テスト | 適合 | DB を使うテストは実際の Postgres で行い、Go のテストはテーブル駆動にする。`task test`・`task lint` を通す |
| VII. Spec 駆動開発 | 適合 | specify → clarify（5 問）→ plan の順。ADR 0006 の「後続の Spec で決めること」はすべて clarify で決めた |
| 技術スタック・新しい基盤技術には ADR | 適合 | 基盤技術は変えない。追加の依存は小さなライブラリだけ（research R16）。ADR 0006 に clarify の決定を追記する |
| UI: Tailwind CSS、shadcn/ui は操作の伴う部品だけ | 適合 | 追加する shadcn/ui の部品は `dropdown-menu` だけ。画面は既存の共通入力部品で組む |
| 認証（ADR 0006） | 適合 | 招待コード・ログイン ID・パスワード、メールなし、rewrites での中継、DB セッション、`HttpOnly`・`Secure`・`SameSite=Lax`、JWT なし、平文で保存しない |
| シートの公開範囲 | 該当なし | シートの操作はない。シートの更新での所有者の確認は、シートの共通基盤の機能で行う |
| シークレット | 適合 | 新しいシークレットはない。`API_URL` は秘密ではないが、ブラウザに出さないため `NEXT_PUBLIC_` を付けない |
| コスト | 適合 | 新しいクラウドのリソースはない。常時起動の仕組みも置かない |
| ツールとタスク | 適合 | ツールのバージョンは変えない。管理者のコマンドの入口は Taskfile の `admin:*` だけ |
| 旧システム | 該当なし | 旧リポジトリは参照しない |

違反はないため、Complexity Tracking は空のまま。

### Phase 1 後の再確認

設計（[data-model.md](data-model.md)、[contracts/](contracts/)、[quickstart.md](quickstart.md)）を
反映して再確認した。判定は上の表から変わらない。

- GraphQL の型は [contracts/graphql.md](contracts/graphql.md) のスキーマの案から始め、ログイン ID は
  `Viewer`（本人）にだけ置く。公開する `User` の型は、シートの所有者を表示する機能で追加する
  （FR-034、SC-008）。
- DB の `CHECK` 制約と一意の索引で、入力の規則とログイン ID の一意性を、コマンドからの操作でも守る。
- constitution の「認証」の節は、パスワード・招待コード・セッション ID を平文で保存しないと定めて
  いる。この機能では再設定コードも同じく扱う。constitution の文言に再設定コードを足すかどうかは、
  ADR 0006 の追記と一緒に決める（文言の修正のため PATCH）。

## Project Structure

### Documentation (this feature)

```text
specs/002-auth/
├── spec.md
├── plan.md              # このファイル
├── research.md          # Phase 0: 技術の選定と調査
├── data-model.md        # Phase 1: 表、制約、状態の移り変わり
├── quickstart.md        # Phase 1: 動作確認の手順
├── contracts/
│   ├── graphql.md       # GraphQL のスキーマの案とエラーのコード
│   ├── http-and-cookie.md # 中継、Cookie、受け付ける HTTP リクエスト、環境変数
│   ├── admin-commands.md  # Taskfile の admin:* とコマンドの入出力
│   └── web-routes.md      # 画面、ヘッダー、アクセスの制御
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks で作る
```

### Source Code (repository root)

```text
schema/graphql/
└── auth.graphqls                    # Viewer、認証のミューテーション（新規）

apps/api/
├── cmd/
│   ├── server/main.go               # auth のサービスとミドルウェアを組み込む
│   └── admin/main.go                # 管理者のコマンド（新規）
├── db/
│   ├── migrations/00002_auth.sql    # 5 つの表（新規、task db:new で作る）
│   └── queries/                     # sqlc 用の SQL（表ごとに新規: users.sql、sessions.sql、
│                                    #   invitations.sql、password_reset_codes.sql、auth_failures.sql）
├── internal/
│   ├── auth/                        # 認証の処理（新規）
│   │   ├── password.go              # Argon2id（PHC 形式）
│   │   ├── token.go                 # コードとセッション ID の生成・正規化・ハッシュ
│   │   ├── validate.go              # ログイン ID・パスワード・表示名の規則
│   │   ├── service.go               # 登録、ログイン、ロック、再設定、無効化など
│   │   ├── session.go               # セッションの読み込みと延長
│   │   ├── middleware.go            # Cookie から Viewer を読み、context に入れる
│   │   ├── cookie.go                # __Host- の Cookie の書き込み・削除
│   │   └── *_test.go
│   ├── config/config.go             # CORS_ALLOWED_ORIGINS を外す
│   ├── db/                          # sqlc の生成コード（auth.sql.go などが増える）
│   ├── graph/
│   │   ├── resolver.go              # auth のサービスを持たせる
│   │   ├── auth.resolvers.go        # gqlgen が作る。実装を書く
│   │   └── generated.go, model/     # gqlgen の生成コード
│   └── server/server.go             # CORS を外し、認証のミドルウェアを入れる
├── sqlc.yaml                        # uuid・timestamptz の型の指定
└── go.mod                           # x/crypto、x/term を追加、go-chi/cors を外す

apps/web/
├── next.config.ts                   # /graphql の rewrites
├── .env.example                     # NEXT_PUBLIC_API_URL を API_URL に置き換える
├── package.json                     # server-only、dev:https のスクリプト
└── src/
    ├── app/
    │   ├── layout.tsx               # ヘッダーにアカウントのメニュー（Suspense の中）
    │   ├── login/page.tsx           # ログイン（新規）
    │   ├── register/page.tsx        # 登録（新規）
    │   ├── reset-password/page.tsx  # パスワードの再設定（新規）
    │   ├── account/page.tsx         # アカウントの設定（新規）
    │   └── admin/invitations/page.tsx # 招待コードの発行（新規、管理者だけ）
    ├── components/
    │   ├── auth/                    # 各画面のフォーム、アカウントのメニュー（新規）
    │   └── ui/dropdown-menu.tsx     # shadcn/ui から追加
    ├── lib/
    │   ├── api/server.ts            # Server Components から api を呼ぶ（server-only）
    │   ├── auth/viewer.ts           # getViewer()、requireViewer()、requireAdmin()
    │   ├── auth/rules.ts            # 入力の規則（api と同じ）
    │   ├── auth/messages.ts         # エラーのコードから文言へ
    │   ├── auth/safe-next.ts        # ログイン後の戻り先の検証
    │   └── urql-provider.tsx        # 接続先を /graphql に変える
    └── gql/                         # GraphQL Code Generator の生成コード

Taskfile.yml                         # admin:*、dev:web:https を追加
.github/workflows/ci-web.yml         # build の環境変数を API_URL にする
.github/workflows/deploy-api.yml     # CORS_ALLOWED_ORIGINS を外す
infra/gcp.md, infra/vercel.md        # 環境変数の手順を更新
apps/api/.env.example                # CORS_ALLOWED_ORIGINS を外す
docs/adr/0006-authentication.md      # 「後続の Spec で決めること」の決定を追記
README.md                            # 管理者の作り方、admin:* のタスク
```

**Structure Decision**: 既存のモノレポの構成（`apps/api` と `apps/web`、スキーマは
`schema/graphql/`）に合わせる。api の認証の処理は `internal/auth` にまとめ、resolver
（`internal/graph`）と管理者のコマンド（`cmd/admin`）の両方から使う。こうすると、コマンドと
GraphQL で、入力の規則・ハッシュ・セッションの削除の処理が 1 か所になる。

## 実装の順序

User Story の優先度と、FR-004（中継の確認を最初に行う）に合わせて、次の順にする。
細かいタスクは `/speckit-tasks` で分ける。

1. **中継と Cookie の確認（FR-004）**: rewrites と `API_URL`、urql の接続先、仮の Cookie を返す
   最小の処理で、Chrome・Firefox・Safari の Cookie の往復を確かめる。届かなければ、ここで止めて
   方針を見直す（research R1 の代案）。
2. **基盤**: マイグレーション、sqlc のクエリ、`internal/auth` のハッシュ・トークン・検証、
   認証のミドルウェア、CORS の撤去、CSRF の条件のテスト。
3. **US1（P1）**: 管理者の作成のコマンド、`viewer`・`login`・`logout`、ヘッダーのアカウントのメニュー、
   ログイン画面。
4. **US2（P1）**: `issueInvitation` と招待コードの発行画面、`register` と登録画面。
5. **US3（P2）**: ロック。
6. **US4（P2）**: アカウントの設定（表示名とパスワードの変更）。
7. **US5（P2）**: 再設定コードのコマンドと再設定の画面。
8. **US6（P2）**: 無効化・有効化・ユーザーの一覧のコマンド。
9. **仕上げ**: 環境変数と手順書、ADR 0006 の追記、README、quickstart の全項目の確認。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

違反はない。
