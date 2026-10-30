# Cloudflare R2（画像）

画像はブラウザから R2 へ直接アップロードし、DB には R2 のキーだけを保存する。
Go の API が署名付き URL を発行する。詳細な設計は「画像アップロード」の Spec で決める。

## 1. バケット

| バケット | 用途 | 公開 |
|---|---|---|
| `chara-q-public` | 公開してよい画像（公開シートの立ち絵など） | カスタムドメインで公開（CDN 経由） |
| `chara-q-private` | 非公開の画像 | 公開しない。表示は署名付き URL |

- ロケーションヒントは **Asia-Pacific (APAC)** にする。
- 1 つのバケットにまとめてキーの接頭辞で分ける案もある。どちらにするかは Spec で決める。

## 2. 公開用ドメイン

- `chara-q-public` に Cloudflare 管理下のカスタムドメイン（例: `img.<ドメイン>`）を接続する。
- `r2.dev` のサブドメインは開発用でレート制限があるため、本番では使わない。

## 3. CORS（ブラウザからの直接アップロード用）

両方のバケットに次のような CORS を設定する。

```json
[
  {
    "AllowedOrigins": ["https://<Vercel のドメイン>", "http://localhost:3000"],
    "AllowedMethods": ["GET", "PUT", "HEAD"],
    "AllowedHeaders": ["Content-Type"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3600
  }
]
```

- 本番のバケットに `localhost` を許可するかは Spec で決める（本番と開発でバケットを分ける場合は不要）。

## 4. API トークン

- R2 の「API トークンを管理」から、**対象バケットに限定した** オブジェクトの読み書き権限の
  トークンを作る。
- 発行されたアクセスキー ID とシークレットは Secret Manager に登録し、Cloud Run に渡す
  （[gcp.md](gcp.md) の 6）。リポジトリには書かない。
- S3 互換エンドポイントは `https://<アカウント ID>.r2.cloudflarestorage.com`。

## ローカル開発

ローカルでは `compose.yaml` の RustFS（S3 互換）を使う。エンドポイントは `http://localhost:9000`、
バケットは `chara-q-local`（[ADR 0003](../docs/adr/0003-local-middleware.md)）。
