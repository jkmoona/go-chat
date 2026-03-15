<template>
    <div class="bg-background text-foreground h-dvh flex items-center justify-center dark">

        <!-- Join form -->
        <div v-if="!connected" class="w-full max-w-sm px-6">
            <div v-if="roomError" class="text-center space-y-2">
                <p class="text-lg font-semibold">Room not found</p>
                <p class="text-sm text-muted-foreground">This room may have expired or never existed.</p>
            </div>

            <template v-else-if="room">
                <div class="mb-6 text-center">
                    <h1 class="text-2xl font-bold">{{ room.name }}</h1>
                    <p class="text-sm text-muted-foreground mt-1">{{ formatExpiry(room.expires_at) }}</p>
                </div>

                <form @submit.prevent="joinRoom" class="space-y-3">
                    <Input
                        v-model="guestName"
                        placeholder="Your name"
                        maxlength="30"
                        required
                        autofocus
                    />
                    <Input
                        v-if="room.has_pin"
                        v-model="pin"
                        type="text"
                        inputmode="numeric"
                        maxlength="4"
                        placeholder="Room PIN"
                    />
                    <p v-if="joinError" class="text-sm text-destructive">{{ joinError }}</p>
                    <LoadingButton type="submit" class="w-full" :loading="joining" loading-text="Joining...">
                        Join Room
                    </LoadingButton>
                </form>
            </template>

            <div v-else class="text-center text-muted-foreground text-sm">Loading...</div>
        </div>

        <!-- Chat -->
        <div v-else class="flex flex-col h-full w-full max-w-2xl p-4">

            <!-- Header -->
            <div class="flex justify-between items-center mb-4 shrink-0">
                <div>
                    <h1 class="text-xl font-bold">{{ room?.name }}</h1>
                    <p class="text-xs text-muted-foreground">Guest: {{ guestName }}</p>
                </div>
                <div class="flex items-center gap-2">
                    <Button
                        variant="ghost"
                        size="sm"
                        @click="showUsers = !showUsers"
                        :class="{ 'text-primary': showUsers }"
                    >
                        <Users class="size-4" />
                        <span class="ml-1 text-xs">{{ onlineUsers.length }}</span>
                    </Button>
                    <Button variant="destructive" size="sm" @click="leaveRoom">
                        <LogOut class="size-4" /> Leave
                    </Button>
                </div>
            </div>

            <!-- Online users panel -->
            <div v-if="showUsers && onlineUsers.length > 0" class="mb-3 shrink-0 rounded-md border border-border p-2">
                <p class="text-xs text-muted-foreground mb-1">Online</p>
                <div class="flex flex-wrap gap-1">
                    <span
                        v-for="u in onlineUsers"
                        :key="u.id"
                        class="text-xs bg-muted rounded px-2 py-0.5"
                    >
                        {{ u.username }}<span v-if="u.is_guest" class="text-muted-foreground"> (guest)</span>
                    </span>
                </div>
            </div>

            <!-- Messages -->
            <Card class="flex-1 gap-1 overflow-y-auto p-3 mb-4 space-y-2 bg-[oklch(0.22_0.006_286)] rounded-lg">
                <ChatMessageCard
                    v-for="(msg, index) in messages"
                    :key="index"
                    :message="msg"
                    :current-user="guestName"
                />
                <div ref="bottomEl"></div>
            </Card>

            <!-- Countdown banner -->
            <div
                v-if="remaining > 0"
                class="mb-2 shrink-0 text-center text-sm rounded-md border border-border py-1.5 px-3"
                :class="remaining < 60 ? 'text-destructive border-destructive/50' : 'text-muted-foreground'"
            >
                Room expires in {{ formattedRemaining }}
            </div>

            <form @submit.prevent="send" class="flex gap-2 shrink-0">
                <input
                    v-model="newMessage"
                    ref="inputEl"
                    placeholder="Type a message..."
                    class="flex-1 rounded border px-3 py-2 focus:outline-none"
                />
                <Button :disabled="!newMessage.trim()">Send</Button>
            </form>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from "vue";
import { useRoute } from "vue-router";
import { toast } from "vue-sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { LogOut, Users } from "lucide-vue-next";
import LoadingButton from "@/components/LoadingButton.vue";
import ChatMessageCard from "@/components/ChatMessageCard.vue";

interface RoomInfo {
    id: string;
    name: string;
    ttl: number;
    expires_at: string;
    has_pin: boolean;
    clients: number;
}

interface ChatMessage {
    type: string;
    username?: string;
    content: string;
}

interface ClientInfo {
    id: string;
    username: string;
    is_guest: boolean;
}

interface WsMessage {
    type: string;
    content?: string;
    username?: string;
    roomId?: string;
    clients?: ClientInfo[];
    remaining?: number;
}

const route = useRoute();
const roomId = route.params.roomId as string;

const room = ref<RoomInfo | null>(null);
const roomError = ref(false);
const guestName = ref("");
const pin = ref("");
const joinError = ref("");
const joining = ref(false);
const connected = ref(false);
const showUsers = ref(false);

const newMessage = ref("");
const messages = ref<ChatMessage[]>([]);
const onlineUsers = ref<ClientInfo[]>([]);
const remaining = ref(0);
const bottomEl = ref<HTMLElement | null>(null);
const inputEl = ref<HTMLInputElement | null>(null);
let socket: WebSocket | null = null;
let redirectTimer: ReturnType<typeof setTimeout> | null = null;

const formattedRemaining = computed(() => {
    const m = Math.floor(remaining.value / 60);
    const s = remaining.value % 60;
    return `${m}:${s.toString().padStart(2, "0")}`;
});

function formatExpiry(expiresAt: string): string {
    const diff = new Date(expiresAt).getTime() - Date.now();
    if (diff <= 0) return "expired";
    const h = Math.floor(diff / 3600000);
    const m = Math.floor((diff % 3600000) / 60000);
    if (h > 0) return `Expires in ${h}h ${m}m`;
    return `Expires in ${m}m`;
}

function scrollToBottom() {
    bottomEl.value?.scrollIntoView({ behavior: "smooth" });
}

async function fetchRoom() {
    try {
        const res = await fetch(`/api/ws/room/${roomId}`);
        if (!res.ok) {
            roomError.value = true;
            return;
        }
        room.value = await res.json();
    } catch {
        roomError.value = true;
    }
}

async function joinRoom() {
    if (!guestName.value.trim()) return;
    if (room.value?.has_pin && !pin.value) {
        joinError.value = "PIN is required";
        return;
    }

    joining.value = true;
    joinError.value = "";

    const name = guestName.value.trim();
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    let wsUrl = `${protocol}//${window.location.host}/ws/guest/joinRoom/${roomId}?name=${encodeURIComponent(name)}`;
    if (pin.value) wsUrl += `&pin=${encodeURIComponent(pin.value)}`;

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        joining.value = false;
        connected.value = true;
        nextTick(() => inputEl.value?.focus());
    };

    socket.onerror = () => {
        joining.value = false;
        joinError.value = "Failed to connect. Check your PIN and try again.";
        socket = null;
    };

    socket.onclose = (event) => {
        if (!connected.value) {
            joining.value = false;
            joinError.value = event.code === 1008
                ? "Invalid PIN or room not found."
                : "Connection refused.";
            socket = null;
        }
    };

    socket.onmessage = async (event: MessageEvent) => {
        let msg: WsMessage;
        try {
            msg = JSON.parse(event.data as string);
        } catch {
            msg = { type: "system", content: event.data as string };
        }

        switch (msg.type) {
            case "presence":
                onlineUsers.value = msg.clients ?? [];
                break;
            case "countdown":
                remaining.value = msg.remaining ?? 0;
                break;
            case "system":
                messages.value.push({ type: "system", content: msg.content ?? "" });
                if (msg.content === "room has expired") {
                    toast.error("Room expired");
                    redirectTimer = setTimeout(() => {
                        connected.value = false;
                        messages.value = [];
                        onlineUsers.value = [];
                        remaining.value = 0;
                        socket = null;
                    }, 2000);
                }
                await nextTick();
                scrollToBottom();
                break;
            default:
                messages.value.push({
                    type: msg.type,
                    username: msg.username,
                    content: msg.content ?? "",
                });
                if (messages.value.length > 500) messages.value.shift();
                await nextTick();
                scrollToBottom();
        }
    };
}

async function send() {
    if (!newMessage.value.trim() || !socket || socket.readyState !== WebSocket.OPEN) return;

    socket.send(JSON.stringify({
        type: "chat",
        username: guestName.value,
        content: newMessage.value,
        roomId,
    }));
    newMessage.value = "";
}

function leaveRoom() {
    if (socket) {
        socket.close();
        socket = null;
    }
    connected.value = false;
    messages.value = [];
    onlineUsers.value = [];
    remaining.value = 0;
}

onMounted(fetchRoom);

onBeforeUnmount(() => {
    if (socket) {
        socket.close();
        socket = null;
    }
    if (redirectTimer) {
        clearTimeout(redirectTimer);
        redirectTimer = null;
    }
});
</script>
