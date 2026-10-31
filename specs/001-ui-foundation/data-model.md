# Data Model: UI 基盤

この機能は DB と GraphQL スキーマを変えない。ここでは、web の中で扱う値の形と決まりを書く。

## ThemePreference（テーマの選択）

利用者がブラウザごとに切り替えた表示テーマ。

| 項目 | 型 | 説明 |
|---|---|---|
| 保存値 | `"light"` \| `"dark"` \| なし | `localStorage` のキー `chara-q-theme`。一度も切り替えていなければキーがない |
| 表示中のテーマ | `"light"` \| `"dark"` | 保存値があればそれ、なければ端末の設定（`prefers-color-scheme`） |

**決まり**

- 保存値が `"light"`・`"dark"` 以外（手で書き換えられたなど）のときは、保存値がないものとして扱う。
- 保存先が使えない（読み書きで例外が出る）ときは、保存値がないものとして扱い、エラーは表示しない。
  切り替えはその画面の中では働く。
- サーバーにも Cookie にも送らない（FR-012）。

**状態の移り変わり**

```text
[未選択: 端末の設定に追従]
   │ トグルを押す（表示中の反対を保存）
   ▼
[選択済み: light] ⇄ トグルを押す ⇄ [選択済み: dark]
```

- 未選択の間は、端末の設定が変わると表示中のテーマも変わる（FR-015）。
- 一度選択済みになると、未選択には戻らない（clarify の回答）。端末の設定の変化は反映しない。

## GameSystem（ゲームシステム）

トップページの一覧と、システムごとのフォントの差し替えに使う定義。`src/lib/game-systems.ts` の定数で
持ち、DB には保存しない。

| 項目 | 型 | 説明 |
|---|---|---|
| `id` | `"yuyakekoyake"` \| `"beginning-idol"` | 識別子。`data-game-system` 属性の値と、後のシートの URL に使う |
| `name` | string | 表示名（例: 「ふしぎもののけRPG ゆうやけこやけ」） |
| `description` | string | トップページに出す紹介文 |
| `status` | `"available"` \| `"coming-soon"` | 対応状況。この機能では両方とも `"coming-soon"` |
| `href` | string \| なし | シート画面の入口。`status` が `"available"` のときだけ持つ |
| `fonts` | `Partial<Record<FontRole, FontId>>` | 役割ごとのフォント。持たない役割は共通フォント |

**FontRole（フォントの役割）**: `"heading"`（見出し）、`"brush"`（筆文字）

**FontId（フォント）**: `"mochiy-pop-p-one"`、`"yuji-syuku"`

**この機能での値**

| id | name | fonts |
|---|---|---|
| `yuyakekoyake` | ふしぎもののけRPG ゆうやけこやけ | heading: mochiy-pop-p-one、brush: yuji-syuku |
| `beginning-idol` | 駆け出しアイドルRPG ビギニングアイドル | heading: mochiy-pop-p-one |

**決まり**

- `status` が `"coming-soon"` のシステムは「準備中」と表示し、リンクにしない（FR-009）。
- `id` は小文字の英字とハイフンだけ。旧システムの URL（`/Yuyakekoyake`、`/BiginingIdol`）とは合わせない
  （旧システムの「Bigining」は綴りの誤り）。URL の互換はデータ移行の機能で扱う。
- フォントの差し替えは CSS で行う（[contracts/theme-and-fonts.md](contracts/theme-and-fonts.md)）。
  `fonts` の定義と CSS が食い違わないことを、テストで確かめる。

## FieldCommonProps（入力部品の共通属性）

4 種類の入力部品が共通して受け取る属性。詳細な props は [contracts/ui-components.md](contracts/ui-components.md)。

| 項目 | 型 | 説明 |
|---|---|---|
| `label` | ReactNode | ラベル。必須。見た目で隠す場合も支援技術には読ませる（`hideLabel`） |
| `description` | ReactNode \| なし | 補足説明 |
| `error` | string \| なし | エラー文言。あると入力欄は `aria-invalid="true"` |
| `required` | boolean | 必須。ラベルに「必須」の表示を付け、入力欄は `aria-required="true"` |
| `disabled` | boolean | 無効 |

## NumberField の値

| 項目 | 型 | 説明 |
|---|---|---|
| `value` | number \| null | `null` は未入力（0 とは区別する） |
| `min` / `max` | number \| なし | 範囲 |
| `allowNegative` | boolean | 既定は `false` |
| `allowDecimal` | boolean | 既定は `false` |

**入力の決まり**

1. 全角の数字・マイナス・ピリオド（`０-９`、`－`、`ー`、`．`）は半角に置き換える。
2. 置き換えた後、数字・`-`・`.` 以外の文字を含む入力は受け付けず、入力欄は直前の文字列のままにする
   （貼り付けも同じ）。
3. 入力中は、数値として読める文字列（例: `12`、`-3`、`1.5`）のときだけ `onValueChange` に数値を渡す。
   途中の形（`-`、`1.`）や空欄は、空欄なら `null` を渡し、途中の形なら値を変えない。
4. 入力欄を離れたとき（blur）に部品単位の検証を行い、内部のエラーを表示する。
   - `allowNegative` が `false` で負の数 → 「0 以上の数を入力してください」
   - `allowDecimal` が `false` で小数 → 「整数を入力してください」
   - `min` 未満・`max` 超過 → 「{min} 以上 {max} 以下で入力してください」（片方だけのときは片方の文言）
5. 画面側が `error` を渡したときは、内部のエラーより画面側のエラーを優先して表示する。
