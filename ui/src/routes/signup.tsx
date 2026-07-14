import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import type { FormEvent } from "react";
import { signupUser } from "../lib/session";

export const Route = createFileRoute("/signup")({ component: SignupPage });

function SignupPage() {
  const navigate = useNavigate();
  const [managerName, setManagerName] = useState("");
  const [clubName, setClubName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const result = await signupUser({
      managerName,
      clubName,
      email,
      password,
    });

    if (!result.ok) {
      setError(result.message);
      return;
    }

    navigate({ to: "/dashboard" });
  }

  return (
    <main className="page-wrap page-gutter">
      <section className="pitch-card mx-auto max-w-2xl p-8 sm:p-10">
        <p className="eyebrow">Onboarding</p>
        <h1 className="title-lg mt-2">Criar clube e conta</h1>
        <p className="muted mt-2">
          Defina seu nome de treinador e o nome do clube. O elenco inicial sera
          gerado com jogadores ruins, medianos e alguns poucos bons.
        </p>

        <form className="mt-7 grid gap-4 sm:grid-cols-2" onSubmit={onSubmit}>
          <div>
            <label className="field-label" htmlFor="managerName">
              Seu nome (treinador)
            </label>
            <input
              id="managerName"
              className="field-input"
              value={managerName}
              onChange={(event) => setManagerName(event.target.value)}
              placeholder="Edgar"
              required
            />
          </div>

          <div>
            <label className="field-label" htmlFor="clubName">
              Nome do clube
            </label>
            <input
              id="clubName"
              className="field-input"
              value={clubName}
              onChange={(event) => setClubName(event.target.value)}
              placeholder="Aurora FC"
              required
            />
          </div>

          <div>
            <label className="field-label" htmlFor="email">
              Email
            </label>
            <input
              id="email"
              className="field-input"
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="voce@clube.com"
              required
            />
          </div>

          <div>
            <label className="field-label" htmlFor="password">
              Senha
            </label>
            <input
              id="password"
              className="field-input"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="••••••••"
              minLength={8}
              required
            />
          </div>

          {error ? (
            <p className="sm:col-span-2 text-sm text-rose-300">{error}</p>
          ) : null}

          <button type="submit" className="btn-primary sm:col-span-2">
            Criar conta e ir para o dashboard
          </button>
        </form>

        <p className="muted mt-6 text-sm">
          Ja tem conta?{" "}
          <Link className="link-inline" to="/login">
            Fazer login
          </Link>
        </p>
      </section>
    </main>
  );
}
