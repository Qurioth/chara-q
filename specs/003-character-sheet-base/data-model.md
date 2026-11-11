# Data Model: キャラクターシートの共通基盤

マイグレーションは `task db:new -- character_sheets` で作る（`apps/api/db/migrations/00003_character_sheets.sql`）。
選んだ理由は [research.md](research.md) の R1・R5〜R9。

## character_sheets

1 行が 1 枚のシート。共通の項目は列、システム固有の部分は `data`（JSONB）に置く。

| 列 | 型 | 制約 | 説明 |
|---|---|---|---|
| `id` | `uuid` | PK、`DEFAULT uuidv7()` | URL の `/sheets/<id>` に使う |
| `system` | `text` | NOT NULL、`CHECK (system IN ('yuyakekoyake', 'beginning_idol'))` | 作成のあとは変えない |
| `owner_id` | `uuid` | NOT NULL、`REFERENCES users (id)` | 所有者。アプリからは変えない（DB 操作でだけ移す） |
| `character_name` | `text` | NOT NULL、1〜100 文字、前後の空白なし、制御文字なし | 必須 |
| `player_name` | `text` | NOT NULL `DEFAULT ''`、0〜50 文字、前後の空白なし、制御文字なし | 自由な文字列。所有者の表示名とは別 |
| `age` | `text` | NOT NULL `DEFAULT ''`、0〜20 文字、前後の空白なし、制御文字なし | 「17 歳」「不明」など |
| `sex` | `text` | NOT NULL `DEFAULT ''`、0〜20 文字、前後の空白なし、制御文字なし | ビギニングアイドルでは空のまま |
| `memo` | `text` | NOT NULL `DEFAULT ''`、0〜10,000 文字 | 改行（`\n`）とタブを許す |
| `data` | `jsonb` | NOT NULL `DEFAULT '{}'`、`CHECK (jsonb_typeof(data) = 'object')` | システム固有の部分。中身は各システムの Spec が決める |
| `revision` | `integer` | NOT NULL `DEFAULT 1`、`CHECK (revision >= 1)` | 更新のたびに 1 増やす（R6） |
| `created_at` | `timestamptz` | NOT NULL `DEFAULT now()` | 複製のときは複製した時刻 |
| `updated_at` | `timestamptz` | NOT NULL `DEFAULT now()` | 更新のたびに `now()` |
| `deleted_at` | `timestamptz` | NULL | NULL なら削除されていない（R7） |

- 文字列の制約は、002-auth の `users.display_name` と同じ書き方にする
  （`char_length(x) BETWEEN 0 AND n AND x = btrim(x) AND x !~ '[[:cntrl:]  ]'`）。
  `memo` だけは前後の空白を許し、制御文字のうち `\n` と `\t` を許す（`\r` は保存の前に Go で `\n` に
  そろえる）。
- `system` の値は DB では `snake_case`、GraphQL では enum（`YUYAKEKOYAKE`・`BEGINNING_IDOL`）、
  URL では `kebab-case`（`yuyakekoyake`・`beginning-idol`、web の `GameSystemId` と同じ）にする。
  変換は api と web のそれぞれ 1 か所にまとめる。
- システム固有の部分の大きさ（256KB 以下）は Go で確かめる。DB には制約を置かない（R11）。

### 索引

すべて削除されていない行だけの部分索引（`WHERE deleted_at IS NULL`）。

| 名前 | 列 | 使うところ |
|---|---|---|
| `character_sheets_updated_idx` | `(updated_at DESC, id DESC)` | 既定の並び順 |
| `character_sheets_created_idx` | `(created_at DESC, id DESC)` | 作成日時の並び順 |
| `character_sheets_name_idx` | `(character_name COLLATE "ja-x-icu", id)` | キャラクター名の並び順（R9） |
| `character_sheets_owner_idx` | `(owner_id, updated_at DESC)` | 「自分のシートだけ」 |

システムでの絞り込みとキーワードの検索には、この機能では索引を作らない（R8。10,000 件なら順次の
走査で足りる）。

### 状態の移り変わり

```text
            作成・複製
               │
               ▼
 ┌──────── 有効（deleted_at が NULL）◀──── 取り消し（管理者の DB 操作）
 │             │  ▲                              │
 │       更新  │  │ revision + 1、updated_at     │
 │             └──┘                              │
 │                                               │
 └─ 削除（所有者）──▶ 削除済み（deleted_at に時刻）─┘
```

- 有効なシートだけが、一覧・閲覧・更新・複製・削除の対象になる。
- 削除しても `updated_at` と `revision` は変えない。取り消すと、削除前と同じ内容・同じ並び順に戻る。

### 削除の取り消し（管理者の DB 操作）

```sql
-- 対象を確かめる
SELECT id, character_name, owner_id, deleted_at FROM character_sheets WHERE id = '<ID>';
-- 取り消す（1 行だけ変わることを確かめる）
UPDATE character_sheets SET deleted_at = NULL WHERE id = '<ID>' AND deleted_at IS NOT NULL;
```

## users（002-auth の表、変更しない）

- シートの所有者として参照する。公開するのは `id` と `display_name` だけ（GraphQL の `User`）。
- DataLoader は `SELECT id, display_name FROM users WHERE id = ANY(@ids::uuid[])` でまとめて引く。
  無効なユーザー（`disabled_at` が NULL でない）も含める。

## sqlc のクエリ（`apps/api/db/queries/character_sheets.sql`）

| 名前 | 種類 | 内容 |
|---|---|---|
| `CreateCharacterSheet` | `:one` | 行を作り、作った行を返す |
| `GetCharacterSheet` | `:one` | 有効なシートを ID で読む |
| `GetCharacterSheetState` | `:one` | 更新・削除が 0 行だったときの判定用。削除済みを含めて `owner_id`・`revision`・`deleted_at` を読む |
| `CountCharacterSheets` | `:one` | 絞り込みに合う有効なシートの件数 |
| `ListCharacterSheetsByUpdated` | `:many` | 絞り込み、`updated_at DESC, id DESC`、`LIMIT`・`OFFSET` |
| `ListCharacterSheetsByUpdatedAsc` | `:many` | 同上、`updated_at ASC, id ASC` |
| `ListCharacterSheetsByCreated` | `:many` | 同上、`created_at DESC, id DESC` |
| `ListCharacterSheetsByCreatedAsc` | `:many` | 同上、`created_at ASC, id ASC` |
| `ListCharacterSheetsByName` | `:many` | 同上、`character_name COLLATE "ja-x-icu", id` |
| `ListCharacterSheetsByNameDesc` | `:many` | 同上、`character_name COLLATE "ja-x-icu" DESC, id DESC` |
| `UpdateCharacterSheet` | `:one` | `id`・`owner_id`・`revision` が合う有効な行だけを更新し、`revision + 1`、`updated_at = now()` |
| `CopyCharacterSheet` | `:one` | `INSERT ... SELECT` で複製する（名前は R12） |
| `DeleteCharacterSheet` | `:execrows` | `id`・`owner_id` が合う有効な行の `deleted_at` を `now()` にする |
| `ListUsersByIDs` | `:many` | DataLoader 用（`users.sql` に足す） |

- 絞り込みは、6 つの一覧と件数で同じ `WHERE` を使う。
  `(sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system))`、
  `(sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id))`、
  キーワードは R8 の式。
- 並び順ごとに SQL を分けるのは、`ORDER BY` を引数で切り替えると索引が使われにくく、sqlc で
  型も付かないため。逆の順は、同じ索引を逆向きに読む（R9）。

## web 側のデータ

DB には保存しない。

- **一覧の検索条件**（`lib/sheets/search-params.ts`）: URL のクエリから読む。
  `{ system: GameSystemId | null; q: string; mine: boolean; sort: SheetSort; page: number }`。
  `SheetSort` は `"updated" | "updated-asc" | "created" | "created-asc" | "name" | "name-desc"`（R14）。
  知らない値は既定（`null`・`""`・`false`・`"updated"`・`1`）にする。URL を作る関数も同じ場所に置き、
  既定の値はクエリに含めない（URL を短く保つ）。
- **システムの表示用の部品の表**（`lib/sheets/system-views.ts`）: `GameSystemId` ごとに、閲覧画面の
  システム固有の部分と所有者の操作を描く部品を持つ。この機能では、どちらも準備中の表示だけ。
