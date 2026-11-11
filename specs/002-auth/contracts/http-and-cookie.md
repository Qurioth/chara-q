# Contract: 通信経路、Cookie、HTTP の受け付け条件、環境変数

## 経路

```text
ブラウザ ──POST /graphql──▶ web（Next.js rewrites）──▶ api ${API_URL}/graphql
   ▲                                                       │
   └──────────── Set-Cookie（web のオリジンの Cookie になる）◀┘

ブラウザ ──GET /（画面）──▶ web（Server Components）
                              │ cookies() からセッションの Cookie だけを取り出す
                              └──POST ${API_URL}/graphql（Cookie: __Host-chara_q_session=…）──▶ api
```

- ミューテーションは、ブラウザから rewrites を通して送る（Server Components と Server Actions
  からは送らない）。api の `Set-Cookie` がブラウザに届くのは、この経路だけ。
- Server Components は `viewer` のクエリだけを送る。応答の `Set-Cookie` は使わない
  （セッションの延長は DB の中で行う。research R2、R10）。

## Cookie

| 項目 | 値 |
|---|---|
| 名前 | `__Host-chara_q_session` |
| 値 | `crypto/rand.Text()` の 26 文字（research R7） |
| 属性 | `Path=/; HttpOnly; Secure; SameSite=Lax`。`Domain` は付けない |
| `Max-Age` | 34560000（400 日）。有効かどうかは DB の `expires_at` で決める |
| 消すとき | 同じ名前・属性で `Max-Age=0` |

- 値が不正・セッションが存在しない・期限切れ・ユーザーが無効のときは、ログインしていない
  扱いにする。rewrites を通したリクエストなら、api は Cookie も消す。

## api が受け付ける `/graphql` のリクエスト

| メソッド | Content-Type | 受け付ける操作 |
|---|---|---|
| `POST` | `application/json`（`charset` などの引数は可） | クエリとミューテーション |
| `GET` | なし | クエリだけ（ミューテーションは「GET requests only allow query operations」） |
| `OPTIONS` | なし | 空の応答（CORS の許可のヘッダーは返さない） |
| それ以外 | — | 拒否（gqlgen の transport が受け付けない） |

- `POST` で `text/plain`・`application/x-www-form-urlencoded`・`multipart/form-data` は拒否する。
- どの応答にも `Access-Control-Allow-Origin` を付けない（CORS の許可を外す）。
- 上の条件は、api の HTTP のテストで確かめる（research R4）。

## 環境変数

### web（`apps/web/.env`、Vercel）

| 名前 | 例 | 説明 |
|---|---|---|
| `API_URL` | `http://localhost:8080` | api のオリジン（パスなし、末尾の `/` なし）。rewrites の中継先と、Server Components の呼び出し先に使う。ビルドのときにも必要。`NEXT_PUBLIC_` を付けない |

- `NEXT_PUBLIC_API_URL` は削除する（`.env.example`、`ci-web.yml`、`infra/vercel.md`）。
- `API_URL` が未設定のときは、ビルドと起動を失敗させる（気付かずに既定値で動かさない）。

### api（`apps/api/.env`、Cloud Run）

| 名前 | 変更 |
|---|---|
| `CORS_ALLOWED_ORIGINS` | 削除する（`.env.example`、`config.go`、`deploy-api.yml`、`infra/gcp.md`） |
| `DATABASE_URL` | 変更なし。管理者のコマンドも使う |

### GitHub

| 名前 | 変更 |
|---|---|
| 変数 `CLOUD_RUN_CORS_ALLOWED_ORIGINS` | `deploy-api.yml` から使わなくなる。GitHub に設定済みなら、利用者が削除する（`infra/gcp.md` に書く） |
