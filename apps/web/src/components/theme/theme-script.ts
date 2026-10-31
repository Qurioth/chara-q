import { THEME_STORAGE_KEY } from "./theme";

/**
 * <head> で最初の描画の前に実行するスクリプト。theme.ts と同じ決まりで html の dark / light クラスと
 * color-scheme を設定し、保存値がない間だけ端末の設定の変化に追従する。
 * React より先に動くため、theme.ts を読み込めない。決まりを変えるときは両方を直す。
 */
export const themeScript = `(function () {
  try {
    var key = ${JSON.stringify(THEME_STORAGE_KEY)};
    var html = document.documentElement;
    var media = window.matchMedia("(prefers-color-scheme: dark)");
    var stored = function () {
      try {
        var value = localStorage.getItem(key);
        return value === "light" || value === "dark" ? value : null;
      } catch (e) {
        return null;
      }
    };
    var apply = function (theme) {
      html.classList.toggle("dark", theme === "dark");
      html.classList.toggle("light", theme === "light");
      html.style.colorScheme = theme;
    };
    var theme = stored();
    apply(theme || (media.matches ? "dark" : "light"));
    if (!theme) {
      media.addEventListener("change", function (event) {
        if (!stored()) apply(event.matches ? "dark" : "light");
      });
    }
  } catch (e) {}
})();`;
