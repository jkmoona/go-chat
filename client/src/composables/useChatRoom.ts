import { ref, computed, onBeforeUnmount } from "vue";

export interface ChatMessage {
    type: string;
    username?: string;
    content: string;
    roomId?: string;
    timestamp: number;
}

export interface ClientInfo {
    id: string;
    username: string;
    is_guest: boolean;
}

export type ConnectionStatus =
    | "idle"
    | "connecting"
    | "connected"
    | "disconnected"
    | "kicked"
    | "expired";

const MAX_ATTEMPTS = 10;
const MAX_DELAY = 30_000;

interface ServerMessage {
    type: string;
    clients?: ClientInfo[];
    remaining?: number;
    content?: string;
    username?: string;
}

function parseMessage(data: string): ServerMessage | null {
    try {
        return JSON.parse(data) as ServerMessage;
    } catch {
        return null;
    }
}

export function useChatRoom(roomId: string, getUsername: () => string) {
    const messages = ref<ChatMessage[]>([]);
    const onlineUsers = ref<ClientInfo[]>([]);
    const remaining = ref(0);
    const totalRemaining = ref(0);
    const status = ref<ConnectionStatus>("idle");
    const closeCode = ref(0);
    const closeReason = ref("");

    let ws: WebSocket | null = null;
    let intentionalClose = false;
    let attempts = 0;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let countdownTimer: ReturnType<typeof setInterval> | null = null;
    let currentUrl = "";

    const formattedRemaining = computed(() => {
        const total = remaining.value;
        const h = Math.floor(total / 3600);
        const m = Math.floor((total % 3600) / 60);
        const s = total % 60;
        if (h > 0) return `${h}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
        return `${m}:${s.toString().padStart(2, "0")}`;
    });

    function push(type: string, content: string, sender?: string) {
        messages.value.push({ type, content, username: sender, roomId, timestamp: Date.now() });
        if (messages.value.length > 500) messages.value.shift();
    }

    function connect(url: string) {
        cleanup();
        currentUrl = url;
        intentionalClose = false;
        status.value = "connecting";

        ws = new WebSocket(url);

        ws.onopen = () => {
            status.value = "connected";
            attempts = 0;
        };

        ws.onmessage = ({ data }: MessageEvent) => {
            const msg = parseMessage(data);
            if (!msg) return;

            switch (msg.type) {
                case "presence":
                    onlineUsers.value = msg.clients ?? [];
                    break;
                case "countdown": {
                    const secs = msg.remaining ?? 0;
                    remaining.value = secs;
                    if (!totalRemaining.value || secs > totalRemaining.value) totalRemaining.value = secs;
                    if (!countdownTimer) {
                        countdownTimer = setInterval(() => {
                            if (remaining.value > 0) remaining.value--;
                        }, 1000);
                    }
                    break;
                }
                case "kicked":
                    push("system", msg.content ?? "");
                    status.value = "kicked";
                    intentionalClose = true;
                    ws!.close();
                    ws = null;
                    break;
                case "expired":
                case "deleted":
                    push("system", msg.content ?? "");
                    status.value = "expired";
                    intentionalClose = true;
                    break;
                case "system":
                    push("system", msg.content ?? "");
                    break;
                default:
                    push(msg.type, msg.content ?? "", msg.username);
            }
        };

        ws.onclose = (e: CloseEvent) => {
            if (intentionalClose) return;

            const wasConnected = status.value === "connected";

            if (wasConnected && attempts < MAX_ATTEMPTS) {
                status.value = "connecting";
                const delay = Math.min(1000 * 2 ** attempts, MAX_DELAY);
                attempts++;
                reconnectTimer = setTimeout(() => connect(currentUrl), delay);
            } else {
                closeCode.value = e.code;
                closeReason.value = e.reason;
                status.value = "disconnected";
            }
        };
    }

    function disconnect() {
        intentionalClose = true;
        cleanup();
    }

    function cleanup() {
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }
        if (countdownTimer) {
            clearInterval(countdownTimer);
            countdownTimer = null;
        }
        if (ws) {
            ws.close();
            ws = null;
        }
    }

    function send(content: string): boolean {
        if (!ws || ws.readyState !== WebSocket.OPEN) return false;
        const name = getUsername();
        ws.send(JSON.stringify({ type: "chat", username: name, content, roomId }));
        push("chat", content, name);
        return true;
    }

    function seedTotal(seconds: number) {
        if (!totalRemaining.value) totalRemaining.value = seconds;
    }

    function retry() {
        attempts = 0;
        connect(currentUrl);
    }

    onBeforeUnmount(disconnect);

    return {
        messages,
        onlineUsers,
        remaining,
        totalRemaining,
        status,
        formattedRemaining,
        closeCode,
        closeReason,
        connect,
        disconnect,
        send,
        retry,
        seedTotal,
    };
}
