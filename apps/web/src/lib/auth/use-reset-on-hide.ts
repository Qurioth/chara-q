"use client";

import { useLayoutEffect, useRef } from "react";

/**
 * 画面を離れたときに reset を呼ぶ。
 * cacheComponents では、離れた画面は <Activity> で隠されるだけで、入力欄の値や表示した内容が DOM に残る。
 * パスワード・招待コード・再設定コードを残さず、古い失敗の文言も見せないようにする
 * （apps/web/node_modules/next/dist/docs/01-app/02-guides/preserving-ui-state.md）。
 */
export function useResetOnHide(reset: () => void) {
  const resetRef = useRef(reset);
  useLayoutEffect(() => {
    resetRef.current = reset;
  });
  // Activity が隠すときにも、アンマウントと同じく cleanup が同期的に走る
  useLayoutEffect(() => () => resetRef.current(), []);
}
