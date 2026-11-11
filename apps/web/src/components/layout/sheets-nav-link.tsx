"use client";

import { ScrollTextIcon } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/**
 * ヘッダーのシートの一覧への入口（specs/003-character-sheet-base の FR-034、research R16）。
 * 幅の狭い画面ではアイコンだけにする。/sheets 以下を表示しているときは aria-current を付ける。
 * usePathname() を読むため、<Suspense> の中に置き、fallback には SheetsNavLinkView を使う（cacheComponents）。
 */
export function SheetsNavLink() {
  const pathname = usePathname();
  const current = pathname === "/sheets" || pathname?.startsWith("/sheets/") === true;
  return <SheetsNavLinkView current={current} />;
}

/** 現在のパスを読まない見た目だけのリンク */
export function SheetsNavLinkView({ current }: { current: boolean }) {
  return (
    <Link
      href="/sheets"
      aria-label="シート一覧"
      aria-current={current ? "page" : undefined}
      // ヘッダーの「ログイン」と同じ見た目にする。幅が狭い画面ではアイコンだけ
      className={cn(buttonVariants({ variant: "ghost" }), "max-sm:px-2 aria-[current=page]:font-bold")}
    >
      <ScrollTextIcon aria-hidden />
      <span className="hidden sm:inline">シート一覧</span>
    </Link>
  );
}
