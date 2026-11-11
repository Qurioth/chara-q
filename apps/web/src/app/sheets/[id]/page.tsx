import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { Suspense } from "react";

import { SheetCommonView } from "@/components/sheets/sheet-common-view";
import { getViewer } from "@/lib/auth/viewer";
import { getCharacterSheet } from "@/lib/sheets/get-sheet";

// シートの閲覧画面（specs/003-character-sheet-base の US2）。ログインなしで見られる。

export async function generateMetadata({ params }: PageProps<"/sheets/[id]">): Promise<Metadata> {
  const { id } = await params;
  const sheet = await getCharacterSheet(id);
  return { title: sheet ? sheet.characterName : "シートが見つかりません" };
}

export default function SheetPage({ params }: PageProps<"/sheets/[id]">) {
  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-8 px-4 py-8">
      {/* params と cookies() はリクエストごとに変わり、シートもキャッシュしないため <Suspense> の中で読む（cacheComponents） */}
      <Suspense>
        <SheetContent params={params} />
      </Suspense>
    </div>
  );
}

async function SheetContent({ params }: Pick<PageProps<"/sheets/[id]">, "params">) {
  const { id } = await params;
  const [sheet, viewer] = await Promise.all([getCharacterSheet(id), getViewer()]);
  if (!sheet) {
    notFound();
  }
  return <SheetCommonView sheet={sheet} isOwner={viewer?.id === sheet.owner.id} />;
}
