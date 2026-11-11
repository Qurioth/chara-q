-- 認証（specs/002-auth）。表の設計は specs/002-auth/data-model.md を参照する。
-- パスワードは Argon2id、招待コード・再設定コード・セッション ID は SHA-256 のハッシュだけを保存する（ADR 0006）。

-- +goose Up

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    -- 入力のとおりの大文字・小文字で保存する。一意性は lower(login_id) で確かめる。本人にだけ見せる
    login_id text NOT NULL CHECK (login_id ~ '^[A-Za-z0-9_-]{3,32}$'),
    -- 公開する。重複してよい
    display_name text NOT NULL CHECK (
        char_length(display_name) BETWEEN 1 AND 32
        AND display_name = btrim(display_name)
        AND display_name !~ '[[:cntrl:]\u2028\u2029]'
    ),
    -- Argon2id の PHC 形式の文字列
    password_hash text NOT NULL,
    is_admin boolean NOT NULL DEFAULT false,
    -- NULL なら有効
    disabled_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_login_id_key ON users (lower(login_id));

CREATE TABLE sessions (
    -- セッション ID の SHA-256
    token_hash bytea PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

CREATE TABLE invitations (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    -- 正規化した招待コードの SHA-256
    code_hash bytea NOT NULL UNIQUE,
    issued_by uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    used_by uuid UNIQUE REFERENCES users (id),
    CHECK ((used_at IS NULL) = (used_by IS NULL))
);

CREATE TABLE password_reset_codes (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- 正規化した再設定コードの SHA-256
    code_hash bytea NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    used_at timestamptz
);

CREATE INDEX password_reset_codes_user_id_idx ON password_reset_codes (user_id);

-- ログインとパスワードの再設定の、続けて失敗した記録。存在しないログイン ID も記録する
CREATE TABLE auth_failures (
    purpose text NOT NULL CHECK (purpose IN ('login', 'password_reset')),
    -- 前後の空白を除き、小文字にしたログイン ID
    login_id_key text NOT NULL,
    failure_count integer NOT NULL CHECK (failure_count >= 0),
    last_failed_at timestamptz NOT NULL,
    locked_until timestamptz,
    PRIMARY KEY (purpose, login_id_key)
);

CREATE INDEX auth_failures_last_failed_at_idx ON auth_failures (last_failed_at);

-- +goose Down

DROP TABLE auth_failures;
DROP TABLE password_reset_codes;
DROP TABLE invitations;
DROP TABLE sessions;
DROP TABLE users;
