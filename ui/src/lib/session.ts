export type Club = {
  id: string;
  name: string;
};

export type SessionUser = {
  id: string;
  email: string;
  managerName: string;
};

export type RosterPlayer = {
  id: string;
  name: string;
  age: number;
  position: string;
  overall: number;
  potential: number;
  salary_eur: string;
  contract_ends_at: string;
  primary_strength: string;
};

export type PlayerDetail = {
  id: string;
  name: string;
  age: number;
  position: string;
  overall: number;
  potential: number;
  contract: {
    salary_eur: string;
    release_clause_eur?: string;
    starts_at: string;
    ends_at: string;
  };
  attributes: {
    pace: number;
    passing: number;
    shooting: number;
    altura: number;
    peso: number;
    impulso: number;
    explosao: number;
    fisico: number;
    fisical_status: number;
    cabeceio: number;
    cruzamento: number;
    habilidade: number;
    finalizacao: number;
    dominio: number;
    temperamento: number;
  };
  summary: {
    games: number;
    goals: number;
    assists: number;
    avg_rating: string;
    minutes_played: number;
  };
  matches: Array<{
    match_id: string;
    played_at: string;
    minutes_played: number;
    goals: number;
    assists: number;
    rating: string;
    passes_completed: number;
    shots: number;
    tackles: number;
    saves: number;
    scoreline: string;
  }>;
};

const TOKEN_KEY = "soccer-manager:auth-token";
const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL?.toString().replace(/\/$/, "") ||
  "http://localhost:8080";

function canUseStorage() {
  return (
    typeof window !== "undefined" && typeof window.localStorage !== "undefined"
  );
}

export function getAuthToken() {
  if (!canUseStorage()) {
    return null;
  }

  return window.localStorage.getItem(TOKEN_KEY);
}

function setAuthToken(token: string) {
  if (!canUseStorage()) {
    return;
  }

  window.localStorage.setItem(TOKEN_KEY, token);
}

export function logout() {
  if (!canUseStorage()) {
    return;
  }

  window.localStorage.removeItem(TOKEN_KEY);
}

async function apiRequest<T>(
  path: string,
  init?: RequestInit,
  useAuth = true,
): Promise<T> {
  const headers = new Headers(init?.headers ?? {});
  headers.set("Content-Type", "application/json");

  if (useAuth) {
    const token = getAuthToken();
    if (!token) {
      throw new Error("Sessao expirada. Faca login novamente.");
    }

    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || "Falha na requisicao.");
  }

  return (await response.json()) as T;
}

export async function signupUser(input: {
  email: string;
  password: string;
  managerName: string;
  clubName: string;
}) {
  try {
    const result = await apiRequest<{ token: string }>(
      "/auth/signup",
      {
        method: "POST",
        body: JSON.stringify({
          email: input.email,
          password: input.password,
          manager_name: input.managerName,
          club_name: input.clubName,
          club_short_name: input.clubName,
        }),
      },
      false,
    );

    if (!result.token) {
      return { ok: false as const, message: "Token nao retornado no signup." };
    }

    setAuthToken(result.token);
    return { ok: true as const };
  } catch (error) {
    return {
      ok: false as const,
      message:
        error instanceof Error
          ? error.message
          : "Falha ao criar conta no servidor.",
    };
  }
}

export async function loginUser(input: { email: string; password: string }) {
  try {
    const result = await apiRequest<{ token: string }>(
      "/auth/signin",
      {
        method: "POST",
        body: JSON.stringify(input),
      },
      false,
    );

    if (!result.token) {
      return { ok: false as const, message: "Token nao retornado no login." };
    }

    setAuthToken(result.token);
    return { ok: true as const };
  } catch (error) {
    return {
      ok: false as const,
      message: error instanceof Error ? error.message : "Falha ao fazer login.",
    };
  }
}

export async function getSessionUser() {
  try {
    return await apiRequest<SessionUser>("/auth/me", { method: "GET" });
  } catch {
    return null;
  }
}

export async function listMyClubs() {
  return apiRequest<Club[]>("/clubs", { method: "GET" });
}

export async function ensureMyClub() {
  return apiRequest<{
    club_id: string;
    club_name: string;
    club_created: boolean;
  }>("/clubs/ensure", {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export async function listClubPlayers(clubId: string) {
  const result = await apiRequest<{ players: RosterPlayer[] }>(
    `/clubs/${clubId}/players`,
    { method: "GET" },
  );
  return result.players;
}

export async function ensureClubSquad(clubId: string) {
  return apiRequest<{
    club_id: string;
    squad_created: boolean;
    players_existing: number;
  }>(`/clubs/${clubId}/ensure-squad`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export async function getClubPlayerDetail(clubId: string, playerId: string) {
  return apiRequest<PlayerDetail>(`/clubs/${clubId}/players/${playerId}`, {
    method: "GET",
  });
}
