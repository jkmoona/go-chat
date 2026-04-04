import { defineStore } from "pinia";
import { apiFetch, refreshToken } from "../services/api";
import { parseApiError } from "../utils/parseError";
interface User {
    id: string;
    username: string;
}

export const useAuthStore = defineStore("auth", {
    state: () => ({
        user: JSON.parse(localStorage.getItem('user') || 'null') as User | null,
        isAuthenticated: !!localStorage.getItem('user'),
    }),
    actions: {
        async login(username: string, password: string) {
            try {
                const res = await apiFetch("login", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        username: username.trim(),
                        password,
                    }),
                });

                if (!res.ok) {
                    throw new Error(await parseApiError(res, "Login failed"));
                }

                const data = await res.json();
                this.user = data;
                this.isAuthenticated = true;
                localStorage.setItem('user', JSON.stringify(data));
            } catch (err: unknown) {
                this.user = null;
                this.isAuthenticated = false;
                if (err instanceof TypeError) {
                    throw new Error("Network error. Please check your connection.");
                }
                throw err instanceof Error ? err : new Error("An unexpected error occurred.");
            }
        },
        async register(username: string, password: string) {
            try {
                const res = await apiFetch("register", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        username: username.trim(),
                        password,
                    }),
                });

                if (!res.ok) {
                    throw new Error(await parseApiError(res, "Registration failed"));
                }

                this.user = null;
                this.isAuthenticated = false;
            } catch (err: unknown) {
                this.user = null;
                this.isAuthenticated = false;
                if (err instanceof TypeError) {
                    throw new Error("Network error. Please check your connection.");
                }
                throw err instanceof Error ? err : new Error("An unexpected error occurred.");
            }
        },
        async logout() {
            try {
                await apiFetch("logout", { method: "GET" });
            } catch {
                // still clear local state even if server logout fails
            }
            this.user = null;
            this.isAuthenticated = false;
            localStorage.removeItem('user');
        },
        async tryRefresh() {
            const ok = await refreshToken();
            if (ok) {
                this.isAuthenticated = true;
                return true;
            }
            return false;
        },
    },
});
