# Contract: GraphQL（`schema/graphql/auth.graphqls`）

既存の `schema/graphql/schema.graphqls` の `type Query` を `extend` し、最初の `type Mutation` を
追加する。gqlgen は `follow-schema` のため、resolver は `internal/graph/auth.resolvers.go` にできる。
次は実装の出発点になる案で、名前と説明は実装のときに整えてよい。型の追加・削除をするときは、
この文書も更新する。

```graphql
extend type Query {
  "ログインしている本人。ログインしていなければ null"
  viewer: Viewer
}

"ログインしている本人。ログイン ID は本人にだけ返す（他のユーザーの情報には loginId を持たせない）"
type Viewer {
  id: ID!
  loginId: String!
  displayName: String!
  isAdmin: Boolean!
}

type Mutation {
  "招待コードで登録し、そのままログインする"
  register(input: RegisterInput!): RegisterPayload!
  "ログインする。成功したらセッションの Cookie を設定する"
  login(input: LoginInput!): LoginPayload!
  "ログアウトする。ログインしていなくても成功する"
  logout: Boolean!
  "自分の表示名を変える（ログインが必要）"
  updateDisplayName(input: UpdateDisplayNameInput!): UpdateDisplayNamePayload!
  "自分のパスワードを変える（ログインが必要）。ほかのセッションはすべて切れる"
  changePassword(input: ChangePasswordInput!): ChangePasswordPayload!
  "再設定コードでパスワードを設定し直す。ログインはしない"
  resetPassword(input: ResetPasswordInput!): ResetPasswordPayload!
  "招待コードを発行する（管理者だけ）"
  issueInvitation: IssueInvitationPayload!
}

input RegisterInput {
  invitationCode: String!
  loginId: String!
  password: String!
  displayName: String!
}

type RegisterPayload {
  "成功したときだけ値がある"
  viewer: Viewer
  errors: [InputError!]!
}

input LoginInput {
  loginId: String!
  password: String!
}

type LoginPayload {
  viewer: Viewer
  errors: [InputError!]!
  "errors に LOCKED があるときだけ値がある"
  lockedUntil: Time
}

input UpdateDisplayNameInput {
  displayName: String!
}

type UpdateDisplayNamePayload {
  viewer: Viewer
  errors: [InputError!]!
}

input ChangePasswordInput {
  currentPassword: String!
  newPassword: String!
}

type ChangePasswordPayload {
  viewer: Viewer
  errors: [InputError!]!
}

input ResetPasswordInput {
  loginId: String!
  resetCode: String!
  newPassword: String!
}

type ResetPasswordPayload {
  ok: Boolean!
  errors: [InputError!]!
  lockedUntil: Time
}

type IssueInvitationPayload {
  "発行したときにだけ返す。4 文字ごとにハイフンで区切った形"
  code: String!
  expiresAt: Time!
}

"利用者が入力を直せば解決する失敗"
type InputError {
  "どの項目の失敗か。フォーム全体の失敗なら null"
  field: InputField
  code: InputErrorCode!
}

enum InputField {
  INVITATION_CODE
  LOGIN_ID
  PASSWORD
  CURRENT_PASSWORD
  NEW_PASSWORD
  DISPLAY_NAME
  RESET_CODE
}

enum InputErrorCode {
  "未入力"
  REQUIRED
  "短すぎる"
  TOO_SHORT
  "長すぎる"
  TOO_LONG
  "使えない文字を含む"
  INVALID_CHARACTERS
  "招待コードが存在しない・使用済み・期限切れ（区別しない）"
  INVITATION_UNAVAILABLE
  "ログイン ID が使われている"
  LOGIN_ID_TAKEN
  "ログイン ID・パスワードが違う、またはユーザーが無効（区別しない）"
  INVALID_CREDENTIALS
  "続けて失敗したため一時的にロックしている"
  LOCKED
  "現在のパスワードが違う"
  WRONG_CURRENT_PASSWORD
  "再設定コードが存在しない・使用済み・期限切れ・別のユーザー用、またはユーザーが無効（区別しない）"
  RESET_CODE_UNAVAILABLE
}
```

## ふるまい

| 操作 | ログイン | 成功したとき | 主な失敗 |
|---|---|---|---|
| `viewer` | 不要 | 有効なセッションなら Viewer、なければ null。エラーにはしない（FR-023） | なし |
| `register` | 不要 | ユーザーを作り、Cookie を設定し、`viewer` を返す | 入力の規則の失敗（項目ごと）、`INVITATION_UNAVAILABLE`（`INVITATION_CODE`）、`LOGIN_ID_TAKEN`（`LOGIN_ID`） |
| `login` | 不要 | 前のセッションを消して新しいセッションを作り、Cookie を設定する | `INVALID_CREDENTIALS`（`field` は null）、`LOCKED` と `lockedUntil` |
| `logout` | 不要 | セッションを消し、Cookie を消す（`Max-Age=0`）。常に `true` | なし |
| `updateDisplayName` | 必要 | 表示名を変え、`viewer` を返す | 表示名の規則の失敗 |
| `changePassword` | 必要 | パスワードを変え、すべてのセッションを消し、新しいセッションの Cookie を設定する | `WRONG_CURRENT_PASSWORD`、新しいパスワードの規則の失敗 |
| `resetPassword` | 不要 | パスワードを変え、すべてのセッションと失敗の記録を消す。`ok: true` | `RESET_CODE_UNAVAILABLE`（`RESET_CODE`）、`LOCKED` と `lockedUntil`、新しいパスワードの規則の失敗 |
| `issueInvitation` | 管理者 | 招待コードを作り、値と期限を返す | GraphQL のエラー（下の表） |

- 入力の規則の失敗は、ほかの確認（コードの照合など）の前に返す。ただし `register` では、
  規則に合っていても招待コードが使えなければ `INVITATION_UNAVAILABLE` だけを返し、ログイン ID が
  使われているかどうかは答えない（research R8）。
- `login` の入力のログイン ID が規則に合わないときも、`INVALID_CREDENTIALS` を返す（規則を
  教えて、存在しないことを知らせない）。

## GraphQL のエラー（`errors[].extensions.code`）

| コード | 場面 |
|---|---|
| `UNAUTHENTICATED` | ログインが必要な操作に、有効なセッションなしでアクセスした |
| `FORBIDDEN` | 管理者だけの操作に、管理者でないユーザーがアクセスした |

- 想定外のエラーの本文には、入力の値（パスワードやコード）を含めない（FR-010）。
- web は `UNAUTHENTICATED` を受け取ったら、ログイン画面に移動する（戻り先を付ける）。
