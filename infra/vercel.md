# Vercel（Web）

## 1. プロジェクト

- GitHub の `Qurioth/chara-q` をインポートする。
- **Root Directory**: `apps/web`
- Framework Preset: Next.js
- Install / Build Command は既定のまま（pnpm workspace を自動で検出する）。
- Node.js のバージョンは 24.x にする。
- pnpm のバージョンはルートの `package.json` の `packageManager` で指定している。
  Vercel が使う pnpm のバージョンがこれと合わない場合は、環境変数
  `ENABLE_EXPERIMENTAL_COREPACK=1` を設定する（設定時に動作を確認する）。

## 2. 関係のない変更ではビルドしない

Settings → Git → **Ignored Build Step** に次のコマンドを設定する。
Root Directory（`apps/web`）で実行されるため、パスはそこからの相対パスで書く。

```sh
git diff --quiet HEAD^ HEAD -- . ../../schema ../../package.json ../../pnpm-workspace.yaml ../../pnpm-lock.yaml
```

- 終了コード 0（差分なし）ならビルドを省き、1（差分あり）ならビルドする。
- `apps/api` だけの変更ではビルドされない。

## 3. 環境変数

| 名前 | 環境 | 値 |
|---|---|---|
| `API_URL` | Production | `https://<Cloud Run の URL>`（パスなし、末尾の `/` なし） |
| `API_URL` | Preview | 本番環境の構築の Spec で決める |

- ブラウザは web とだけ通信し、`/graphql` は `next.config.ts` の `rewrites` で `API_URL` に中継する
  （[ADR 0006](../docs/adr/0006-authentication.md)、`specs/002-auth`）。Server Components もこの URL に直接つなぐ。
- `API_URL` はブラウザに出さないため、`NEXT_PUBLIC_` を付けない。`rewrites` はビルドのときに評価されるため、
  ビルドでも必要になる（未設定ならビルドが失敗する）。
- 以前の手順で `NEXT_PUBLIC_API_URL` を登録していた場合は、使わなくなったので削除する。

## 4. ドメイン

- 独自ドメインは使わず、Vercel のドメイン（`*.vercel.app`）をそのまま使う（ADR 0006）。
- Vercel のドメインが決まったら、R2 の CORS に追加する（ブラウザから R2 へ直接アップロードするため）。
- api は Vercel 経由で呼ぶため、api の CORS の許可は外した（認証の Spec）。
