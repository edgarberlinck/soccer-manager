import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { getClubPlayerDetail } from "../lib/session";
import type { PlayerDetail } from "../lib/session";

export const Route = createFileRoute("/dashboard/player/$clubId/$playerId")({
  component: PlayerDetailsPage,
});

function PlayerDetailsPage() {
  const { clubId, playerId } = Route.useParams();
  const navigate = useNavigate();
  const [player, setPlayer] = useState<PlayerDetail | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const load = async () => {
      try {
        const data = await getClubPlayerDetail(clubId, playerId);
        setPlayer(data);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Falha ao carregar jogador.",
        );
      }
    };

    void load();
  }, [clubId, playerId]);

  if (error) {
    return (
      <main className="page-wrap page-gutter">
        <section className="pitch-card p-8">
          <p className="text-rose-300">{error}</p>
          <button
            type="button"
            className="btn-ghost mt-4"
            onClick={() => navigate({ to: "/dashboard" })}
          >
            Voltar ao dashboard
          </button>
        </section>
      </main>
    );
  }

  if (!player) {
    return (
      <main className="page-wrap page-gutter">
        <section className="pitch-card p-8">
          <p className="muted">Carregando ficha do jogador...</p>
        </section>
      </main>
    );
  }

  const attributes = [
    ["Pace", player.attributes.pace],
    ["Passing", player.attributes.passing],
    ["Shooting", player.attributes.shooting],
    ["Altura", player.attributes.altura],
    ["Peso", player.attributes.peso],
    ["Impulso", player.attributes.impulso],
    ["Explosao", player.attributes.explosao],
    ["Fisico", player.attributes.fisico],
    ["Fisical", player.attributes.fisical_status],
    ["Cabeceio", player.attributes.cabeceio],
    ["Cruzamento", player.attributes.cruzamento],
    ["Habilidade", player.attributes.habilidade],
    ["Finalizacao", player.attributes.finalizacao],
    ["Dominio", player.attributes.dominio],
    ["Temperamento", player.attributes.temperamento],
  ] as const;

  const contractStartsAt = new Date(
    player.contract.starts_at,
  ).toLocaleDateString();
  const contractEndsAt = new Date(player.contract.ends_at).toLocaleDateString();

  return (
    <main className="page-wrap page-gutter">
      <section className="dashboard-hero p-8 sm:p-10">
        <button
          type="button"
          className="btn-ghost"
          onClick={() => navigate({ to: "/dashboard" })}
        >
          Voltar
        </button>
        <p className="eyebrow mt-5">Ficha do jogador</p>
        <h1 className="title-lg mt-2">{player.name}</h1>
        <p className="muted mt-2">
          {player.position} · {player.age} anos · OVR {player.overall} · POT{" "}
          {player.potential}
        </p>
        <div className="mt-4 flex flex-wrap gap-2">
          <span className="stat-chip">
            Salario: {player.contract.salary_eur}
          </span>
          <span className="stat-chip">Inicio: {contractStartsAt}</span>
          <span className="stat-chip">Contrato ate: {contractEndsAt}</span>
          {player.contract.release_clause_eur ? (
            <span className="stat-chip">
              Clausula: {player.contract.release_clause_eur}
            </span>
          ) : null}
        </div>
      </section>

      <section className="mt-6 grid gap-4 lg:grid-cols-3">
        <article className="pitch-card p-5 lg:col-span-2">
          <h2 className="m-0 text-lg font-semibold text-white">Atributos</h2>
          <div className="mt-4 grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {attributes.map(([name, value]) => (
              <div key={name} className="feature-tile p-3">
                <p className="muted text-xs uppercase tracking-wide">{name}</p>
                <p className="m-0 text-lg font-bold text-white">{value}</p>
              </div>
            ))}
          </div>
        </article>

        <article className="pitch-card p-5">
          <h2 className="m-0 text-lg font-semibold text-white">Resumo</h2>
          <div className="mt-4 grid gap-2 text-sm">
            <p className="muted">Jogos: {player.summary.games}</p>
            <p className="muted">Gols: {player.summary.goals}</p>
            <p className="muted">Assistencias: {player.summary.assists}</p>
            <p className="muted">Minutos: {player.summary.minutes_played}</p>
            <p className="muted">Nota media: {player.summary.avg_rating}</p>
          </div>
        </article>
      </section>

      <section className="mt-6 pitch-card p-5">
        <h2 className="m-0 text-lg font-semibold text-white">
          Desempenho por partida
        </h2>
        {player.matches.length === 0 ? (
          <p className="muted mt-3">
            Sem partidas registradas para este jogador.
          </p>
        ) : (
          <div className="mt-4 overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-[var(--sea-ink-soft)]">
                  <th className="py-2">Data</th>
                  <th className="py-2">Placar</th>
                  <th className="py-2">Min</th>
                  <th className="py-2">G/A</th>
                  <th className="py-2">Nota</th>
                  <th className="py-2">Passes</th>
                  <th className="py-2">Finalizacoes</th>
                  <th className="py-2">Desarmes</th>
                  <th className="py-2">Defesas</th>
                </tr>
              </thead>
              <tbody>
                {player.matches.map((match) => (
                  <tr
                    key={match.match_id}
                    className="border-t border-[var(--line)] text-[var(--sea-ink)]"
                  >
                    <td className="py-2">
                      {new Date(match.played_at).toLocaleDateString()}
                    </td>
                    <td className="py-2">{match.scoreline}</td>
                    <td className="py-2">{match.minutes_played}</td>
                    <td className="py-2">
                      {match.goals}/{match.assists}
                    </td>
                    <td className="py-2">{match.rating}</td>
                    <td className="py-2">{match.passes_completed}</td>
                    <td className="py-2">{match.shots}</td>
                    <td className="py-2">{match.tackles}</td>
                    <td className="py-2">{match.saves}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
}
