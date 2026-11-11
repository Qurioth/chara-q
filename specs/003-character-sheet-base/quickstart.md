# Quickstart: キャラクターシートの共通基盤の動作確認

この機能では作成・更新・複製・削除の画面がないため、それらは web 経由の `/graphql` に
GraphQL のリクエストを送って確かめる。ログインが要らないクエリは、GraphQL のプレイグラウンド
（`http://localhost:8080/playground`、開発時だけ）でも確かめられる。
API の形は [contracts/graphql.md](contracts/graphql.md)、画面は [contracts/web-routes.md](contracts/web-routes.md)。

## 前提

```sh
task db:up
task db:migrate
task generate
task dev
```

- 002-auth の手順で、ユーザーを 2 人（A と B）作っておく。A は管理者でもよい（管理者でも
  自分のシートしか更新できないことを確かめるため、片方は管理者にする）。
- ログインが要る操作は、web（`http://localhost:3000`）でログインしたタブの DevTools のコンソールから
  `fetch("/graphql", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ query, variables }) })`
  で送る。セッションの Cookie は web のオリジンにあるため、プレイグラウンド（api のオリジン）からは
  送れない。
- 照合順序の確認（research R9）:

  ```sql
  SELECT collname FROM pg_collation WHERE collname IN ('ja-x-icu', 'und-x-icu');
  ```

## 1. 作成と閲覧（US3、US2）

1. A でログインし、`createCharacterSheet` で ゆうやけこやけ のシートを作る（`data` は `{}`）。
   → `revision` が 1、`owner.displayName` が A の表示名。
2. ログアウトして `/sheets/<ID>` を開く。→ 共通の項目・システム・所有者の表示名・作成日時・
   更新日時が表示される。ブラウザのタブの題名にキャラクター名が出る。
3. 応答（DevTools のネットワーク、ページのソース）に A のログイン ID が含まれないこと（SC-005）。
4. 存在しない UUID・`/sheets/abc` を開く。→ どちらも「シートが見つかりません」。
5. ログインしていない状態で `createCharacterSheet` を送る。→ `UNAUTHENTICATED`。
6. キャラクター名を空・101 文字にして作る。→ `CHARACTER_NAME` の `REQUIRED`・`TOO_LONG`。
   `data` に配列を渡す → `DATA` の `NOT_OBJECT`。

## 2. 一覧と検索（US1）

1. A と B で、システム・キャラクター名・プレイヤー名の違うシートを 25 件以上作る（全角の英字の
   名前、`%` を含む名前も入れる）。
2. ログインせずに `/sheets` を開く。→ 更新日時の新しい順に 20 件、「全 N 件」、ページ送り。
3. システムを選んで検索 → そのシステムだけ。キーワードに半角の小文字を入れ、全角の大文字の名前が
   見つかること。`%` で検索して、`%` を含む名前だけが見つかること。
4. 並び順を「キャラクター名の順」にする → 五十音の順、ページは 1 に戻る。表の見出しの
   「キャラクター名」をもう一度選ぶ → 逆の順（`?sort=name-desc`）、見出しに下向きの矢印。
   「更新日時」を選ぶ → 新しい順（`sort` なし）、もう一度選ぶ → 古い順（`?sort=updated-asc`）。
5. 2 ページ目の URL を別のブラウザで開く → 同じ結果（SC-006）。`?page=999`・`?sort=foo`・
   `?system=bar` → エラーにならず、最後のページ・既定の並び順・すべてのシステム。
6. A でログインし、「自分のシートだけ」→ A のシートだけ。ログアウトすると、この欄は表示されない。
7. ヘッダーの「シート一覧」から `/sheets` に移動できる。幅 360px で横スクロールが出ない。

## 3. 更新（US4）

1. A で、自分のシートを `updateCharacterSheet`（`revision: 1`）で更新する → `revision` が 2、
   `updatedAt` が新しくなり、一覧の先頭に移る。`createdAt` は変わらない。
2. もう一度 `revision: 1` で更新する → `REVISION_CONFLICT`、内容は 1 の結果のまま（SC-007）。
3. B（管理者を含む）とログインしていない状態で、A のシートを更新する → `FORBIDDEN` と
   `UNAUTHENTICATED`、内容は変わらない（SC-004）。

## 4. 複製（US5）

1. B で A のシートを `copyCharacterSheet` → B が所有者の新しいシート、名前の末尾に `（コピー）`。
2. 複製を B が更新しても、元のシートは変わらない。100 文字の名前のシートを複製すると、100 文字に
   収まる。

## 5. 削除（US6）

1. B で A のシートを削除する → `FORBIDDEN`。
2. A で削除する → 一覧に出ず、`/sheets/<ID>` は「シートが見つかりません」。もう一度削除・更新・
   複製する → `NOT_FOUND`。
3. DB で `deleted_at` が入り、内容が残っていることを確かめる。[data-model.md](data-model.md) の
   「削除の取り消し」の SQL で戻す → 同じ URL・同じ内容で、一覧の元の位置に戻る。

## 6. 自動テスト

```sh
task lint
task test
task check:generated
```

- 所有者の表示名の問い合わせ回数のテスト（SC-003）と、権限のテスト（SC-004）が通ること。
