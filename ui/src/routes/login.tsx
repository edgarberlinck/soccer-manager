import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import type { FormEvent } from "react";
import { loginUser } from "../lib/session";

export const Route = createFileRoute("/login")({ component: LoginPage });

function LoginPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const result = await loginUser({ email, password });

    if (!result.ok) {
      setError(result.message);
      return;
    }

    navigate({ to: "/dashboard" });
  }

  return (
    <main className="page-wrap page-gutter">
      <section className="pitch-card mx-auto max-w-xl p-8 sm:p-10">
        <p className="eyebrow">Onboarding</p>
        <h1 className="title-lg mt-2">Login do treinador</h1>
        <p className="muted mt-2">
          Entre para gerenciar seu elenco, treinos e as proximas partidas.
        </p>

        <form className="mt-7 space-y-4" onSubmit={onSubmit}>
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
            required
          />

          {error ? <p className="text-sm text-rose-300">{error}</p> : null}

          <button type="submit" className="btn-primary w-full">
            Entrar no dashboard
          </button>
        </form>

        <p className="muted mt-6 text-sm">
          Ainda nao tem clube?{" "}
          <Link className="link-inline" to="/signup">
            Criar conta e montar o time
          </Link>
        </p>
      </section>
    </main>
  );
}
