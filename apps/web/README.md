# @chara-q/web

chara-q の Web フロントエンド（Next.js App Router）。
セットアップとコマンドはリポジトリ直下の [README](../../README.md) と [CLAUDE.md](../../CLAUDE.md) を参照する。

- GraphQL の型は `task generate:web` で `src/gql/` に生成する（手で編集しない）
- api の URL は `.env` の `API_URL` で切り替える。ブラウザは `/graphql` を web に送り、`next.config.ts` の rewrites で api に中継する（ADR 0006）
