import "server-only";

import type { TypedDocumentNode } from "@graphql-typed-document-node/core";
import { print } from "graphql";
import { cookies } from "next/headers";

/** api が発行するセッションの Cookie の名前（apps/api/internal/auth/cookie.go と同じ） */
export const SESSION_COOKIE = "__Host-chara_q_session";

type GraphQLResponse<T> = {
  data?: T | null;
  errors?: { message: string; extensions?: { code?: string } }[];
};

/**
 * Server Components から api の GraphQL を呼ぶ（specs/002-auth の research R3）。
 * ブラウザから受け取った Cookie のうち、セッションの Cookie だけを付けて api に直接送る。
 * セッションの Cookie がなければ api を呼ばずに null を返す。
 * api に接続できない・GraphQL のエラーのときは例外を投げる。
 */
export async function queryWithSession<TResult, TVariables extends Record<string, unknown>>(
  document: TypedDocumentNode<TResult, TVariables>,
  variables?: TVariables,
): Promise<TResult | null> {
  const session = (await cookies()).get(SESSION_COOKIE)?.value;
  if (!session) {
    return null;
  }

  const res = await fetch(`${process.env.API_URL}/graphql`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Cookie: `${SESSION_COOKIE}=${session}`,
    },
    body: JSON.stringify({ query: print(document), variables }),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`api の呼び出しに失敗しました（${res.status}）`);
  }
  const body = (await res.json()) as GraphQLResponse<TResult>;
  if (body.errors?.length) {
    throw new Error(`api の GraphQL のエラー: ${body.errors.map((e) => e.message).join(", ")}`);
  }
  return body.data ?? null;
}
