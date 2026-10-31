type Scheme = "light" | "dark";

/**
 * window.matchMedia の偽物を入れる。`(prefers-color-scheme: dark)` だけを端末の設定として扱い、
 * setScheme で端末の設定を変えると、購読しているリスナーに change を送る。
 */
export function mockColorScheme(initial: Scheme) {
  let scheme = initial;
  const listeners = new Set<(event: MediaQueryListEvent) => void>();

  const matches = (query: string) =>
    query.includes("prefers-color-scheme: dark") ? scheme === "dark" : false;

  window.matchMedia = (query: string) =>
    ({
      media: query,
      get matches() {
        return matches(query);
      },
      onchange: null,
      addEventListener: (_: string, listener: (event: MediaQueryListEvent) => void) =>
        listeners.add(listener),
      removeEventListener: (_: string, listener: (event: MediaQueryListEvent) => void) =>
        listeners.delete(listener),
      addListener: (listener: (event: MediaQueryListEvent) => void) => listeners.add(listener),
      removeListener: (listener: (event: MediaQueryListEvent) => void) =>
        listeners.delete(listener),
      dispatchEvent: () => true,
    }) as MediaQueryList;

  return {
    setScheme(next: Scheme) {
      scheme = next;
      for (const listener of listeners) {
        listener({ matches: next === "dark", media: "(prefers-color-scheme: dark)" } as MediaQueryListEvent);
      }
    },
  };
}
