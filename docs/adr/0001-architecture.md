# ADR 0001: 再構築後のアーキテクチャ

- ステータス: 採用
- 日付: 2026-10-30

## 背景

chara-q を 1 つのモノレポで構成し、型の安全性、検索のしやすさ、運用コストを重視した
アーキテクチャを決める。

## 決定

```
ブラウザ
 ├─ Next.js（Vercel）
 │    └─ GraphQL → Go API（Cloud Run／asia-southeast1）
 │                    └─ Postgres（Neon／シンガポール）
 └─ 画像の直接アップロード・表示 → Cloudflare R2
```

- **フロントエンド**: Next.js（App Router）を Vercel でホストする。GraphQL Code Generator で
  クエリから型を生成し、クライアントは urql（または Server Components からの fetch）を使う。
- **バックエンド**: Go + gqlgen（スキーマファースト）、ルーターは chi。
  - シートは `interface CharacterSheet` を軸に、システムごとの型で実装する。
  - 最初は共通項目だけを型にし、残りは `JSON` スカラーで返す。
  - 関連データは DataLoader で解決し、N+1 を避ける。
- **API の契約**: `schema/graphql/*.graphqls` を唯一の正とし、gqlgen と GraphQL Code Generator の
  両方がここを読む。
- **認証**: ~~Go 側で JWT を発行し、Next.js はそれを渡すだけの単純な形から始める。~~
  → [ADR 0006](0006-authentication.md) で置き換え（招待コードでの登録、Vercel 経由の中継、DB セッション）。
- **DB**: Neon（Postgres）。
  - 検索に使う共通項目（所有者・システム・キャラクター名・更新日時など）は通常のカラム、
    シート本体は JSONB に保存する。
  - アクセスは sqlc + pgx。ORM は使わない。
- **インフラ**: Cloud Run（asia-southeast1、DB と同じリージョン）。イメージは Artifact Registry、
  デプロイは GitHub Actions（Workload Identity 連携）。予算アラートを必ず設定する。
- **画像**: Cloudflare R2。Go が署名付き URL を発行し、ブラウザから R2 へ直接アップロードする。
  DB には R2 のキーだけを保存する。WebP 変換・縮小を行う。公開画像は CDN 経由、非公開画像は
  署名付き URL で表示する。

## 未決定事項

次は後続の Spec で `/speckit-clarify` を使って決める。

- ~~認証の詳細（ログイン方法、トークンの保存場所）~~ → [ADR 0006](0006-authentication.md) で決定
- WebP 変換をどこで行うか（ブラウザ / Go / Worker）
- ~~UI ライブラリ~~ → [ADR 0004](0004-ui-library.md) で決定（Tailwind CSS と shadcn/ui）

## 結果

- スキーマから両側の型を生成するため、フロントエンドとバックエンドのずれを型エラーで検出できる。
- JSONB とカラムの併用で、システムごとの差を吸収しつつ、検索にはインデックスを使える。
- Cloud Run の最小インスタンス数 0 と Neon のスケールトゥゼロで、利用がないときの費用を抑えられる。
  その代わり、コールドスタートの遅延を受け入れる。
- 既存データの移行が必要になる（後続の Spec で扱う）。
