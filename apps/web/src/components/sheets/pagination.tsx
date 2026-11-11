import Link from "next/link";

type PaginationProps = {
  page: number;
  totalPages: number;
  /** ページ番号から、そのページの URL を作る（ほかの条件は保つ） */
  hrefFor: (page: number) => string;
};

/**
 * 表示するページ番号。最初・最後・現在の前後 2 ページを残し、間は "gap" にする。
 * 間が 1 ページだけなら、"gap" ではなくそのページ番号を出す（「…」が 1 つの番号より長くならないように）。
 */
export function pageItems(page: number, totalPages: number): (number | "gap")[] {
  const shown = new Set<number>([1, totalPages]);
  for (let p = page - 2; p <= page + 2; p++) {
    if (p >= 1 && p <= totalPages) shown.add(p);
  }
  const sorted = [...shown].sort((a, b) => a - b);
  const items: (number | "gap")[] = [];
  sorted.forEach((p, i) => {
    const prev = sorted[i - 1];
    if (i > 0 && p - prev === 2) items.push(prev + 1);
    if (i > 0 && p - prev > 2) items.push("gap");
    items.push(p);
  });
  return items;
}

const linkClass =
  "inline-flex h-9 min-w-9 items-center justify-center rounded-md border border-border px-3 text-sm outline-none hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50";

/** 一覧のページ送り。1 ページだけなら何も表示しない */
export function Pagination({ page, totalPages, hrefFor }: PaginationProps) {
  if (totalPages <= 1) {
    return null;
  }
  return (
    <nav aria-label="ページ">
      <ul className="flex flex-wrap items-center justify-center gap-2">
        {page > 1 && (
          <li>
            <Link href={hrefFor(page - 1)} className={linkClass}>
              前へ
            </Link>
          </li>
        )}
        {pageItems(page, totalPages).map((item, i) =>
          item === "gap" ? (
            <li key={`gap-${i}`} aria-hidden className="px-1 text-muted-foreground">
              …
            </li>
          ) : (
            <li key={item}>
              <Link
                href={hrefFor(item)}
                aria-current={item === page ? "page" : undefined}
                aria-label={`${item} ページ目`}
                className={`${linkClass} ${item === page ? "bg-primary text-primary-foreground hover:bg-primary" : ""}`}
              >
                {item}
              </Link>
            </li>
          ),
        )}
        {page < totalPages && (
          <li>
            <Link href={hrefFor(page + 1)} className={linkClass}>
              次へ
            </Link>
          </li>
        )}
      </ul>
    </nav>
  );
}
