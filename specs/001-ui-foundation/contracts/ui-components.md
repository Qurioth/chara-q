# Contract: 共通部品

後の機能（認証、シート）の画面が使う部品の外から見た形とふるまい。props の型は実装時に TypeScript で
定義し、ここに書いた名前と意味を変えない。変えるときはこのファイルも直す。

## 共通レイアウト（ルートレイアウトが配置する）

画面側は、ページの中身だけを書く。

```text
<html lang="ja" class="[dark] [フォントの変数のクラス]">
  <body class="min-h-dvh flex flex-col">
    <SkipLink />        先頭。フォーカスしたときだけ見える「本文へスキップ」→ #main-content
    <SiteHeader />      アプリ名「CharaQ」（ロゴの書体 font-logo、トップページへのリンク）、右端に操作の領域（ThemeToggle）
    <main id="main-content" tabIndex={-1} class="flex-1">{children}</main>
    <SiteFooter />      アプリ名「CharaQ」と「© 2026 Qurioth」
  </body>
</html>
```

- `metadata.title` は `{ template: "%s | CharaQ", default: "CharaQ" }`。各ページは `title` だけを書く（FR-007）。
- `SiteHeader` の右端の操作の領域は、後の機能でログインの状態とメニューを並べる場所にする。幅が足りない
  ときはアプリ名を省略表示にし、操作の領域は縮めない（Edge Cases）。

## ThemeToggle

| prop | 型 | 説明 |
|---|---|---|
| `className` | string? | 位置の調整用 |

- `<button type="button">`。押すと表示中のテーマの反対を保存し、画面に反映する。
- 読み上げ用の名前は「ダークモードに切り替える」/「ライトモードに切り替える」（押した後のテーマ）。
  ハイドレーション前は「テーマを切り替える」。
- アイコン（月/太陽）は CSS の `dark:` で出し分け、`aria-hidden`。

## GameSystemScope

| prop | 型 | 説明 |
|---|---|---|
| `system` | `GameSystemId` | `"yuyakekoyake"` \| `"beginning-idol"` |
| `as` | `"div"` \| `"section"` \| `"article"` | 既定は `"div"` |
| `className` / `children` | | |

- 描画するのは `<{as} data-game-system={system}>` だけ。中の要素は `font-heading`・`font-brush` で役割の
  フォントになる（[theme-and-fonts.md](theme-and-fonts.md)）。
- 入れ子にしたときは、内側の指定が優先される。

## 入力部品の共通の枠（FormField）

4 つの部品は、次の共通の props を持ち、同じ構造で描画する。

| prop | 型 | 既定 | 説明 |
|---|---|---|---|
| `label` | ReactNode | 必須 | ラベル |
| `hideLabel` | boolean | `false` | ラベルを見た目だけ隠す（シートの表のマス目など）。支援技術には読ませる |
| `description` | ReactNode? | | 補足説明 |
| `error` | string? | | エラー文言 |
| `required` | boolean | `false` | 必須 |
| `disabled` | boolean | `false` | 無効 |
| `id` | string? | 自動 | 入力欄の `id`。省略時は `useId()` で採番 |
| `name` | string? | | フォームのライブラリとつなぐときに使う |
| `className` | string? | | 外枠（ラベル・入力欄・説明・エラーを含む）の見た目 |
| `inputClassName` | string? | | 入力欄そのものの見た目（大きな文字、マス目に合わせた幅など。FR-031） |

**描画とアクセシビリティの決まり**

- ラベルは `<label for={id}>`。`required` のときは、ラベルの後ろに「必須」の文字を付ける（色だけに頼らない）。
- 入力欄に付ける属性:
  - `aria-describedby`: 補足説明の `id` と、エラーがあればエラーの `id`（この順、空白区切り）
  - `aria-invalid="true"`: エラーがあるとき
  - `aria-required="true"`: `required` のとき
  - `disabled`: `disabled` のとき
- エラーは入力欄の下に表示し、先頭にエラーのアイコンと、`destructive` の色を使う。文言そのものがエラーだと
  わかる書き方にする（色だけに頼らない）。エラーが出たときに読み上げられるよう、エラーの領域は
  `role="alert"` を付ける。
- フォーカスは `ring` の色の 2px 以上の輪郭で示す（両テーマ）。

## TextField

| prop | 型 | 説明 |
|---|---|---|
| `value` | string | |
| `onValueChange` | `(value: string) => void` | |
| `onBlur` | `() => void`? | |
| `type` | `"text"` \| `"password"` \| `"search"` | 既定 `"text"`（`"password"` は認証の機能で使う） |
| `maxLength` / `placeholder` / `autoComplete` | | そのまま入力欄へ渡す |

## NumberField

| prop | 型 | 説明 |
|---|---|---|
| `value` | `number \| null` | `null` は未入力 |
| `onValueChange` | `(value: number \| null) => void` | 数値として読める入力のたびに呼ぶ。範囲外・許可外の値も渡す（部品はエラーを表示するだけで、値を捨てない） |
| `onBlur` | `() => void`? | |
| `min` / `max` | number? | |
| `allowNegative` | boolean | 既定 `false` |
| `allowDecimal` | boolean | 既定 `false` |
| `placeholder` | string? | |

- 入力欄は `type="text"`、`inputMode` は `allowDecimal` なら `"decimal"`、それ以外は `"numeric"`。
- 入力の変換・拒否・blur 時の検証の決まりは [data-model.md](../data-model.md) の「NumberField の値」。
- 画面側の `value` が外から変わったときは、入力欄の文字列もその値に合わせる（入力中の途中の形は除く）。

## SelectField

| prop | 型 | 説明 |
|---|---|---|
| `value` | `string \| null` | `null` は未選択 |
| `onValueChange` | `(value: string \| null) => void` | |
| `options` | `{ value: string; label: ReactNode; disabled?: boolean }[]` | |
| `placeholder` | string? | 未選択のときの表示。既定は「選択してください」 |

- shadcn/ui（Base UI）の Select を使う。キーボード: フォーカスして Enter・Space・上下キーで開く、上下キーで
  移動、Enter で決定、Escape で閉じる。閉じたらフォーカスはトリガーへ戻る。
- 選択肢の一覧は画面の高さを超えないようにし、超える分はスクロールする（Edge Cases）。
- 選択後の表示は 1 行で、入力欄の幅を超える分は省略記号で切る。

## TextareaField

| prop | 型 | 説明 |
|---|---|---|
| `value` | string | 改行を含む |
| `onValueChange` | `(value: string) => void` | |
| `onBlur` | `() => void`? | |
| `rows` | number? | 既定 3 |
| `maxLength` / `placeholder` | | |

- 改行は `\n` のまま保つ（FR-030）。

## 部品確認ページ（`/dev/ui`）

- 本番のビルド（`NODE_ENV === "production"`）では 404（FR-032）。
- 表示するもの: 4 つの入力部品の通常・エラー・無効・必須・ラベルを隠した状態、`GameSystemScope` で囲んだ
  見出しと筆文字（両システムと、囲まない場合）、API の接続状態（既存の `HealthStatus`）。
- テーマはヘッダーのトグルで切り替えて確かめる。
