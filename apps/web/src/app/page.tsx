import Link from "next/link";

import { gameSystems, type GameSystem } from "@/lib/game-systems";

export default function Home() {
  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-8 px-4 py-8">
      <section className="flex flex-col gap-3">
        <h1 className="text-2xl font-bold">TRPGのキャラクターをつくろう</h1>
        <p className="leading-relaxed text-muted-foreground">
          キャラクターシートを作成するシステムを選択してください。
        </p>
      </section>

      <section
        aria-labelledby="systems-heading"
        className="flex flex-col gap-4"
      >
        <h2 id="systems-heading" className="text-xl font-bold">
          対応予定のゲームシステム
        </h2>
        <ul
          aria-labelledby="systems-heading"
          className="grid grid-cols-1 gap-4 sm:grid-cols-2"
        >
          {gameSystems.map((system) => (
            <li key={system.id}>
              <GameSystemCard system={system} />
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}

function GameSystemCard({ system }: { system: GameSystem }) {
  return (
    <div className="flex h-full flex-col gap-2 rounded-lg border border-border bg-card p-4 text-card-foreground">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <h3 className="font-bold">
          {system.status === "available" ? (
            <Link
              href={system.href}
              className="underline-offset-4 hover:underline"
            >
              {system.name}
            </Link>
          ) : (
            system.name
          )}
        </h3>
        {system.status === "coming-soon" && (
          <span className="shrink-0 rounded-md bg-muted px-2 py-0.5 text-xs font-medium text-foreground">
            準備中
          </span>
        )}
      </div>
      <p className="text-sm leading-relaxed text-muted-foreground">
        {system.description}
      </p>
    </div>
  );
}
