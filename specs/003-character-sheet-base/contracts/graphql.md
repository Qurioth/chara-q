# Contract: GraphQL（`schema/graphql/character_sheet.graphqls`）

スキーマの案。実装ではこのファイルを `schema/graphql/` に置き、`task generate` で gqlgen と
GraphQL Code Generator のコードを作る（constitution I・II）。設計の理由は [research.md](../research.md)
の R2〜R6、R9〜R12。

```graphql
# キャラクターシートの共通基盤（specs/003-character-sheet-base）。
# 共通の項目は interface CharacterSheet に置き、システム固有の部分は各システムの型の data（JSON）で返す
# （constitution III）。所有者は公開用の User で返し、ログイン ID は含めない（002-auth の FR-034）。

extend type Query {
  "シートを 1 枚返す。存在しない・削除された・形の正しくない ID なら null"
  characterSheet(id: ID!): CharacterSheet
  "削除されていないシートの一覧。ログインは不要"
  characterSheets(
    filter: CharacterSheetFilter
    orderBy: CharacterSheetOrder! = UPDATED_AT_DESC
    "1 から始まるページ番号。範囲の外なら、範囲の中の最も近いページにする"
    page: Int! = 1
  ): CharacterSheetPage!
}

extend type Mutation {
  "シートを作る（ログインが必要）。所有者はログインしている本人"
  createCharacterSheet(input: CreateCharacterSheetInput!): CharacterSheetPayload!
  "シートを更新する（所有者だけ）"
  updateCharacterSheet(input: UpdateCharacterSheetInput!): CharacterSheetPayload!
  "シートを複製する（ログインが必要）。複製の所有者はログインしている本人"
  copyCharacterSheet(id: ID!): CharacterSheetPayload!
  "シートを削除する（所有者だけ）。一覧と閲覧から消えるが、データは残る"
  deleteCharacterSheet(id: ID!): DeleteCharacterSheetPayload!
}

"シートのゲームシステム"
enum GameSystem {
  YUYAKEKOYAKE
  BEGINNING_IDOL
}

"公開されるユーザーの情報。ログイン ID は含めない"
type User {
  id: ID!
  displayName: String!
}

"全システム共通のシートの項目"
interface CharacterSheet {
  id: ID!
  system: GameSystem!
  owner: User!
  "必須。1〜100 文字"
  characterName: String!
  "未入力は空文字列。0〜50 文字"
  playerName: String!
  "未入力は空文字列。0〜20 文字"
  age: String!
  "未入力は空文字列。0〜20 文字"
  sex: String!
  "未入力は空文字列。0〜10,000 文字。改行を含む"
  memo: String!
  "更新のたびに 1 増える。更新の入力に、読み込んだときの値を渡す"
  revision: Int!
  createdAt: Time!
  updatedAt: Time!
}

"ふしぎもののけRPG ゆうやけこやけ のシート"
type YuyakekoyakeSheet implements CharacterSheet {
  id: ID!
  system: GameSystem!
  owner: User!
  characterName: String!
  playerName: String!
  age: String!
  sex: String!
  memo: String!
  revision: Int!
  createdAt: Time!
  updatedAt: Time!
  "システム固有の部分（JSON のオブジェクト）。中身は ゆうやけこやけ の Spec で決める"
  data: JSON!
}

"駆け出しアイドルRPG ビギニングアイドル のシート"
type BeginningIdolSheet implements CharacterSheet {
  id: ID!
  system: GameSystem!
  owner: User!
  characterName: String!
  playerName: String!
  age: String!
  sex: String!
  memo: String!
  revision: Int!
  createdAt: Time!
  updatedAt: Time!
  "システム固有の部分（JSON のオブジェクト）。中身は ビギニングアイドル の Spec で決める"
  data: JSON!
}

input CharacterSheetFilter {
  "null ならすべてのシステム"
  system: GameSystem
  "キャラクター名・プレイヤー名の部分一致。大文字と小文字、全角と半角を区別しない。空なら条件なし"
  keyword: String
  "true なら、ログインしている本人のシートだけ。ログインしていなければ無視する"
  mine: Boolean
}

enum CharacterSheetOrder {
  "更新日時の新しい順（既定）"
  UPDATED_AT_DESC
  "更新日時の古い順"
  UPDATED_AT_ASC
  "作成日時の新しい順"
  CREATED_AT_DESC
  "作成日時の古い順"
  CREATED_AT_ASC
  "キャラクター名の順"
  CHARACTER_NAME_ASC
  "キャラクター名の逆の順"
  CHARACTER_NAME_DESC
}

type CharacterSheetPage {
  nodes: [CharacterSheet!]!
  "絞り込みに合うシートの件数"
  totalCount: Int!
  "実際に返したページ（範囲の外を丸めた後）"
  page: Int!
  "1 ページの件数（20）"
  pageSize: Int!
  "全ページ数。0 件なら 1"
  totalPages: Int!
}

"作成と更新で共通の入力"
input CharacterSheetFieldsInput {
  characterName: String!
  playerName: String! = ""
  age: String! = ""
  sex: String! = ""
  memo: String! = ""
  "JSON のオブジェクト。256KB まで"
  data: JSON! = {}
}

input CreateCharacterSheetInput {
  system: GameSystem!
  fields: CharacterSheetFieldsInput!
}

input UpdateCharacterSheetInput {
  id: ID!
  "読み込んだときの revision。現在の値と違えば REVISION_CONFLICT"
  revision: Int!
  fields: CharacterSheetFieldsInput!
}

type CharacterSheetPayload {
  "成功したときだけ値がある"
  characterSheet: CharacterSheet
  errors: [InputError!]!
}

type DeleteCharacterSheetPayload {
  "削除したシートの ID"
  deletedId: ID!
}
```

002-auth の `auth.graphqls` の enum に、次の値を足す。

```graphql
enum InputField {
  # 既存の値はそのまま
  CHARACTER_NAME
  PLAYER_NAME
  AGE
  SEX
  MEMO
  DATA
}

enum InputErrorCode {
  # 既存の値（REQUIRED、TOO_LONG、INVALID_CHARACTERS など）はそのまま使う
  "JSON のオブジェクトでない"
  NOT_OBJECT
  "大きすぎる（システム固有の部分）"
  TOO_LARGE
  "読み込んだ後に、別の場所で更新された"
  REVISION_CONFLICT
}
```

## ふるまい

| 操作 | ログイン | 成功したとき | 主な失敗 |
|---|---|---|---|
| `characterSheet` | 不要 | 有効なシートを返す | 見つからなければ `null`（エラーにしない） |
| `characterSheets` | 不要 | 絞り込み・並び順・ページに合うシートを返す。`mine` はログインしているときだけ効く | なし（知らない値は GraphQL の検証で弾かれる） |
| `createCharacterSheet` | 必要 | 本人が所有者のシートを作り、返す。`revision` は 1 | 入力の規則の失敗（項目ごと） |
| `updateCharacterSheet` | 所有者 | 共通の項目と `data` を置き換え、`revision` を 1 増やし、`updatedAt` を更新する | 入力の規則の失敗、`REVISION_CONFLICT`（`field` は null）、`FORBIDDEN`、`NOT_FOUND` |
| `copyCharacterSheet` | 必要 | 本人が所有者の複製を作り、返す。名前の末尾に `（コピー）` | `NOT_FOUND` |
| `deleteCharacterSheet` | 所有者 | `deleted_at` を記録し、ID を返す | `FORBIDDEN`、`NOT_FOUND`（削除済みも含む） |

- 文字列の入力は、前後の空白を取り除いてから検証する（`memo` は除く。`memo` の `\r\n` と `\r` は
  `\n` にそろえる）。
- 入力の規則の失敗は、項目ごとに `InputError` を返す。1 つの項目に複数の失敗があっても、返すのは
  最初の 1 つ（`REQUIRED` → `TOO_LONG` → `INVALID_CHARACTERS` の順）。
- `updateCharacterSheet` は、入力の規則を確かめた後に、所有者とリビジョンを 1 文で確かめて
  更新する（R6）。更新できなかったときの判定の順は、`NOT_FOUND`（存在しない・削除済み）→
  `FORBIDDEN`（所有者でない）→ `REVISION_CONFLICT`。
- システムは更新の入力に含めないため、変えられない（FR-005）。所有者も入力に含めない（FR-021、FR-031）。
- `characterSheets` の `nodes` の `owner` は DataLoader でまとめて引く（R4）。

## GraphQL のエラー（`errors[].extensions.code`）

002-auth のコードに `NOT_FOUND` を足す。`FORBIDDEN` の文言は「この操作は行えません」に変え、
管理者の操作とシートの操作の両方で使う。

| コード | 場面 |
|---|---|
| `UNAUTHENTICATED` | 作成・更新・複製・削除に、有効なセッションなしでアクセスした |
| `FORBIDDEN` | 所有者でないユーザー（管理者を含む）が、更新・削除しようとした |
| `NOT_FOUND` | 更新・複製・削除の対象が、存在しない・削除済み・形の正しくない ID だった |

- 確認の順は `UNAUTHENTICATED` → `NOT_FOUND` → `FORBIDDEN`。
- web は、この機能では作成・更新の画面を持たないため、これらのエラーを画面で扱うのは各システムの
  Spec から。
