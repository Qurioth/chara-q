import "server-only";

import { cache } from "react";

import { graphql } from "@/gql";
import { queryApi } from "@/lib/api/server";

import { toGameSystemId } from "./systems";
import type { SheetDetail } from "./types";

const CharacterSheetQuery = graphql(`
  query CharacterSheet($id: ID!) {
    characterSheet(id: $id) {
      __typename
      id
      system
      characterName
      playerName
      age
      sex
      memo
      revision
      createdAt
      updatedAt
      owner {
        id
        displayName
      }
      ... on YuyakekoyakeSheet {
        data
      }
      ... on BeginningIdolSheet {
        data
      }
    }
  }
`);

/**
 * シートを 1 枚読む。存在しない・削除された・形の正しくない ID なら null。
 * 閲覧の本体と generateMetadata で共有し、1 回のリクエストの中では api の呼び出しを 1 回にする。
 */
export const getCharacterSheet = cache(async (id: string): Promise<SheetDetail | null> => {
  const data = await queryApi(CharacterSheetQuery, { id });
  const s = data?.characterSheet;
  if (!s) {
    return null;
  }
  return {
    id: s.id,
    system: toGameSystemId(s.system),
    characterName: s.characterName,
    playerName: s.playerName,
    age: s.age,
    sex: s.sex,
    memo: s.memo,
    revision: s.revision,
    createdAt: s.createdAt,
    updatedAt: s.updatedAt,
    owner: s.owner,
    data: s.data,
  };
});
