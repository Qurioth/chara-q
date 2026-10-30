# インフラのセットアップ手順

本番環境を作るための手順書。**このリポジトリのコマンドやタスクはクラウドのリソースを作らない。**
ここに書いた手順は、本番環境の構築（Spec「本番環境の構築とデプロイ」）のときに手で実行する。

| サービス | 用途 | 手順 |
|---|---|---|
| Google Cloud | Cloud Run（API）、Artifact Registry、Secret Manager、予算アラート | [gcp.md](gcp.md) |
| Neon | Postgres | [neon.md](neon.md) |
| Cloudflare R2 | 画像の保存と配信 | [cloudflare-r2.md](cloudflare-r2.md) |
| Vercel | Next.js（Web） | [vercel.md](vercel.md) |

## 進める順番

1. **GCP のプロジェクトと予算アラート**（課金を有効にしたら、最初に予算アラートを設定する）
2. Neon のプロジェクト（接続文字列を Secret Manager に登録するため）
3. Cloudflare R2 のバケットと API トークン
4. GCP の残り（Artifact Registry、Workload Identity 連携、Secret Manager、Cloud Run）
5. GitHub Actions の変数を設定し、`deploy-api.yml` を手動で実行する
6. Vercel（API の URL が決まってから）
7. API の `CORS_ALLOWED_ORIGINS` に Vercel のドメインを追加する

## 共通の決まり

- リージョンは **シンガポール** にそろえる（Cloud Run は `asia-southeast1`、Neon は
  `AWS Asia Pacific (Singapore)`）。
- シークレットはリポジトリに置かない。API は Secret Manager、Web は Vercel の環境変数を使う。
- 常時起動のリソースは置かない（Cloud Run の最小インスタンス数は 0）。
