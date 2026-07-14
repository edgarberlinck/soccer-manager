export default function Footer() {
  const year = new Date().getFullYear();

  return (
    <footer className="mt-20 border-t border-[var(--line)] px-4 pb-14 pt-10 text-[var(--sea-ink-soft)]">
      <div className="page-wrap flex flex-col items-center justify-between gap-4 text-center sm:flex-row sm:text-left">
        <p className="m-0 text-sm">
          &copy; {year} Soccer Manager. All rights reserved.
        </p>
        <p className="island-kicker m-0">Football strategy, gamer vibe.</p>
      </div>
      <div className="mt-4 flex justify-center gap-4">
        <a
          href="/signup"
          target="_self"
          rel="noreferrer"
          className="rounded-xl p-2 text-[var(--sea-ink-soft)] transition hover:bg-[var(--link-bg-hover)] hover:text-[var(--sea-ink)]"
        >
          <span className="sr-only">Create your club</span>
          <svg viewBox="0 0 16 16" aria-hidden="true" width="32" height="32">
            <path
              fill="currentColor"
              d="M8 0L2 2.5v6.4c0 3.6 2.6 6.9 6 7.1 3.4-.2 6-3.5 6-7.1V2.5L8 0zm0 2.2l4 1.7v5c0 2.5-1.7 4.9-4 5.2-2.3-.3-4-2.7-4-5.2v-5l4-1.7z"
            />
          </svg>
        </a>
        <a
          href="/dashboard"
          target="_self"
          rel="noreferrer"
          className="rounded-xl p-2 text-[var(--sea-ink-soft)] transition hover:bg-[var(--link-bg-hover)] hover:text-[var(--sea-ink)]"
        >
          <span className="sr-only">Open dashboard</span>
          <svg viewBox="0 0 16 16" aria-hidden="true" width="32" height="32">
            <path
              fill="currentColor"
              d="M8 1.2a6.8 6.8 0 1 0 0 13.6A6.8 6.8 0 0 0 8 1.2zm0 1.6A5.2 5.2 0 1 1 8 13.2 5.2 5.2 0 0 1 8 2.8zm0 1.3a3.9 3.9 0 1 0 0 7.8 3.9 3.9 0 0 0 0-7.8z"
            />
          </svg>
        </a>
      </div>
    </footer>
  );
}
