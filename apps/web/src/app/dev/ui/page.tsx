import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { GameSystemScope } from "@/components/game-system/game-system-scope";

import { FormPreview } from "./form-preview";
import { HealthStatus } from "./health-status";

// 開発時だけ使う部品の確認ページ。本番のビルドでは、タイトルを含めて常に 404 にする（FR-032）
function notFoundInProduction() {
  if (process.env.NODE_ENV === "production") {
    notFound();
  }
}

export function generateMetadata(): Metadata {
  notFoundInProduction();
  return { title: "部品の確認", robots: { index: false } };
}

export default function UiPreviewPage() {
  notFoundInProduction();

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-4 py-8">
      <h1 className="text-2xl font-bold">部品の確認</h1>

      <section aria-labelledby="form-heading" className="flex flex-col gap-4">
        <h2 id="form-heading" className="text-lg font-bold">
          入力部品
        </h2>
        <FormPreview />
      </section>

      <section aria-labelledby="font-heading" className="flex flex-col gap-4">
        <h2 id="font-heading" className="text-lg font-bold">
          ゲームシステムのフォント
        </h2>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <FontSample title="領域の外（共通フォント）" />
          <GameSystemScope system="yuyakekoyake">
            <FontSample title="ゆうやけこやけ" />
          </GameSystemScope>
          <GameSystemScope system="beginning-idol">
            <FontSample title="ビギニングアイドル" />
          </GameSystemScope>
        </div>
      </section>

      <section aria-labelledby="health-heading" className="flex flex-col gap-2">
        <h2 id="health-heading" className="text-lg font-bold">
          API の接続状態
        </h2>
        <HealthStatus />
      </section>
    </div>
  );
}

function FontSample({ title }: { title: string }) {
  return (
    <div className="flex flex-col gap-2 rounded-lg border border-border p-4">
      <p className="text-xs text-muted-foreground">{title}</p>
      <p className="font-heading text-xl">見出し: 能力値とつながり</p>
      <p className="font-brush text-xl">筆文字: 夕焼け小焼けの赤とんぼ</p>
      <p className="font-brush">旧字体と記号: 國學 ♪ ★ ①</p>
    </div>
  );
}
