import { render } from "@testing-library/react";
import type { ReactElement } from "react";
import { vi } from "vitest";

import { UrqlProvider } from "@/lib/urql-provider";

export type GraphQLRequest = {
  query: string;
  variables?: Record<string, unknown>;
  operationName?: string;
};

type Handler = (req: GraphQLRequest) => unknown;

/**
 * ブラウザから /graphql への fetch を差し替える。handler は GraphQL の応答の本文（{ data, errors }）を返す。
 * 本物の urql の Client を使うため、送るリクエストの形も確かめられる。
 */
export function mockGraphQL(handler: Handler) {
  const fetchMock = vi.fn(async (url: string, init?: RequestInit) => {
    if (url !== "/graphql") {
      throw new Error(`想定外の URL: ${url}`);
    }
    const body = handler(JSON.parse(String(init?.body)) as GraphQLRequest);
    return new Response(JSON.stringify(body), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

/** 送られたリクエストを取り出す */
export function requestsOf(fetchMock: ReturnType<typeof mockGraphQL>): GraphQLRequest[] {
  return fetchMock.mock.calls.map(([, init]) => JSON.parse(String(init?.body)) as GraphQLRequest);
}

/** urql の Provider で包んで描画する */
export function renderWithUrql(ui: ReactElement) {
  return render(<UrqlProvider>{ui}</UrqlProvider>);
}
