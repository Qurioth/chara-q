---

description: "Task list for キャラクターシートの共通基盤"
---

# Tasks: キャラクターシートの共通基盤

**Input**: Design documents from `specs/003-character-sheet-base/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/](contracts/), [quickstart.md](quickstart.md)

**Tests**: constitution VI（DB を触るテストは実際の Postgres、`task test` が通らない変更はマージしない）と
[research.md](research.md) の R17 に従い、各ストーリーにテストのタスクを含める。api はテーブル駆動、DB を
使うテストは `TEST_DATABASE_URL` の Postgres、GraphQL は `httptest`、問い合わせの回数は pgx の
`QueryTracer`。web は Vitest + Testing Library + `user-event`。幅 360px の表示と照合順序は
[quickstart.md](quickstart.md) で手で確かめる。

**Organization**: ユーザーストーリーごとに分ける。番号は spec.md のもの（US1〜US6）。
フェーズは、作成（US3）があると他のストーリーのデータを用意できるため、P1 の中を
US3 → US2 → US1 → US4 の順にし、続けて P2 の US5・US6 にする。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 並行して進められる（別のファイルで、未完了のタスクに依存しない）
- **[Story]**: どのユーザーストーリーのタスクか（US1〜US6）
- パスはリポジトリのルートからの相対パス

## 全タスク共通の注意

- web のコードを書く前に、`apps/web/AGENTS.md` の指示に従い、`apps/web/node_modules/next/dist/docs/` の
  該当ガイドを読む（`01-app/01-getting-started/08-caching.md`、`01-app/03-api-reference/03-file-conventions/page.md`
  の `params`・`searchParams`、`not-found.md`、`01-app/03-api-reference/02-components/form.md`、
  `01-app/03-api-reference/04-functions/generate-metadata.md` の「With Cache Components」）。
- 生成コードは手で直さない。スキーマ・SQL・`gqlgen.yml` を変えたら `task generate` を実行し、生成結果を
  元の変更と同じコミットに入れる（CLAUDE.md）。`apps/api/internal/graph/model/sheet.go` は手で書く
  ファイルで、生成コードではない（research R3）。
- マイグレーションは `task db:new -- character_sheets` で作る。ORM は使わない。
- Go の依存は `apps/api` で `go get` で入れる。web には依存を足さない。
- DB を使うテストは `authtest.New(t)` で始める。`TRUNCATE users ... CASCADE` で `character_sheets` も
  空になる。`t.Parallel()` は使わない（`go test -p 1`）。
- 所有者のログイン ID を、シートの応答・画面・ログに出さない（FR-019、SC-005）。
- 文言は日本語。web のエラーの文言は `apps/web/src/lib/auth/messages.ts` に、シートの画面の文言は
  各部品に書く。
- コミットは CLAUDE.md の形式（`Refs: specs/003-character-sheet-base`）で、論理的なまとまりごとに分ける。

---

## Phase 1: Setup（依存と照合順序の確認）

**Purpose**: 追加の依存を入れ、設計の前提（`ja-x-icu` の照合順序）を確かめる

- [X] T001 `apps/api` で `go get github.com/vikstrous/dataloadgen@v0.0.10` を実行し、`apps/api/go.mod` と `apps/api/go.sum` を更新する（research R18）
- [X] T002 `task db:up` の後、ローカルの DB で `SELECT collname FROM pg_collation WHERE collname IN ('ja-x-icu', 'und-x-icu');` を実行し、結果を [research.md](research.md) の R9 の「確かめること」の下に「確認結果（ローカル）」として書く。`ja-x-icu` がなければ、R9 のとおり `und-x-icu`、それもなければ既定の照合順序に決め、[data-model.md](data-model.md) の `character_sheets_name_idx` と `ListCharacterSheetsByName` の記述も合わせて直す。Neon での確認は、デプロイの手順書（`infra/`）の作業として R9 に「未確認」と書いておく

**Checkpoint**: `go build ./...`（`apps/api`）が通り、キャラクター名の並び順に使う照合順序が決まっている

---

## Phase 2: Foundational（全ストーリーの前提）

**Purpose**: 表・クエリ・スキーマ・型・DataLoader・エラーのコードなど、どのストーリーにも要るもの

**⚠️ CRITICAL**: このフェーズが終わるまで、ユーザーストーリーの作業は始めない

### api: DB の表と sqlc

- [X] T003 `task db:new -- character_sheets` で `apps/api/db/migrations/00003_character_sheets.sql` を作り、[data-model.md](data-model.md) の `character_sheets` を書く。ファイルの先頭に「表の設計は specs/003-character-sheet-base/data-model.md」とコメントを書く。列と制約は次のとおり:
  - `id uuid PRIMARY KEY DEFAULT uuidv7()`
  - `system text NOT NULL CHECK (system IN ('yuyakekoyake', 'beginning_idol'))`
  - `owner_id uuid NOT NULL REFERENCES users (id)`
  - `character_name text NOT NULL`、「1〜100 文字、前後の空白なし、制御文字なし」（`char_length(character_name) BETWEEN 1 AND 100 AND character_name = btrim(character_name) AND character_name !~ '[[:cntrl:]  ]'`）
  - `player_name text NOT NULL DEFAULT ''`、「0〜50 文字、前後の空白なし、制御文字なし」
  - `age text NOT NULL DEFAULT ''`、「0〜20 文字、前後の空白なし、制御文字なし」
  - `sex text NOT NULL DEFAULT ''`、「0〜20 文字、前後の空白なし、制御文字なし」
  - `memo text NOT NULL DEFAULT ''`、「0〜10,000 文字」。制御文字は `\n` と `\t` だけを許す（`memo !~ '[\x01-\x08\x0b-\x1f\x7f  ]'`。`\r` は Go で `\n` にそろえるため許さない）
  - `data jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(data) = 'object')`
  - `revision integer NOT NULL DEFAULT 1 CHECK (revision >= 1)`
  - `created_at timestamptz NOT NULL DEFAULT now()`、`updated_at timestamptz NOT NULL DEFAULT now()`、`deleted_at timestamptz`（NULL なら削除されていない）
  - 部分索引（すべて `WHERE deleted_at IS NULL`）: `character_sheets_updated_idx (updated_at DESC, id DESC)`、`character_sheets_created_idx (created_at DESC, id DESC)`、`character_sheets_name_idx (character_name COLLATE "ja-x-icu", id)`（照合順序は T002 の結果）、`character_sheets_owner_idx (owner_id, updated_at DESC)`
  - `-- +goose Down` は `DROP TABLE character_sheets;`
  - `task db:migrate` と `task db:test:prepare` で適用できることを確かめる
- [X] T004 `apps/api/db/queries/character_sheets.sql` を作り、[data-model.md](data-model.md) の「sqlc のクエリ」の 10 個を書く。すべての読み書きに `deleted_at IS NULL` を条件に含める（`GetCharacterSheetState` だけは削除済みも読む）。時刻は `@now::timestamptz` を引数で受け取り（002-auth のクエリと同じく、テストの時計で決められるようにする）、`created_at`・`updated_at`・`deleted_at` に使う。
  - 絞り込みの `WHERE` は、件数と 3 つの一覧で同じにする: `(sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system))`、`(sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id))`、キーワードは `sqlc.narg(keyword)::text IS NULL OR lower(normalize(character_name, NFKC)) LIKE <パターン> ESCAPE '\' OR lower(normalize(player_name, NFKC)) LIKE <パターン> ESCAPE '\'`。`<パターン>` は `'%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword), NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%'`（正規化の後にエスケープする、research R8）
  - 並び順: `ListCharacterSheetsByUpdated` は `ORDER BY updated_at DESC, id DESC`、`ListCharacterSheetsByCreated` は `ORDER BY created_at DESC, id DESC`、`ListCharacterSheetsByName` は `ORDER BY character_name COLLATE "ja-x-icu" ASC, id ASC`。どれも `LIMIT @limit OFFSET @offset`
  - `UpdateCharacterSheet`: `WHERE id = @id AND owner_id = @owner_id AND deleted_at IS NULL AND revision = @revision` で、共通の項目と `data` を置き換え、`revision = revision + 1`、`updated_at = @now` にして `RETURNING *`
  - `CopyCharacterSheet`: `INSERT INTO character_sheets (system, owner_id, character_name, player_name, age, sex, memo, data, created_at, updated_at) SELECT system, @owner_id, left(character_name, 95) || '（コピー）', player_name, age, sex, memo, data, @now, @now FROM character_sheets WHERE id = @id AND deleted_at IS NULL RETURNING *`（research R12）
  - `DeleteCharacterSheet`（`:execrows`）: `UPDATE ... SET deleted_at = @now WHERE id = @id AND owner_id = @owner_id AND deleted_at IS NULL`。`updated_at` と `revision` は変えない
- [X] T005 [P] `apps/api/db/queries/users.sql` に `-- name: ListUsersByIDs :many` を足す: `SELECT id, display_name FROM users WHERE id = ANY(@ids::uuid[]);`（無効なユーザーも含める、data-model）
- [X] T006 `task generate` で sqlc の生成コード（`apps/api/internal/db/character_sheets.sql.go` など）を作り、`apps/api/internal/db/models.go` に `CharacterSheet` が増えたことを確かめる。`data` は `[]byte`、`deleted_at` は `*time.Time` になる（`sqlc.yaml` の既存の指定）

### api: スキーマと gqlgen の型

- [X] T007 [contracts/graphql.md](contracts/graphql.md) のスキーマの案を `schema/graphql/character_sheet.graphqls` に置く。`schema/graphql/auth.graphqls` の `InputField` に `CHARACTER_NAME`・`PLAYER_NAME`・`AGE`・`SEX`・`MEMO`・`DATA` を、`InputErrorCode` に `NOT_OBJECT`・`TOO_LARGE`・`REVISION_CONFLICT`（説明の文字列も契約のとおり）を足す。`CharacterSheetFieldsInput.data` の既定値 `{}` が gqlgen と GraphQL Code Generator の両方で受け付けられなければ、既定値を外して `data: JSON!` を必須にし、契約も直す
- [X] T008 `apps/api/internal/graph/model/sheet.go` を手で書く（research R3）: 共通の項目の構造体 `SheetCommon`（`ID string`、`System GameSystem`、`OwnerID uuid.UUID`、`CharacterName`・`PlayerName`・`Age`・`Sex`・`Memo string`、`Revision int`、`CreatedAt`・`UpdatedAt time.Time`）と、それを埋め込む `YuyakekoyakeSheet`・`BeginningIdolSheet`（`Data scalar.JSON` を持つ）。gqlgen が生成する `CharacterSheet` のインターフェースのメソッド（`IsCharacterSheet()` と各項目の getter）を実装する。ファイルの先頭に「gqlgen の生成コードではない。gqlgen.yml で結び付けている」とコメントを書く
- [X] T009 `apps/api/gqlgen.yml` の `models` に `YuyakekoyakeSheet` と `BeginningIdolSheet` を `github.com/Qurioth/chara-q/apps/api/internal/graph/model.YuyakekoyakeSheet` などに結び付け、`owner` の項目に `resolver: true` を指定する。`task generate` で `apps/api/internal/graph/generated.go`・`model/models_gen.go`・`character_sheet.resolvers.go`（`panic("not implemented")` の雛形）ができ、`go build ./...` が通ることを確かめる。web の `apps/web/src/gql/` も同じコミットに入れる

### api: シートの共通の処理（`apps/api/internal/sheet`）

- [X] T010 [P] `apps/api/internal/sheet/system.go` に、システムの値の変換を書く: DB の `yuyakekoyake`・`beginning_idol` と GraphQL の `model.GameSystemYuyakekoyake`・`model.GameSystemBeginningIdol` の相互の変換（知らない値はエラー）。DB の行（`db.CharacterSheet`）から `model.YuyakekoyakeSheet` か `model.BeginningIdolSheet` を作る関数 `ToModel(row db.CharacterSheet) (model.CharacterSheet, error)` もここに置く（変換を 1 か所にする、research R3）
- [X] T011 [P] `apps/api/internal/sheet/validate.go` に、共通の項目と `data` の検証を書く（FR-006、FR-007、contracts/graphql.md の「ふるまい」）。`auth.InputErrors`（`apps/api/internal/auth/validate.go`）と同じ形の項目ごとの失敗を返し、1 項目につき最初の 1 つだけ（`REQUIRED` → `TOO_LONG` → `INVALID_CHARACTERS` の順）。規則:
  - キャラクター名: 「前後の空白を除いて 1〜100 文字。必須」
  - プレイヤー名: 「0〜50 文字」、年齢・性別: 「それぞれ 0〜20 文字の自由な文字列」。前後の空白を除き、制御文字（` `・` ` を含む）は `INVALID_CHARACTERS`
  - メモ: 「0〜10,000 文字。改行を含めてよい」。前後の空白は除かない。`\r\n` と `\r` を `\n` にそろえてから数え、`\n` と `\t` 以外の制御文字は `INVALID_CHARACTERS`
  - `data`: JSON のオブジェクトでなければ `DATA` の `NOT_OBJECT`、262,144 バイトを超えたら `DATA` の `TOO_LARGE`
  - 文字数はコードポイントで数える（`utf8.RuneCountInString`、Postgres の `char_length` と同じ）
  - 検証の後の値（空白を除いた文字列、そろえた改行）を返し、保存にはその値を使う
- [X] T012 [P] `apps/api/internal/sheet/validate_test.go` に、T011 のテーブル駆動のテストを書く: 各項目の境界（0・1・上限・上限+1 文字）、前後の空白、空白だけのキャラクター名、改行・タブ・` ` を含む名前、絵文字と日本語、メモの `\r\n`、`data` の配列・文字列・`null`・ちょうど 262,144 バイトと 1 バイト超え
- [X] T013 `apps/api/internal/sheet/service.go` に `Service` の骨組みを書く: `NewService(pool *pgxpool.Pool, logger *slog.Logger)`、テストで時計を差し替える `SetClock(func() time.Time)`（`auth.Service` と同じ形）、エラーの値 `ErrNotFound`・`ErrForbidden`・`ErrRevisionConflict`。各操作のメソッドは、それぞれのストーリーのタスクで足す
- [X] T014 [P] `apps/api/internal/sheet/sheettest/sheettest.go` に、DB を使うテストの補助を書く: `authtest.Env` を受け取り、`CreateSheet(t, owner uuid.UUID, system, name, playerName string)` で行を作る関数、`SetTimes(t, id, createdAt, updatedAt)` で日時を直接書き換える関数（並び順のテスト用）、`Row(t, id)` で削除済みを含めて行を読む関数

### api: DataLoader、エラー、サーバー

- [X] T015 `apps/api/internal/loader/loader.go` に DataLoader を書く（research R4）: `dataloadgen` で `UserByID`（キーは `uuid.UUID`、値は `db.ListUsersByIDsRow`）を作る `Loaders` 型、リクエストごとに新しい `Loaders` を context に入れる `Middleware(queries *db.Queries) func(http.Handler) http.Handler`、context から取り出す `For(ctx)`。見つからないユーザーは `nil` ではなくエラーにする（所有者は外部キーで必ず存在するため）
- [X] T016 `apps/api/internal/graph/errors.go` に `codeNotFound = "NOT_FOUND"` と `errNotFound()`（文言「シートが見つかりません」）を足し、`errForbidden()` の文言を「この操作は行えません」に変える。`apps/api/internal/auth/service.go` の `ErrForbidden` の文言も「この操作は行えません」に変え、文言を確かめているテスト（`apps/api/internal/auth/*_test.go`、`apps/api/internal/graph/auth_resolvers_test.go`）があれば直す。sheet の入力の失敗を payload の `errors` に変える関数を、002-auth の `inputErrors` と同じ形で足す
- [X] T017 `apps/api/internal/graph/resolver.go` の `Resolver` に `Sheets *sheet.Service` を足し、`apps/api/cmd/server/main.go` で `sheet.NewService(pool, logger)` を渡す。`apps/api/internal/server/server.go` の `/graphql` に、認証のミドルウェアの後で `loader.Middleware(resolver.Queries)` を入れ、リクエストの本体を `http.MaxBytesReader` で 1MB（1,048,576 バイト）までに制限する（research R11）
- [X] T018 `apps/api/internal/graph/auth_resolvers_test.go` の `newClient` を、`Sheets` を含む `Resolver` を作る形に変え、シートのテストからも使えるよう、`apps/api/internal/graph/client_test.go` に移す（`client`・`gqlResponse`・`do`・`errorCode` を移し、002-auth のテストはそのまま通ることを確かめる）
- [X] T019 [P] `apps/api/internal/server/server_test.go` に、1MB を超える本体の `/graphql` への POST が処理されずに失敗することのテストを足す

### web: サーバー側の土台

- [X] T020 `apps/web/src/lib/api/server.ts` に `queryApi` を足す（research R13、contracts/web-routes.md）: セッションの Cookie があれば `__Host-chara_q_session=<値>` だけを付け、なければ付けずに `${process.env.API_URL}/graphql` に POST する。`cache: "no-store"`。失敗の扱いは `queryWithSession` と同じ（例外を投げる）。共通の部分は内部の関数にまとめ、`queryWithSession` の動きは変えない
- [X] T021 [P] `apps/web/src/lib/sheets/systems.ts` に、`GameSystemId`（`apps/web/src/lib/game-systems.ts`、`yuyakekoyake`・`beginning-idol`）と GraphQL の `GameSystem`（`YUYAKEKOYAKE`・`BEGINNING_IDOL`）の相互の変換と、URL の値が `GameSystemId` かどうかを確かめる関数を書く。`apps/web/src/lib/sheets/systems.test.ts` にテストを書く
- [X] T022 [P] `apps/web/src/lib/sheets/format.ts` に、日時を `Asia/Tokyo` の `YYYY/MM/DD HH:mm` にする関数を書き（`Intl.DateTimeFormat("ja-JP", { timeZone: "Asia/Tokyo", ... })`）、`apps/web/src/lib/sheets/format.test.ts` で、UTC の日付の変わり目（`2026-11-09T15:30:00Z` → `2026/11/10 00:30`）を確かめる

**Checkpoint**: マイグレーションが適用でき、`task generate` の後に `go build ./...`・`pnpm --filter @chara-q/web typecheck` が通る。`go test ./internal/sheet/...` の検証のテストと、002-auth のテストが通る

---

## Phase 3: User Story 3 - ログインしてシートを作る (Priority: P1)

**Goal**: ログインしたユーザーが、API でシートを作れる（FR-020〜FR-022、FR-001〜FR-007）

**Independent Test**: ログインしたユーザーで `createCharacterSheet` を送り、所有者が自分・`revision` が 1 のシートができること、ログインしていなければ `UNAUTHENTICATED` になること、入力の規則の失敗が項目ごとに返ることを確かめる

### Tests for User Story 3

- [X] T023 [P] [US3] `apps/api/internal/sheet/create_test.go` に、DB を使うテストを書く: 作った行の `owner_id`・`system`・`revision = 1`・`created_at = updated_at = 時計の時刻`、空白を除いた値で保存されること、`data` がそのまま保存されること
- [X] T024 [P] [US3] `apps/api/internal/graph/character_sheet_create_test.go` に、`httptest` のテストを書く: ログインして作る → `characterSheet` の `__typename` がシステムに合った型（`YuyakekoyakeSheet`・`BeginningIdolSheet`）、`owner { id displayName }` が本人。ログインなし → `UNAUTHENTICATED` で行が増えない。キャラクター名が空 → `errors` に `CHARACTER_NAME` の `REQUIRED`、`data` が配列 → `DATA` の `NOT_OBJECT`。応答の本文に作成者のログイン ID が含まれない

### Implementation for User Story 3

- [X] T025 [US3] `apps/api/internal/sheet/create.go` に `Service.Create(ctx, owner uuid.UUID, system string, fields Fields) (db.CharacterSheet, error)` を書く: T011 の検証の後、`CreateCharacterSheet` で作る。所有者は引数だけから決め、入力からは受け取らない（FR-020）
- [X] T026 [US3] `apps/api/internal/graph/character_sheet.resolvers.go` に `createCharacterSheet` を実装する: `auth.ViewerFrom(ctx)` が nil なら `errUnauthenticated()`、入力の失敗は payload の `errors`、成功したら `sheet.ToModel` で返す。あわせて、両方のシステムの型の `owner` の resolver を、`loader.For(ctx).UserByID.Load` から `model.User{ID, DisplayName}` を返す形で実装する

**Checkpoint**: quickstart の「1. 作成と閲覧」の 1・5・6 が API で確かめられる

---

## Phase 4: User Story 2 - ログインせずにシートを見る (Priority: P1)

**Goal**: 誰でも `/sheets/<ID>` でシートの共通部分を見られる（FR-015〜FR-020）

**Independent Test**: シートを用意し、ログインしていない状態で `/sheets/<ID>` を開く。共通の項目・システム・所有者の表示名・日時が表示され、存在しない・形の正しくない ID では「シートが見つかりません」になることを確かめる

### Tests for User Story 2

- [X] T027 [P] [US2] `apps/api/internal/graph/character_sheet_get_test.go` に、`httptest` のテストを書く: ログインなしで `characterSheet(id:)` が共通の項目と `owner.displayName` を返す。存在しない UUID・`"abc"`・削除済み（`sheettest` で `deleted_at` を入れた行）は、どれもエラーなしの `null`。大文字の UUID でも同じシートを返す。所有者が無効（`disabled_at` あり）でも同じように返す。応答に所有者のログイン ID が含まれない
- [X] T028 [P] [US2] `apps/web/src/components/sheets/sheet-common-view.test.tsx` に、閲覧の共通部分のテストを書く: 見出しがキャラクター名、システムの名前・所有者の表示名・作成日時・更新日時、空の項目が「未入力」、メモの改行が保たれる、HTML の文字列がそのまま文字として表示される、システム固有の部分の場所に「このシステムの項目は準備中です」。axe の検査（`apps/web/src/test/axe.ts`）

### Implementation for User Story 2

- [X] T029 [US2] `apps/api/internal/sheet/get.go` に `Service.Get(ctx, id string) (*db.CharacterSheet, error)` を書く: `uuid.Parse` に失敗したら `nil`（エラーにしない、research R5）、`GetCharacterSheet` が行なしなら `nil`
- [X] T030 [US2] `apps/api/internal/graph/character_sheet.resolvers.go` に `characterSheet` を実装する（`nil` なら `null`）
- [X] T031 [P] [US2] `apps/web/src/lib/sheets/system-views.ts` に、`GameSystemId` ごとの部品の表を書く（data-model の「web 側のデータ」）: `body`（システム固有の部分）と `ownerActions`（所有者の操作）。この機能では、どちらのシステムも `body` は「このシステムの項目は準備中です」を表示する部品、`ownerActions` は何も描かない部品にする。各システムの Spec がここに登録すると、ファイルの先頭のコメントに書く
- [X] T032 [US2] `apps/web/src/components/sheets/sheet-common-view.tsx` に、閲覧の共通部分を書く（contracts/web-routes.md の「閲覧の画面」）: 見出し（キャラクター名）、システムの名前（`getGameSystem`）、所有者の表示名、作成日時・更新日時（T022）、プレイヤー名・年齢・性別・メモ（`whitespace-pre-wrap`）、空の項目は「未入力」。`system-views.ts` の `body` と `ownerActions` を差し込む場所を持つ。見出しは `GameSystemScope`（`apps/web/src/components/game-system/game-system-scope.tsx`）でシステムのフォントにする。幅 360px で横スクロールを出さない
- [X] T033 [US2] `apps/web/src/lib/sheets/get-sheet.ts`（`server-only`）に、React の `cache()` で包んだ `getCharacterSheet(id)` を書く: `queryApi` で `characterSheet(id:)` を読み、共通の項目・`__typename`・`owner { id displayName }` を返す（GraphQL Code Generator の `graphql()` でクエリを書く）
- [X] T034 [US2] `apps/web/src/app/sheets/[id]/page.tsx` を作る: `params` を読む部分を `<Suspense>` の中に置き（`cacheComponents`）、`getCharacterSheet` が `null` なら `notFound()`、あれば `SheetCommonView` を表示する。`generateMetadata` で `title` をキャラクター名にする（同じ `getCharacterSheet` を使い、見つからなければ `title` を「シートが見つかりません」にする）。所有者かどうかの判定は、`getViewer()` の ID と `owner.id` を比べて `ownerActions` に渡す
- [X] T035 [P] [US2] `apps/web/src/app/sheets/[id]/not-found.tsx` を作る: 見出し「シートが見つかりません」、説明「URL が正しいか確かめてください」。削除されたことがわかる文言は書かない（FR-018）。一覧（`/sheets`）へのリンク。`apps/web/src/app/sheets/[id]/not-found.test.tsx` に表示と axe のテストを書く

**Checkpoint**: quickstart の「1. 作成と閲覧」の 2〜4 が確かめられる

---

## Phase 5: User Story 1 - ログインせずにシートを探す (Priority: P1) 🎯 MVP

**Goal**: 誰でも `/sheets` で、システム・キーワード・「自分のシートだけ」で絞り込み、並び順を変え、ページを送ってシートを探せる（FR-008〜FR-014、FR-032、FR-034）

**Independent Test**: 25 件以上のシートを用意し、ログインしていない状態とログインした状態で `/sheets` を開く。絞り込み・並び順・ページ送り・URL の共有、ヘッダーのリンクを確かめる

### Tests for User Story 1

- [X] T036 [P] [US1] `apps/api/internal/sheet/list_test.go` に、DB を使うテーブル駆動のテストを書く: システムの絞り込み、キーワード（半角の小文字で全角の大文字の名前が見つかる、プレイヤー名でも見つかる、`%`・`_`・`\`・全角の `％` が文字どおりに扱われる、前後の空白だけのキーワードは条件なし）、所有者の絞り込み、条件の組み合わせ、削除済みが含まれない、3 つの並び順（`sheettest.SetTimes` で日時をそろえ、同じ日時・同じ名前では ID の順になること）、ページの丸め（0・負の数は 1、全ページ数を超えたら最後のページ、0 件なら `page = 1`・`totalPages = 1`）、`totalCount`
- [X] T037 [P] [US1] `apps/api/internal/loader/loader_test.go` に、問い合わせの回数のテストを書く（SC-003）: pgx の `QueryTracer` で `users` への問い合わせを数えるプールを作り、所有者の違うシートを 1 件と 20 件にして `characterSheets { nodes { owner { displayName } } }` を `httptest` で送り、`users` への問い合わせがどちらも 1 回であることを確かめる
- [X] T038 [P] [US1] `apps/api/internal/graph/character_sheet_list_test.go` に、`httptest` のテストを書く: ログインなしで `characterSheets` が返る、`mine: true` はログインなしでは無視される・ログインすると本人のシートだけ、`pageSize` が 20、応答に所有者のログイン ID が含まれない
- [X] T039 [P] [US1] `apps/web/src/lib/sheets/search-params.test.ts` に、URL のクエリの読み取りと組み立てのテストを書く（contracts/web-routes.md の「一覧の URL のクエリ」の表のすべての行）: 知らない `system`・`sort`、`mine` が `1` 以外、`page` が 0・負・小数・文字、同じキーが複数（最初の値）、既定の値は組み立てた URL に含まれないこと
- [X] T040 [P] [US1] `apps/web/src/components/sheets/sheet-search-form.test.tsx` に、検索のフォームのテストを書く: 送信すると条件が URL のクエリになり `page` が含まれない、ログインしていなければ「自分のシートだけ」が表示されない、システムのプルダウンの選択肢が「すべて」・ゆうやけこやけ・ビギニングアイドル、条件があるときだけ「条件を外す」が出る。axe の検査
- [X] T041 [P] [US1] `apps/web/src/components/sheets/pagination.test.tsx` に、ページ送りのテストを書く: 1 ページだけなら表示しない、現在のページに `aria-current="page"`、最初と最後のページで「前へ」「次へ」がない、ページが多いときに「…」で省略する（最初・最後・現在の前後 2 ページ）、各リンクの URL が条件を保つ
- [X] T042 [P] [US1] `apps/web/src/components/sheets/sheet-list.test.tsx` に、一覧のテストを書く: 各行にキャラクター名（`/sheets/<ID>` へのリンク）・プレイヤー名・システム・所有者の表示名・更新日時、0 件で条件ありなら「条件に合うシートはありません」と「条件を外す」、条件なしなら「まだシートはありません」、「全 N 件」
- [X] T043 [P] [US1] `apps/web/src/components/layout/site-header.test.tsx` に、「シート一覧」のリンク（`/sheets`、狭い画面ではアイコンだけで `aria-label="シート一覧"`）のテストを足す

### Implementation for User Story 1

- [X] T044 [US1] `apps/api/internal/sheet/search.go` に `Service.List(ctx, filter Filter, order Order, page int, viewer *uuid.UUID) (Page, error)` を書く（research R8〜R10）: キーワードの前後の空白を除き、空なら `nil`。`mine` はログインしているときだけ `owner_id` にする。`CountCharacterSheets` で件数を数え、`page` を 1〜全ページ数に丸め（0 件なら 1）、並び順に合った `ListCharacterSheetsBy*` を `LIMIT 20 OFFSET (page - 1) * 20` で呼ぶ。返す値は `Rows`・`TotalCount`・`Page`・`PageSize`（20）・`TotalPages`
- [X] T045 [US1] `apps/api/internal/graph/character_sheet.resolvers.go` に `characterSheets` を実装する（各行は `sheet.ToModel`、所有者は DataLoader で解決される）
- [X] T046 [P] [US1] `apps/web/src/lib/sheets/search-params.ts` に、URL のクエリの読み取り（`{ system: GameSystemId | null; q: string; mine: boolean; sort: "updated" | "created" | "name"; page: number }`、知らない値は既定）と、条件から `/sheets?...` を組み立てる関数（既定の値は含めない）を書く。`sort` と GraphQL の `CharacterSheetOrder`（`UPDATED_AT_DESC`・`CREATED_AT_DESC`・`CHARACTER_NAME_ASC`）の変換もここに置く
- [X] T047 [US1] `apps/web/src/components/sheets/sheet-search-form.tsx`（Client Component）を書く（research R14、contracts/web-routes.md の「一覧の画面」）: `next/form` の `<Form action="/sheets">`。キーワード（`TextField`、`name="q"`、ラベル「キーワード」、補足「キャラクター名・プレイヤー名」）、システム（`SelectField`、`name="system"`、「すべて」と 2 つのシステム。「すべて」は空の値にして URL に含めない）、並び順（`SelectField`、`name="sort"`、「更新日時の新しい順」・「作成日時の新しい順」・「キャラクター名の順」）、ログインしているときだけ「自分のシートだけ」のネイティブのチェックボックス（`name="mine" value="1"`、Tailwind で整え、ラベルと関連付ける）、「検索」のボタン、条件があるときだけ「条件を外す」（`/sheets` へのリンク）。空の値の項目がクエリに残る場合は、送信の前に取り除く（`next/form` の動きをドキュメントで確かめる）
- [X] T048 [P] [US1] `apps/web/src/components/sheets/pagination.tsx` を書く: `page`・`totalPages` と URL を組み立てる関数を受け取り、「前へ」「次へ」とページ番号のリンク（`next/link`）を表示する。省略の規則は T041 のとおり。`nav` に `aria-label="ページ"`
- [X] T049 [P] [US1] `apps/web/src/components/sheets/sheet-list.tsx` を書く: 「全 N 件」、各行（キャラクター名のリンク・プレイヤー名・システムの名前・所有者の表示名・更新日時）。`sm` 以上は表の形、`sm` 未満は縦に積んだカードの形にして、幅 360px で横スクロールを出さない。0 件の表示は T042 のとおり
- [X] T050 [US1] `apps/web/src/app/sheets/page.tsx` を作る: 見出し「キャラクターシート一覧」、`metadata` の `title` は「キャラクターシート一覧」。`searchParams` と `getViewer()` を読む部分を `<Suspense>` の中に置き、`search-params.ts` で条件を読み、`queryApi` で `characterSheets` を呼ぶ（`filter: { system, keyword, mine }`、`orderBy`、`page`）。api が丸めた `page` を使って `SheetSearchForm`・`SheetList`・`Pagination` を表示する
- [X] T051 [US1] `apps/web/src/components/layout/site-header.tsx` に、右側の操作の領域の先頭に「シート一覧」のリンク（`/sheets`）を足す（research R16。最初はロゴの右に置いたが、ずれて見えたため移した）: `lucide-react` のアイコンと文字、`sm` 未満は文字を隠して `aria-label="シート一覧"`。`/sheets` 以下のときに `aria-current="page"` を付けるため、現在のパスの判定が要るなら、その部分だけを小さな Client Component（`usePathname`）にする。幅 360px で、ロゴ・リンク・テーマの切り替え・アカウントのメニューが重ならないことを確かめる

**Checkpoint**: quickstart の「2. 一覧と検索」が確かめられる。ここまでで、シートを API で作り、誰でも探して見られる（MVP）

---

## Phase 6: User Story 4 - 所有者がシートを更新する (Priority: P1)

**Goal**: 所有者だけが、API でシートを更新できる。古い内容をもとにした更新は受け付けない（FR-025〜FR-028）

**Independent Test**: 2 人のユーザーで、所有者が更新できること、もう 1 人（管理者を含む）とログインしていない訪問者は更新できないこと、古い `revision` での更新が `REVISION_CONFLICT` になることを確かめる

### Tests for User Story 4

- [X] T052 [P] [US4] `apps/api/internal/sheet/update_test.go` に、DB を使うテーブル駆動のテストを書く: 所有者の更新で `revision` が 1 増え `updated_at` が時計の時刻になり、`created_at`・`owner_id`・`system` が変わらない。所有者でない → `ErrForbidden`、存在しない・削除済み → `ErrNotFound`、古い `revision` → `ErrRevisionConflict`（判定の順は `NOT_FOUND` → `FORBIDDEN` → `REVISION_CONFLICT`）。同じ `revision` の更新を 2 つの goroutine から同時に送ると、成功は 1 回だけ
- [X] T053 [P] [US4] `apps/api/internal/graph/character_sheet_update_test.go` に、`httptest` のテストを書く: 所有者は成功、管理者を含む所有者でないユーザーは `FORBIDDEN`、ログインなしは `UNAUTHENTICATED`、古い `revision` は `errors` に `field: null` の `REVISION_CONFLICT`、どの失敗でも行が変わらない（SC-004、SC-007）

### Implementation for User Story 4

- [X] T054 [US4] `apps/api/internal/sheet/update.go` に `Service.Update(ctx, owner uuid.UUID, id string, revision int, fields Fields) (db.CharacterSheet, error)` を書く（research R6）: 検証の後、`UpdateCharacterSheet` の 1 文で更新し、0 行なら `GetCharacterSheetState` で判定する（行なし・削除済み → `ErrNotFound`、所有者が違う → `ErrForbidden`、それ以外 → `ErrRevisionConflict`）。形の正しくない ID は `ErrNotFound`
- [X] T055 [US4] `apps/api/internal/graph/character_sheet.resolvers.go` に `updateCharacterSheet` を実装する: ログインなし → `errUnauthenticated()`、入力の失敗と `ErrRevisionConflict` → payload の `errors`、`ErrNotFound` → `errNotFound()`、`ErrForbidden` → `errForbidden()`

**Checkpoint**: quickstart の「3. 更新」が確かめられる

---

## Phase 7: User Story 5 - 他の人のシートを複製して自分のものにする (Priority: P2)

**Goal**: ログインしたユーザーが、API でどのシートでも複製でき、複製は本人の所有になる（FR-023〜FR-025）

**Independent Test**: ユーザー A のシートをユーザー B が複製し、B の所有で名前の末尾に `（コピー）` が付くこと、複製と元が互いに影響しないことを確かめる

### Tests for User Story 5

- [X] T056 [P] [US5] `apps/api/internal/sheet/copy_test.go` に、DB を使うテストを書く: 複製の `owner_id` が複製した人、`system`・共通の項目・`data` が同じ、`created_at = updated_at = 時計の時刻`、`revision = 1`。名前が 95 文字以下ならそのまま `（コピー）` が付き、100 文字の名前は先頭 95 文字 + `（コピー）` の 100 文字。複製を更新しても元が変わらない。削除済みの元 → `ErrNotFound`。自分のシートも複製できる
- [X] T057 [P] [US5] `apps/api/internal/graph/character_sheet_copy_test.go` に、`httptest` のテストを書く: ログインなし → `UNAUTHENTICATED`、存在しない・削除済み → `NOT_FOUND`、成功したら `owner` が本人

### Implementation for User Story 5

- [X] T058 [US5] `apps/api/internal/sheet/copy.go` に `Service.Copy(ctx, owner uuid.UUID, id string) (db.CharacterSheet, error)` を書く: 形の正しくない ID と、`CopyCharacterSheet` が行なしなら `ErrNotFound`
- [X] T059 [US5] `apps/api/internal/graph/character_sheet.resolvers.go` に `copyCharacterSheet` を実装する

**Checkpoint**: quickstart の「4. 複製」が確かめられる

---

## Phase 8: User Story 6 - 所有者がシートを削除する (Priority: P2)

**Goal**: 所有者だけが、API でシートを削除できる。データは残り、管理者が DB 操作で戻せる（FR-029、FR-030）

**Independent Test**: 所有者が削除すると、一覧・閲覧・更新・複製の対象から外れ、DB には内容と削除した日時が残ること、所有者以外は削除できないこと、取り消しの SQL で元に戻ることを確かめる

### Tests for User Story 6

- [X] T060 [P] [US6] `apps/api/internal/sheet/delete_test.go` に、DB を使うテストを書く: 所有者の削除で `deleted_at` が時計の時刻になり、`updated_at`・`revision`・内容は変わらない。所有者でない → `ErrForbidden`、もう一度削除 → `ErrNotFound`。削除の後、`Get` が `nil`、`List` に含まれない、`Update`・`Copy` が `ErrNotFound`。[data-model.md](data-model.md) の取り消しの SQL を実行すると、`Get` と `List` の元の位置に戻る
- [X] T061 [P] [US6] `apps/api/internal/graph/character_sheet_delete_test.go` に、`httptest` のテストを書く: 管理者を含む所有者でないユーザーは `FORBIDDEN`、ログインなしは `UNAUTHENTICATED`、成功したら `deletedId` が返り、`characterSheet` が `null`

### Implementation for User Story 6

- [X] T062 [US6] `apps/api/internal/sheet/delete.go` に `Service.Delete(ctx, owner uuid.UUID, id string) error` を書く: `DeleteCharacterSheet` が 0 行なら `GetCharacterSheetState` で判定する（行なし・削除済み → `ErrNotFound`、所有者が違う → `ErrForbidden`）。形の正しくない ID は `ErrNotFound`
- [X] T063 [US6] `apps/api/internal/graph/character_sheet.resolvers.go` に `deleteCharacterSheet` を実装する

**Checkpoint**: quickstart の「5. 削除」が確かめられる

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: 決定の記録、文書、全体の確認

- [X] T064 [P] `docs/adr/0005-sheet-access-and-legacy-owner.md` の「後続の Spec で決めること」の「シートの削除を許すか」に、決定（所有者だけが削除できる、データは消さずに `deleted_at` を記録する、取り消しは管理者の DB 操作、specs/003-character-sheet-base の Clarifications）を追記し、「公開範囲」の表に「削除: シートの所有者だけ」の行を足す
- [X] T065 [P] `.specify/memory/constitution.md` の「シートの公開範囲」に「削除はシートの所有者だけができる」を足し、バージョンを 1.4.0、`Last Amended` を作業した日にする（原則の節の拡張のため MINOR、plan の「Phase 1 後の再確認」）
- [X] T066 [P] `README.md` と `apps/api/README.md`（あれば）に、シートの画面（`/sheets`）と、削除の取り消しの手順（[data-model.md](data-model.md) へのリンク）を書く。書く場所がなければ、このタスクは何もしない
- [X] T067 `task lint`・`task test`・`task check:generated` を実行し、すべて通ることを確かめる（実施メモ: `task lint`・`task test` は通った（api はすべて ok、web は 263 件）。`task check:generated` はコミット済みの内容と比べるため、コミットの後に実行する。代わりに `task generate` をもう一度実行し、生成コードが変わらないことを確かめた。`next build` で、ヘッダーのリンクの `usePathname()` が `<Suspense>` の外にあり事前描画を止める不具合を見つけて直した）
- [X] T068 [quickstart.md](quickstart.md) の 1〜6 をすべて実施し、結果をこのタスクに追記する。幅 360px の `/sheets` と `/sheets/<ID>`、ライトとダークの両方のテーマで、横スクロールと重なりがないことを内蔵ブラウザで確かめる（実施メモ: 内蔵ブラウザ（Chromium）で、api 8091 番・web 3001 番（`next build` + `next start`）・テスト用の DB（27 件のシートを投入）で確かめた。1〜5 はすべて期待どおり（作成・更新・複製・削除と権限の拒否は、ログインしたタブから web 経由の `/graphql` に送った）。照合順序はローカルの `ja-x-icu` を確認済み、Neon は未確認。幅 360px の `/sheets`（ライト）と `/sheets/<ID>`（ダーク）で横スクロールなし。Firefox・Safari は未確認）

---

## Phase 10: 逆の順と見出しでの並び替え（2026-11-11 に追加、FR-011）

**Goal**: 一覧の並び順に逆の順を加え、表の見出し（キャラクター名・更新日時）を選んで並び替えられるようにする（spec の Clarifications の Session 2026-11-11、research R9・R14）

- [X] T069 `schema/graphql/character_sheet.graphqls` の `CharacterSheetOrder` に `UPDATED_AT_ASC`・`CREATED_AT_ASC`・`CHARACTER_NAME_DESC` を足す。`apps/api/db/queries/character_sheets.sql` に `ListCharacterSheetsByUpdatedAsc`・`ByCreatedAsc`・`ByNameDesc` を足し、`task generate` で再生成する
- [X] T070 `apps/api/internal/sheet/search.go` の `Order` に 3 つの値を足して、新しいクエリにつなぐ。`internal/graph/sheet_convert.go` の `toSheetOrder` も足す。`internal/sheet/list_test.go` の並び順と同順のテストに、逆の順を足す
- [X] T071 [P] `apps/web/src/lib/sheets/search-params.ts` の `SheetSort` に `updated-asc`・`created-asc`・`name-desc` を足し、見出しを選んだときの次の並び順を返す関数を作る。テストを足す
- [X] T072 `apps/web/src/components/sheets/sheet-list.tsx` の表の見出しのうち、キャラクター名と更新日時を並び替えのリンクにする（`aria-sort`・矢印・視覚的に隠した文）。`app/sheets/page.tsx` から条件を渡す。テストを足す
- [X] T073 `task lint`・`task test`・`next build` を通し、内蔵ブラウザで quickstart の 4 を確かめる（実施メモ: `task lint`・api のテスト（DB あり）・web のテスト 285 件・`next build` が通った。内蔵ブラウザ（api 8091 番・web 3001 番・テスト用の DB に 25 件を投入）で、見出しを選ぶと名前の順→逆の順に切り替わり、システムの絞り込みは保たれ、ページは 1 に戻り、プルダウンも同じ並び順になることを確かめた）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup（Phase 1）**: すぐに始められる。T002 の結果は T003・T004 の照合順序に使う
- **Foundational（Phase 2）**: Setup の後。すべてのユーザーストーリーの前提
  - T003 → T004・T005 → T006 → T007 → T008 → T009 の順（生成コードが前のタスクに依存する）
  - T010〜T014 は T009 の後。T010・T011・T012・T014 は並行できる。T013 は T010 の後
  - T015 は T006 の後、T016〜T018 は T009 と T013 の後、T019 は T017 の後
  - web の T020〜T022 は T009（GraphQL Code Generator の生成）の後に並行できる
- **US3（Phase 3）**: Foundational の後。他のストーリーのテストのデータは `sheettest` で作れるため、US2・US1 は US3 を待たずに始めてもよいが、resolver のファイル（`character_sheet.resolvers.go`）と `owner` の resolver（T026）を共有するため、順に進めるのが安全
- **US2（Phase 4）**: Foundational の後（T026 の `owner` の resolver を使う）
- **US1（Phase 5）**: Foundational の後（T026 の `owner` の resolver、T029 は不要）
- **US4（Phase 6）**: Foundational の後
- **US5・US6（Phase 7・8）**: Foundational の後。US6 のテスト（T060）は `Update`・`Copy`・`List` を使うため、US4・US5・US1 の後
- **Polish（Phase 9）**: すべてのストーリーの後

### User Story Dependencies

- **US3（P1）**: 依存なし。`owner` の resolver（T026）を他のストーリーが使う
- **US2（P1）**: T026 に依存。web の画面は US3 がなくても `sheettest` のデータで確かめられる
- **US1（P1）**: T026 に依存。一覧の行から US2 の閲覧画面にリンクする（画面がなくても一覧は動く）
- **US4（P1）**: 依存なし（行は `sheettest` で作る）
- **US5（P2）**: 依存なし
- **US6（P2）**: 削除の後の確認に US1（`List`）・US4（`Update`）・US5（`Copy`）を使う

### Within Each User Story

- テストを先に書き、失敗することを確かめてから実装する
- `internal/sheet` の処理 → resolver → web の部品 → 画面 の順
- `character_sheet.resolvers.go` は複数のストーリーが同じファイルを編集するため、[P] を付けない

### Parallel Opportunities

- Foundational の T010・T011・T012・T014、web の T020〜T022
- 各ストーリーのテストのタスク（[P] のもの）は、互いに別のファイルなので並行できる
- US1 の web の部品（T046・T048・T049）と api の処理（T044）は並行できる

---

## Parallel Example: User Story 1

```text
# テストをまとめて書く:
Task: "T036 apps/api/internal/sheet/list_test.go の絞り込み・並び順・ページの丸めのテスト"
Task: "T037 apps/api/internal/loader/loader_test.go の問い合わせの回数のテスト"
Task: "T039 apps/web/src/lib/sheets/search-params.test.ts の URL のクエリのテスト"
Task: "T041 apps/web/src/components/sheets/pagination.test.tsx のページ送りのテスト"

# 実装の中で並行できるもの:
Task: "T044 apps/api/internal/sheet/search.go の Service.List"
Task: "T048 apps/web/src/components/sheets/pagination.tsx"
Task: "T049 apps/web/src/components/sheets/sheet-list.tsx"
```

## Parallel Example: Foundational

```text
# T009 の後に並行できるもの:
Task: "T010 apps/api/internal/sheet/system.go のシステムの値の変換"
Task: "T011 apps/api/internal/sheet/validate.go の共通の項目の検証"
Task: "T014 apps/api/internal/sheet/sheettest/sheettest.go のテストの補助"
Task: "T021 apps/web/src/lib/sheets/systems.ts のシステムの値の変換"
Task: "T022 apps/web/src/lib/sheets/format.ts の日時の表示"
```

---

## Implementation Strategy

### MVP First（US3 → US2 → US1）

1. Phase 1: Setup（照合順序の確認）
2. Phase 2: Foundational（全ストーリーの前提）
3. Phase 3〜5: 作成（API）・閲覧・一覧と検索
4. **STOP and VALIDATE**: quickstart の 1・2 を確かめる。ここで「シートを作り、誰でも探して見られる」状態になる

### Incremental Delivery

1. Setup + Foundational → 土台ができる
2. US3 → API でシートを作れる
3. US2 → 誰でも URL でシートを見られる
4. US1 → 誰でも一覧で探せる（MVP）
5. US4 → 所有者が更新できる（各システムの Spec の編集画面の前提）
6. US5 → 複製できる
7. US6 → 削除できる
8. Polish → ADR・constitution の追記、全体の確認

---

## Notes

- [P] のタスクは別のファイルで、未完了のタスクに依存しない
- [Story] のラベルで、タスクとユーザーストーリーを対応させる
- テストは先に書き、失敗することを確かめる
- タスクか論理的なまとまりごとにコミットする。生成コードは元の変更と同じコミットに入れる
- 作成・編集・複製・削除の画面は、この機能では作らない（各システムの Spec で作る、FR-033）
