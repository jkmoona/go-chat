export async function apiFetch(url, options = {}) {
    const res = await fetch(url, { ...options, credentials: "include" });
    if (res.status === 401) {
        // Try refresh
        const refreshed = await refreshToken();
        if (refreshed) {
            // Retry original request
            return fetch(url, { ...options, credentials: "include" });
        }
    }
    return res;
}

export async function refreshToken() {
    const res = await fetch("http://localhost:8080/refresh", {
        method: "POST",
        credentials: "include",
    });
    return res.ok;
}