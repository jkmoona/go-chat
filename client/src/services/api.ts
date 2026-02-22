export async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
    let res = await fetch(`api/${path}`, { ...options, credentials: "include" });
    if (res.status === 401) {
        const refreshed = await refreshToken();
        if (refreshed) {
            res = await fetch(`api/${path}`, { ...options, credentials: "include" });
        }
    }
    return res;
}

export async function refreshToken(): Promise<boolean> {
    const res = await fetch(`api/refresh`, {
        method: "POST",
        credentials: "include",
    });
    return res.ok;
}
