import type { ReactNode } from "react";

/** 認証の画面（ログイン、登録など）の共通の枠 */
export function AuthPage({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="mx-auto flex w-full max-w-sm flex-col gap-6 px-4 py-8">
      <h1 className="text-2xl font-bold">{title}</h1>
      {children}
    </div>
  );
}
