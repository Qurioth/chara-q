import { unstable_rethrow } from "next/navigation";

import { getViewer } from "@/lib/auth/viewer";

import { AccountMenu } from "./account-menu";

/**
 * ヘッダーのログイン状態を読み込む前・読み込めなかったときに出す領域。
 * ログインの操作は出さない（ログインしているのにログインしていない表示が一瞬出ないようにする。SC-009）。
 */
export function AccountAreaFallback() {
  return <div aria-hidden className="h-8 w-24" />;
}

/** ヘッダーのログイン状態。cookies() を読むため <Suspense> の中に置く */
export async function AccountArea() {
  let viewer;
  try {
    viewer = await getViewer();
  } catch (error) {
    // cookies() などが Next.js の制御のために投げるエラーは、そのまま Next.js に返す
    unstable_rethrow(error);
    // api に接続できないときは、ログインしていない扱いにせず、ログインの操作も出さない
    console.error("ログイン状態を読めません", error);
    return <AccountAreaFallback />;
  }
  return (
    <AccountMenu viewer={viewer ? { displayName: viewer.displayName, isAdmin: viewer.isAdmin } : null} />
  );
}
