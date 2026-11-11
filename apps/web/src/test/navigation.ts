import { vi } from "vitest";

/**
 * next/navigation の偽物。テストのファイルで次のように使う:
 *
 *   vi.mock("next/navigation", () => import("@/test/navigation"));
 *
 * 呼び出しの記録は beforeEach で vi.clearAllMocks() を呼んで消す。
 */
export const router = {
  push: vi.fn(),
  replace: vi.fn(),
  refresh: vi.fn(),
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
};

export const navigation = { pathname: "/" };

export function useRouter() {
  return router;
}

export function usePathname() {
  return navigation.pathname;
}
