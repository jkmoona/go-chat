import { SERVER_URL } from "./constants";

export async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
    let res = await fetch(`${SERVER_URL}${path}`, { ...options, credentials: "include" });
    if (res.status === 401) {
        const refreshed = await refreshToken();
        if (refreshed) {
            console.log("why refresh??");
            res = await fetch(`${SERVER_URL}${path}`, { ...options, credentials: "include" });
        }
    }
    return res;
}

export async function refreshToken(): Promise<boolean> {
    const res = await fetch(`${SERVER_URL}/refresh`, {
        method: "POST",
        credentials: "include",
    });
    return res.ok;
}