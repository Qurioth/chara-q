"use client";

import { ChevronDownIcon, LogInIcon } from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useMutation } from "urql";

import { buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLinkItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { graphql } from "@/gql";
import { cn } from "@/lib/utils";

const LogoutMutation = graphql(`
  mutation Logout {
    logout
  }
`);

/** ヘッダーに渡す本人の情報。ログイン ID は渡さない（他の画面に出さないため） */
export type AccountMenuViewer = {
  displayName: string;
  isAdmin: boolean;
};

/** ヘッダーの右端のログイン状態（specs/002-auth/contracts/web-routes.md） */
export function AccountMenu({ viewer }: { viewer: AccountMenuViewer | null }) {
  const pathname = usePathname();
  const router = useRouter();
  const [, logout] = useMutation(LogoutMutation);

  if (!viewer) {
    return (
      <Link
        href={`/login?next=${encodeURIComponent(pathname)}`}
        className={buttonVariants({ variant: "ghost" })}
      >
        <LogInIcon aria-hidden />
        ログイン
      </Link>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label={`アカウントのメニュー（${viewer.displayName}）`}
        className={cn(buttonVariants({ variant: "ghost" }), "max-w-32 sm:max-w-48")}
      >
        <span className="truncate">{viewer.displayName}</span>
        <ChevronDownIcon aria-hidden />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-auto min-w-44">
        <DropdownMenuLinkItem render={<Link href="/account" />}>アカウントの設定</DropdownMenuLinkItem>
        {viewer.isAdmin && (
          <DropdownMenuLinkItem render={<Link href="/admin/invitations" />}>
            招待コードの発行
          </DropdownMenuLinkItem>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={async () => {
            await logout({});
            // ヘッダー（Server Components）を描き直す。ログインが必要な画面なら、その画面がログイン画面へ移す
            router.refresh();
          }}
        >
          ログアウト
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
