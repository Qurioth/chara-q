import type { NextConfig } from "next";

// api のオリジン（例: http://localhost:8080）。ブラウザには出さないため NEXT_PUBLIC_ を付けない（ADR 0006）。
// rewrites はビルドのときに評価されるため、ビルドでも必要になる
const apiUrl = process.env.API_URL;
if (!apiUrl) {
  throw new Error("API_URL が設定されていません（例: http://localhost:8080）");
}
if (apiUrl.endsWith("/")) {
  throw new Error(`API_URL の末尾に / を付けないでください: ${apiUrl}`);
}

const nextConfig: NextConfig = {
  cacheComponents: true,
  partialPrefetching: true,
  // ブラウザは web とだけ通信し、/graphql は api に中継する（ADR 0006）
  async rewrites() {
    return [{ source: "/graphql", destination: `${apiUrl}/graphql` }];
  },
};

export default nextConfig;
