---

description: "Task list for 認証（招待コードでの登録、ログイン・ログアウト）"
---

# Tasks: 認証（招待コードでの登録、ログイン・ログアウト）

**Input**: Design documents from `specs/002-auth/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/](contracts/), [quickstart.md](quickstart.md)

**Tests**: constitution VI（DB を触るテストは実際の Postgres、`task test` が通らない変更はマージしない）と
[research.md](research.md) の R15 に従い、各ストーリーにテストのタスクを含める。api はテーブル駆動、DB を
使うテストは `TEST_DATABASE_URL` の Postgres、HTTP は `httptest`。web は Vitest + Testing Library。
中継を通した Cookie、ブラウザごとの動作、幅 360px の表示は [quickstart.md](quickstart.md) で手で確かめる。

**Organization**: ユーザーストーリーごとに分ける。番号は spec.md のもの（US1〜US6）。
フェーズは優先度の順（P1: US1・US2 → P2: US3〜US6）。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 並行して進められる（別のファイルで、未完了のタスクに依存しない）
- **[Story]**: どのユーザーストーリーのタスクか（US1〜US6）
- パスはリポジトリのルートからの相対パス

## 全タスク共通の注意

- web のコードを書く前に、`apps/web/AGENTS.md` の指示に従い、`apps/web/node_modules/next/dist/docs/` の
  該当ガイドを読む（rewrites: `01-app/03-api-reference/05-config/01-next-config-js/rewrites.md`、
  Cache Components と認証: `01-app/02-guides/authentication-with-cache-components.md`、cookies:
  `01-app/03-api-reference/04-functions/cookies.md`、`useSearchParams`・`useRouter`・`redirect`・`notFound` の
  各リファレンス）。
- 生成コードは手で直さない。スキーマ・SQL・sqlc の設定を変えたら `task generate` を実行し、生成結果を
  元の変更と同じコミットに入れる（CLAUDE.md）。
- マイグレーションは `task db:new -- auth` で作る。ORM は使わない。
- Go の依存は `apps/api` で `go get`、web の依存は `pnpm --filter @chara-q/web add` で入れ、
  `package.json` には厳密なバージョン（`^` なし）で書く。
- パスワード・招待コード・再設定コード・セッション ID の値を、ログ・エラーの本文・テストの失敗の
  メッセージに出さない（FR-010）。
- 文言は日本語。web のエラーの文言は `apps/web/src/lib/auth/messages.ts` だけに書く。
- コミットは CLAUDE.md の形式（`Refs: specs/002-auth`）で、論理的なまとまりごとに分ける。

---

## Phase 1: Setup（中継の経路と、Cookie の往復の確認）

**Purpose**: ブラウザが web とだけ通信する経路を作り、Cookie が往復することを最初に確かめる（FR-001、FR-004、FR-048）

- [X] T001 `apps/web/next.config.ts` に `rewrites()` を追加し、`/graphql` を `${process.env.API_URL}/graphql` に中継する（[contracts/http-and-cookie.md](contracts/http-and-cookie.md)）。`API_URL` が未設定、または末尾が `/` のときは、設定の読み込みで `Error` を投げてビルドと起動を失敗させる。既存の `cacheComponents: true`・`partialPrefetching: true` は残す
- [X] T002 [P] `apps/web/src/lib/urql-provider.tsx` の接続先を `NEXT_PUBLIC_API_URL` から相対パスの `"/graphql"` に変える。`apps/web/.env.example` の `NEXT_PUBLIC_API_URL=…` を `API_URL=http://localhost:8080` に置き換え、説明のコメント（ブラウザに出さないため `NEXT_PUBLIC_` を付けない）を書く。`apps/web/README.md` の `NEXT_PUBLIC_API_URL` の記述も `API_URL` に直す
- [X] T003 [P] `.github/workflows/ci-web.yml` の build の手順の環境変数を `NEXT_PUBLIC_API_URL: http://localhost:8080/graphql` から `API_URL: http://localhost:8080` に変える
- [X] T004 確認用の一時的な処理を `apps/api/internal/server/server.go` に入れる: `cfg.Development` のときだけ、`/graphql` へのリクエストのヘッダー `X-Cookie-Check: set` があれば `Set-Cookie: __Host-chara_q_cookie_check=ok; Path=/; HttpOnly; Secure; SameSite=Lax` を付け、リクエストに `__Host-chara_q_cookie_check` の Cookie があれば、その有無を `logger.Info("cookie check", "received", true)` でログに出す（値は出さない）。コメントに「T006 で消す」と書く
- [X] T005 [quickstart.md](quickstart.md) の「1. 中継と Cookie の確認」の 1〜3・5 を、T004 の処理で確かめる（ブラウザのコンソールから `fetch("/graphql", { method: "POST", headers: { "Content-Type": "application/json", "X-Cookie-Check": "set" }, body: JSON.stringify({ query: "{ health }" }) })` を実行し、続けてヘッダーなしで同じリクエストを送る）。Chrome と Firefox で、`localhost:3000` に Cookie が保存され、api のログに `received: true` が出ることを記録する。Safari は `next dev --experimental-https` を使うかどうかを利用者に確認してから行う。**どれかが確かめられなければ、ここで止めて利用者に報告し、research R1 の代案（Route Handler での中継）を検討する**（実施メモ: 内蔵ブラウザ（Chromium）で、`next build` + `next start` の 3001 番と api の 8091 番で確かめた。既に動いていた `next dev` と重ならないようにするため。Firefox と Safari は未確認）
- [X] T006 T004 の確認用の処理を `apps/api/internal/server/server.go` から消す（T005 で確かめられた後）

**Checkpoint**: ブラウザの `/graphql` への通信が `localhost:3000` だけになり、Chrome と Firefox で Cookie の往復が確かめられている。`pnpm --filter @chara-q/web build`（`API_URL` を設定して）が通る

---

## Phase 2: Foundational（全ストーリーの前提）

**Purpose**: CORS の撤去、DB の表、認証の共通の処理、スキーマ、Viewer の読み込み、web のサーバー側の土台

**⚠️ CRITICAL**: このフェーズが終わるまで、ストーリーの作業を始めない

### api: CORS の撤去と CSRF の条件（FR-005、FR-047）

- [X] T007 CORS の許可を外す: `apps/api/internal/server/server.go` から `cors.Handler` のミドルウェアと `github.com/go-chi/cors` の import を消す。`apps/api/internal/config/config.go` から `CORSAllowedOrigins` と `CORS_ALLOWED_ORIGINS` の読み込みを消し、`apps/api/internal/config/config_test.go` の該当のケースを消す。`apps/api` で `go mod tidy` を実行し、`apps/api/go.mod` から `github.com/go-chi/cors` が消えることを確かめる。`apps/api/.env.example` から `CORS_ALLOWED_ORIGINS=…` の行を消す
- [X] T008 [P] `.github/workflows/deploy-api.yml` の `env_vars` から `CORS_ALLOWED_ORIGINS=${{ vars.CLOUD_RUN_CORS_ALLOWED_ORIGINS }}` の行を消す。`infra/gcp.md` の環境変数の表から `CORS_ALLOWED_ORIGINS`、GitHub 変数の表から `CLOUD_RUN_CORS_ALLOWED_ORIGINS` を消し、「設定済みなら GitHub の Settings → Secrets and variables → Actions → Variables で削除する」と書く
- [X] T009 `apps/api/internal/server/server_test.go` を作り、`httptest` で次をテーブル駆動で確かめる（DB は使わない。resolver は `health` だけが要る最小のものを渡すか、DB が要らない形で `server.New` を組み立てる）: `GET /graphql?query=mutation{__typename}` が「GET requests only allow query operations」を返し処理されないこと、`POST` で `Content-Type: text/plain`・`application/x-www-form-urlencoded`・`multipart/form-data` が拒否されること、`POST` で `application/json; charset=utf-8` が受け付けられること、`OPTIONS` と通常の応答に `Access-Control-Allow-Origin` が付かないこと（[contracts/http-and-cookie.md](contracts/http-and-cookie.md)）。ミューテーションの例はスキーマに `Mutation` ができてから（T016 の後、T034 で）足す

### api: DB の表と sqlc

- [X] T010 `task db:new -- auth` でマイグレーションを作り、`apps/api/db/migrations/00002_auth.sql` に [data-model.md](data-model.md) の 5 つの表を書く。`-- +goose Down` では逆の順に `DROP TABLE`。制約は data-model のとおりに書く:
  - `users`: `id uuid PRIMARY KEY DEFAULT uuidv7()`、`login_id text NOT NULL CHECK (login_id ~ '^[A-Za-z0-9_-]{3,32}$')`、`display_name text NOT NULL CHECK (char_length(display_name) BETWEEN 1 AND 32)` と、`display_name = btrim(display_name)` かつ制御文字（`[[:cntrl:]]`）・U+2028・U+2029 を含まない `CHECK`、`password_hash text NOT NULL`、`is_admin boolean NOT NULL DEFAULT false`、`disabled_at timestamptz NULL`、`created_at`・`updated_at timestamptz NOT NULL DEFAULT now()`、一意の索引 `users_login_id_key ON users (lower(login_id))`
  - `sessions`: `token_hash bytea PRIMARY KEY`、`user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE`、`created_at`・`last_used_at timestamptz NOT NULL DEFAULT now()`、`expires_at timestamptz NOT NULL`、索引 `sessions (user_id)` と `sessions (expires_at)`
  - `invitations`: `id uuid PRIMARY KEY DEFAULT uuidv7()`、`code_hash bytea NOT NULL UNIQUE`、`issued_by uuid NOT NULL REFERENCES users(id)`、`created_at timestamptz NOT NULL DEFAULT now()`、`expires_at timestamptz NOT NULL`、`used_at timestamptz NULL`、`used_by uuid NULL UNIQUE REFERENCES users(id)`、`CHECK ((used_at IS NULL) = (used_by IS NULL))`
  - `password_reset_codes`: `id uuid PRIMARY KEY DEFAULT uuidv7()`、`user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE`、`code_hash bytea NOT NULL UNIQUE`、`created_at timestamptz NOT NULL DEFAULT now()`、`expires_at timestamptz NOT NULL`、`used_at timestamptz NULL`、索引 `password_reset_codes (user_id)`
  - `auth_failures`: `purpose text NOT NULL CHECK (purpose IN ('login', 'password_reset'))`、`login_id_key text NOT NULL`、`failure_count integer NOT NULL CHECK (failure_count >= 0)`、`last_failed_at timestamptz NOT NULL`、`locked_until timestamptz NULL`、`PRIMARY KEY (purpose, login_id_key)`、索引 `auth_failures (last_failed_at)`
  - `00001_init.sql` のコメント（「テーブル設計は最初の Spec で行う」）は変えない。`task db:migrate` と `task db:reset` が通ることを確かめる
- [X] T011 `apps/api/sqlc.yaml` の `gen.go` に `overrides` を足し、`uuid` を `github.com/google/uuid.UUID`、NULL を許さない `timestamptz` を `time.Time`、NULL を許す `timestamptz` を `*time.Time` にする（`emit_pointers_for_null_types: true` または列ごとの指定）。`github.com/google/uuid` を `apps/api/go.mod` の直接の依存にする
- [X] T012 `apps/api/db/queries/users.sql` と `apps/api/db/queries/sessions.sql` に、全ストーリーで使う基本のクエリを書く: `CreateUser`（`login_id`・`display_name`・`password_hash`・`is_admin` を受け取り行を返す）、`GetUserByLoginID`（`lower(login_id) = lower($1)`）、`CreateSession`（`token_hash`・`user_id`・`expires_at`）、`GetSessionViewer`（`sessions` と `users` を結合し、`token_hash = $1 AND expires_at > now() AND users.disabled_at IS NULL` の行の `user_id`・`login_id`・`display_name`・`is_admin`・`last_used_at` を 1 つのクエリで返す）、`TouchSession`（`last_used_at = now()`、`expires_at = $2`）、`DeleteSession`（`token_hash`）、`DeleteUserSessions`（`user_id`、消した件数を返す `:execrows`）。`task generate` を実行し、`apps/api/internal/db/` の生成コードをコミットに含める

### api: 認証の共通の処理（`apps/api/internal/auth`）

- [X] T013 [P] `apps/api/internal/auth/password.go` に Argon2id のハッシュと照合を書く（research R6）: `golang.org/x/crypto/argon2` の `IDKey`、m = 19456（KiB）、t = 2、p = 1、ソルト 16 バイト（`crypto/rand`）、出力 32 バイト、PHC 形式 `$argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>`（base64 は `RawStdEncoding`）。照合は保存された文字列のパラメーターで行い、`crypto/subtle.ConstantTimeCompare` で比べる。存在しないログイン ID 用に、起動時に 1 回作る捨てのハッシュで照合する関数 `VerifyDummy(password)` も用意する。`apps/api/internal/auth/password_test.go` にテーブル駆動のテスト（同じパスワードが照合できる、違うパスワードが照合できない、壊れた文字列はエラー、同じパスワードでもソルトが違えばハッシュが違う）を書く
- [X] T014 [P] `apps/api/internal/auth/token.go` にトークンの処理を書く（research R7）: `NewToken()` は `crypto/rand.Text()` の 26 文字、`HashToken(s)` は SHA-256 の 32 バイト、`NormalizeCode(s)` は前後の空白・途中の空白・ハイフンを取り除いて大文字にする、`FormatCode(s)` は 4 文字ごとにハイフンで区切る。`apps/api/internal/auth/token_test.go` にテーブル駆動のテスト（`NormalizeCode(FormatCode(x)) == x`、小文字・空白・ハイフン入りの入力が同じ値になる、`NewToken` が 26 文字で毎回違う）を書く
- [X] T015 [P] `apps/api/internal/auth/validate.go` に入力の規則を書く（research R11、FR-016）: ログイン ID は前後の空白を除いて `^[A-Za-z0-9_-]{3,32}$`、パスワードは前後の空白を取り除かずに 8〜128 文字（`utf8.RuneCountInString`）、表示名は前後の空白を除いて 1〜32 文字（コードポイント）で、`unicode.IsControl` の文字と U+2028・U+2029 を含まない。失敗は [contracts/graphql.md](contracts/graphql.md) の `InputField` と `InputErrorCode`（`REQUIRED`・`TOO_SHORT`・`TOO_LONG`・`INVALID_CHARACTERS`）に対応する Go の型で返し、複数の項目の失敗をまとめて返せるようにする。`apps/api/internal/auth/validate_test.go` にテーブル駆動のテスト（境界値: 2・3・32・33 文字のログイン ID、7・8・128・129 文字のパスワード、日本語・絵文字の表示名、改行入りの表示名、前後の空白だけの表示名）を書く
- [X] T016 `schema/graphql/auth.graphqls` を [contracts/graphql.md](contracts/graphql.md) のスキーマの案のとおりに作る（`extend type Query { viewer }`、`Viewer`、`type Mutation` の 7 つ、入力と payload、`InputError`・`InputField`・`InputErrorCode`）。`task generate` を実行し、gqlgen が作る `apps/api/internal/graph/auth.resolvers.go` の各 resolver を、ストーリーで実装するまで `errors.New("未実装")` を返す形にしておく。web の `apps/web/src/gql/` の生成コードもコミットに含める
- [X] T017 `apps/api/internal/auth/context.go` と `apps/api/internal/auth/cookie.go` を作る（research R13、[contracts/http-and-cookie.md](contracts/http-and-cookie.md)）: `Viewer` の型（`UserID`・`LoginID`・`DisplayName`・`IsAdmin`）、リクエストごとの状態（Viewer、送られてきたセッション ID のハッシュ、`http.ResponseWriter`）を `context` に入れる・取り出す関数、`SetSessionCookie(ctx, token)`（名前 `__Host-chara_q_session`、`Path=/`、`HttpOnly`、`Secure`、`SameSite=Lax`、`Domain` なし、`Max-Age=34560000`）と `ClearSessionCookie(ctx)`（同じ名前と属性で `Max-Age=0`）。`apps/api/internal/auth/cookie_test.go` で `httptest.ResponseRecorder` の `Set-Cookie` の属性を確かめる
- [X] T018 `apps/api/internal/auth/session.go` と `apps/api/internal/auth/middleware.go` を作る（research R10）: ミドルウェアは Cookie の `__Host-chara_q_session` を読み、`HashToken` で `GetSessionViewer` を引く。見つかれば Viewer を `context` に入れ、`last_used_at` が 1 時間より前なら `TouchSession` で `expires_at = now() + 30 日` にする。Cookie があるのに有効なセッションがなければ `ClearSessionCookie` を呼ぶ（FR-023: エラーにしない）。DB のエラーは 500 にし、ログには Cookie の値を出さない。時刻は差し替えられる `func() time.Time` で受け取る（テストで期限を確かめるため）
- [X] T019 `apps/api/internal/auth/service.go` に `Service` の型を作る。`*pgxpool.Pool`（トランザクション用）、`*db.Queries`、時刻の関数を持つ。トランザクションの補助関数 `withTx(ctx, func(*db.Queries) error)` を用意する。ストーリーの処理はこの型のメソッドとして足していく
- [X] T020 `apps/api/internal/auth/authtest/authtest.go` に DB を使うテストの補助を作る: `TEST_DATABASE_URL` が未設定なら `t.Skip`（`apps/api/internal/db/db_test.go` と同じ）、接続して `TRUNCATE users, sessions, invitations, password_reset_codes, auth_failures CASCADE` してから `Service` を返す `NewService(t)`、テスト用のユーザーを作る `CreateUser(t, svc, loginID, password, isAdmin)`。DB を使うテストのパッケージでは `t.Parallel()` を使わないことをコメントに書く（実施メモ: ローカルではテスト用の DB `chara_q_test` を分け、`task db:test:prepare` を追加した。research R15）
- [X] T021 `apps/api/internal/auth/middleware_test.go` に DB を使うテストを書く: Cookie なしなら Viewer なし、有効なセッションなら Viewer が入る、期限切れ・存在しない・無効なユーザーのセッションなら Viewer なしで Cookie を消す `Set-Cookie` が返る、`last_used_at` が 1 時間より前なら `expires_at` が延びて、1 時間以内なら更新しない
- [X] T022 `apps/api/internal/graph/resolver.go` の `Resolver` に `Auth *auth.Service` を足す。`apps/api/internal/server/server.go` で `/graphql` に認証のミドルウェアを掛ける（`/healthz` と `/playground` には掛けない）。`apps/api/cmd/server/main.go` で `auth.Service` を作って渡す。`apps/api` で `go get golang.org/x/crypto` を実行し、`go mod tidy` する
- [X] T023 `apps/api/internal/graph/errors.go` に、GraphQL のエラーを作る関数 `errUnauthenticated()`（`extensions.code = "UNAUTHENTICATED"`）と `errForbidden()`（`"FORBIDDEN"`）、`auth` の入力の失敗を GraphQL の `[InputError!]!` に変換する関数を書く。想定外のエラーは本文に入力の値を含めない汎用の文言にする（[contracts/graphql.md](contracts/graphql.md)、FR-010）

### api: 管理者のコマンドの土台

- [X] T024 `apps/api/cmd/admin/main.go` を作る（[contracts/admin-commands.md](contracts/admin-commands.md)、research R14）: 最初の引数でサブコマンドを選び、残りを標準の `flag` で読む。`DATABASE_URL` で接続し、実行の最初に `接続先: <ホスト>/<データベース名>` を標準エラー出力に出す（ユーザー名とパスワードは出さない）。結果は標準出力、メッセージは標準エラー出力、失敗したら終了コード 1。パスワードの入力の関数（端末なら `golang.org/x/term` の `ReadPassword` で 2 回、端末でなければ標準入力から 2 行）を用意する。サブコマンドの中身はストーリーで足す。`apps/api` で `go get golang.org/x/term` を実行する

### web: サーバー側の土台

- [X] T025 [P] `apps/web` に `server-only` を入れる。`apps/web/src/lib/api/server.ts`（先頭で `import "server-only"`）に、Server Components から api を呼ぶ関数を作る（research R3）: `cookies()` から `__Host-chara_q_session` だけを取り出し、なければ api を呼ばずに `null` を返す。あれば `fetch(\`${process.env.API_URL}/graphql\`, { method: "POST", headers: { "Content-Type": "application/json", Cookie: "__Host-chara_q_session=<値>" }, body, cache: "no-store" })` で送る。クエリは GraphQL Code Generator の `TypedDocumentNode` を `graphql` の `print` で文字列にして送り、結果に型を付けて返す。api に接続できない・GraphQL のエラーのときは例外を投げる
- [X] T026 `apps/web/src/lib/auth/viewer.ts`（`server-only`）に、`graphql()` で `query Viewer { viewer { id loginId displayName isAdmin } }` を定義し、`getViewer()`・`requireViewer(path)`・`requireAdmin(path)` を作る（[contracts/web-routes.md](contracts/web-routes.md)）。`getViewer` は React の `cache()` で 1 回のリクエストの中の呼び出しをまとめる。`requireViewer` はログインしていなければ `redirect("/login?next=" + encodeURIComponent(path))`、`requireAdmin` は加えて管理者でなければ `notFound()`。`task generate:web` で型を生成する（T025 の後）
- [X] T027 [P] `apps/web/src/lib/auth/rules.ts` に api と同じ入力の規則（T015）を書き、`InputField` と `InputErrorCode` の値（生成された型を使う）で失敗を返す。`apps/web/src/lib/auth/rules.test.ts` に T015 と同じ境界値のテストを書く
- [X] T028 [P] `apps/web/src/lib/auth/messages.ts` に `InputErrorCode` から日本語の文言を返す関数を書く。項目ごとの文言（例: ログイン ID の `INVALID_CHARACTERS` は「半角英数字・_・- だけで入力してください」、`TOO_SHORT` は「3 文字以上で入力してください」）、`INVITATION_UNAVAILABLE` は「この招待コードは使えません。管理者に確認してください」、`INVALID_CREDENTIALS` は「ログイン ID またはパスワードが違います」、`LOCKED` は `lockedUntil` を受け取って「ログインに続けて失敗したため、HH:mm ごろまでログインできません」、`RESET_CODE_UNAVAILABLE` は「この再設定コードは使えません。管理者に確認してください」など。`apps/web/src/lib/auth/messages.test.ts` で、すべての `InputErrorCode` に文言があることを確かめる
- [X] T029 [P] `apps/web/src/lib/auth/safe-next.ts` に戻り先の検証を書き（[contracts/web-routes.md](contracts/web-routes.md)）、`apps/web/src/lib/auth/safe-next.test.ts` にテーブル駆動のテスト（`/account` は通す、`https://example.com`・`//example.com`・`/\example.com`・`javascript:alert(1)`・空・`/login?next=/x` は `/` になる）を書く
- [X] T030 [P] `apps/web` で `pnpm dlx shadcn@4.23.0 add dropdown-menu` を実行し、`apps/web/src/components/ui/dropdown-menu.tsx` を追加する。追加された依存があれば `apps/web/package.json` で厳密なバージョンに直す
- [X] T031 `apps/web/src/test/urql.tsx` に、テストで urql のミューテーションの結果を差し替える補助（`fromValue` を返す偽の `Client` で包む `renderWithUrql`）を作り、`next/navigation` の `useRouter`（`push`・`replace`・`refresh`）を `vi.fn()` にする補助を作る

**Checkpoint**: `task lint`・`task test`・`task check:generated` が通る。`viewer` は常に null を返し、ほかのミューテーションは「未実装」のエラーを返す

---

## Phase 3: User Story 1 - 管理者を作り、ログイン・ログアウトする (Priority: P1) 🎯 MVP

**Goal**: コマンドで管理者を作り、ログイン画面でログインし、ヘッダーに表示名が出て、ログアウトできる

**Independent Test**: 空の DB で `task admin:create` を実行し、[quickstart.md](quickstart.md) の「2. 最初の管理者とログイン」をすべて確かめる

### Tests for User Story 1

- [X] T032 [P] [US1] `apps/api/internal/auth/login_test.go` に DB を使うテーブル駆動のテストを書く: 正しいログイン ID とパスワードでセッションができる、ログイン ID の大文字・小文字の違いでもログインできる、誤ったパスワード・存在しないログイン ID・形式に合わないログイン ID はどれも同じ `INVALID_CREDENTIALS`、前のセッションのハッシュを渡すとそのセッションが消える（FR-019）、ログアウトでセッションが消える
- [X] T033 [P] [US1] `apps/api/internal/auth/admin_test.go` に DB を使うテストを書く: `CreateAdmin` で `is_admin = true` のユーザーができる、既に使われているログイン ID（大文字・小文字だけ違うものを含む）なら失敗して既存のユーザーは変わらない、規則に合わない入力は失敗する、`ListUsers` が作成日時の順に返す
- [X] T034 [P] [US1] `apps/api/internal/graph/auth_resolvers_test.go` に、`server.New` を `httptest` で動かす DB を使うテストを書く: `login` の成功で `Set-Cookie` に `__Host-chara_q_session` と `HttpOnly`・`Secure`・`SameSite=Lax`・`Path=/` があり `Domain` がない、その Cookie を付けた `viewer` が本人の `loginId`・`displayName`・`isAdmin` を返す、`logout` の後は同じ Cookie で `viewer` が null、Cookie なしの `viewer` は null でエラーにならない、`GET` での `login` のミューテーションは処理されない（T009 にもミューテーションの例を足す）
- [X] T035 [P] [US1] `apps/web/src/components/auth/account-menu.test.tsx` に、ログインしていない（「ログイン」のリンクが `/login?next=<今のパス>`）、ログインしている（表示名のボタン、読み上げの名前「アカウントのメニュー（<表示名>）」、メニューに「アカウントの設定」と「ログアウト」、管理者でなければ「招待コードの発行」がない）、管理者（「招待コードの発行」がある）、長い表示名が省略のクラスを持つ、キーボードでメニューを開いて選べる、のテストを書く。`expectNoAxeViolations` も確かめる
- [X] T036 [P] [US1] `apps/web/src/components/auth/login-form.test.tsx` に、空の送信で項目のエラー、`INVALID_CREDENTIALS` でフォームの上に文言、成功で `router.refresh()` の後に戻り先へ移動、`?next=` が外部の URL なら `/` へ移動、入力欄の `autocomplete`（`username`・`current-password`）と `type="password"` のテストを書く

### Implementation for User Story 1

- [X] T037 [US1] `apps/api/internal/auth/admin.go` に `CreateAdmin(ctx, loginID, displayName, password)`（T015 の規則で検証し、`is_admin = true` で作る。一意の制約の違反は「ログイン ID が使われている」のエラーにする）と `ListUsers(ctx)` を書く。`apps/api/db/queries/users.sql` に `ListUsers`（`login_id`・`display_name`・`is_admin`・`disabled_at`・`created_at` を `created_at` の順）を足して `task generate`
- [X] T038 [US1] `apps/api/internal/auth/login.go` に `Login(ctx, loginID, password, previousTokenHash)` と `Logout(ctx, tokenHash)` を書く: ログイン ID の形式に合わなければ `VerifyDummy` を呼んでから `INVALID_CREDENTIALS`、ユーザーが見つからない・無効なら `VerifyDummy` を呼んでから `INVALID_CREDENTIALS`、照合に失敗したら `INVALID_CREDENTIALS`。成功したら前のセッション（あれば）を消し、`NewToken` で新しいセッションを `expires_at = now() + 30 日` で作り、トークンと Viewer を返す。ロックは US3 で足す
- [X] T039 [US1] `apps/api/internal/graph/auth.resolvers.go` に `Viewer`・`Login`・`Logout` を実装する。`Login` は成功したら `auth.SetSessionCookie` を呼ぶ。`Logout` はセッションがあれば消し、常に `auth.ClearSessionCookie` を呼んで `true` を返す（T032〜T034 が通ること）
- [X] T040 [US1] `apps/api/cmd/admin/main.go` に `create-admin`（`-login-id`・`-display-name`、パスワードは T024 の関数で 2 回入力）と `users`（表の形で出す）のサブコマンドを足す。`Taskfile.yml` に `admin:create`（`dir: "{{.API_DIR}}"`、`dotenv: [.env]`、`interactive: true`、`go run ./cmd/admin create-admin {{.CLI_ARGS}}`）と `admin:users` を足す。シェルで設定した `DATABASE_URL` が `dotenv` で上書きされないことを確かめ、上書きされる場合は利用者に報告して Taskfile の書き方を直す（research R14）
- [X] T041 [US1] `apps/web/src/components/auth/account-menu.tsx`（Client Component）を作る: props は `viewer: { displayName: string; isAdmin: boolean } | null`。ログインしていなければ「ログイン」のリンク（`usePathname` で今のパスを `next` に付ける）。ログインしていれば `dropdown-menu` のトリガーに表示名（最大の幅を決めて `truncate`）、項目に「アカウントの設定」（`/account`）、管理者だけ「招待コードの発行」（`/admin/invitations`）、「ログアウト」（urql で `mutation Logout { logout }` を送り、`router.refresh()`）。T035 が通ること
- [X] T042 [US1] `apps/web/src/components/auth/account-area.tsx`（Server Component）で `getViewer()` を呼び、`displayName` と `isAdmin` だけを `AccountMenu` に渡す（`loginId` は渡さない）。`apps/web/src/app/layout.tsx` のヘッダーの `actions` を `<><ThemeToggle /><Suspense fallback={<AccountAreaFallback />}><AccountArea /></Suspense></>` にする。`AccountAreaFallback` はボタンと同じ大きさの空の領域で、ログインの操作を出さない（SC-009）。`AccountArea` の中で api のエラーが起きたときも、ログインの操作を出さずに空の領域にする（error boundary か try/catch）。`UrqlProvider` がヘッダーのメニューも包むように、`apps/web/src/app/layout.tsx` の `UrqlProvider` の位置を `body` の中身全体に移す
- [X] T043 [US1] `apps/web/src/components/auth/login-form.tsx`（Client Component）と `apps/web/src/app/login/page.tsx` を作る: 共通入力部品（`apps/web/src/components/form/text-field.tsx`）でログイン ID とパスワード、送信は urql の `login` のミューテーション。成功したら `router.refresh()` してから `safeNext(searchParams.get("next"))` へ `router.replace`。失敗は T028 の文言で表示する。登録（`/register`）と再設定（`/reset-password`）への入口を置く。ページのタイトルは「ログイン」。ログインしていればページ（Server Component）が `getViewer()` を見て戻り先へ `redirect` する（`<Suspense>` の中で読む）。`useSearchParams` を使う部分は `<Suspense>` で包む。T036 が通ること
- [X] T044 [US1] [quickstart.md](quickstart.md) の「1. 中継と Cookie の確認」の 4（再読み込みで Server Components のヘッダーに表示名が出る）と「2. 最初の管理者とログイン」を Chrome と Firefox で確かめる（実施メモ: Chromium で確認。Firefox は未確認）

**Checkpoint**: 管理者を作ってログイン・ログアウトでき、ヘッダーの表示が 3 つの状態で正しい。ここまでで MVP として動く

---

## Phase 4: User Story 2 - 招待コードで登録する (Priority: P1)

**Goal**: 管理者が招待コードの発行画面でコードを発行し、招待された人がそのコードで登録してログインした状態になる

**Independent Test**: [quickstart.md](quickstart.md) の「3. 招待コードと登録」をすべて確かめる

### Tests for User Story 2

- [X] T045 [P] [US2] `apps/api/internal/auth/invitation_test.go` に DB を使うテーブル駆動のテストを書く: `IssueInvitation` が 26 文字のコード（`FormatCode` の形）と `created_at + 7 日` の期限を返し、DB には `code_hash` だけがある（コードの平文が表のどの列にもない）、`Register` が未使用で期限内のコードで成功してセッションを返す、使用済み・期限切れ（時刻の関数を進める）・存在しないコードは同じ `INVITATION_UNAVAILABLE`、コードが使えないときはログイン ID が使われていても `INVITATION_UNAVAILABLE` だけを返す、ログイン ID が使われている（大文字・小文字だけ違うものを含む）・入力の規則に合わないときはコードが未使用のまま残る（FR-014）、小文字・空白・ハイフン入りで入力したコードでも登録できる
- [X] T046 [P] [US2] `apps/api/internal/auth/invitation_concurrency_test.go` に、別々の接続から同じ招待コードで 10 人が同時に登録し、成功が 1 回だけで、残りが `INVITATION_UNAVAILABLE` になるテストを書く（SC-005）
- [X] T047 [P] [US2] `apps/api/internal/graph/auth_resolvers_test.go` に、`register` の成功で `Set-Cookie` が付き `viewer` が返る、管理者でないユーザーの `issueInvitation` が `FORBIDDEN`、ログインしていない `issueInvitation` が `UNAUTHENTICATED`、無効にした管理者のセッションでは発行できない、のテストを足す
- [X] T048 [P] [US2] `apps/web/src/components/auth/register-form.test.tsx` に、4 項目の規則のエラーが項目の近くに出る、`INVITATION_UNAVAILABLE` が招待コードの項目に出る、`LOGIN_ID_TAKEN` がログイン ID の項目に出る、成功で `router.refresh()` の後に `/` へ移動、メールアドレスの項目がない、パスワードの `autocomplete="new-password"`、のテストを書く
- [X] T049 [P] [US2] `apps/web/src/components/auth/invitation-issuer.test.tsx` に、ボタンを押すとコードと期限が表示される、コピーのボタンで `navigator.clipboard.writeText` がコードで呼ばれる、`FORBIDDEN` のときに文言が出る、のテストを書く

### Implementation for User Story 2

- [X] T050 [US2] `apps/api/db/queries/invitations.sql` に `CreateInvitation`（`code_hash`・`issued_by`・`expires_at`）、`LockUsableInvitation`（`code_hash = $1 AND used_at IS NULL AND expires_at > now()` の行を `FOR UPDATE`）、`MarkInvitationUsed`（`used_at = now()`、`used_by`）を書いて `task generate`
- [X] T051 [US2] `apps/api/internal/auth/invitation.go` に `IssueInvitation(ctx, issuer Viewer)`（`expires_at = now() + 7 日`、平文のコードは戻り値だけで返す）と `Register(ctx, input)` を書く（research R8）: 入力の規則で検証 → パスワードをハッシュ（トランザクションの外）→ トランザクションで `LockUsableInvitation`（なければ `INVITATION_UNAVAILABLE`）→ `CreateUser`（`is_admin = false`。一意の制約の違反なら `LOGIN_ID_TAKEN` でロールバック）→ `MarkInvitationUsed` → `CreateSession`。T045・T046 が通ること
- [X] T052 [US2] `apps/api/internal/graph/auth.resolvers.go` に `Register`（成功したら `SetSessionCookie`）と `IssueInvitation`（Viewer がなければ `errUnauthenticated()`、管理者でなければ `errForbidden()`）を実装する。T047 が通ること
- [X] T053 [US2] `apps/web/src/components/auth/register-form.tsx` と `apps/web/src/app/register/page.tsx` を作る: 招待コード・ログイン ID・パスワード・表示名の 4 項目（共通入力部品）、送信前に T027 の規則で確かめ、urql の `register` を送る。成功したら `router.refresh()` して `/` へ。ログインへの入口を置く。ログインしていればページが `/` へ `redirect` する。タイトルは「ユーザー登録」。T048 が通ること
- [X] T054 [US2] `apps/web/src/components/auth/invitation-issuer.tsx` と `apps/web/src/app/admin/invitations/page.tsx` を作る: ページは `requireAdmin("/admin/invitations")`（`<Suspense>` の中）。「招待コードを発行する」ボタンで urql の `issueInvitation` を送り、コード（等幅で表示）、有効期限（`yyyy/MM/dd HH:mm`）、コピーのボタン、「この画面を離れるとコードは再表示できません」の注意を出す。タイトルは「招待コードの発行」、`robots: { index: false }`。T049 が通ること
- [X] T055 [US2] [quickstart.md](quickstart.md) の「3. 招待コードと登録」を確かめる

**Checkpoint**: 招待した人が登録でき、US1 と合わせて、管理者と一般のユーザーの両方が使える

---

## Phase 5: User Story 3 - 続けて失敗したログインをロックする (Priority: P2)

**Goal**: 同じログイン ID で 5 回続けて失敗したら 15 分ロックする。存在しないログイン ID も同じように扱う

**Independent Test**: [quickstart.md](quickstart.md) の「4. ロック」をすべて確かめる

### Tests for User Story 3

- [X] T056 [P] [US3] `apps/api/internal/auth/lock_test.go` に DB を使うテーブル駆動のテストを書く（時刻の関数を進めて確かめる）: 4 回失敗の後の成功で回数が 0 に戻る、5 回目の失敗で `LOCKED` と `locked_until = 5 回目の時刻 + 15 分`、ロック中は正しいパスワードでも `LOCKED` で回数もロックの時間も変わらない、15 分後に正しいパスワードでログインでき行が消える、15 分後の最初の失敗は 1 回目として数える、存在しないログイン ID も 5 回で `LOCKED`、形式に合わないログイン ID は記録されない、別のユーザーのログインに影響しない、ログインの成功で期限切れのセッションと 1 日以上前でロック中でない失敗の記録が消える
- [X] T057 [P] [US3] `apps/api/internal/auth/lock_concurrency_test.go` に、別々の接続から同じログイン ID で同時に 5 回失敗したとき、回数が 5 になりロックされるテストを書く
- [X] T058 [P] [US3] `apps/web/src/components/auth/login-form.test.tsx` に、`LOCKED` と `lockedUntil` を受け取ると「HH:mm ごろまでログインできません」が出るテストを足す

### Implementation for User Story 3

- [X] T059 [US3] `apps/api/db/queries/auth_failures.sql` に `GetLockedUntil`（`purpose`・`login_id_key`、`locked_until > $3` のときだけ返す）、`RecordFailure`（[data-model.md](data-model.md) の判定の規則のとおり、`INSERT … ON CONFLICT (purpose, login_id_key) DO UPDATE` 1 文で、ロックが過ぎていれば回数を 1、そうでなければ 1 足し、5 になったら `locked_until = $now + interval '15 minutes'`。結果の `failure_count` と `locked_until` を返す）、`ClearFailures`（`purpose`・`login_id_key`）、`DeleteStaleFailures`（`last_failed_at < $1 - interval '1 day' AND (locked_until IS NULL OR locked_until <= $1)`）を書く。`apps/api/db/queries/sessions.sql` に `DeleteExpiredSessions`（`expires_at <= $1`）を足す。`task generate`
- [X] T060 [US3] `apps/api/internal/auth/lock.go` に、目的（`login` / `password_reset`）を受け取るロックの処理（`checkLock`、`recordFailure`、`clearFailures`）を書き、`apps/api/internal/auth/login.go` の `Login` に組み込む: 形式に合うログイン ID はロック中なら照合せずに `LOCKED` と `locked_until`、失敗なら `RecordFailure`（5 回目なら `LOCKED` を返す）、成功なら `ClearFailures` と、`DeleteExpiredSessions`・`DeleteStaleFailures` の掃除。T056・T057 が通ること
- [X] T061 [US3] `apps/api/internal/graph/auth.resolvers.go` の `Login` で `LOCKED` のときに `lockedUntil` を返す。`apps/web/src/components/auth/login-form.tsx` で `LOCKED` の文言を `lockedUntil` 付きで出す。T058 が通ること
- [X] T062 [US3] [quickstart.md](quickstart.md) の「4. ロック」を確かめる（15 分の待ちは、DB の `auth_failures.locked_until` を手で過去にして代えてよい）

**Checkpoint**: 総当たりのログインが 5 回で止まる

---

## Phase 6: User Story 4 - 表示名とパスワードを変える (Priority: P2)

**Goal**: アカウントの設定画面で、自分のログイン ID を確かめ、表示名とパスワードを変えられる。パスワードを変えると、ほかの端末のログイン状態が切れる

**Independent Test**: [quickstart.md](quickstart.md) の「5. アカウントの設定」をすべて確かめる

### Tests for User Story 4

- [X] T063 [P] [US4] `apps/api/internal/auth/account_test.go` に DB を使うテーブル駆動のテストを書く: `UpdateDisplayName` が規則に合う表示名で変わり `updated_at` が進む、規則に合わない表示名は失敗する、`ChangePassword` が現在のパスワードの一致で成功し、そのユーザーのセッションがすべて消えて新しいセッションが 1 つだけになる、現在のパスワードが違えば `WRONG_CURRENT_PASSWORD` でパスワードもセッションも変わらない、新しいパスワードが規則に合わなければ失敗する
- [X] T064 [P] [US4] `apps/api/internal/graph/auth_resolvers_test.go` に、ログインしていない `updateDisplayName`・`changePassword` が `UNAUTHENTICATED`、`changePassword` の成功で新しい `Set-Cookie` が付き、古い Cookie での `viewer` が null、別のユーザーとして `viewer` 以外の応答に他人の `loginId` が含まれない（SC-008）、のテストを足す
- [X] T065 [P] [US4] `apps/web/src/components/auth/account-settings.test.tsx` に、ログイン ID が表示だけで変えられない、表示名の保存の成功で `router.refresh()` と成功の表示、`WRONG_CURRENT_PASSWORD` が現在のパスワードの項目に出る、パスワードの変更の成功で入力欄が空になり成功の表示、`UNAUTHENTICATED` のエラーで `/login?next=/account` へ移動、のテストを書く

### Implementation for User Story 4

- [X] T066 [US4] `apps/api/db/queries/users.sql` に `UpdateDisplayName`（`display_name`、`updated_at = now()`）と `UpdatePasswordHash`（`password_hash`、`updated_at = now()`）を足して `task generate`。`apps/api/internal/auth/account.go` に `UpdateDisplayName(ctx, viewer, displayName)` と `ChangePassword(ctx, viewer, current, next)` を書く: `ChangePassword` は現在のパスワードを照合し、トランザクションで `UpdatePasswordHash` → `DeleteUserSessions` → `CreateSession` して新しいトークンを返す（FR-033）。T063 が通ること
- [X] T067 [US4] `apps/api/internal/graph/auth.resolvers.go` に `UpdateDisplayName` と `ChangePassword` を実装する（Viewer がなければ `errUnauthenticated()`、`ChangePassword` の成功で `SetSessionCookie`）。T064 が通ること
- [X] T068 [US4] `apps/web/src/components/auth/account-settings.tsx` と `apps/web/src/app/account/page.tsx` を作る: ページは `requireViewer("/account")`（`<Suspense>` の中）で、ログイン ID（表示だけ、「他の人には表示されません」の補足）、表示名の変更のフォーム、パスワードの変更のフォーム（現在のパスワード `autocomplete="current-password"`、新しいパスワード `autocomplete="new-password"`）を置く。送信は urql の `updateDisplayName`・`changePassword`。成功したら `router.refresh()` して成功を表示する（`role="status"`）。`UNAUTHENTICATED` を受け取ったら `/login?next=/account` へ移動する。タイトルは「アカウントの設定」。T065 が通ること
- [X] T069 [US4] [quickstart.md](quickstart.md) の「5. アカウントの設定」を 2 つのブラウザで確かめる

**Checkpoint**: 表示名とパスワードを自分で変えられる

---

## Phase 7: User Story 5 - 再設定コードでパスワードを再設定する (Priority: P2)

**Goal**: 管理者がコマンドで再設定コードを発行し、ユーザーがそのコードで新しいパスワードを設定する

**Independent Test**: [quickstart.md](quickstart.md) の「6. パスワードの再設定」をすべて確かめる

### Tests for User Story 5

- [X] T070 [P] [US5] `apps/api/internal/auth/reset_test.go` に DB を使うテーブル駆動のテストを書く: `IssueResetCode` が 26 文字のコードと `created_at + 24 時間` の期限を返し、DB には `code_hash` だけがある、同じユーザーに 2 回発行すると 1 回目のコードは使えない、無効なユーザーには発行できない、`ResetPassword` の成功でパスワードが変わり、そのユーザーのセッションがすべて消え、`login` と `password_reset` の失敗の記録が消える（ロック中でもすぐにログインできる）、使用済み・期限切れ・存在しない・別のユーザー用のコードと、無効なユーザーは同じ `RESET_CODE_UNAVAILABLE`、再設定の失敗が 5 回続くと `password_reset` で `LOCKED` になり、ログインのロックとは独立している、新しいパスワードが規則に合わなければ失敗してコードは未使用のまま
- [X] T071 [P] [US5] `apps/api/internal/auth/reset_concurrency_test.go` に、同じ再設定コードを別々の接続から 10 回同時に使い、成功が 1 回だけのテストを書く（SC-005）
- [X] T072 [P] [US5] `apps/web/src/components/auth/reset-password-form.test.tsx` に、3 項目の規則のエラー、`RESET_CODE_UNAVAILABLE` が再設定コードの項目に出る、`LOCKED` の文言、成功で `/login?reset=1` へ移動、のテストを書く。`apps/web/src/components/auth/login-form.test.tsx` に、`?reset=1` のときに「パスワードを再設定しました。新しいパスワードでログインしてください」が出るテストを足す

### Implementation for User Story 5

- [X] T073 [US5] `apps/api/db/queries/password_reset_codes.sql` に `DeleteUnusedResetCodes`（`user_id`、`used_at IS NULL`）、`CreateResetCode`（`user_id`・`code_hash`・`expires_at`）、`LockUsableResetCode`（`user_id = $1 AND code_hash = $2 AND used_at IS NULL AND expires_at > now()` の行を `FOR UPDATE`）、`MarkResetCodeUsed` を書いて `task generate`
- [X] T074 [US5] `apps/api/internal/auth/reset.go` に `IssueResetCode(ctx, loginID)`（無効なユーザーならエラー。トランザクションで `DeleteUnusedResetCodes` → `CreateResetCode`、`expires_at = now() + 24 時間`）と `ResetPassword(ctx, loginID, code, newPassword)` を書く: 新しいパスワードの規則 → `password_reset` のロックの確認 → ユーザーを引く（見つからない・無効なら失敗を記録して `RESET_CODE_UNAVAILABLE`）→ パスワードをハッシュ → トランザクションで `LockUsableResetCode`（なければ失敗を記録して `RESET_CODE_UNAVAILABLE`）→ `UpdatePasswordHash` → `MarkResetCodeUsed` → `DeleteUserSessions` → 両方の目的の `ClearFailures`。T070・T071 が通ること
- [X] T075 [US5] `apps/api/internal/graph/auth.resolvers.go` に `ResetPassword` を実装する（ログインはさせず、Cookie も設定しない。`LOCKED` なら `lockedUntil` を返す）。`apps/api/cmd/admin/main.go` に `reset-code <ログイン ID>` のサブコマンド（コードを `FormatCode` の形で、有効期限と一緒に標準出力へ）を足し、`Taskfile.yml` に `admin:reset-code` を足す
- [X] T076 [US5] `apps/web/src/components/auth/reset-password-form.tsx` と `apps/web/src/app/reset-password/page.tsx` を作る: ログイン ID・再設定コード・新しいパスワード（`autocomplete="new-password"`）の 3 項目、「再設定コードは管理者に発行してもらってください」の補足、urql の `resetPassword`。成功したら `/login?reset=1` へ移動する。ログインへの入口を置く。タイトルは「パスワードの再設定」。`apps/web/src/components/auth/login-form.tsx` で `?reset=1` の表示を足す。T072 が通ること
- [X] T077 [US5] [quickstart.md](quickstart.md) の「6. パスワードの再設定」を確かめる

**Checkpoint**: パスワードを忘れたユーザーが、管理者の助けで戻ってこられる

---

## Phase 8: User Story 6 - 管理者がユーザーを無効にする (Priority: P2)

**Goal**: 管理者がコマンドでユーザーを無効・有効にでき、無効にしたユーザーのログイン状態がすぐに切れる

**Independent Test**: [quickstart.md](quickstart.md) の「7. ユーザーの無効化」をすべて確かめる

### Tests for User Story 6

- [X] T078 [P] [US6] `apps/api/internal/auth/disable_test.go` に DB を使うテーブル駆動のテストを書く: `DisableUser` で `disabled_at` が入り、そのユーザーのセッションがすべて消えて件数が返る、無効なユーザーは正しいパスワードでも `INVALID_CREDENTIALS`、既に無効なら何もせず成功する、`EnableUser` で元のパスワードでログインできる、既に有効なら何もせず成功する、存在しないログイン ID はエラー
- [X] T079 [P] [US6] `apps/api/cmd/admin/main_test.go` に、DB を使うテストで、標準入力からパスワードを渡した `create-admin`、`users`、`reset-code`、`disable`、`enable` の終了コードと出力（標準出力に結果、標準エラー出力に接続先とメッセージ、パスワードとコードのハッシュが出力にない）を確かめるテストを書く（`run(args, stdin, stdout, stderr) int` の形に分けて呼ぶ）

### Implementation for User Story 6

- [X] T080 [US6] `apps/api/db/queries/users.sql` に `DisableUser`（`disabled_at = now()`、`updated_at = now()`、`disabled_at IS NULL` のときだけ）と `EnableUser`（`disabled_at = NULL`）を足して `task generate`。`apps/api/internal/auth/disable.go` に `DisableUser(ctx, loginID)`（トランザクションで `DisableUser` → `DeleteUserSessions`、切ったセッションの件数を返す）と `EnableUser(ctx, loginID)` を書く。T078 が通ること
- [X] T081 [US6] `apps/api/cmd/admin/main.go` に `disable <ログイン ID>` と `enable <ログイン ID>` のサブコマンドを足し（出力は [contracts/admin-commands.md](contracts/admin-commands.md) のとおり）、`Taskfile.yml` に `admin:disable` と `admin:enable` を足す。`main` を `run(args, stdin, stdout, stderr) int` に分ける。T079 が通ること
- [X] T082 [US6] [quickstart.md](quickstart.md) の「7. ユーザーの無効化」を確かめる

**Checkpoint**: すべてのユーザーストーリーが動く

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: 手順書と決定の記録、全体の確認

- [X] T083 [P] `infra/vercel.md` の環境変数の表を `API_URL`（Production: `https://<Cloud Run の URL>`、Preview: 「本番環境の構築の機能で決める」）に書き換え、`NEXT_PUBLIC_API_URL` の記述と「認証の Spec で置き換える」の注意を消す。§4 の「api の CORS は不要になる見込み」を「api の CORS は外した（認証の Spec）」に直す
- [X] T084 [P] `docs/adr/0006-authentication.md` の「後続の Spec で決めること」の各項目に、[spec.md](spec.md) の Clarifications の決定（入力の規則、ロックと有効期限、管理者の操作の分け方、招待コードとシートの紐づけを先送りしたこと、Server Components からの Cookie の渡し方、CORS を外したこと）と、`specs/002-auth` へのリンクを追記する。「結果」の CORS と `NEXT_PUBLIC_API_URL` の項目も、決定に合わせて直す
- [ ] T085 [P] `.specify/memory/constitution.md` の「認証」の節で、平文で保存しないものに再設定コードを足す（文言の修正のため PATCH: Version を 1.3.1、Last Amended を作業日にする）。足す前に利用者に確認する
- [X] T086 [P] `README.md` に、最初の管理者の作り方（`task admin:create`）と `admin:*` のタスクの一覧、`API_URL` の説明を足す。`CLAUDE.md` の「よく使うコマンド」に `task admin:create` を足す
- [X] T087 [P] `Taskfile.yml` に `dev:web:https`（`pnpm --filter {{.WEB_PKG}} dev --experimental-https`、desc に「Safari で Cookie を確かめるとき。自己署名の証明書を端末に登録する」）を足す
- [X] T088 `task lint`、`task test`、`task check:generated`、`task build` を実行し、すべて通ることを確かめる
- [X] T089 [quickstart.md](quickstart.md) の「8. 表示と安全性」を確かめる（幅 360px のヘッダー、api のログに秘密が出ていない、DB に平文がない、`OPTIONS` に `Access-Control-Allow-Origin` がない）。確かめられなかった項目（Safari など）は、理由と一緒に利用者に報告する（実施メモ: 幅 360px、ログと DB に平文がないこと、CORS のヘッダーがないこと（server_test.go）を確かめた。Safari・Firefox は未確認）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup（Phase 1）**: すぐに始められる。T005 で Cookie の往復が確かめられなければ、以降に進まない
- **Foundational（Phase 2）**: Setup の後。すべてのストーリーを止める
- **User Stories（Phase 3〜8）**: Foundational の後
- **Polish（Phase 9）**: 必要なストーリーが終わった後

### User Story Dependencies

- **US1（P1）**: Foundational の後。ほかのストーリーに依存しない
- **US2（P1）**: Foundational の後。発行画面の確認には、管理者としてログインできること（US1 の `login` とヘッダー）が要る
- **US3（P2）**: US1 の `Login` に組み込む
- **US4（P2）**: US1 の `login` とヘッダーに依存する
- **US5（P2）**: US3 のロックの処理（`password_reset` の目的）と、US4 の `UpdatePasswordHash` を使う。US4 より先に行う場合は、`UpdatePasswordHash` のクエリを US5 で作る
- **US6（P2）**: US1 の管理者のコマンドの土台に依存する。US2〜US5 とは独立

### Within Each User Story

- テストを先に書き、失敗することを確かめてから実装する
- SQL（`task generate`）→ `internal/auth` のサービス → resolver・コマンド → web の画面 → quickstart での確認の順

### Parallel Opportunities

- Phase 1: T002・T003 は並行できる
- Phase 2: T013・T014・T015 と、web の T025・T027・T028・T029・T030 は並行できる。T010〜T012 は順番に行う
- 各ストーリーの「Tests」の [P] のタスクは並行できる
- US6 は US2〜US5 と並行できる（`apps/api/cmd/admin/main.go` と `Taskfile.yml` を触るタスクは、同時に進めない）

---

## Parallel Example: User Story 1

```text
# User Story 1 のテストをまとめて書く:
Task: "T032 apps/api/internal/auth/login_test.go にログインとログアウトのテスト"
Task: "T033 apps/api/internal/auth/admin_test.go に管理者の作成と一覧のテスト"
Task: "T034 apps/api/internal/graph/auth_resolvers_test.go に Cookie と viewer のテスト"
Task: "T035 apps/web/src/components/auth/account-menu.test.tsx にヘッダーの 3 つの状態のテスト"
Task: "T036 apps/web/src/components/auth/login-form.test.tsx にログイン画面のテスト"
```

## Parallel Example: Foundational

```text
Task: "T013 apps/api/internal/auth/password.go（Argon2id）"
Task: "T014 apps/api/internal/auth/token.go（トークンとコード）"
Task: "T015 apps/api/internal/auth/validate.go（入力の規則）"
Task: "T027 apps/web/src/lib/auth/rules.ts（web の入力の規則）"
Task: "T029 apps/web/src/lib/auth/safe-next.ts（戻り先の検証）"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: Setup（Cookie の往復の確認まで）
2. Phase 2: Foundational
3. Phase 3: User Story 1
4. **止めて確かめる**: [quickstart.md](quickstart.md) の 1 と 2
5. 管理者だけがログインできる状態で、ほかの機能（シートの共通基盤）の作業を始められる

### Incremental Delivery

1. Setup + Foundational → 土台ができる
2. US1 → 管理者がログインできる（MVP）
3. US2 → 招待した人が登録できる。ここで利用者に使ってもらえる
4. US3 → ロック。US2 と同じ時期に公開するのが望ましい
5. US4・US5・US6 → アカウントの管理と運用の手段
6. Polish → 手順書と決定の記録

---

## Notes

- [P] のタスクは別のファイルで、未完了のタスクに依存しない
- `apps/api/internal/graph/auth.resolvers.go`、`apps/api/internal/graph/auth_resolvers_test.go`、`apps/api/cmd/admin/main.go`、`Taskfile.yml`、`apps/api/db/queries/users.sql`、`apps/web/src/components/auth/login-form.tsx` は複数のストーリーで触るため、そのタスクは並行にしない
- 生成コードは、元になったスキーマ・SQL の変更と同じコミットに入れる（どのコミットでも `task check:generated` が通るように）
- 各チェックポイントで止めて、ストーリーを単独で確かめられる
- `git push` とクラウドのリソースの操作は、利用者の指示があるときだけ行う
