import type { Metadata } from "next";
import { Suspense } from "react";

import { Pagination } from "@/components/sheets/pagination";
import { SheetList } from "@/components/sheets/sheet-list";
import { SheetSearchForm } from "@/components/sheets/sheet-search-form";
import { graphql } from "@/gql";
import { queryApi } from "@/lib/api/server";
import { getViewer } from "@/lib/auth/viewer";
import {
  hasSheetConditions,
  parseSheetSearch,
  sheetSearchHref,
  toSheetOrder,
} from "@/lib/sheets/search-params";
import { toGameSystem, toGameSystemId } from "@/lib/sheets/systems";

// シートの一覧と検索（specs/003-character-sheet-base の US1）。ログインなしで見られる。

export const metadata: Metadata = { title: "キャラクターシート一覧" };

const CharacterSheetsQuery = graphql(`
  query CharacterSheets($filter: CharacterSheetFilter, $orderBy: CharacterSheetOrder!, $page: Int!) {
    characterSheets(filter: $filter, orderBy: $orderBy, page: $page) {
      nodes {
        id
        system
        characterName
        playerName
        updatedAt
        owner {
          displayName
        }
      }
      totalCount
      page
      totalPages
    }
  }
`);

export default function SheetsPage({ searchParams }: PageProps<"/sheets">) {
  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-4 py-8">
      <h1 className="text-2xl font-bold">キャラクターシート一覧</h1>
      {/* searchParams と cookies() はリクエストごとに変わり、一覧もキャッシュしないため <Suspense> の中で読む（cacheComponents） */}
      <Suspense>
        <SheetsContent searchParams={searchParams} />
      </Suspense>
    </div>
  );
}

async function SheetsContent({ searchParams }: Pick<PageProps<"/sheets">, "searchParams">) {
  const search = parseSheetSearch(await searchParams);
  const [viewer, data] = await Promise.all([
    getViewer(),
    queryApi(CharacterSheetsQuery, {
      filter: {
        system: search.system ? toGameSystem(search.system) : null,
        keyword: search.q || null,
        mine: search.mine,
      },
      orderBy: toSheetOrder(search.sort),
      page: search.page,
    }),
  ]);
  if (!data) {
    throw new Error("シートの一覧を読めません");
  }
  const { nodes, totalCount, page, totalPages } = data.characterSheets;
  // ログインしていなければ「自分のシートだけ」は無視される（api と同じ扱い）
  const effective = { ...search, mine: viewer ? search.mine : false, page };

  return (
    <>
      {/* 条件が変わったら（「条件を外す」など）、フォームの入力を URL の条件に戻す */}
      <SheetSearchForm key={sheetSearchHref({ ...effective, page: 1 })} search={effective} loggedIn={viewer !== null} />
      <SheetList
        sheets={nodes.map((n) => ({ ...n, system: toGameSystemId(n.system) }))}
        totalCount={totalCount}
        hasConditions={hasSheetConditions(effective)}
        sort={effective.sort}
        sortHref={(sort) => sheetSearchHref({ ...effective, sort, page: 1 })}
      />
      <Pagination page={page} totalPages={totalPages} hrefFor={(p) => sheetSearchHref({ ...effective, page: p })} />
    </>
  );
}
