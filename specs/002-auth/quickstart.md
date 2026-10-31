# Quickstart: 認証の動作確認

実装の後に、この手順で受け入れシナリオを確かめる。細かい型・表・属性は
[contracts/](contracts/) と [data-model.md](data-model.md) を見る。

## 前提

```sh
mise install
task setup            # .env を作る（既にある場合は下の「既存の .env」を見る）
task db:up
task db:migrate
task generate
task dev              # api :8080、web :3000
```

- `apps/web/.env` に `API_URL=http://localhost:8080` があること。

既存の `.env` は `task setup` で上書きされないため、手で次のように直す。

- `apps/web/.env`: `NEXT_PUBLIC_API_URL=…` を消し、`API_URL=http://localhost:8080` を足す。
- `apps/api/.env`: `CORS_ALLOWED_ORIGINS=…` を消す。`TEST_DATABASE_URL` のデータベース名を
  `chara_q_test` にする（テストは表を空にするため、開発用の DB と分ける）。
- ブラウザでは `http://localhost:3000` を開く。api の `http://localhost:8080` には直接つながない。

## 1. 中継と Cookie の確認（FR-004、最初に行う）

実装の最初の段階（ログインの処理ができる前）は、仮の Cookie を返す処理で確かめる。

1. Chrome の開発者ツールの Network で、`/graphql` へのリクエストが `localhost:3000` 宛てで、
   `localhost:8080` 宛てのリクエストがないことを確かめる。
2. ログイン（または仮の処理）の応答に `Set-Cookie: __Host-chara_q_session=…; Path=/; HttpOnly;
   Secure; SameSite=Lax` があり、Application → Cookies の `localhost:3000` に保存されていることを
   確かめる。`Domain` は空であること。
3. 次の `/graphql` のリクエストの `Cookie` ヘッダーに `__Host-chara_q_session` が付き、api の
   ログ（または `viewer` の応答）で受け取れていることを確かめる。
4. 画面を再読み込みし、Server Components が描いたヘッダーに表示名が出ることを確かめる
   （Server Components から api に Cookie が届いている）。
5. Firefox でも 1〜4 を確かめる。
6. Safari で確かめるときは `task dev:web:https` で `https://localhost:3000` を使う（research R2）。
   自己署名の証明書を端末に登録するため、使うかどうかは開発者が決める。使わない場合は、本番環境の
   構築の機能で Vercel の URL を使って確かめる。

どれかが確かめられなければ、以降の実装に進まず、research R1 の代案（Route Handler での中継）を
検討する。

## 2. 最初の管理者とログイン（US1）

```sh
task admin:create -- -login-id admin -display-name 管理者
# パスワードを 2 回入力する（画面には出ない）
task admin:users
```

| 確かめること | 期待する結果 |
|---|---|
| ログインしていない状態で `/` を開く | ヘッダーに「ログイン」がある |
| `/login` で `admin` とパスワードでログインする | `/` に戻り、ヘッダーに「管理者」が出る |
| 再読み込みする・タブを閉じて開き直す | 最初の描画から「管理者」が出る（「ログイン」が一瞬出ない） |
| 誤ったパスワード・存在しないログイン ID でログインする | どちらも同じ文言になる |
| メニューの「ログアウト」 | 「ログイン」に戻る。ログアウト前の Cookie の値を手で戻しても、ログインしていない扱いになる |
| `ADMIN` のように大文字でログインする | ログインできる |

## 3. 招待コードと登録（US2）

1. 管理者でログインし、メニューの「招待コードの発行」で発行する。コードと期限（7 日後）が出る。
2. 別のブラウザ（またはシークレットウィンドウ）で `/register` を開き、そのコードで登録する。
   ログインした状態で `/` に移り、ヘッダーに表示名が出る。
3. 同じコードでもう一度登録すると、「招待コードが使えない」の文言になる。
4. 新しいコードで、既に使われているログイン ID（大文字・小文字だけ変えたもの）で登録すると、
   ログイン ID の項目にエラーが出る。そのコードはまだ使える。
5. 管理者でないユーザーで `/admin/invitations` を開くと 404 になり、メニューにも入口がない。
6. 管理者でないユーザーのブラウザの開発者ツールのコンソールから、`fetch("/graphql", { method: "POST",
   headers: { "Content-Type": "application/json" }, body: JSON.stringify({ query: "mutation { issueInvitation { code } }" }) })`
   を実行すると、`FORBIDDEN` のエラーが返る。

## 4. ロック（US3）

1. 登録したユーザーのログイン ID で、誤ったパスワードで 5 回ログインする。5 回目に、ロックと
   解ける時刻の目安が出る。
2. 正しいパスワードでもログインできない。
3. 存在しないログイン ID で 5 回失敗しても、同じロックの表示になる。
4. 15 分待つ（またはテストで時刻を進める）と、正しいパスワードでログインできる。

## 5. アカウントの設定（US4）

1. ブラウザ A と B で同じユーザーとしてログインする。
2. A の `/account` で表示名を変えると、A のヘッダーにすぐ反映される。
3. A でパスワードを変えると、A はログインしたまま、B は再読み込みでログインしていない表示になる。
4. ログインしていない状態で `/account` を開くと、ログイン画面に移り、ログイン後に `/account` に戻る。

## 6. パスワードの再設定（US5）

```sh
task admin:reset-code -- <ログイン ID>
```

1. `/reset-password` で、ログイン ID・出力されたコード・新しいパスワードを入力する。ログイン画面に
   移り、新しいパスワードでログインできる。古いパスワードではログインできない。
2. 同じコードをもう一度使うと、「再設定コードが使えない」の文言になる。
3. 同じユーザーに 2 回発行すると、1 回目のコードは使えない。
4. ロック中のユーザーでも、再設定するとすぐにログインできる。

## 7. ユーザーの無効化（US6）

```sh
task admin:disable -- <ログイン ID>
task admin:enable -- <ログイン ID>
```

1. ログインしているユーザーを無効にすると、そのユーザーの次の画面の表示でログインしていない
   表示になる。正しいパスワードでもログインできず、文言は誤ったパスワードと同じ。
2. 無効なユーザーへの `admin:reset-code` は失敗する。
3. 有効に戻すと、元のパスワードでログインできる。

## 8. 表示と安全性

| 確かめること | 期待する結果 |
|---|---|
| 幅 360px で、ログインしていない・長い表示名でログインしている状態のヘッダー | 横スクロールがなく、要素が重ならない（表示名は省略される） |
| ログインしていない状態で `viewer` 以外の応答を見る | ほかのユーザーのログイン ID が含まれない（SC-008） |
| api のログ | パスワード・コード・セッション ID の値が出ていない（FR-010） |
| DB の `users`・`sessions`・`invitations`・`password_reset_codes` | 平文のパスワード・コード・セッション ID がない（SC-007） |
| `curl -i -X OPTIONS -H "Origin: https://example.com" http://localhost:8080/graphql` | `Access-Control-Allow-Origin` がない |

## 9. 自動テスト

```sh
task lint
task test             # api の DB を使うテストには task db:up が必要（テスト用の DB は test:api が用意する）
task check:generated
```

- api: 入力の規則、ハッシュ、登録・ログイン・ロック・再設定・無効化、同時の使用（10 回で成功 1 回）、
  Cookie の属性、GET のミューテーション・`text/plain` の POST の拒否が通ること。
- web: フォームの検証と文言、ヘッダーの 3 つの状態、戻り先の検証が通ること。
