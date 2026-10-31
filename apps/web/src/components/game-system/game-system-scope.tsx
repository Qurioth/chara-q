import type { ComponentPropsWithoutRef, ReactNode } from "react";

import type { GameSystemId } from "@/lib/game-systems";

type GameSystemScopeProps = {
  system: GameSystemId;
  as?: "div" | "section" | "article";
  className?: string;
  children: ReactNode;
} & Omit<ComponentPropsWithoutRef<"div">, "children" | "className">;

/**
 * 中の要素をゲームシステムの領域にする。font-heading・font-brush が、そのシステムのフォントになる
 * （globals.css の [data-game-system]）。
 */
export function GameSystemScope({
  system,
  as: Component = "div",
  className,
  children,
  ...props
}: GameSystemScopeProps) {
  return (
    <Component {...props} data-game-system={system} className={className}>
      {children}
    </Component>
  );
}
