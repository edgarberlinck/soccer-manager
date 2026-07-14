import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({ component: App });

function App() {
  return (
    <main className="page-wrap page-gutter">
      <section className="hero-grid gap-6 lg:grid lg:grid-cols-[1.2fr_0.8fr] lg:items-center">
        <article className="pitch-card p-8 sm:p-10">
          <p className="eyebrow">Soccer Manager</p>
          <h1 className="title-xl mt-3">
            O manager de futebol com energia gamer e foco em evolucao de elenco.
          </h1>
          <p className="muted mt-5 max-w-2xl">
            Comece com um time improvisado, monte sua estrategia e transforme
            promessas em craques. O objetivo e dominar sua liga com boa gestao,
            treino e escolhas inteligentes.
          </p>

          <div className="mt-7 flex flex-wrap gap-3">
            <a href="/signup" className="btn-primary">
              Criar clube
            </a>
            <a href="/login" className="btn-ghost">
              Ja tenho conta
            </a>
          </div>
        </article>

        <article className="pitch-card p-6 sm:p-8">
          <h2 className="title-md">Resumo do jogo</h2>
          <ul className="mt-4 space-y-3 pl-5 text-sm text-[var(--text-muted)]">
            <li>Monte seu clube e assuma o papel de treinador.</li>
            <li>Receba um elenco inicial com qualidade aleatoria.</li>
            <li>Gerencie titulares, reservas e desenvolvimento individual.</li>
            <li>Cresca da base para desafiar os melhores times.</li>
          </ul>
        </article>
      </section>

      <section className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[
          ["Onboarding rapido", "Login, signup e criacao do clube em minutos."],
          [
            "Elenco dinamico",
            "Jogadores ruins, medianos e bons no time inicial.",
          ],
          ["Dashboard central", "Visao geral de status do clube e atletas."],
          [
            "Estetica gamer",
            "Interface com clima de arena noturna e energia neon.",
          ],
        ].map(([title, text]) => (
          <article key={title} className="feature-tile p-5">
            <h3 className="m-0 text-base font-semibold text-white">{title}</h3>
            <p className="muted mt-2 text-sm">{text}</p>
          </article>
        ))}
      </section>
    </main>
  );
}
