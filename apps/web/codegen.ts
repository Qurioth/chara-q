import type { CodegenConfig } from "@graphql-codegen/cli";

// スキーマはリポジトリ直下の schema/graphql を直接読む（API の唯一の正）
const config: CodegenConfig = {
  schema: "../../schema/graphql/*.graphqls",
  documents: ["src/**/*.{ts,tsx}", "!src/gql/**/*"],
  ignoreNoDocuments: true,
  generates: {
    "./src/gql/": {
      preset: "client",
      config: {
        useTypeImports: true,
        enumsAsTypes: true,
        scalars: {
          JSON: "unknown",
          Time: "string",
        },
      },
    },
  },
};

export default config;
