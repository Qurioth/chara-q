// シートの画面で使うデータの形（GraphQL の結果から作る）。
import type { GameSystemId } from "@/lib/game-systems";

/** 閲覧画面の共通部分に使うシート */
export type SheetDetail = {
  id: string;
  system: GameSystemId;
  characterName: string;
  playerName: string;
  age: string;
  sex: string;
  memo: string;
  revision: number;
  createdAt: string;
  updatedAt: string;
  owner: { id: string; displayName: string };
  /** システム固有の部分（JSON のオブジェクト）。中身は各システムの Spec が決める */
  data: unknown;
};

/** 一覧の 1 行 */
export type SheetSummary = {
  id: string;
  system: GameSystemId;
  characterName: string;
  playerName: string;
  updatedAt: string;
  owner: { displayName: string };
};
