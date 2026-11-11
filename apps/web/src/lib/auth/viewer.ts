import "server-only";

import { notFound, redirect } from "next/navigation";
import { cache } from "react";

import { graphql } from "@/gql";
import { queryWithSession } from "@/lib/api/server";

const ViewerQuery = graphql(`
  query Viewer {
    viewer {
      id
      loginId
      displayName
      isAdmin
    }
  }
`);

/**
 * ログインしている本人を返す。ログインしていなければ null。
 * ヘッダーと画面の両方から呼ぶため、1 回のリクエストの中では api の呼び出しを 1 回にまとめる。
 * cookies() を読むため、呼び出す側は <Suspense> の中に置く（cacheComponents）。
 */
export const getViewer = cache(async () => {
  const data = await queryWithSession(ViewerQuery);
  return data?.viewer ?? null;
});

function loginPath(path: string) {
  return `/login?next=${encodeURIComponent(path)}`;
}

/** ログインが必要な画面で使う。ログインしていなければログイン画面へ移動する */
export async function requireViewer(path: string) {
  const viewer = await getViewer();
  if (!viewer) {
    redirect(loginPath(path));
  }
  return viewer;
}

/** 管理者だけの画面で使う。ログインしていなければログイン画面へ、管理者でなければ 404 にする */
export async function requireAdmin(path: string) {
  const viewer = await requireViewer(path);
  if (!viewer.isAdmin) {
    notFound();
  }
  return viewer;
}
