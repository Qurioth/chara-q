"use client";

import { useMemo, type ReactNode } from "react";
import { Client, Provider, cacheExchange, fetchExchange } from "urql";

// ブラウザは web とだけ通信する。/graphql は next.config.ts の rewrites で api に中継される（ADR 0006）
const apiUrl = "/graphql";

export function UrqlProvider({ children }: { children: ReactNode }) {
  const client = useMemo(
    () => new Client({ url: apiUrl, exchanges: [cacheExchange, fetchExchange] }),
    [],
  );
  return <Provider value={client}>{children}</Provider>;
}
