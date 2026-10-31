import { Hina_Mincho, Mochiy_Pop_P_One, Yuji_Syuku } from "next/font/google";

// next/font の引数はビルド時に読まれるため、変数を使わずリテラルで書く。

// アプリ名（CharaQ）のロゴの書体。全画面のヘッダーに出るため、使う英字の範囲（latin）だけを先読みする。
const hinaMincho = Hina_Mincho({
  weight: "400",
  subsets: ["latin"],
  variable: "--font-hina-mincho",
  display: "swap",
  fallback: ["Hiragino Mincho ProN", "Yu Mincho", "serif"],
});

// ゲームシステム用のフォント。globals.css の [data-game-system] の中で役割（--font-heading など）に割り当てる。
// preload: false なので、その書体で描く文字が画面に出るまでダウンロードしない（FR-022）。
// 日本語の文字は Google Fonts 側で範囲ごとに分割されているため、subsets は指定しない。

const mochiyPopPOne = Mochiy_Pop_P_One({
  weight: "400",
  variable: "--font-mochiy-pop-p-one",
  preload: false,
  display: "swap",
  fallback: ["Hiragino Sans", "Yu Gothic UI", "Meiryo", "sans-serif"],
});

const yujiSyuku = Yuji_Syuku({
  weight: "400",
  variable: "--font-yuji-syuku",
  preload: false,
  display: "swap",
  fallback: ["Hiragino Sans", "Yu Gothic UI", "Meiryo", "sans-serif"],
});

/** html の className に付ける、フォントの CSS 変数のクラス */
export const fontVariables = [hinaMincho.variable, mochiyPopPOne.variable, yujiSyuku.variable].join(
  " ",
);
