# Research: 認証（招待コードでの登録、ログイン・ログアウト）

Phase 0 の調査結果。Technical Context の不明点と、技術の選び方を決める。
Next.js は `apps/web/node_modules/next/dist/docs/` の 16.4.3 のドキュメントで確かめた
（`apps/web/AGENTS.md` の指示）。gqlgen はモジュールキャッシュの v0.17.98 のソースで確かめた。

## R1. ブラウザから api への経路（rewrites）

- **Decision**: `next.config.ts` の `rewrites()` で、`/graphql` を `${API_URL}/graphql` に中継する
  （外部 URL への rewrite）。`API_URL` はサーバー側の環境変数で、`NEXT_PUBLIC_` を付けない。
  urql の接続先は相対パスの `/graphql` にする。
- **Rationale**: ADR 0006 の決定のとおり。rewrite はブラウザから見て同じオリジンのため、api の
  `Set-Cookie` はファーストパーティの Cookie になり、Safari と Firefox の既定の設定でも
  ブロックされない。
- **注意**:
  - `rewrites()` はビルドのときに評価される。Vercel と CI のビルドで `API_URL` を設定する。
  - 中継で `Set-Cookie` が落ちない・`Cookie` が api に届くことは、実装の最初に確かめる（FR-004）。
    確かめ方は [quickstart.md](quickstart.md) の「1. 中継と Cookie の確認」。
- **Alternatives considered**:
  - Route Handler（`app/graphql/route.ts`）で `fetch` して中継する: ヘッダーの受け渡しを自分で書く
    必要があり、`Set-Cookie` の複数行の扱いなどを間違えやすい。rewrites で届かなかった場合の
    代わりの案として残す。
  - Server Actions で api を呼び、web が `cookies().set()` で Cookie を付け直す: Cookie の属性を
    web と api の 2 か所で持つことになり、ADR 0006 の「api が発行する」形から外れる。

## R2. セッションの Cookie

- **Decision**: 名前は `__Host-chara_q_session`。属性は `Path=/; HttpOnly; Secure; SameSite=Lax`、
  `Domain` は付けない。値は `crypto/rand.Text()`（128 ビット、base32 の 26 文字）。
  `Max-Age` は 400 日（Chrome の上限）にし、有効かどうかはサーバー側の期限だけで決める。
- **Rationale**:
  - `__Host-` 接頭辞にすると、ブラウザが `Secure`・`Path=/`・`Domain` なしを強制する。
    ADR 0006 の属性を、設定の書き間違いがあっても守れる。
  - セッションの期限は「最後に使ってから 30 日」で延びる（FR-022）。ところが、Server Components
    から api を呼んだ応答の `Set-Cookie` はブラウザに届かない。Cookie の `Max-Age` を 30 日にすると、
    Server Components の画面だけを見ている人は、サーバーのセッションが生きていても Cookie が先に
    消えてしまう。Cookie の値はサーバーの期限が切れたら役に立たないため、Cookie を長く持たせても
    危険は増えない。
- **ローカル開発**: Chrome と Firefox は `http://localhost` を安全な接続とみなし、`Secure` の Cookie を
  受け付ける。Safari は受け付けない場合がある。Safari で確かめるときは、`next dev --experimental-https`
  （`task dev:web:https`）で `https://localhost:3000` を使う。この方法は自己署名の証明書を作って
  端末に登録するため、使うかどうかは開発者が決める。
- **Alternatives considered**:
  - `Max-Age` 30 日で、api の応答のたびに付け直す: 上の理由で、Server Components の画面だけでは
    延びない。
  - Cookie の属性をローカルだけ `Secure` なしにする: 本番と違う設定になり、本番だけで起きる
    不具合を見落とす（spec の Edge Cases）。

## R3. Server Components から api を呼ぶ（Clarifications Q5）

- **Decision**: `server-only` のモジュール `lib/api/server.ts` で、`cookies()` からセッションの Cookie
  だけを取り出し、`Cookie: __Host-chara_q_session=<値>` を付けて `${API_URL}/graphql` に POST する。
  Cookie がなければ api を呼ばずに「ログインしていない」とする。`fetch` はキャッシュしない。
- **cacheComponents との関係**: `next.config.ts` で `cacheComponents: true` のため、`cookies()` を読む
  コンポーネントは `<Suspense>` の中に置く必要がある（外にあるとビルドエラー）。ヘッダーの
  ログイン状態は `<Suspense>` の中で表示する。
  - SC-009（ログインしていない表示が一瞬出ない）を守るため、`fallback` は「ログイン」のボタンでは
    なく、同じ幅の空の領域にする。ログイン状態は、静的な外枠と同じ応答の中でストリーミングされる。
  - レイアウトの一番上で `await` しない（ドキュメントの「Push dynamic access down」）。
- **Alternatives considered**:
  - web 自身の `/graphql`（rewrite）を Server Components から呼ぶ: 通信が 1 段増え、自分の URL を
    サーバー側で知る必要がある。
  - Cookie をすべて転送する: テーマなど関係のない Cookie まで api に渡る。Clarifications で除外した。

## R4. CSRF の対策（FR-005）

- **Decision**: gqlgen の transport を `Options`・`GET`・`POST` だけにする（今のまま）。
  - gqlgen v0.17.98 の `transport.POST.Supports` は、`POST` かつ `Content-Type` のメディアタイプが
    `application/json` のときだけ受け付ける。
  - `transport.GET` は、クエリ以外の操作に「GET requests only allow query operations」を返す。
  - `MultipartForm`・`UrlEncodedForm` などの transport は追加しない。
  この条件を、api の HTTP のテストで固定する（transport を足したときに気付けるようにする）。
- **Rationale**: `text/plain` やフォームの形式は、他サイトから CORS の事前確認なしで送れる。
  `application/json` だけにすれば、他サイトからは事前確認が必要になり、CORS の許可を外すため
  通らない。さらに `SameSite=Lax` で、他サイトからの POST には Cookie が付かない。
- **Alternatives considered**: `Origin`・`Sec-Fetch-Site` ヘッダーの確認を足す。上の 2 つで足りるため、
  この機能では入れない。

## R5. CORS の許可を外す（Clarifications Q5）

- **Decision**: `go-chi/cors` のミドルウェアと依存、`config.Config.CORSAllowedOrigins`、
  `CORS_ALLOWED_ORIGINS`（`.env.example`、`deploy-api.yml`、`infra/gcp.md`）、GitHub 変数
  `CLOUD_RUN_CORS_ALLOWED_ORIGINS`（`infra/gcp.md` の手順）を取り除く。
- **Rationale**: ブラウザは web とだけ通信する。api の `/playground`（開発時だけ）は api 自身の
  オリジンから `/graphql` を呼ぶため、CORS はいらない。
- **注意**: GitHub に設定済みの変数そのものの削除は、利用者が GitHub の画面で行う（クラウドの
  リソースは操作しない）。手順書に書く。

## R6. パスワードのハッシュ

- **Decision**: `golang.org/x/crypto/argon2` の `IDKey`。パラメーターは OWASP の推奨の最低値の
  m = 19 MiB（19456 KiB）、t = 2、p = 1。ソルトは 16 バイト、出力は 32 バイト。保存は PHC 形式の文字列
  `$argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>`（base64 はパディングなし）。照合は
  `crypto/subtle.ConstantTimeCompare`。保存された文字列のパラメーターで照合するため、後で値を
  変えても古いハッシュを照合できる。
- **Rationale**: Cloud Run のメモリ（既定 512 MiB）でも、同時に数件のログインなら足りる。
  ユーザーが少ないため、メモリより照合の速さ（数十ミリ秒）を気にする必要はない。
- **存在しないログイン ID**: 起動時に作った捨てのハッシュで照合を 1 回行い、応答時間からログイン
  ID の有無がわからないようにする（FR-018）。
- **Alternatives considered**: bcrypt（72 バイトを超えるパスワードを切り捨てる。ADR 0006 で Argon2id
  に決めている）。

## R7. 招待コード・再設定コード・セッション ID の生成と保存

- **Decision**: どれも `crypto/rand.Text()`（Go 1.24 から。128 ビット、base32 の大文字と数字の
  26 文字）で作り、DB には SHA-256 のハッシュ（`bytea`）だけを保存する。入力されたコードは、
  前後の空白・空白・ハイフンを取り除き、大文字にしてからハッシュにする。画面とコマンドでは、
  読みやすいように 4 文字ごとにハイフンで区切って表示する。
- **Rationale**: 128 ビットの乱数は総当たりできないため、パスワードのような遅いハッシュやソルトは
  いらない（FR-008）。ハッシュが一致するかを DB の一意の索引で引ける。
- **Alternatives considered**: Argon2id でハッシュにする（索引で引けず、コードごとに照合が必要になる）。

## R8. 1 回だけ使えるコードの同時の使用（FR-012、FR-037、SC-005）

- **Decision**: 1 つのトランザクションの中で、`SELECT … FOR UPDATE` で「未使用で期限内」のコードの
  行を押さえ、使用済みにする。2 つ目のトランザクションは行のロックを待ち、READ COMMITTED で
  条件を再評価するため、使用済みになった行は返らない。
  - 登録は「招待コードを押さえる → ユーザーを作る → 招待コードを使用済みにする → セッションを
    作る」の順にする。ログイン ID の重複（一意の制約の違反）ならロールバックし、招待コードは未使用の
    まま残る（FR-014）。
  - 招待コードを先に確かめるのは、使えない招待コードの持ち主に、ログイン ID が使われているか
    どうかを知られないため。
- **Alternatives considered**: `UPDATE … WHERE used_at IS NULL RETURNING` を先に実行する（ユーザーを
  作る前に使用済みの印が要るため、使った人の列を後から埋めることになり、制約で守りにくい）。

## R9. ログインの失敗とロック（FR-027〜FR-030、FR-040）

- **Decision**: `auth_failures` の表に、目的（`login` / `password_reset`）とログイン ID（小文字に
  そろえた値）ごとに、続けて失敗した回数とロックが解ける時刻を持つ。
  - 判定の順: ロック中なら照合せずに `LOCKED` を返す → 照合 → 失敗なら回数を足し、5 回目で
    `locked_until = now() + 15 分` にする → 成功なら行を消す。
  - ロックの時間が過ぎた後の最初の失敗は、1 回目から数え直す。
  - 回数の加算は `INSERT … ON CONFLICT DO UPDATE` 1 文で行い、同時の失敗でも数え漏れない。
  - ログイン ID の形式（R11）に合わない入力は、存在し得ないため記録せず、同じ失敗の文言を返す。
  - パスワードの再設定の失敗は `password_reset` として別に数える。ログインのロックとは独立させ、
    再設定でログインのロックを解けるようにする（FR-030）。
- **掃除**: ログインに成功したときに、期限切れのセッションと、最後の失敗から 1 日以上たった
  失敗の記録をまとめて消す（存在しないログイン ID の行が増え続けないようにする）。
  常時起動の仕組み（cron）は置かない（constitution のコスト）。
- **Alternatives considered**: メモリに持つ（Cloud Run が複数インスタンスになるため ADR 0006 で不可）。

## R10. セッションの延長（FR-022）

- **Decision**: 認証のミドルウェアがセッションを読んだとき、`last_used_at` が 1 時間より前なら
  `last_used_at = now()`、`expires_at = now() + 30 日` に更新する。
- **Rationale**: リクエストのたびに書き込むと、Neon への書き込みが無駄に増える。1 時間の誤差は
  30 日の期限に対して問題にならない。

## R11. 入力の規則の検証（Clarifications Q1）

- **Decision**: api が正とし、web でも同じ規則で先に確かめる（送信前に項目の近くに表示するため）。
  - ログイン ID: 前後の空白を除き、`^[A-Za-z0-9_-]{3,32}$`。保存は入力のとおりで、一意性は
    `lower(login_id)` の一意の索引で確かめる。
  - パスワード: 8〜128 文字（Unicode のコードポイントで数える）。前後の空白は取り除かない。
  - 表示名: 前後の空白を除いて 1〜32 文字（コードポイント）。制御文字（Unicode の `Cc`）と、
    行区切り・段落区切り（U+2028、U+2029）を含むものは受け付けない。
  DB にも同じ内容の `CHECK` 制約を置き、コマンドからの操作でも規則を外れないようにする。
- **Alternatives considered**: 書記素（見た目の 1 文字）で数える。絵文字の数え方が環境で変わるため、
  コードポイントで数える。

## R12. GraphQL のエラーの返し方

- **Decision**:
  - 利用者が直せる失敗（入力の規則、使えないコード、ログインの失敗、ロック）は、ミューテーションの
    payload の `errors: [InputError!]!` で、項目（`field`）とコード（`code`）を返す。文言は web が
    コードから決める。
  - ログインが要る操作にログインせずにアクセスしたら、GraphQL のエラーで `extensions.code` を
    `UNAUTHENTICATED`、管理者でなければ `FORBIDDEN` にする。
- **Rationale**: 画面の文言を web の 1 か所で管理でき、api の応答に日本語の文言を持たせずに済む。
  型付きの payload にすると、GraphQL Code Generator の型で漏れなく扱える。
- **Alternatives considered**: ミューテーションごとの union 型（`LoginSuccess | LoginError`）。型が
  増えるわりに、この機能の画面では得るものが少ない。

## R13. resolver から Cookie を書く

- **Decision**: 認証のミドルウェアが、リクエストごとに「ログインしている人（Viewer）」と、Cookie を
  書くための関数（`http.ResponseWriter` を包んだもの）を `context` に入れる。resolver はそれを使って
  `Set-Cookie` を付ける。gqlgen は resolver が終わってから本文を書くため、ヘッダーの追加が間に合う。
- **N+1**: Viewer はリクエストごとに 1 回だけ、セッションとユーザーを結合した 1 つのクエリで読む。
  この機能に、一覧の中で関連データを引く resolver はないため、DataLoader はまだ使わない
  （constitution V に違反しない。シートの所有者の表示名を引くときに DataLoader を入れる）。

## R14. 管理者のコマンド（Clarifications Q3）

- **Decision**: `apps/api/cmd/admin` に 1 つのコマンドを作り、サブコマンドで分ける。Taskfile の
  `admin:*` から呼ぶ。引数は標準の `flag` で読む。
  - パスワードは `golang.org/x/term` の `ReadPassword` で、画面に出さずに 2 回入力させる。端末で
    ないとき（テスト）は標準入力から 1 行ずつ読む。コマンドの引数や環境変数では受け取らない（FR-041）。
  - 接続先は `DATABASE_URL`。実行の前に、接続先のホストとデータベース名を表示する（本番の DB に
    つないでいることに気付けるようにする）。
  - Taskfile のタスクには `interactive: true` を付け、パスワードの入力を受け付ける。
- **確かめた結果**: Task の `dotenv` は、シェルで設定した環境変数を上書きしない（`PORT` と
  `DATABASE_URL` で確かめた）。本番の DB に対しては、シェルで `DATABASE_URL` を設定して実行すればよい。
- **Alternatives considered**: `psql` で決まった SQL を流す（Argon2id のハッシュを作れない）。

## R15. テストの方針（constitution VI）

- **api**:
  - 入力の検証、コードの正規化、パスワードのハッシュはテーブル駆動の単体テスト。
  - DB を使う処理（登録、ログイン、ロック、再設定、無効化、セッションの延長）は実際の Postgres に
    対して実行する（`TEST_DATABASE_URL`）。テストごとに `auth` 関係の表を `TRUNCATE` する。DB を
    使うテストのパッケージは並列に実行しない（`go test -p 1`。Taskfile の `test:api` と CI で指定）。
  - 表を空にするため、ローカルではテスト用の DB を開発用と分ける（`chara_q_test`。`task db:test:prepare`
    が作ってマイグレーションを適用し、`test:api` が先に実行する）。`TEST_DATABASE_URL` が `DATABASE_URL` と
    同じならテストを失敗させる。CI は使い捨てのサービスコンテナなので分けない。
  - 同時の使用は、別々の接続から 10 回同時に使い、成功が 1 回だけであることを確かめる（SC-005）。
  - HTTP のテスト（`httptest`）で、`Set-Cookie` の属性、GET のミューテーションと `text/plain`・
    フォームの POST が拒否されること、CORS のヘッダーが返らないことを確かめる。
- **web**: Vitest で、フォームの検証と文言、ログイン状態の表示（3 つの状態）、戻り先の URL の検証、
  urql のミューテーションを差し替えた送信の流れを確かめる。
- **手で確かめるもの**: 中継を通した Cookie（Chrome・Firefox・Safari）、最初の描画のログイン状態、
  幅 360px での表示（[quickstart.md](quickstart.md)）。

## R17. 画面を離れた後に残る入力（実装中に見つけたこと）

- **Decision**: 認証のフォーム（ログイン、登録、再設定、アカウントの設定）と招待コードの発行画面は、
  画面を離れたときに入力・表示した内容・失敗の文言を消す（`lib/auth/use-reset-on-hide.ts`）。
- **Rationale**: `cacheComponents` では、離れた画面はアンマウントされず React の `<Activity>` で隠される
  だけで、入力欄の値や表示した内容が DOM に残る
  （`01-app/02-guides/preserving-ui-state.md`）。ログアウトした後も、前のパスワードや発行した招待コードが
  DOM に残り、古い失敗の文言も戻ってくる。`useLayoutEffect` の cleanup は Activity が隠すときにも
  同期的に走るため、そこで状態を戻す。
- **Alternatives considered**: `router.bfcacheId` を `key` にする（戻る・進むでは状態を残すため、
  パスワードが残る場面がある）。

## R16. 追加する依存

| 依存 | 場所 | 用途 |
|---|---|---|
| `golang.org/x/crypto` | `apps/api/go.mod` | Argon2id |
| `golang.org/x/term` | `apps/api/go.mod` | 管理者のコマンドでのパスワードの入力 |
| `github.com/google/uuid`（間接から直接へ） | `apps/api/go.mod` | sqlc の `uuid` の型 |
| `server-only` | `apps/web/package.json` | サーバー専用のモジュールを Client Components から読ませない |
| shadcn/ui の `dropdown-menu`（Base UI の Menu） | `apps/web/src/components/ui/` | ヘッダーのアカウントのメニュー |

取り除く依存: `github.com/go-chi/cors`。

基盤技術の変更はないため新しい ADR は作らず、ADR 0006 に「後続の Spec で決めること」の決定を
追記する（spec の Assumptions）。
