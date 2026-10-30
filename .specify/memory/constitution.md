# chara-q Constitution

## Core Principles

### I. スキーマファースト

- `schema/graphql/*.graphqls` が API の唯一の正である。API の型・クエリ・ミューテーションの
  変更は、必ずこのスキーマの変更から始める（MUST）。
- Go 側の gqlgen と Web 側の GraphQL Code Generator は、どちらもこのディレクトリを直接読む
  （MUST）。スキーマのコピーや、アプリ側でのスキーマの再定義は禁止する。

理由: フロントエンドとバックエンドが同じ定義から型を得ることで、両者のずれを型エラーとして
早く検出できる。

### II. 生成コードはコミットする

- gqlgen・sqlc・GraphQL Code Generator の出力はリポジトリにコミットする（MUST）。
- 生成コードは手で編集しない（MUST）。変更が必要な場合は、入力（スキーマ・SQL・設定）を直して
  再生成する。
- CI で再生成を行い、差分が出たら失敗させる（MUST）。

理由: レビューで生成結果の変化を確認でき、生成ツールなしでもビルドできる。CI のチェックで
「入力を変えたのに再生成し忘れた」状態を防ぐ。

### III. キャラクターシートのモデル

- GraphQL では `interface CharacterSheet` に全システム共通の項目（名前・プレイヤー名・年齢・
  性別・メモなど）を置き、システムごとの型がこれを実装する（MUST）。
- システム固有の部分は、最初は `JSON` スカラーで返す。型として切り出すのは、Spec で必要性を
  確認してからにする（SHOULD）。
- DB では、検索や一覧に使う共通項目（所有者・システム・キャラクター名・更新日時など）を通常の
  カラムにし、シート本体を JSONB に保存する（MUST）。

理由: システムの追加や項目の変更に、スキーマとテーブルの大きな変更なしで対応しつつ、
検索に必要な項目だけはインデックスを効かせられる。

### IV. DB アクセスは SQL ファースト

- SQL は手で書き、sqlc で型付きの Go コードを生成し、pgx で実行する（MUST）。ORM は使わない。
- スキーマ変更はマイグレーションファイル（`apps/api/db/migrations/`）で管理する（MUST）。
  本番やローカルの DB を手作業で変更しない。

理由: 実行される SQL がそのままレビュー対象になり、型安全性も保てる。

### V. N+1 禁止

- GraphQL の resolver で関連データを解決するときは、必ず DataLoader を通す（MUST）。
  resolver からループ内で個別にクエリを発行してはならない。

理由: GraphQL はネストしたクエリでクエリ数が爆発しやすく、Neon の接続数と応答時間に
直接響くため。

### VI. テスト

- Go のテストはテーブル駆動で書く（SHOULD）。
- DB を触るテストは、モックではなく実際の Postgres に対して実行する（MUST）。ローカルは
  Docker（`compose.yaml`）、CI は service コンテナを使う。
- `task test` と `task lint` が通らない変更はマージしない（MUST）。

理由: SQL と JSONB の扱いはモックでは検証できず、本番と同じエンジンで確かめる必要がある。

### VII. Spec 駆動開発

- 機能は `/speckit-specify → /speckit-clarify → /speckit-plan → /speckit-tasks →
  /speckit-implement` の順で進め、`specs/` に設計書を残す（MUST）。
- 未決定事項に関わる機能は、`/speckit-clarify` で決めてから plan に進む（MUST）。推測で実装
  しない。

理由: 設計の判断と経緯を残し、後から読める状態を保つため。

## 技術・運用上の制約

- **決定済みの技術スタック**: Next.js（App Router、Vercel）、Go + gqlgen + chi（Cloud Run）、
  Neon（Postgres）、Cloudflare R2、GitHub Actions。これらの変更や新しい基盤技術の追加には
  ADR（`docs/adr/`）を必須とする（MUST）。
- **リージョン**: Cloud Run と Neon は同じシンガポール（Cloud Run は `asia-southeast1`）に置く
  （MUST）。
- **シークレット**: シークレットや実際の接続文字列をリポジトリに置かない（MUST）。ローカルは
  gitignore 済みの `.env`、本番は Secret Manager と Vercel の環境変数を使う。`.env.example`
  には例の値だけを書く。
- **コスト**: GCP には必ず予算アラートを設定する（MUST）。常時起動のリソースは置かず、
  Cloud Run の最小インスタンス数は 0 にする（MUST）。
- **画像**: 画像はブラウザから R2 へ署名付き URL で直接アップロードし、DB には R2 のキーだけを
  保存する（MUST）。
- **ツールとタスク**: 開発ツールのバージョンは `mise.toml` で固定し、手元と CI で同じものを
  使う（MUST）。タスクの入口は `Taskfile.yml` だけにし、mise の tasks は使わない（MUST）。
  Turborepo などのビルドオーケストレーターは、JS パッケージが増えて必要になるまで入れない。
- **旧システム**: `E:\develop\chara-q-old` 配下の旧リポジトリは参照専用とし、変更しない（MUST）。

## 未決定事項

次の事項は意図的に未決定とする。関連する機能の Spec で `/speckit-clarify` を使って決め、
決まったら本ファイルと ADR を更新する。それまでは推測で実装しない（MUST）。

- **認証の詳細**: ログイン方法、トークンの保存場所（Go が JWT を発行し、Next.js はそれを渡す
  だけ、という大枠のみ決定済み）
- **WebP 変換の実行場所**: ブラウザ / Go / Cloudflare Worker のいずれか
- **UI ライブラリ**: 旧システムの MUI と Tailwind を引き継ぐかどうか

## Governance

- 本 constitution は、他の開発ルールや慣習より優先する。`CLAUDE.md` は日々の作業手順の
  ガイドであり、本ファイルと矛盾してはならない。
- 改定は PR で行い、変更理由を記載する。技術選定に関わる改定は ADR を伴う。
- バージョンはセマンティックバージョニングに従う。原則の削除・再定義は MAJOR、原則や節の追加・
  大きな拡張は MINOR、表現の修正は PATCH とする。
- `/speckit-plan` の Constitution Check で各原則への適合を確認する。違反が必要な場合は、
  plan の Complexity Tracking に理由と、より単純な代替案を採らなかった理由を書く。

**Version**: 1.0.0 | **Ratified**: 2026-10-29 | **Last Amended**: 2026-10-29
