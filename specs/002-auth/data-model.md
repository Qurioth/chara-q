# Data Model: 認証

Postgres の表を 5 つ追加する（`apps/api/db/migrations/00002_auth.sql`、`task db:new -- auth` で作る）。
日時はすべて `timestamptz`。ID は Postgres 18 の `uuidv7()` を既定値にする。

## users

アプリを使う人。後の機能でキャラクターシートの所有者になる。

| 列 | 型 | 制約 | 説明 |
|---|---|---|---|
| `id` | `uuid` | PK、既定 `uuidv7()` | |
| `login_id` | `text` | NOT NULL、`CHECK (login_id ~ '^[A-Za-z0-9_-]{3,32}$')` | 入力のとおりの大文字・小文字で保存する。本人にだけ見せる |
| `display_name` | `text` | NOT NULL、`CHECK (char_length(display_name) BETWEEN 1 AND 32)`、前後の空白と制御文字を含まない | 公開する。重複してよい |
| `password_hash` | `text` | NOT NULL | Argon2id の PHC 形式の文字列（research R6） |
| `is_admin` | `boolean` | NOT NULL、既定 `false` | |
| `disabled_at` | `timestamptz` | NULL | NULL なら有効。値があれば無効にした日時 |
| `created_at` | `timestamptz` | NOT NULL、既定 `now()` | |
| `updated_at` | `timestamptz` | NOT NULL、既定 `now()` | 表示名・パスワード・状態を変えたら更新する |

- 一意の索引: `users_login_id_key ON users (lower(login_id))`（大文字と小文字を区別しない一意性。FR-015）
- 表示名の `CHECK` は、`display_name = btrim(display_name)` と、制御文字（`[[:cntrl:]]`）・
  U+2028・U+2029 を含まないことを確かめる（research R11）。

**状態の移り変わり**

```text
（作成: コマンドで管理者 / 招待コードで登録）
        │
        ▼
     有効 ──（admin:disable）──▶ 無効
       ▲                         │
       └────（admin:enable）──────┘
```

- 無効にするときは、同じトランザクションでそのユーザーの `sessions` をすべて消す（FR-044）。
- 無効のユーザーはログインできず、再設定コードも発行・使用できない。
- 有効に戻しても、パスワードは変わらない（US6 の受け入れシナリオ 4）。

## sessions

ログイン状態。Cookie にはセッション ID だけを入れ、DB にはその SHA-256 を保存する。

| 列 | 型 | 制約 | 説明 |
|---|---|---|---|
| `token_hash` | `bytea` | PK | セッション ID の SHA-256（32 バイト） |
| `user_id` | `uuid` | NOT NULL、FK `users(id)` ON DELETE CASCADE | |
| `created_at` | `timestamptz` | NOT NULL、既定 `now()` | |
| `last_used_at` | `timestamptz` | NOT NULL、既定 `now()` | 1 時間より古ければ、使ったときに更新する（research R10） |
| `expires_at` | `timestamptz` | NOT NULL | 作成・更新のときに `now() + 30 日` |

- 索引: `sessions (user_id)`（ユーザーのセッションをまとめて消す）、`sessions (expires_at)`（掃除）
- 有効なセッション: `expires_at > now()` かつ、ユーザーが有効（`disabled_at IS NULL`）。
- 削除するとき:

| 契機 | 消すもの |
|---|---|
| ログアウト | そのセッション |
| ログイン | ブラウザが送ってきた前のセッション（あれば）。新しいセッションを作る（FR-019） |
| パスワードの変更 | そのユーザーのすべて。変更したブラウザには新しいセッションを作る（FR-033） |
| パスワードの再設定 | そのユーザーのすべて（FR-039） |
| ユーザーの無効化 | そのユーザーのすべて（FR-044） |
| ログインの成功のたび | 全体の期限切れのセッション（掃除。research R9） |

## invitations

登録を許可する、1 回だけ使えるコード。

| 列 | 型 | 制約 | 説明 |
|---|---|---|---|
| `id` | `uuid` | PK、既定 `uuidv7()` | |
| `code_hash` | `bytea` | NOT NULL、UNIQUE | 正規化した招待コードの SHA-256（research R7） |
| `issued_by` | `uuid` | NOT NULL、FK `users(id)` | 発行した管理者 |
| `created_at` | `timestamptz` | NOT NULL、既定 `now()` | |
| `expires_at` | `timestamptz` | NOT NULL | `created_at + 7 日` |
| `used_at` | `timestamptz` | NULL | |
| `used_by` | `uuid` | NULL、UNIQUE、FK `users(id)` | 登録されたユーザー |

- `CHECK ((used_at IS NULL) = (used_by IS NULL))`
- 使えるコード: `used_at IS NULL AND expires_at > now()`。

**状態の移り変わり**

```text
未使用 ──（登録の成功）──▶ 使用済み
  │
  └──（7 日が過ぎる）──▶ 期限切れ（行は残る。判定は expires_at で行う）
```

- 登録の手順（research R8）: 1 つのトランザクションで、コードの行を `FOR UPDATE` で押さえる →
  ユーザーを作る → `used_at`・`used_by` を入れる → セッションを作る。どこかで失敗したら
  ロールバックし、招待コードは未使用のまま残る（FR-014）。

## password_reset_codes

特定のユーザーのパスワードを再設定するための、1 回だけ使えるコード。コマンドで発行するため、
発行した管理者は記録しない（コマンドはログインしたユーザーとして動かない）。

| 列 | 型 | 制約 | 説明 |
|---|---|---|---|
| `id` | `uuid` | PK、既定 `uuidv7()` | |
| `user_id` | `uuid` | NOT NULL、FK `users(id)` ON DELETE CASCADE | 対象のユーザー |
| `code_hash` | `bytea` | NOT NULL、UNIQUE | 正規化した再設定コードの SHA-256 |
| `created_at` | `timestamptz` | NOT NULL、既定 `now()` | |
| `expires_at` | `timestamptz` | NOT NULL | `created_at + 24 時間` |
| `used_at` | `timestamptz` | NULL | |

- 索引: `password_reset_codes (user_id)`
- 発行するときは、同じトランザクションでそのユーザーの未使用のコードを消してから作る
  （前のコードを使えなくする。FR-043）。
- 使うときは、ログイン ID で引いたユーザーの `user_id` と `code_hash` が一致し、未使用で期限内の
  行を `FOR UPDATE` で押さえる。別のユーザー用のコードは一致しない（FR-038）。

**状態の移り変わり**

```text
未使用 ──（再設定の成功）──▶ 使用済み
  ├──（24 時間が過ぎる）──▶ 期限切れ
  └──（同じユーザーに新しいコードを発行）──▶ 削除
```

## auth_failures

ログインとパスワードの再設定の、続けて失敗した記録。存在しないログイン ID も記録する
（FR-029）。

| 列 | 型 | 制約 | 説明 |
|---|---|---|---|
| `purpose` | `text` | `CHECK (purpose IN ('login', 'password_reset'))` | |
| `login_id_key` | `text` | NOT NULL | 前後の空白を除き、小文字にしたログイン ID。形式に合うものだけを記録する |
| `failure_count` | `integer` | NOT NULL、`CHECK (failure_count >= 0)` | 続けて失敗した回数 |
| `last_failed_at` | `timestamptz` | NOT NULL | |
| `locked_until` | `timestamptz` | NULL | ロックが解ける日時 |

- PK: `(purpose, login_id_key)`
- 索引: `auth_failures (last_failed_at)`（掃除）

**判定の規則**（research R9）

| 場面 | 処理 |
|---|---|
| 試みの前 | `locked_until > now()` なら、照合せずに `LOCKED` と `locked_until` を返す。回数は変えない |
| 失敗 | ロックが過ぎていれば回数を 1 に、そうでなければ 1 足す。5 になったら `locked_until = now() + 15 分`。1 つの `INSERT … ON CONFLICT DO UPDATE` で行う |
| ログインの成功 | `login` の行を消す |
| 再設定の成功 | `login` と `password_reset` の両方の行を消す（FR-030） |
| 掃除 | ログインの成功のたびに、`last_failed_at < now() - 1 日` で、ロック中でない行を消す |

## web 側のデータ

- web は DB を持たない。ログイン状態は、Server Components で `viewer` を問い合わせて得る
  （[contracts/http-and-cookie.md](contracts/http-and-cookie.md)）。
- Client Components に渡すのは `Viewer` の `displayName` と `isAdmin` だけにする。`loginId` は
  アカウントの設定画面で本人に表示するときだけ使う。
