-- キャラクターシート（specs/003-character-sheet-base/data-model.md）。
-- 削除済み（deleted_at が NULL でない）の行は、GetCharacterSheetState 以外では読み書きしない。
--
-- 一覧と件数の絞り込みは、7 つのクエリで同じ WHERE を使う。キーワードは、列とキーワードの両方を
-- lower(normalize(x, NFKC)) で正規化して部分一致で比べ、正規化の後に \ % _ をエスケープする（research R8）。

-- name: CreateCharacterSheet :one
INSERT INTO character_sheets (
    system, owner_id, character_name, player_name, age, sex, memo, data, created_at, updated_at
) VALUES (
    @system, @owner_id, @character_name, @player_name, @age, @sex, @memo, @data, @now, @now
)
RETURNING *;

-- name: GetCharacterSheet :one
SELECT * FROM character_sheets
WHERE id = @id AND deleted_at IS NULL;

-- name: GetCharacterSheetState :one
-- 更新・削除が 0 行だったときに、理由を判定するために読む（削除済みも含める）。
SELECT owner_id, revision, deleted_at FROM character_sheets
WHERE id = @id;

-- name: CountCharacterSheets :one
SELECT count(*) FROM character_sheets
WHERE deleted_at IS NULL
  AND (sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system)::text)
  AND (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id)::uuid)
  AND (
    sqlc.narg(keyword)::text IS NULL
    OR lower(normalize(character_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
    OR lower(normalize(player_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
  );

-- name: ListCharacterSheetsByUpdated :many
SELECT * FROM character_sheets
WHERE deleted_at IS NULL
  AND (sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system)::text)
  AND (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id)::uuid)
  AND (
    sqlc.narg(keyword)::text IS NULL
    OR lower(normalize(character_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
    OR lower(normalize(player_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
  )
ORDER BY updated_at DESC, id DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: ListCharacterSheetsByUpdatedAsc :many
-- 更新日時の古い順。character_sheets_updated_idx を逆向きに読む（research R9）。
SELECT * FROM character_sheets
WHERE deleted_at IS NULL
  AND (sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system)::text)
  AND (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id)::uuid)
  AND (
    sqlc.narg(keyword)::text IS NULL
    OR lower(normalize(character_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
    OR lower(normalize(player_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
  )
ORDER BY updated_at ASC, id ASC
LIMIT @row_limit OFFSET @row_offset;

-- name: ListCharacterSheetsByCreated :many
SELECT * FROM character_sheets
WHERE deleted_at IS NULL
  AND (sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system)::text)
  AND (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id)::uuid)
  AND (
    sqlc.narg(keyword)::text IS NULL
    OR lower(normalize(character_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
    OR lower(normalize(player_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
  )
ORDER BY created_at DESC, id DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: ListCharacterSheetsByCreatedAsc :many
-- 作成日時の古い順。character_sheets_created_idx を逆向きに読む（research R9）。
SELECT * FROM character_sheets
WHERE deleted_at IS NULL
  AND (sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system)::text)
  AND (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id)::uuid)
  AND (
    sqlc.narg(keyword)::text IS NULL
    OR lower(normalize(character_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
    OR lower(normalize(player_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
  )
ORDER BY created_at ASC, id ASC
LIMIT @row_limit OFFSET @row_offset;

-- name: ListCharacterSheetsByName :many
-- 照合順序は ja-x-icu（ひらがなとカタカナは五十音の順、漢字は文字の順、research R9）。
SELECT * FROM character_sheets
WHERE deleted_at IS NULL
  AND (sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system)::text)
  AND (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id)::uuid)
  AND (
    sqlc.narg(keyword)::text IS NULL
    OR lower(normalize(character_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
    OR lower(normalize(player_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
  )
ORDER BY character_name COLLATE "ja-x-icu" ASC, id ASC
LIMIT @row_limit OFFSET @row_offset;

-- name: ListCharacterSheetsByNameDesc :many
-- キャラクター名の逆の順。character_sheets_name_idx を逆向きに読む（research R9）。
SELECT * FROM character_sheets
WHERE deleted_at IS NULL
  AND (sqlc.narg(system)::text IS NULL OR system = sqlc.narg(system)::text)
  AND (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id)::uuid)
  AND (
    sqlc.narg(keyword)::text IS NULL
    OR lower(normalize(character_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
    OR lower(normalize(player_name, NFKC)) LIKE '%' || replace(replace(replace(lower(normalize(sqlc.narg(keyword)::text, NFKC)), '\', '\\'), '%', '\%'), '_', '\_') || '%' ESCAPE '\'
  )
ORDER BY character_name COLLATE "ja-x-icu" DESC, id DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: UpdateCharacterSheet :one
-- 所有者とリビジョンの確認と更新を 1 文で行う（research R6）。0 行なら GetCharacterSheetState で理由を判定する。
UPDATE character_sheets SET
    character_name = @character_name,
    player_name = @player_name,
    age = @age,
    sex = @sex,
    memo = @memo,
    data = @data,
    revision = revision + 1,
    updated_at = @now
WHERE id = @id
  AND owner_id = @owner_id
  AND deleted_at IS NULL
  AND revision = @revision
RETURNING *;

-- name: CopyCharacterSheet :one
-- 名前は先頭 95 文字に「（コピー）」（5 文字）を付け、上限の 100 文字に収める（research R12）。
INSERT INTO character_sheets (
    system, owner_id, character_name, player_name, age, sex, memo, data, created_at, updated_at
)
SELECT system, sqlc.arg(owner_id)::uuid, left(character_name, 95) || '（コピー）', player_name, age, sex, memo, data,
       sqlc.arg(now)::timestamptz, sqlc.arg(now)::timestamptz
FROM character_sheets
WHERE id = sqlc.arg(id)::uuid AND deleted_at IS NULL
RETURNING *;

-- name: DeleteCharacterSheet :execrows
-- updated_at と revision は変えない。取り消すと削除前と同じ並び順に戻る（research R7）。
UPDATE character_sheets SET deleted_at = sqlc.arg(now)::timestamptz
WHERE id = @id AND owner_id = @owner_id AND deleted_at IS NULL;
