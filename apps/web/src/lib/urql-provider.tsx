"use client";

import { useMemo, type ReactNode } from "react";
import { Client, Provider, cacheExchange, fetchExchange } from "urql";

const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/graphql";

export function UrqlProvider({ children }: { children: ReactNode }) {
  const client = useMemo(
    () => new Client({ url: apiUrl, exchanges: [cacheExchange, fetchExchange] }),
    [],
  );
  return <Provider value={client}>{children}</Provider>;
}
