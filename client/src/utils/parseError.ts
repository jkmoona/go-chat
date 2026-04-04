export async function parseApiError(res: Response, fallback: string): Promise<string> {
  try {
    const data = await res.json();
    if (data?.error && typeof data.error === "string") {
      return data.error;
    }
  } catch {
    // not JSON
  }
  return `${fallback} (${res.status})`;
}
