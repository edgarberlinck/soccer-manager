import { Link, useNavigate } from "@tanstack/react-router";
import { useEffect, useMemo, useRef, useState } from "react";
import { getAuthToken, getSessionUser, logout } from "../lib/session";
import ThemeToggle from "./ThemeToggle";

export default function Header() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [loadingUser, setLoadingUser] = useState(true);
  const [userName, setUserName] = useState("");
  const navigate = useNavigate();
  const menuRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const loadUser = async () => {
      const token = getAuthToken();
      if (!token) {
        setLoadingUser(false);
        setUserName("");
        return;
      }

      const user = await getSessionUser();
      setUserName(user?.managerName || user?.email || "Manager");
      setLoadingUser(false);
    };

    void loadUser();
  }, []);

  useEffect(() => {
    const onClick = (event: MouseEvent) => {
      if (!menuRef.current) {
        return;
      }

      if (!menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    };

    document.addEventListener("mousedown", onClick);
    return () => {
      document.removeEventListener("mousedown", onClick);
    };
  }, []);

  const isLoggedIn = !loadingUser && userName !== "";
  const avatarLabel = useMemo(() => {
    const trimmed = userName.trim();
    if (!trimmed) {
      return "??";
    }

    const parts = trimmed.split(/\s+/).filter(Boolean);
    if (parts.length === 1) {
      return parts[0].slice(0, 2).toUpperCase();
    }

    return (parts[0][0] + parts[1][0]).toUpperCase();
  }, [userName]);

  return (
    <header className="sticky top-0 z-50 border-b border-[var(--line)] bg-[var(--header-bg)] px-4 backdrop-blur-lg">
      <nav className="page-wrap flex flex-wrap items-center gap-x-3 gap-y-2 py-3 sm:py-4">
        <h2 className="m-0 flex-shrink-0 text-base font-semibold tracking-tight">
          <Link
            to="/"
            className="inline-flex items-center gap-2 rounded-full border border-[var(--chip-line)] bg-[var(--chip-bg)] px-3 py-1.5 text-sm text-[var(--sea-ink)] no-underline shadow-[0_8px_24px_rgba(30,90,72,0.08)] sm:px-4 sm:py-2"
          >
            <span className="h-2 w-2 rounded-full bg-[linear-gradient(90deg,#adff2f,#00e676)]" />
            Soccer Manager
          </Link>
        </h2>

        <div className="order-3 flex w-full flex-wrap items-center gap-x-4 gap-y-1 pb-1 text-sm font-semibold sm:order-none sm:w-auto sm:flex-nowrap sm:pb-0">
          <Link
            to="/"
            className="nav-link"
            activeProps={{ className: "nav-link is-active" }}
          >
            Home
          </Link>
          {isLoggedIn ? (
            <>
              <Link
                to="/dashboard"
                className="nav-link"
                activeProps={{ className: "nav-link is-active" }}
              >
                Dashboard
              </Link>
              <Link
                to="/about"
                className="nav-link"
                activeProps={{ className: "nav-link is-active" }}
              >
                About
              </Link>
            </>
          ) : (
            <>
              <Link
                to="/signup"
                className="nav-link"
                activeProps={{ className: "nav-link is-active" }}
              >
                Signup
              </Link>
              <Link
                to="/login"
                className="nav-link"
                activeProps={{ className: "nav-link is-active" }}
              >
                Login
              </Link>
            </>
          )}
        </div>

        <div className="ml-auto flex items-center gap-1.5 sm:gap-2">
          <a
            href="https://www.fifa.com/"
            target="_blank"
            rel="noreferrer"
            className="hidden rounded-xl p-2 text-[var(--sea-ink-soft)] transition hover:bg-[var(--link-bg-hover)] hover:text-[var(--sea-ink)] sm:block"
          >
            <span className="sr-only">Visit FIFA website</span>
            <svg viewBox="0 0 16 16" aria-hidden="true" width="24" height="24">
              <path
                fill="currentColor"
                d="M8 0L2 2.5v6.4c0 3.6 2.6 6.9 6 7.1 3.4-.2 6-3.5 6-7.1V2.5L8 0zm0 2.2l4 1.7v5c0 2.5-1.7 4.9-4 5.2-2.3-.3-4-2.7-4-5.2v-5l4-1.7z"
              />
            </svg>
          </a>
          <a
            href="https://www.uefa.com/"
            target="_blank"
            rel="noreferrer"
            className="hidden rounded-xl p-2 text-[var(--sea-ink-soft)] transition hover:bg-[var(--link-bg-hover)] hover:text-[var(--sea-ink)] sm:block"
          >
            <span className="sr-only">Visit UEFA website</span>
            <svg viewBox="0 0 16 16" aria-hidden="true" width="24" height="24">
              <path
                fill="currentColor"
                d="M8 1.2a6.8 6.8 0 1 0 0 13.6A6.8 6.8 0 0 0 8 1.2zm0 1.6A5.2 5.2 0 1 1 8 13.2 5.2 5.2 0 0 1 8 2.8zm0 1.3a3.9 3.9 0 1 0 0 7.8 3.9 3.9 0 0 0 0-7.8z"
              />
            </svg>
          </a>

          {isLoggedIn ? (
            <div className="user-menu" ref={menuRef}>
              <button
                type="button"
                className="avatar-button"
                onClick={() => setMenuOpen((value) => !value)}
                aria-expanded={menuOpen}
                aria-haspopup="menu"
              >
                <span className="avatar-pill">{avatarLabel}</span>
                <span className="hidden text-sm font-semibold text-[var(--sea-ink)] sm:inline">
                  {userName}
                </span>
              </button>

              {menuOpen ? (
                <div className="user-menu-panel" role="menu">
                  <button
                    type="button"
                    className="user-menu-item"
                    onClick={() => {
                      setMenuOpen(false);
                      navigate({ to: "/settings" });
                    }}
                  >
                    Settings
                  </button>
                  <button
                    type="button"
                    className="user-menu-item"
                    onClick={() => {
                      setMenuOpen(false);
                      navigate({ to: "/about" });
                    }}
                  >
                    About
                  </button>
                  <button
                    type="button"
                    className="user-menu-item user-menu-item-danger"
                    onClick={() => {
                      logout();
                      setMenuOpen(false);
                      navigate({ to: "/" });
                    }}
                  >
                    Logout
                  </button>
                </div>
              ) : null}
            </div>
          ) : null}

          <ThemeToggle />
        </div>
      </nav>
    </header>
  );
}
