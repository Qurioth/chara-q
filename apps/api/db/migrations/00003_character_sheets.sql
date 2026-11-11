-- キャラクターシートの共通基盤（specs/003-character-sheet-base）。
-- 表の設計は specs/003-character-sheet-base/data-model.md を参照する。
-- 共通の項目は列、システム固有の部分は data（JSONB）に置く（constitution III、research R1）。

-- +goose Up

CREATE TABLE character_sheets (
    -- URL の /sheets/<id> に使う
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    -- 作成のあとは変えない
    system text NOT NULL CHECK (system IN ('yuyakekoyake', 'beginning_idol')),
    -- 所有者。アプリからは変えない（DB 操作でだけ移す、ADR 0005）
    owner_id uuid NOT NULL REFERENCES users (id),
    character_name text NOT NULL CHECK (
        char_length(character_name) BETWEEN 1 AND 100
        AND character_name = btrim(character_name)
        AND character_name !~ '[[:cntrl:]  ]'
    ),
    -- 自由な文字列。所有者の表示名とは別
    player_name text NOT NULL DEFAULT '' CHECK (
        char_length(player_name) <= 50
        AND player_name = btrim(player_name)
        AND player_name !~ '[[:cntrl:]  ]'
    ),
    age text NOT NULL DEFAULT '' CHECK (
        char_length(age) <= 20
        AND age = btrim(age)
        AND age !~ '[[:cntrl:]  ]'
    ),
    -- ビギニングアイドルでは空のまま
    sex text NOT NULL DEFAULT '' CHECK (
        char_length(sex) <= 20
        AND sex = btrim(sex)
        AND sex !~ '[[:cntrl:]  ]'
    ),
    -- 改行（\n）とタブだけを許す。\r は保存の前に \n にそろえる
    memo text NOT NULL DEFAULT '' CHECK (
        char_length(memo) <= 10000
        AND memo !~ '[\x01-\x08\x0b-\x1f\x7f\u0080-\u009f  ]'
    ),
    -- システム固有の部分。中身は各システムの Spec が決める。大きさ（256KB）は Go で確かめる
    data jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(data) = 'object'),
    -- 更新のたびに 1 増やす（同時の更新の検出、research R6）
    revision integer NOT NULL DEFAULT 1 CHECK (revision >= 1),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    -- NULL なら削除されていない（論理削除、research R7）
    deleted_at timestamptz
);

-- 並び順と絞り込みの索引は、削除されていない行だけの部分索引にする
CREATE INDEX character_sheets_updated_idx ON character_sheets (updated_at DESC, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX character_sheets_created_idx ON character_sheets (created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX character_sheets_name_idx ON character_sheets (character_name COLLATE "ja-x-icu", id)
    WHERE deleted_at IS NULL;
CREATE INDEX character_sheets_owner_idx ON character_sheets (owner_id, updated_at DESC)
    WHERE deleted_at IS NULL;

-- +goose Down

DROP TABLE character_sheets;
