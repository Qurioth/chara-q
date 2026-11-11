import Link from "next/link";
import { Suspense, type ReactNode } from "react";

import { LogoMark } from "@/components/logo/logo-mark";

import { SheetsNavLink, SheetsNavLinkView } from "./sheets-nav-link";

type SiteHeaderProps = {
  /** 右側の操作の領域の先頭に置く、アイコンだけの補助の操作（テーマの切り替えなど） */
  utilities?: ReactNode;
  /** 右側の操作の領域の末尾（右端）に置く操作（ログインの状態など） */
  actions?: ReactNode;
};

/**
 * 共通のヘッダー。左端はアプリ名、右端は操作の領域。
 * 操作の領域は「補助の操作（テーマ）・シートの一覧への入口・アカウント」の順に並べる。
 * アイコンだけの操作を、アイコンと文字の付いた操作の間に挟まないようにするため
 * （specs/003-character-sheet-base の FR-034、research R16）。
 */
export function SiteHeader({ utilities, actions }: SiteHeaderProps) {
  return (
    <header className="border-b border-border">
      <div className="mx-auto flex h-14 w-full max-w-5xl items-center justify-between gap-2 px-4">
        <Link
          href="/"
          className="flex min-w-0 items-center gap-2 rounded-md font-logo text-[22px] tracking-[0.04em] outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
        >
          <LogoMark className="size-8 shrink-0 text-seal" />
          <span className="truncate">CharaQ</span>
        </Link>
        <div className="flex shrink-0 items-center gap-1">
          {utilities}
          <nav aria-label="メイン">
            {/* 現在のパスはリクエストごとに変わるため、外枠とは別にストリーミングする（cacheComponents） */}
            <Suspense fallback={<SheetsNavLinkView current={false} />}>
              <SheetsNavLink />
            </Suspense>
          </nav>
          {actions}
        </div>
      </div>
    </header>
  );
}
