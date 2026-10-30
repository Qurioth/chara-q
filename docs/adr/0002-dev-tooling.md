# ADR 0002: 開発ツールの管理（mise と Taskfile、Turborepo は採用しない）

- ステータス: 採用
- 日付: 2026-10-30

## 背景

モノレポには Go（apps/api）と Node.js（apps/web）が同居し、gqlgen・sqlc・goose・
golangci-lint・GraphQL Code Generator などのツールを使う。開発は主に Windows（PowerShell）で行い、
CI は Linux で動く。手元と CI でツールのバージョンをそろえ、どちらでも同じコマンドで作業したい。

## 決定

1. **ツールのバージョンは mise で固定する。**
   - ルートの `mise.toml` の `[tools]` に node・pnpm・go・uv・task・golangci-lint・sqlc を書く。
   - CI も `jdx/mise-action` で同じ `mise.toml` を使う。
   - `package.json` の `packageManager` と `go.mod` の `go` は `mise.toml` と同じバージョンにする。
   - corepack は使わない（Node.js 25 以降は同梱されないため）。
2. **タスクの入口は Taskfile だけにする。** mise の tasks は使わない。
   - Task は組み込みのシェル（mvdan/sh）と coreutils で動くため、Windows でも同じ書き方で動く。
     bash 専用の書き方は使わない。
3. **Go のツールは `tool` ディレクティブで管理する。**
   - gqlgen は本体の `apps/api/go.mod` で管理する（実行時ライブラリと同じバージョンにするため）。
   - goose は `apps/api/tools/go.mod` で管理し、`go tool -modfile=tools/go.mod goose` で実行する。
     goose のコマンドは全 DB ドライバ（MySQL、ClickHouse、MSSQL など）に依存しており、本体の
     go.mod に入れると使わない依存が大量に増えるため。
   - sqlc と golangci-lint は mise で公式バイナリを入れる。sqlc は cgo の都合で Windows では
     `go tool` 経由のビルドが通らないことがあり、golangci-lint は公式に `go tool` での導入を
     推奨していないため。
4. **Turborepo などのビルドオーケストレーターは採用しない。**

## Turborepo を採用しない理由

- Turborepo の並列実行やキャッシュが効くのは pnpm workspace 内の JS パッケージだけで、
  現状は JS アプリが `apps/web` の 1 つしかない。Go の api は対象外になる。
- タスクの入口が Taskfile と二重になる。
- Vercel で関係のない変更のビルドを省く設定は、Turborepo なしで実現できる（infra/vercel.md）。

### 採用を考え直す条件

`packages/*`（共有 UI、共有の型など）が増え、パッケージ間のビルドの依存関係やキャッシュが
実際に効くようになったとき。その場合は新しい ADR を書いて追加する。

## 結果

- `mise install` だけで必要なツールがそろい、手元と CI のバージョンが一致する。
- 新しいツールを追加するときは、まず `mise.toml` に追加できるかを検討する。
- Windows では、以前から入っている Go などが PATH 上で mise のシムより先に見つかることがある。
  その場合も `go.mod` の `go` の指定により、Go は必要なバージョンへ自動で切り替わる。
