import Link from "next/link";
import type { ReactNode } from "react";

import { LogoMark } from "@/components/logo/logo-mark";

type SiteHeaderProps = {
  /** 右端の操作の領域（テーマの切り替え、後の機能でログインの状態など） */
  actions?: ReactNode;
};

export function SiteHeader({ actions }: SiteHeaderProps) {
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
        <div className="flex shrink-0 items-center gap-1">{actions}</div>
      </div>
    </header>
  );
}
