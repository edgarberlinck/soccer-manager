import {
  createFileRoute,
  Link,
  Outlet,
  useNavigate,
  useRouterState,
} from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";
import {
  ensureClubSquad,
  ensureMyClub,
  getSessionUser,
  listMyClubs,
  listClubPlayers,
  logout,
} from "../lib/session";
import type { RosterPlayer } from "../lib/session";

export const Route = createFileRoute("/dashboard")({
  component: DashboardPage,
});

function DashboardPage() {
  const [loaded, setLoaded] = useState(false);
  const [userName, setUserName] = useState("");
  const [clubName, setClubName] = useState("");
  const [clubId, setClubId] = useState("");
  const [players, setPlayers] = useState<RosterPlayer[]>([]);
  const [error, setError] = useState("");
  const [progressMessage, setProgressMessage] = useState("");
  const navigate = useNavigate();
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  const isPlayerDetailsRoute = pathname.startsWith("/dashboard/player/");

  useEffect(() => {
    const load = async () => {
      const user = await getSessionUser();
      if (!user) {
        navigate({ to: "/login" });
        return;
      }

      try {
        setProgressMessage("Carregando dados do clube...");
        let clubs = await listMyClubs();
        if (clubs.length === 0) {
          setProgressMessage("aguarde, criando seu time");
          const created = await ensureMyClub();
          clubs = [{ id: created.club_id, name: created.club_name }];
        }

        const activeClub = clubs[0];
        setProgressMessage("Carregando elenco...");
        let roster = await listClubPlayers(activeClub.id);

        if (roster.length === 0) {
          setProgressMessage("aguarde, criando seu time");
          await ensureClubSquad(activeClub.id);
          roster = await listClubPlayers(activeClub.id);
        }

        setUserName(user.managerName || user.email);
        setClubName(activeClub.name);
        setClubId(activeClub.id);
        setPlayers(roster);
        setProgressMessage("");
        setLoaded(true);
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "Falha ao carregar os dados do clube.",
        );
        setProgressMessage("");
        setLoaded(true);
      }
    };

    void load();
  }, [navigate]);

  const stats = useMemo(() => {
    const total = players.length;
    const avgOverall =
      total > 0
        ? Math.round(
            players.reduce((acc, player) => acc + player.overall, 0) / total,
          )
        : 0;
    const avgAge =
      total > 0
        ? Math.round(
            players.reduce((acc, player) => acc + player.age, 0) / total,
          )
        : 0;

    return { total, avgOverall, avgAge };
  }, [players]);

  if (isPlayerDetailsRoute) {
    return <Outlet />;
  }

  if (!loaded) {
    return (
      <main className="page-wrap page-gutter">
        <section className="pitch-card p-10 text-center">
          <p className="muted">Carregando seu clube...</p>
          {progressMessage ? (
            <p className="muted mt-2">{progressMessage}</p>
          ) : null}
        </section>
      </main>
    );
  }

  return (
    <main className="page-wrap page-gutter">
      <section className="dashboard-hero p-8 sm:p-10">
        <p className="eyebrow">Dashboard do treinador</p>
        <h1 className="title-lg mt-2">{clubName}</h1>
        <p className="muted mt-2">Comando tecnico: {userName}</p>

        <div className="mt-6 flex flex-wrap gap-3">
          <span className="stat-chip">Elenco: {stats.total}</span>
          <span className="stat-chip">OVR medio: {stats.avgOverall}</span>
          <span className="stat-chip">Idade media: {stats.avgAge}</span>
        </div>

        <button
          type="button"
          className="btn-ghost mt-6"
          onClick={() => {
            logout();
            navigate({ to: "/" });
          }}
        >
          Sair
        </button>
      </section>

      <section className="mt-6">
        <div className="mb-3 flex items-end justify-between gap-2">
          <h2 className="m-0 text-xl font-semibold text-white">
            Jogadores do time
          </h2>
          <p className="muted text-sm">
            Clique em um jogador para abrir a ficha.
          </p>
        </div>

        {error ? <p className="mb-3 text-sm text-rose-300">{error}</p> : null}
        {progressMessage ? (
          <p className="mb-3 text-sm text-[var(--text-muted)]">
            {progressMessage}
          </p>
        ) : null}

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {players.map((player) => (
            <Link
              key={player.id}
              to="/dashboard/player/$clubId/$playerId"
              params={{ clubId, playerId: player.id }}
              className="player-card block p-4 text-left"
            >
              <div className="flex items-center justify-between gap-2">
                <h3 className="m-0 text-base font-semibold text-white">
                  {player.name}
                </h3>
                <span className="stat-chip">{player.position}</span>
              </div>
              <p className="muted mt-2 text-sm">
                {player.age} anos · Forca principal: {player.primary_strength}
              </p>
              <div className="mt-3 flex flex-wrap gap-2 text-xs">
                <span className="stat-chip">OVR {player.overall}</span>
                <span className="stat-chip">POT {player.potential}</span>
                <span className="stat-chip">{player.salary_eur}</span>
              </div>
            </Link>
          ))}
        </div>
      </section>

      <section className="mt-8 text-center">
        <p className="muted text-sm">
          Quer recomeçar a carreira?{" "}
          <Link className="link-inline" to="/signup">
            Criar novo clube
          </Link>
        </p>
      </section>
    </main>
  );
}
