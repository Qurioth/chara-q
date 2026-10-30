# chara-q

TRPG キャラクターシート管理アプリのモノレポ。

- 対応予定のシステム: ゆうやけこやけ、駆け出しアイドルRPG ビギニングアイドル
- 原則: [.specify/memory/constitution.md](.specify/memory/constitution.md)
- アーキテクチャ: [docs/adr/0001-architecture.md](docs/adr/0001-architecture.md)

## 構成

```
ブラウザ
 ├─ Next.js（Vercel）                         apps/web
 │    └─ GraphQL → Go API（Cloud Run）         apps/api
 │                    └─ Postgres（Neon）
 └─ 画像の直接アップロード・表示 → Cloudflare R2
```

| ディレクトリ      | 内容                                                      |
| ----------------- | --------------------------------------------------------- |
| `schema/graphql/` | GraphQL スキーマ（API の唯一の正）                        |
| `apps/api/`       | Go（gqlgen + chi + pgx + sqlc、マイグレーションは goose） |
| `apps/web/`       | Next.js（App Router、urql、GraphQL Code Generator）       |
| `specs/`          | Spec Kit の設計書                                         |
| `docs/adr/`       | アーキテクチャ決定記録                                    |
| `infra/`          | クラウドのセットアップ手順（手順書のみ）                  |

## セットアップ

### 必要なもの

- [Git](https://git-scm.com/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)（ローカルの Postgres と S3 互換ストレージ用）
- [mise](https://mise.jdx.dev/)（Node.js、pnpm、Go、Task などはすべて mise で入れる）

Windows では、PowerShell のプロファイルに `mise activate pwsh` を追加するか、
mise のシム（`%LOCALAPPDATA%\mise\shims`）を PATH に追加しておく。

### 手順

```sh
mise install     # mise.toml に固定したツールを入れる
task setup       # .env の作成と依存パッケージのインストール
task db:up       # Postgres と S3 互換ストレージ（RustFS）を起動
task db:migrate  # マイグレーションを適用
task dev         # api（http://localhost:8080）と web（http://localhost:3000）を起動
```

http://localhost:3000 に「API: ok」と表示されれば、web → api → DB までつながっている。
開発時は http://localhost:8080/playground で GraphQL Playground が使える。

ローカル用の接続先は `apps/api/.env` と `apps/web/.env`（`task setup` が `.env.example` から作る）。

| サービス             | URL                   | 認証情報（ローカル専用）                               |
| -------------------- | --------------------- | ------------------------------------------------------ |
| Postgres             | `localhost:5432`      | `chara_q` / `chara_q`、DB `chara_q`                    |
| RustFS（S3 API）     | http://localhost:9000 | `chara_q` / `chara_q_secret`、バケット `chara-q-local` |
| RustFS（コンソール） | http://localhost:9001 | 同上                                                   |

## Taskfile のコマンド

| コマンド                      | 内容                                                           |
| ----------------------------- | -------------------------------------------------------------- |
| `task setup`                  | .env の作成と依存パッケージのインストール                      |
| `task dev`                    | api と web を起動（`dev:api` / `dev:web` で個別に起動）        |
| `task generate`               | gqlgen・sqlc・GraphQL Code Generator で生成コードを更新        |
| `task check:generated`        | 再生成してコミット済みの内容と差分がないか確認                 |
| `task lint`                   | gofmt / goimports、go vet、golangci-lint、ESLint、型チェック   |
| `task fmt`                    | Go のコードを整形                                              |
| `task test`                   | api のテスト（DB を使うテストには `task db:up` が必要）        |
| `task build`                  | api と web をビルド                                            |
| `task docker:build`           | api の Docker イメージをビルド                                 |
| `task db:up` / `task db:down` | ローカルの Postgres と RustFS を起動 / 停止                    |
| `task db:migrate`             | マイグレーションを適用                                         |
| `task db:status`              | マイグレーションの適用状況を表示                               |
| `task db:reset`               | マイグレーションをすべて戻してから適用し直す（データは消える） |
| `task db:new -- <名前>`       | マイグレーションファイルを作る                                 |

`task --list` ですべてのタスクを確認できる。

## 開発の進め方

機能は Spec Kit で、設計書を作りながら 1 つずつ実装する。

1. `/speckit-specify <機能の説明>` — 仕様
2. `/speckit-clarify` — 曖昧な点を決める
3. `/speckit-plan` — 実装計画
4. `/speckit-tasks` — タスク分解
5. `/speckit-implement` — 実装

生成コード（`apps/api/internal/graph/generated.go`、`apps/api/internal/db/`、`apps/web/src/gql/`）は
手で編集せず、`task generate` で再生成してコミットする。CI で差分がないことを確認している。

## 未決定事項

次は該当する機能の Spec で `/speckit-clarify` を使って決める。それまでは推測で実装しない。

- **WebP 変換の実行場所**: ブラウザ / Go / Cloudflare Worker

UI ライブラリは Tailwind CSS と shadcn/ui に決定した（[ADR 0004](docs/adr/0004-ui-library.md)）。
シートの公開範囲と既存シートの所有者も決定済み（[ADR 0005](docs/adr/0005-sheet-access-and-legacy-owner.md)）。
一覧・検索・閲覧は誰でもでき、更新は所有者だけができる。既存シートはシステム管理者ユーザーの所有にする。
認証も決定済み（[ADR 0006](docs/adr/0006-authentication.md)）。ログイン ID・パスワード・管理者が発行した
招待コードで登録し、メールアドレスは持たない。api へは Vercel から中継し、ログイン状態は DB セッションで持つ。

## Spec Kit で進める予定の機能

1. UI 基盤（Tailwind CSS と shadcn/ui の導入、ヘッダー・フッター、ダークモード、フォント、共通の入力部品）
2. 認証（招待コードでの登録、ログイン、Vercel 経由の中継、DB セッション、管理者用のコマンド）
3. キャラクターシートの共通基盤（`interface CharacterSheet`、DB テーブル、一覧・検索・取得・作成・更新）
4. 本番環境の構築とデプロイ（Cloud Run、Artifact Registry、予算アラート、Neon、Vercel）
5. ゆうやけこやけのシート
6. ビギニングアイドルのシート（サイコロ・フィクションの特技表を含む）
7. ココフォリア出力
8. 画像アップロード（R2、署名付き URL、WebP）
9. 既存データの移行（既存シートはシステム管理者ユーザーの所有にする）

## CI / デプロイ

| ワークフロー     | 実行条件                                                  | 内容                                                                       |
| ---------------- | --------------------------------------------------------- | -------------------------------------------------------------------------- |
| `ci-api.yml`     | `apps/api/**`・`schema/**`・`mise.toml` の変更            | gofmt、go vet、golangci-lint、go test、生成コードの差分確認、Docker ビルド |
| `ci-web.yml`     | `apps/web/**`・`schema/**`・pnpm 関連・`mise.toml` の変更 | codegen の差分確認、lint、型チェック、build                                |
| `deploy-api.yml` | 手動（`workflow_dispatch`）                               | Artifact Registry へ push、Cloud Run へデプロイ                            |

本番環境の準備は [infra/README.md](infra/README.md) を参照。
