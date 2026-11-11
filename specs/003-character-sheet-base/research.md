# Research: キャラクターシートの共通基盤

Phase 0 の調査結果。Technical Context の不明点と、技術の選び方を決める。
Next.js は `apps/web/node_modules/next/dist/docs/` の 16.4.3 のドキュメントで確かめた
（`apps/web/AGENTS.md` の指示）。gqlgen・sqlc・pgx は 002-auth と同じ版を使う。

## R1. 共通の項目を、列と JSONB のどちらに置くか（FR-001、FR-004）

- **Decision**: 共通の項目（キャラクター名・プレイヤー名・年齢・性別・メモ）は、すべて
  `character_sheets` の通常の列にする。JSONB の列 `data` には、システム固有の部分だけを入れる。
- **Rationale**:
  - 入力の規則（長さ・制御文字）を DB の `CHECK` 制約でも守れる。002-auth の `users` と同じ形。
  - `interface CharacterSheet` の項目と列が 1 対 1 になり、JSON の中から値を取り出す処理がいらない。
  - 「同じ値を列と JSON の両方に持たない」（FR-004）を、置き場所を分けることで自然に守れる。
  - constitution III の「検索や一覧に使う共通項目を列にし、シート本体を JSONB に」に反しない。
    年齢・性別・メモは検索に使わないが、列にしても困ることはない。
- **Alternatives considered**:
  - 検索に使うキャラクター名・プレイヤー名だけを列にし、年齢・性別・メモは JSONB に入れる:
    共通の項目の置き場所が 2 つに分かれ、検証も DB と Go に分かれる。
  - シート全体を JSONB に入れ、キャラクター名などを生成列で取り出す: 生成列の式が JSON の形に
    依存し、システムごとに形が違うと扱いにくい。

## R2. GraphQL の型の形（FR-003）

- **Decision**:
  - `interface CharacterSheet` に、ID・システム・所有者・共通の項目・作成日時・更新日時・
    リビジョン（R6）を置く。
  - `YuyakekoyakeSheet` と `BeginningIdolSheet` がこれを実装し、システム固有の部分を
    `data: JSON!` として持つ。`data` は interface には置かない。
  - 空でもよい共通の項目は `String!` にし、未入力は空文字列で表す（null を使わない）。
  - 所有者は公開用の `type User { id: ID!, displayName: String! }` で返す。ログイン ID は
    `Viewer` にだけある（002-auth の FR-034）。
- **Rationale**:
  - `data` を各システムの型に置くと、各システムの Spec が型に項目を切り出すときに（constitution III の
    「型として切り出すのは Spec で必要性を確認してから」）、interface を変えずに済む。
  - 空文字列にすると、web で null と空の 2 通りを扱わずに済み、旧データの移行も単純になる。
- **Alternatives considered**:
  - `data: JSON!` を interface に置く: システムが型を切り出したときに、interface の項目が
    使われなくなるか、互換性を壊す変更になる。
  - システムを区別せず 1 つの型 `CharacterSheet` にする: constitution III（MUST）に反する。

## R3. gqlgen での interface と所有者の resolver

- **Decision**:
  - `YuyakekoyakeSheet` と `BeginningIdolSheet` の Go の型は、`internal/graph/model/sheet.go` に
    手で書く（生成コードではない）。共通の項目は埋め込みの構造体 `SheetCommon` にまとめ、
    所有者は `OwnerID uuid.UUID` で持つ。gqlgen は `owner` の項目に resolver を作る。
  - `gqlgen.yml` の `models` で、2 つの型をこの構造体に結び付ける。
  - DB の行からどちらの型を作るかは、`system` の列で決める変換関数 1 つにまとめる。
- **Rationale**: 生成された構造体を使うと `owner` が `*User` の項目になり、DataLoader で後から
  引く形にできない。手で書いた構造体なら、所有者の ID だけを持たせて resolver で引ける。
- **実装で決めたこと**:
  - gqlgen は interface に getter（`GetOwner() *User` など）を作る。所有者を ID だけで持つ手書きの型では
    `GetOwner` が意味を持たないため、`gqlgen.yml` に `omit_getters: true` を指定した。
  - DB の行から GraphQL の型を作る変換（`toSheetModel`）とシステムの enum の変換は、`internal/sheet` が
    GraphQL に依存しないよう、`internal/graph/sheet_convert.go` に置いた。`internal/sheet/system.go` には
    DB の値（`sheet.System`）だけを置く。
  - シートの入力の失敗は `auth.InputErrors` で返し、resolver の変換（`inputErrors`）を認証と共通にした。
  - sqlc が NULL を許す text の引数を `pgtype.Text` にするため、`sqlc.yaml` に `*string` にする指定を足した
    （既存の uuid・timestamptz と同じ扱い）。
- **Alternatives considered**: 生成された構造体に `owner` の `resolver: true` を指定する:
  構造体から `owner` の項目が消えるだけで、所有者の ID を持たせる場所がない。

## R4. 所有者の表示名の DataLoader（FR-032、constitution V、SC-003）

- **Decision**: `github.com/vikstrous/dataloadgen` v0.0.10 を使う。リクエストごとに
  `UserByID` のローダーを作るミドルウェアを `/graphql` に入れ、`owner` の resolver はローダーから
  引く。まとめて引く SQL は `WHERE id = ANY(@ids::uuid[])` の 1 つ。
- **Rationale**:
  - gqlgen のドキュメントが紹介しているライブラリで、ジェネリクスで型が付き、依存が少ない。
  - ローダーをリクエストごとに作るため、別のユーザーのリクエストの結果が混ざらない。
  - 無効にされたユーザーも表示名を返す（002-auth の US6 シナリオ 5）。
- **確かめ方**: テストで pgx の `QueryTracer` を使い、1 ページの件数を 1 件と 20 件にしたときの
  問い合わせの回数が同じであることを確かめる（SC-003）。
- **Alternatives considered**:
  - `github.com/graph-gophers/dataloader/v7`: 機能は十分だが、API が冗長で、gqlgen の案内からも
    外れた。
  - 一覧の SQL で `users` を結合して表示名を返す: 一覧では N+1 にならないが、閲覧や今後の別の
    経路（複製の結果など）で同じ解決を何度も書くことになる。constitution V は DataLoader を求めている。

## R5. ID と URL（FR-002、FR-015、FR-016）

- **Decision**: ID は 002-auth の表と同じく `uuidv7()` の UUID にする。URL は `/sheets/<UUID>`
  （小文字、ハイフン付き）。形の正しくない ID は、API で `null` を返す（エラーにしない）。
- **Rationale**:
  - UUIDv7 は 74 ビットの乱数を含み、推測できない。先頭の時刻は作成日時（公開している）と同じ
    情報なので、漏れても困らない。
  - 時刻順に並ぶため、主キーの索引が断片化しにくい。
  - 存在しない・削除された・形の正しくない ID を同じ `null` にすると、FR-018 の「どの場合も同じ
    表示」をそのまま満たせる。
- **Alternatives considered**: 短い ID（nanoid など）: URL は短くなるが、依存と一意性の扱いが
  増える。URL の短さは要件にない。

## R6. 同時の更新の検出（FR-028、SC-007）

- **Decision**: `character_sheets` に `revision integer` の列を置き、更新のたびに 1 増やす。
  更新の入力に、読み込んだときの `revision` を含めさせ、
  `UPDATE ... WHERE id = @id AND owner_id = @owner AND deleted_at IS NULL AND revision = @revision`
  で 1 文で更新する。0 行だったときだけ、行を読み直して「見つからない」「所有者でない」
  「リビジョンが古い」のどれかを判定する。
- **Rationale**:
  - 確認と更新を 1 文にするため、同時に保存されても、成功するのは 1 回だけになる。
  - 更新日時を比べる方法は、DB はマイクロ秒、GraphQL の `Time` は文字列で、往復で値がずれる
    おそれがある。整数なら正確に比べられる。
- **Alternatives considered**:
  - 後勝ち（何も確かめない）: SC-007 を満たせない。
  - `SELECT ... FOR UPDATE` の後に更新する: 2 文になり、トランザクションが必要になる。

## R7. 削除（FR-029、FR-030、Clarifications Q3）

- **Decision**: `deleted_at timestamptz` の列で表す（論理削除）。一覧・閲覧・更新・複製・削除の
  SQL は、すべて `deleted_at IS NULL` を条件に含める。並べ替えの索引も
  `WHERE deleted_at IS NULL` の部分索引にする。取り消しは、管理者が DB に接続して
  `UPDATE character_sheets SET deleted_at = NULL WHERE id = '<ID>'` を実行する（手順は
  [quickstart.md](quickstart.md) と [data-model.md](data-model.md)）。
- **Rationale**: Clarifications Q3 の決定のとおり。取り消しの操作は API に置かないため、
  コマンドも作らない（Spec の Assumptions）。
- **注意**: 削除しても `updated_at` と `revision` は変えない。取り消したときに、削除前と同じ
  内容・同じ並び順で戻る（FR-030）。

## R8. キーワードの検索（FR-010、SC-002）

- **Decision**: SQL で、キャラクター名とプレイヤー名を `lower(normalize(<列>, NFKC))` で
  正規化し、同じように正規化したキーワードと `LIKE` の部分一致で比べる。キーワードの `\`・`%`・`_`
  は、正規化の後にエスケープする（`ESCAPE '\'`）。キーワードの前後の空白は Go で取り除き、空なら
  条件を付けない。この機能では、検索用の索引（pg_trgm など）は作らない。
- **Rationale**:
  - NFKC で全角の英数字が半角に、`lower` で大文字が小文字になり、Spec の「大文字と小文字、全角と
    半角を区別しない」を満たす。正規化を SQL の 1 か所にすると、Go と DB で Unicode の版がずれる
    心配もない。
  - 全角の `％` は NFKC で `%` になるため、エスケープは正規化の後でないと漏れる。
  - 10,000 件の順次の走査は数十ミリ秒で終わり、SC-002（1 秒）に十分な余裕がある。pg_trgm は
    3 文字未満のキーワードに効かず、日本語の短い名前ではあまり役に立たない。
- **範囲の外**: ひらがなとカタカナの違いは区別する（NFKC では揃わない）。必要になったら別の
  Spec で扱う。
- **Alternatives considered**:
  - 正規化した文字列を生成列に持つ: 生成列の式は IMMUTABLE でなければならず、照合順序に依存する
    `lower` の扱いに注意がいる。件数が少ないうちは不要。
  - Go の `golang.org/x/text/unicode/norm` で正規化した値を列に保存する: 列と元の値のずれを
    防ぐ仕組みが要る。

## R9. 並び順とページ送り（FR-011、FR-012、SC-006）

- **Decision**:
  - 並び順は 6 つ（3 つの項目のそれぞれに順と逆の順、2026-11-11 に追加）。どれも ID を最後の
    基準にして、順序を一意にする。ID の向きは項目の向きに合わせる。
    - `UPDATED_AT_DESC`（既定）: `updated_at DESC, id DESC`
    - `UPDATED_AT_ASC`: `updated_at ASC, id ASC`
    - `CREATED_AT_DESC`: `created_at DESC, id DESC`
    - `CREATED_AT_ASC`: `created_at ASC, id ASC`
    - `CHARACTER_NAME_ASC`: `character_name COLLATE "ja-x-icu" ASC, id ASC`
    - `CHARACTER_NAME_DESC`: `character_name COLLATE "ja-x-icu" DESC, id DESC`
  - 逆の順のために索引は足さない。B-tree の索引は逆向きにも読めるため、すべての列の向きを
    反転した `ORDER BY` には同じ索引が使える。
  - ページは `LIMIT 20 OFFSET (page - 1) * 20`。先に件数を数え、`page` を 1〜全ページ数の範囲に
    丸めてから一覧を読む。丸めた後のページ番号を応答に含め、web はそれを表示する。件数が 0 なら
    ページは 1。
  - 並び順ごとに `WHERE deleted_at IS NULL` の部分索引を作る。
- **Rationale**:
  - ページ番号で URL を共有する要件（FR-013）には、カーソルより OFFSET が合う。10,000 件・
    1 ページ 20 件なら OFFSET の遅さは問題にならない。
  - 範囲外のページを API の側で丸めると、web と API で丸め方がずれない（Spec の Edge Cases）。
  - `ja-x-icu` は、ひらがなとカタカナを五十音の順に並べる。漢字は読みを持たないため、読みの順には
    ならない（文字の順）。読みの項目を持たない以上、これ以上は望めない。
- **確かめること**: ローカルの `postgres:18.6-alpine` と Neon で `ja-x-icu` の照合順序が使えるか
  （`SELECT collname FROM pg_collation WHERE collname = 'ja-x-icu'`）。使えなければ `und-x-icu`、
  それもなければ既定の照合順序にして、この research に記録する。
- **確認結果（ローカル、2026-11-11）**: `postgres:18.6-alpine`（PostgreSQL 18.6）で `ja-x-icu` と
  `und-x-icu` の両方があることを確かめた。`ja-x-icu` を使う。
- **確認結果（Neon）**: 未確認。本番環境の構築の Spec で、マイグレーションを適用するときに確かめる。
- **Alternatives considered**:
  - カーソルでのページ送り: 一覧を見ている間の変化に強いが、ページ番号を指定して開けない。
  - 件数を `count(*) OVER ()` で一覧と同時に数える: 範囲外のページの丸めができない。

## R10. 「自分のシートだけ」（FR-010、Edge Cases）

- **Decision**: 一覧のクエリの入力に `mine: Boolean` を持たせる。API はログインしているときだけ
  `owner_id = <本人の ID>` の条件を付け、ログインしていなければ無視する。URL は `mine=1`。
  web は、ログインしていなければ絞り込みの欄を表示しない。
- **Rationale**: 所有者の ID を URL や入力で受け取らないため、他のユーザーで絞り込む機能を
  作らずに済み、Spec の範囲を超えない。
- **Alternatives considered**: `ownerId: ID` で任意の所有者を絞り込む: Spec にない機能が増える。

## R11. 入力の検証とエラーの返し方（FR-006、FR-007、002-auth の R12）

- **Decision**:
  - 入力の失敗（項目の規則、システム固有の部分の形と大きさ、リビジョンが古い）は、payload の
    `errors: [InputError!]!` で返す。002-auth の `InputError` を使い回し、`InputField` と
    `InputErrorCode` の enum に値を足す（[contracts/graphql.md](contracts/graphql.md)）。
  - ログインしていない・所有者でない・シートが見つからない、は GraphQL のエラー
    （`extensions.code` が `UNAUTHENTICATED`・`FORBIDDEN`・`NOT_FOUND`）で返す。
  - 検証は `internal/sheet` の Go で行い、同じ規則を DB の `CHECK` 制約にも書く。文字数は
    Unicode のコードポイントで数える（Go の `utf8.RuneCountInString` と Postgres の
    `char_length` が同じになる）。
  - システム固有の部分の大きさは、JSON の文字列にしたときのバイト数で 256KB（262,144 バイト）
    以下とする。リクエストの本体は、余裕を持たせて 1MB までに制限する（`http.MaxBytesReader`）。
- **Rationale**: 002-auth と同じ「利用者が直せる失敗は payload、そうでないものは GraphQL の
  エラー」の分け方にすると、web のエラーの扱いを共通にできる。
- **Alternatives considered**: シート用に別の `SheetInputError` の型を作る: 型が増えるだけで、
  web の扱いが分かれる。

## R12. 複製の名前（FR-023）

- **Decision**: 複製は `INSERT ... SELECT` の 1 文で行い、キャラクター名は
  `left(character_name, 95) || '（コピー）'` にする（`（コピー）` は 5 文字）。
- **Rationale**: 95 文字以下の名前はそのまま、96 文字以上の名前だけが短くなり、上限の 100 文字に
  必ず収まる。1 文なので、元のシートを読んでから書くまでの間に変わる心配もない。

## R13. web からの API の呼び出し（公開の画面）

- **Decision**: `lib/api/server.ts` に、セッションの Cookie があれば付け、なければ付けずに
  api を呼ぶ関数を足す（既存の `queryWithSession` は、セッションがなければ api を呼ばない）。
  一覧と閲覧の画面はこの関数を使う。どちらも `cache: "no-store"` で、`"use cache"` は使わない。
- **Rationale**:
  - 一覧と閲覧はログインなしで見られるが、「自分のシートだけ」と所有者の判定にはログイン状態が
    要る。1 つの関数で両方を扱える。
  - シートは所有者が更新したらすぐに反映されるべきで、キャッシュすると古い内容が見える。
- **Next.js の注意**（`01-getting-started/08-caching.md`）: `searchParams`・`params`・`cookies()` を
  読むコンポーネントは `<Suspense>` の中に置く（`cacheComponents`）。

## R14. 一覧の画面と検索のフォーム（FR-010〜FR-013）

- **Decision**:
  - 検索のフォームは `next/form` の `<Form action="/sheets">`（GET）で作る。送信すると URL の
    クエリが変わり、画面の移動になる。ページの値はフォームに含めないため、条件や並び順を変えて
    送信すると最初のページに戻る（FR-011）。
  - システムと並び順は、UI 基盤の共通部品 `SelectField` を使う（Base UI の Select は、`name` を
    渡すとフォームの値として送られる）。「自分のシートだけ」は、ネイティブのチェックボックスを
    Tailwind で整えて使う。
  - URL のクエリは `system`（`yuyakekoyake`・`beginning-idol`）・`q`・`mine=1`・`sort`
    （`updated`・`updated-asc`・`created`・`created-asc`・`name`・`name-desc`）・`page`。
    読み取りと検証は純粋な関数にまとめ、知らない値は既定の値にする（Edge Cases）。
  - ページ送りは、前へ・次へと、ページ番号のリンク（多いときは省略する）。全体の件数を表示する。
- **Rationale**: `next/form` は JavaScript が読み込まれる前でも普通の GET のフォームとして動き、
  読み込んだ後はページを読み直さずに移動する（`02-components/form.md`）。URL がそのまま共有できる。
- **Alternatives considered**: 値が変わるたびに自動で検索する: キーワードの入力のたびに画面が
  移動し、履歴が増える。送信のボタンで十分。
- **実装で変えたこと**: `next/form` の GET は、空の値も `?q=&system=` のようにクエリに残し、
  「既定の値は URL に含めない」（contracts/web-routes.md）と合わなかった。そのため、普通の
  `<form method="get" action="/sheets">` にし、JavaScript が動くときは送信を止めて、
  `sheetSearchHref` で組み立てた URL へ `router.push` で移動する。JavaScript が動かないときも GET の
  フォームとして送られ、空の値は既定の値として読まれる。フォームは条件から作った `key` を持ち、
  「条件を外す」などで URL の条件が変わったら入力を作り直す。
- **逆の順と見出しでの並び替え（2026-11-11 に追加、FR-011）**:
  - `sort` は 1 つのキーのまま、項目の最初の向き（日時は新しい順、名前は名前の順）は接尾辞なし、
    逆の向きは `-asc`・`-desc` を付ける。既存の URL（`sort=created` など）は同じ意味のまま使える。
    プルダウンは 1 つの値を送るだけで済み、向きのための 2 つ目の入力が要らない。
  - 表の見出しのうち、キャラクター名と更新日時を、並び順を変えた一覧の URL（`page` なし）への
    リンクにする。JavaScript がなくても動き、URL で共有できる。並んでいない列は最初の向き、
    並んでいる列は逆の向きへのリンク。
  - 並んでいる列の `<th>` に `aria-sort`（`ascending`・`descending`）を付け、見出しの横に上向き・
    下向きの矢印を表示する。並んでいない列には、並び替えられることを示す上下の矢印を薄く表示する。
    リンクには、選ぶとどの並び順になるかを視覚的に隠した文で添える（例: 「更新日時の古い順に
    並べ替える」）。
  - 作成日時は列にないため、見出しでは選べない。カードの形（`sm` 未満）には見出しがないため、
    プルダウンを残す。
  - **Alternatives considered**: `sort` と向き（`dir`）の 2 つのキーにする: プルダウンが 2 つになるか、
    1 つのプルダウンの値を 2 つのキーに分けるための処理が要る。見出しのクリックをボタンにして
    `router.push` する: JavaScript がないと動かず、リンクで足りる。

## R15. 閲覧の画面と「見つからない」の表示（FR-017、FR-018、FR-020）

- **Decision**:
  - `app/sheets/[id]/page.tsx` で `characterSheet(id:)` を読み、`null` なら `notFound()` を呼ぶ。
    `app/sheets/[id]/not-found.tsx` で「シートが見つかりません」と一覧への移動を表示する。
  - 題名は `generateMetadata` で `<キャラクター名>` にする（ルートの `title.template` でアプリ名が
    付く）。本体と `generateMetadata` は React の `cache()` で包んだ関数から読み、api の呼び出しを
    1 回にする。
  - システム固有の部分と、所有者の操作の場所は、システムごとの部品を登録する表
    （`lib/sheets/system-views.ts`）で差し込む。この機能では、どのシステムも「このシステムの項目は
    準備中です」を表示し、所有者の操作は何も置かない。
  - 日時は `Asia/Tokyo` の `YYYY/MM/DD HH:mm` で表示する。
- **注意**（`03-file-conventions/loading.md`、`not-found.md`）: 画面を流し始めた後の `notFound()` は
  HTTP の状態が 200 のままで、`noindex` の meta が付く。Spec は状態コードを求めていないため、
  これでよいとする。
- **Alternatives considered**: proxy（`proxy.ts`）で ID を確かめて 404 を返す: 毎回 api を
  呼ぶことになり、ドキュメントも重い処理を置かないよう勧めている。

## R16. ヘッダーの一覧への入口（FR-034）

- **Decision**: `SiteHeader` の右側の操作の領域を「テーマの切り替え・シート一覧・アカウント」の順にし、
  `/sheets` への「シート一覧」のリンクをテーマの切り替えとアカウントのメニューの間に置く。`SiteHeader` は、
  ナビの前に置く `utilities`（テーマの切り替え）と、後に置く `actions`（アカウント）を受け取る。見た目は「ログイン」と同じ（`buttonVariants({ variant: "ghost" })`）。
  幅が狭い画面（`sm` 未満）では、アイコンだけにして `aria-label` を付ける。現在のパスで `aria-current` を
  付けるため `usePathname()` を読み、`<Suspense>` の中に置く（fallback は `aria-current` なしの同じリンク）。
- **Rationale**: 幅 360px でも、ロゴ・リンク・テーマの切り替え・アカウントのメニューが重ならない
  ようにする（002-auth の FR-025、この機能の FR-035）。
- **実装で変えたこと（2026-11-11）**: 最初はロゴの右に置いた。位置の計算ではロゴと上下の中心がそろって
  いたが、ロゴの書体は小文字の部分が下寄りに見え、日本語の「シート一覧」が浮いて見えた。同じ見た目の
  操作と並べる右側の領域に移した。左端はアプリ名、右端は操作という UI 基盤のヘッダーの形にも合う。
  移した直後は「シート一覧・テーマ・アカウント」の順で、アイコンだけのテーマの切り替えが文字の付いた
  操作に挟まれて不揃いに見えたため、テーマの切り替えを左端に移した。アカウントのメニューが右端に来る、
  よくある配置にもなる。
  また、`usePathname()` を `<Suspense>` の外で読むと、`next build` で `/sheets/[id]` の事前描画が止まる
  ことがわかった（`cacheComponents`）。

## R17. テストの方針（constitution VI）

- **api**:
  - 入力の規則は、テーブル駆動の単体テスト（`internal/sheet/validate_test.go`）。
  - SQL（検索の正規化とエスケープ、3 つの並び順と同順の扱い、ページの丸め、所有者の条件、
    リビジョン、論理削除、複製の名前）は、実際の Postgres（`TEST_DATABASE_URL`）でのテスト。
  - 権限（作成・複製はログイン、更新・削除は所有者だけ、管理者も例外でない）は、`httptest` で
    GraphQL のリクエストを送るテスト。ログイン ID が応答に含まれないことも確かめる（SC-005）。
  - DataLoader は、pgx の `QueryTracer` で問い合わせの回数を数えるテスト（SC-003）。
- **web**（Vitest + Testing Library）:
  - URL のクエリの読み取りと検証（知らない値、範囲外のページ）。
  - 検索のフォーム（送信される URL、ログインしていないときの「自分のシートだけ」）。
  - ページ送りの部品、一覧の行、閲覧の共通部分、見つからない表示、ヘッダーのリンク。
  - axe による検査（既存の `src/test/axe.ts`）。
- **手で確かめること**: [quickstart.md](quickstart.md)。

## R18. 追加する依存

| 依存 | 場所 | 理由 |
|---|---|---|
| `github.com/vikstrous/dataloadgen` v0.0.10 | `apps/api/go.mod` | 所有者の DataLoader（R4） |

web には依存を足さない（`next/form` は Next.js に含まれる。チェックボックスはネイティブの要素）。
基盤技術は変えないため、ADR は新しく作らない。
