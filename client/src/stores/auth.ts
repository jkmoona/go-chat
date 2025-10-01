import { defineStore } from "pinia";
import { apiFetch, refreshToken } from "../services/api";

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
                const res = await apiFetch("/login", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        username: username.trim(),
                        password: password.trim(),
                    }),
                });
                const data = await res.json();
                if (res.ok) {
                    this.user = data;
                    this.isAuthenticated = true;
                    localStorage.setItem('user', JSON.stringify(data));
                } else {
                    throw new Error(data.error || "Login failed.");
                }
            } catch (err: unknown) {
                this.user = null;
                this.isAuthenticated = false;
                if (err instanceof Error) {
                    throw new Error(err.message || "Network error. Please try again.");
                } else {
                    throw new Error("Network error. Please try again.");
                }
            }
        },
        async register(username: string, password: string) {
            try {
                const res = await apiFetch("/register", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        username: username.trim(),
                        password: password.trim(),
                    }),
                });
                const data = await res.json();
                if (res.ok && data.id && data.username) {
                    this.user = null;
                    this.isAuthenticated = false;
                } else {
                    throw new Error(data.error || "Registration failed.");
                }
            } catch (err: unknown){
                this.user = null;
                this.isAuthenticated = false;
                if (err instanceof Error) {
                    throw new Error(err.message || "Network error. Please try again.");
                } else {
                    throw new Error("Network error. Please try again.");
                }
            }
        },
        async logout() {
            await apiFetch("/logout", {
                method: "GET",
            });
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
            this.isAuthenticated = false;
            this.user = null;
            return false;
        },
    },
});
