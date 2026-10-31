import type { Metadata } from "next";

import { SiteFooter } from "@/components/layout/site-footer";
import { SiteHeader } from "@/components/layout/site-header";
import { SkipLink } from "@/components/layout/skip-link";
import { themeScript } from "@/components/theme/theme-script";
import { ThemeToggle } from "@/components/theme/theme-toggle";
import { fontVariables } from "@/lib/fonts";
import { UrqlProvider } from "@/lib/urql-provider";

import "./globals.css";

export const metadata: Metadata = {
  title: {
    template: "%s | CharaQ",
    default: "CharaQ",
  },
  description:
    "TRPGのキャラクターをつくろう。キャラクターシートを作成するシステムを選択してください。",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    // テーマのスクリプトが描画の前に class と style を変えるため、ハイドレーションの差分を許す
    <html lang="ja" className={fontVariables} suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body className="flex min-h-dvh flex-col">
        <SkipLink />
        <SiteHeader actions={<ThemeToggle />} />
        <main id="main-content" tabIndex={-1} className="flex-1 outline-none">
          <UrqlProvider>{children}</UrqlProvider>
        </main>
        <SiteFooter />
      </body>
    </html>
  );
}
