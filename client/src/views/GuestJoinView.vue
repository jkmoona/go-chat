<template>
    <div class="bg-background text-foreground h-dvh flex items-center justify-center dark">

        <!-- Join form -->
        <div v-if="!joined" class="w-full max-w-sm px-6">
            <div v-if="roomError" class="text-center space-y-2">
                <p class="text-lg font-black">Room not found</p>
                <p class="text-sm text-muted-foreground font-mono">{{ roomErrorMessage }}</p>
            </div>

            <template v-else-if="room">
                <div class="mb-6">
                    <h1 class="text-3xl font-black tracking-tight">{{ room.name }}</h1>
                    <div class="h-1.5 w-10 bg-primary mt-2"></div>
                    <p class="text-xs text-muted-foreground mt-2 font-mono">{{ formatExpiry(room.expires_at) }}</p>
                </div>

                <div class="border-2 border-border p-5 neo-card">
                    <form @submit.prevent="joinRoom" class="space-y-3">
                        <div>
                            <label class="block text-[11px] font-black uppercase tracking-widest mb-1.5 text-muted-foreground">Your name</label>
                            <input
                                v-model="guestName"
                                placeholder="enter a name"
                                maxlength="30"
                                required
                                autofocus
                                class="w-full bg-input border-2 border-border px-3 py-2 text-sm focus:outline-none focus:border-primary font-mono"
                            />
                        </div>
                        <div v-if="room.has_pin">
                            <label class="block text-[11px] font-black uppercase tracking-widest mb-1.5 text-muted-foreground">Room PIN</label>
                            <input
                                v-model="pin"
                                type="text"
                                inputmode="numeric"
                                maxlength="4"
                                placeholder="4 digits"
                                class="w-full bg-input border-2 border-border px-3 py-2 text-sm focus:outline-none focus:border-primary font-mono"
                            />
                        </div>
                        <p v-if="joinError" class="text-xs text-destructive font-mono">{{ joinError }}</p>
                        <button
                            type="submit"
                            :disabled="joining"
                            class="w-full py-2.5 text-sm font-black bg-primary text-primary-foreground border-2 border-primary neo-btn disabled:opacity-60"
                        >{{ joining ? "Joining..." : "Join Room →" }}</button>
                    </form>
                </div>
            </template>

            <div v-else class="text-center text-muted-foreground text-sm font-mono">loading...</div>
        </div>

        <!-- Chat -->
        <div v-else class="flex flex-col h-full w-full max-w-2xl mx-auto px-3 py-3 sm:px-4 sm:py-4">
            <ChatRoom
                :room-name="room!.name"
                :subtitle="`Guest: ${guestName}`"
                :messages="chat.messages.value"
                :online-users="chat.onlineUsers.value"
                :remaining="chat.remaining.value"
                :total-remaining="chat.totalRemaining.value"
                :connection-status="chat.status.value"
                :formatted-remaining="chat.formattedRemaining.value"
                :username="guestName"
                :send-fn="chat.send"
                :retry-fn="chat.retry"
            >
                <template #header-end>
                    <button
                        @click="leaveRoom"
                        class="px-2 py-1 text-xs font-bold border-2 border-destructive text-destructive neo-btn hover:bg-destructive hover:text-white"
                    >Leave</button>
                </template>
            </ChatRoom>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useChatRoom } from "@/composables/useChatRoom";
import { toast } from "vue-sonner";

import ChatRoom from "@/components/ChatRoom.vue";

interface RoomInfo {
    id: string;
    name: string;
    ttl: number;
    expires_at: string;
    has_pin: boolean;
    clients: number;
}

const route = useRoute();
const roomId = route.params.roomId as string;

const room = ref<RoomInfo | null>(null);
const roomError = ref(false);
const roomErrorMessage = ref("");
const guestName = ref("");
const pin = ref("");
const joinError = ref("");
const joining = ref(false);
const joined = ref(false);

const chat = useChatRoom(roomId, () => guestName.value.trim());

function formatExpiry(expiresAt: string): string {
    const diff = new Date(expiresAt).getTime() - Date.now();
    if (diff <= 0) return "expired";
    const h = Math.floor(diff / 3600000);
    const m = Math.floor((diff % 3600000) / 60000);
    if (h > 0) return `expires in ${h}h ${m}m`;
    return `expires in ${m}m`;
}

async function fetchRoom() {
    try {
        const res = await fetch(`/ws/room/${roomId}`);
        if (!res.ok) {
            roomError.value = true;
            roomErrorMessage.value = res.status === 404
                ? "This room may have expired or never existed."
                : "Failed to load room. Please try again later.";
            return;
        }
        room.value = await res.json();
        chat.seedTotal(room.value!.ttl * 60);
    } catch {
        roomError.value = true;
        roomErrorMessage.value = "Network error. Please check your connection.";
    }
}

function joinRoom() {
    if (!guestName.value.trim()) return;
    if (room.value?.has_pin && !pin.value) {
        joinError.value = "PIN is required";
        return;
    }

    joining.value = true;
    joinError.value = "";

    const name = guestName.value.trim();
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    let wsUrl = `${proto}//${window.location.host}/ws/guest/joinRoom/${roomId}?name=${encodeURIComponent(name)}`;
    if (pin.value) wsUrl += `&pin=${encodeURIComponent(pin.value)}`;

    chat.connect(wsUrl);
}

watch(
    () => chat.status.value,
    (s) => {
        if (s === "connected") {
            joining.value = false;
            joined.value = true;
        }
        if (s === "disconnected" && !joined.value) {
            joining.value = false;
            if (room.value?.has_pin) {
                joinError.value = "Invalid PIN. Please try again.";
            } else if (chat.closeCode.value === 1006) {
                joinError.value = "Could not connect to server. Please try again.";
            } else {
                joinError.value = chat.closeReason.value || "Connection refused.";
            }
        }
        if (s === "kicked") {
            toast.error("You have been removed from the room");
            chat.disconnect();
            joined.value = false;
        }
        if (s === "expired") {
            toast.error("Room expired");
            joined.value = false;
        }
    },
);

function leaveRoom() {
    chat.disconnect();
    joined.value = false;
}

fetchRoom();
</script>
