/** ログインの後に戻らない画面（ログインの画面に戻らないようにする） */
const authPaths = ["/login", "/register", "/reset-password"];

/**
 * ログインの後の戻り先（?next=）を検証する。同じアプリの中のパスだけを通し、
 * それ以外はトップページにする（外部の URL へのオープンリダイレクトを防ぐ）。
 */
export function safeNext(next: string | null | undefined): string {
  if (!next || !next.startsWith("/") || next.startsWith("//") || next.startsWith("/\\")) {
    return "/";
  }
  const path = next.split(/[?#]/)[0];
  if (authPaths.some((p) => path === p || path.startsWith(`${p}/`))) {
    return "/";
  }
  return next;
}
