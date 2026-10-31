# Contract: 画面とヘッダー（`apps/web`）

画面は UI 基盤（`001-ui-foundation`）の共通レイアウトと共通入力部品（`components/form/`）で作る。
フォームの送信は、Client Components から urql のミューテーションで `/graphql` に送る。
成功したら `router.refresh()` で Server Components（ヘッダー）を描き直してから移動する。

## 画面

| パス | 表示できる人 | 内容 | 成功したとき |
|---|---|---|---|
| `/login` | 誰でも。ログインしていれば戻り先へ移動 | ログイン ID、パスワード。登録と再設定への入口 | 戻り先（`?next=`）、なければ `/` |
| `/register` | 誰でも。ログインしていれば `/` へ移動 | 招待コード、ログイン ID、パスワード、表示名。ログインへの入口 | `/` |
| `/reset-password` | 誰でも | ログイン ID、再設定コード、新しいパスワード。ログインへの入口 | `/login`（再設定したことを表示する） |
| `/account` | ログインしたユーザー。していなければ `/login?next=/account` | ログイン ID（表示だけ）、表示名の変更、パスワードの変更（現在と新しいもの） | 同じ画面で成功を表示する |
| `/admin/invitations` | 有効な管理者。ログインしていなければ `/login?next=/admin/invitations`、管理者でなければ 404 | 「招待コードを発行する」ボタン。発行したコード、有効期限、コピーのボタン | 同じ画面に表示する（再読み込みすると消える） |

- 各画面にタイトルを付ける（例: `ログイン | CharaQ`）。
- 招待コードと再設定コードの入力欄は、前後の空白・ハイフンを許し、大文字・小文字を区別しない
  （research R7）。
- エラーの文言は `lib/auth/messages.ts` で `InputErrorCode` から決める。項目の失敗は共通入力部品の
  エラーに、フォーム全体の失敗（`INVALID_CREDENTIALS`、`LOCKED`）はフォームの上に表示する。
- `LOCKED` のときは、`lockedUntil` を `HH:mm` の形で「〜ごろまでログインできません」と表示する。
- パスワードの入力欄は `type="password"`、`autocomplete` は登録と変更で `new-password`、ログインで
  `current-password`、ログイン ID は `username` にする（パスワード管理ソフトで使えるようにする）。

## 戻り先（`?next=`）の検証

`lib/auth/safe-next.ts` で、次をすべて満たすときだけ戻り先として使い、それ以外は `/` にする（spec の
Edge Cases）。

- `/` で始まる
- `//` と `/\` で始まらない
- `/login`・`/register`・`/reset-password` ではない（ログインの画面に戻らない）

## ヘッダーのアカウントのメニュー

ルートレイアウトのヘッダーの操作の領域に、テーマの切り替えと並べて置く。ログイン状態は
`<Suspense>` の中で読む（research R3）。

| 状態 | 表示 |
|---|---|
| 読み込み中（`fallback`） | ボタンと同じ大きさの空の領域。ログインの操作は出さない（SC-009） |
| ログインしていない | 「ログイン」のリンク（`/login?next=<今の画面>`） |
| ログインしている | 表示名のボタン（長い表示名は省略）。押すとメニューが開く: 「アカウントの設定」（`/account`）、管理者だけ「招待コードの発行」（`/admin/invitations`）、「ログアウト」 |

- メニューは shadcn/ui の `dropdown-menu`（Base UI の Menu）で、キーボードで開閉・選択できる。
- 表示名のボタンには、読み上げ用に「アカウントのメニュー（<表示名>）」の名前を付ける。
- 幅 360px で、ロゴ・テーマの切り替え・表示名のボタンが重ならない（FR-025）。表示名のボタンの
  最大の幅を決め、超えた分は `…` で省略する。
- ログアウトしたら `router.refresh()` する。ログインが必要な画面にいた場合は、その画面の
  Server Components がログイン画面へ移動させる。

## Server Components 用の関数（`lib/auth/viewer.ts`、`server-only`）

| 関数 | 返すもの | ログインしていないとき | 管理者でないとき |
|---|---|---|---|
| `getViewer()` | `Viewer \| null` | `null` | — |
| `requireViewer(path)` | `Viewer` | `redirect("/login?next=" + path)` | — |
| `requireAdmin(path)` | `Viewer` | `redirect("/login?next=" + path)` | `notFound()` |

- 1 回のリクエストの中では、React の `cache()` で api の呼び出しを 1 回にまとめる（ヘッダーと画面の
  両方で使うため）。
- api に接続できないときは、ログインしていない扱いにせず、エラーにする（ログインしているのに
  ログアウトしたように見せない）。ヘッダーでは、エラーのときにログインの操作を出さない。
