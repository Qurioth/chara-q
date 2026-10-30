import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

import { UrqlProvider } from "@/lib/urql-provider";

import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "chara-q",
  description: "TRPG キャラクターシート管理",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="ja" className={`${geistSans.variable} ${geistMono.variable}`}>
      <body>
        <UrqlProvider>{children}</UrqlProvider>
      </body>
    </html>
  );
}
