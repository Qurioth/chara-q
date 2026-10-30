# CLAUDE.md

TRPG キャラクターシート管理アプリ「chara-q」のモノレポ。
原則は [.specify/memory/constitution.md](.specify/memory/constitution.md)、
技術選定の経緯は [docs/adr/](docs/adr/) を参照する。constitution と矛盾する作業はしない。

## よく使うコマンド

ツールは mise で入れる。タスクの入口は Taskfile だけ（`npm run` や `go` を直接叩く前に Taskfile を確認する）。

```sh
mise install          # ツールを mise.toml のバージョンで入れる（初回とバージョン変更時）
task setup            # .env の作成と依存パッケージのインストール
task db:up            # ローカルの Postgres と S3 互換ストレージ（RustFS）を起動
task db:migrate       # マイグレーションを適用
task generate         # gqlgen・sqlc・GraphQL Code Generator で生成コードを更新
task dev              # api（:8080）と web（:3000）を起動
task lint             # 静的チェック（Go と web）
task test             # テスト（DB を使うテストには db:up が必要）
task check:generated  # 生成コードがコミット済みの内容と一致するか確認
task db:new -- <名前> # マイグレーションファイルを作る
```

`task --list` ですべてのタスクを確認できる。

## ディレクトリ構成

```
schema/graphql/     GraphQL スキーマ（API の唯一の正）
apps/api/           Go（gqlgen + chi + pgx + sqlc）
  cmd/server/       エントリーポイント
  internal/graph/   resolver（*.resolvers.go、resolver.go）と生成コード
  internal/db/      sqlc の生成コード
  internal/config/  環境変数の読み込み
  internal/server/  ルーター
  db/migrations/    goose のマイグレーション
  db/queries/       sqlc 用の SQL
  tools/go.mod      goose 専用の modfile
apps/web/           Next.js（App Router）
  src/gql/          GraphQL Code Generator の生成コード
specs/              Spec Kit の設計書（機能ごと）
docs/adr/           アーキテクチャ決定記録
infra/              クラウドのセットアップ手順（手順書のみ）
```

## 守ること

- **生成コードは手で直さない。** 入力を直して `task generate` で再生成し、生成結果もコミットする。
  - gqlgen: `apps/api/internal/graph/generated.go`、`model/models_gen.go`
    （`*.resolvers.go` は実装を書く場所。生成時に実装は保持される）
  - sqlc: `apps/api/internal/db/`（`*_test.go` を除く）
  - GraphQL Code Generator: `apps/web/src/gql/`
- **API の変更はスキーマから。** `schema/graphql/*.graphqls` を変更してから `task generate`。
- **DB は SQL ファースト。** スキーマ変更は `task db:new` でマイグレーションを作る。ORM は使わない。
- **関連データは DataLoader で解決する**（N+1 禁止）。
- **シークレットを書かない。** `.env` はコミットしない。`.env.example` には例の値だけを書く。
- **ツールのバージョンは `mise.toml` で変える。** `package.json` の `packageManager` と
  `go.mod` の `go` も合わせる。
- **旧リポジトリ（`E:\develop\chara-q-old`）は参照専用。** 変更しない。
- **クラウドのリソースを作らない。** 課金が発生する操作や `git push` は、ユーザーの指示があるときだけ行う。
- web の Next.js は学習データより新しい可能性がある。コードを書く前に
  [apps/web/AGENTS.md](apps/web/AGENTS.md) の指示に従ってドキュメントを確認する。

## Spec Kit の進め方

機能は 1 つずつ、次の順で進める。設計書は `specs/<番号>-<機能名>/` に残る。

1. `/speckit-specify <機能の説明>` — 仕様を書く
2. `/speckit-clarify` — 曖昧な点を質問して決める（constitution の「未決定事項」に関わる機能では必須）
3. `/speckit-plan` — 実装計画（Constitution Check を含む）
4. `/speckit-tasks` — タスクに分解
5. `/speckit-implement` — 実装

Spec Kit のスクリプトは PowerShell 版（`.specify/scripts/powershell/`）を使う。
