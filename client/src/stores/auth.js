import { defineStore } from "pinia";
import { apiFetch, refreshToken } from "../services/api";

export const useAuthStore = defineStore("auth", {
    state: () => ({
        user: null,
        isAuthenticated: false,
    }),
    actions: {
        async login(email, password) {
            try {
                const res = await apiFetch("http://localhost:8080/login", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        email: email.trim(),
                        password: password.trim(),
                    }),
                });
                const data = await res.json();
                if (res.ok) {
                    this.user = data;
                    this.isAuthenticated = true;
                } else {
                    throw new Error(data.error || "Login failed.");
                }
            } catch (err) {
                this.user = null;
                this.isAuthenticated = false;
                throw new Error(err.message || "Network error. Please try again.");
            }
        },
        async register(username, email, password) {
            try {
                const res = await apiFetch("http://localhost:8080/signup", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        username: username.trim(),
                        email: email.trim(),
                        password: password.trim(),
                    }),
                });
                const data = await res.json();
                if (res.ok && data.id && data.username && data.email) {
                    this.user = null;
                    this.isAuthenticated = false;
                } else {
                    throw new Error(data.error || "Registration failed.");
                }
            } catch (err){
                this.user = null;
                this.isAuthenticated = false;
                throw new Error(err.message || "Network error. Please try again.");
            }
        },
        async logout() {
            await apiFetch("http://localhost:8080/logout", {
                method: "GET",
            });
            this.user = null;
            this.isAuthenticated = false;
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
