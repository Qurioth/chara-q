"use client";

import { useQuery } from "urql";

import { graphql } from "@/gql";

const HealthQuery = graphql(`
  query Health {
    health
  }
`);

// API の health をブラウザから問い合わせて表示する（疎通確認用）
export function HealthStatus() {
  const [{ data, error }] = useQuery({
    query: HealthQuery,
    // サーバー側のレンダリングでは問い合わせない
    pause: typeof window === "undefined",
  });

  if (error) {
    return <p role="alert">API: エラー（{error.message}）</p>;
  }
  if (!data) {
    return <p>API: 確認中…</p>;
  }
  return <p>API: {data.health}</p>;
}
