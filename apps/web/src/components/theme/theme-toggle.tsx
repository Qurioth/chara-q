"use client";

import { MoonIcon, SunIcon } from "lucide-react";
import { useSyncExternalStore } from "react";

import { Button } from "@/components/ui/button";

import { getDisplayedTheme, subscribeTheme, toggleTheme } from "./theme";

const labels = {
  light: "ダークモードに切り替える",
  dark: "ライトモードに切り替える",
} as const;

export function ThemeToggle({ className }: { className?: string }) {
  // サーバーでは表示中のテーマがわからないため null にし、ハイドレーション後に確定する
  const theme = useSyncExternalStore(subscribeTheme, getDisplayedTheme, () => null);

  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      className={className}
      aria-label={theme ? labels[theme] : "テーマを切り替える"}
      onClick={() => toggleTheme()}
    >
      {/* アイコンは CSS で出し分けるため、ハイドレーション前もちらつかない */}
      <MoonIcon aria-hidden className="dark:hidden" />
      <SunIcon aria-hidden className="hidden dark:block" />
    </Button>
  );
}
