<template>
    <div class="fixed inset-0 flex flex-col bg-background text-foreground dark overflow-hidden">
        <div class="flex flex-col h-full w-full max-w-2xl mx-auto px-3 py-3 sm:px-4 sm:py-4">
            <div v-if="!chatReady" class="flex-1 flex items-center justify-center">
                <span class="text-muted-foreground font-mono text-sm">connecting...</span>
            </div>
            <ChatRoom v-else
                :room-name="roomStore.currentRoom.name ?? 'Room'"
                :messages="chat.messages.value"
                :online-users="chat.onlineUsers.value"
                :remaining="chat.remaining.value"
                :total-remaining="chat.totalRemaining.value"
                :connection-status="chat.status.value"
                :formatted-remaining="chat.formattedRemaining.value"
                :username="username"
                :client-count="chat.clientCount.value"
                :is-creator="isCreator"
                :current-user-id="auth.user?.id"
                :send-fn="chat.send"
                :retry-fn="chat.retry"
                @kick="kickUser"
            >
                <template #header-actions>
                    <button
                        v-if="isCreator"
                        @click="showManagement = !showManagement"
                        :class="[
                            'px-2 py-1 text-xs font-bold border-2 neo-btn',
                            showManagement
                                ? 'bg-primary text-primary-foreground border-primary'
                                : 'border-border',
                        ]"
                        aria-label="Room settings"
                    >
                        <Settings class="size-3" />
                    </button>
                    <button
                        @click="copyLink"
                        class="px-2 py-1 text-xs font-bold border-2 border-border neo-btn"
                        aria-label="Copy room link"
                    >
                        <Link2 class="size-3" />
                    </button>
                </template>
                <template #header-end>
                    <button
                        @click="leaveRoom"
                        class="px-2 py-1 text-xs font-bold border-2 border-destructive text-destructive neo-btn"
                    >Leave</button>
                </template>
                <template v-if="isCreator && showManagement" #management>
                    <div class="mb-2 shrink-0 border-b-2 border-border pb-2 space-y-1.5">
                        <div class="flex items-center gap-1.5 flex-wrap">
                            <span class="text-xs font-mono text-muted-foreground mr-1">extend:</span>
                            <button
                                v-for="opt in ttlOptions"
                                :key="opt.value"
                                @click="extendTTL = extendTTL === opt.value ? '' : opt.value"
                                :class="[
                                    'px-2.5 py-1 text-xs font-bold border-2 neo-btn',
                                    extendTTL === opt.value
                                        ? 'bg-primary text-primary-foreground border-primary'
                                        : 'border-border',
                                ]"
                            >{{ opt.label }}</button>
                            <button
                                :disabled="!extendTTL"
                                @click="extendRoom"
                                class="px-2.5 py-1 text-xs font-bold border-2 border-border neo-btn disabled:opacity-30 disabled:cursor-not-allowed"
                            >Extend</button>
                            <button
                                @click="handleDeleteClick"
                                :class="[
                                    'ml-auto px-2.5 py-1 text-xs font-bold border-2 neo-btn',
                                    deleteConfirming
                                        ? 'border-destructive text-destructive bg-destructive/10 animate-pulse'
                                        : 'border-border',
                                ]"
                            >{{ deleteConfirming ? "Sure?" : "Delete Room" }}</button>
                        </div>
                    </div>
                </template>
            </ChatRoom>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useRoomStore } from "@/stores/room";
import { apiFetch } from "@/services/api";
import { parseApiError } from "@/utils/parseError";
import { copyToClipboard } from "@/utils/clipboard";
import { useChatRoom } from "@/composables/useChatRoom";
import { toast } from "vue-sonner";

import { Link2, Settings } from "lucide-vue-next";
import ChatRoom from "@/components/ChatRoom.vue";

const route = useRoute();
const router = useRouter();
const roomStore = useRoomStore();
const auth = useAuthStore();

const roomId = route.params.roomId as string;
const username = auth.user?.username ?? "user";
const isCreator = ref(false);
const showManagement = ref(false);
const extendTTL = ref("");
const deleteConfirming = ref(false);
let deleteTimer: ReturnType<typeof setTimeout> | null = null;

const ttlOptions = [
    { value: "15", label: "+15m" },
    { value: "30", label: "+30m" },
    { value: "60", label: "+1h" },
    { value: "360", label: "+6h" },
];

const chat = useChatRoom(roomId, () => username);
const chatReady = computed(() => chat.status.value !== "idle" && chat.status.value !== "connecting");

watch(
    () => chat.status.value,
    (s) => {
        if (s === "kicked") {
            toast.error("You have been removed from the room");
            setTimeout(() => router.push("/"), 1500);
        }
        if (s === "expired") {
            toast.error("Room expired", { description: "Redirecting..." });
            setTimeout(() => router.push("/"), 2000);
        }
    },
);

function buildWsUrl(): string {
    const pinKey = `room-pin:${roomId}`;
    const pin = (window.history.state?.pin as string | undefined) || sessionStorage.getItem(pinKey) || undefined;
    if (pin) sessionStorage.setItem(pinKey, pin);
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    let url = `${proto}//${window.location.host}/ws/joinRoom/${roomId}`;
    if (pin) url += `?pin=${encodeURIComponent(pin)}`;
    return url;
}

async function copyLink() {
    const link = `${window.location.origin}/join/${roomId}`;
    if (await copyToClipboard(link)) {
        toast.success("Link copied");
    } else {
        toast.info(link, { description: "Copy this link manually" });
    }
}

function leaveRoom() {
    chat.disconnect();
    sessionStorage.removeItem(`room-pin:${roomId}`);
    router.push("/");
}

function handleDeleteClick() {
    if (!deleteConfirming.value) {
        deleteConfirming.value = true;
        deleteTimer = setTimeout(() => {
            deleteConfirming.value = false;
        }, 3000);
    } else {
        if (deleteTimer) clearTimeout(deleteTimer);
        deleteRoom();
    }
}

async function extendRoom() {
    const ttl = Number(extendTTL.value);
    if (!ttl) return;

    try {
        const res = await apiFetch(`/ws/room/${roomId}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ ttl }),
        });
        if (!res.ok) throw new Error(await parseApiError(res, "Failed to extend room"));
        toast.success("Room extended");
        extendTTL.value = "";
    } catch (err: unknown) {
        toast.error(err instanceof Error ? err.message : "Failed to extend room");
    }
}

async function deleteRoom() {
    try {
        const res = await apiFetch(`/ws/room/${roomId}`, { method: "DELETE" });
        if (!res.ok) throw new Error(await parseApiError(res, "Failed to delete room"));
        chat.disconnect();
        toast.success("Room deleted");
        router.push("/");
    } catch (err: unknown) {
        toast.error(err instanceof Error ? err.message : "Failed to delete room");
        deleteConfirming.value = false;
    }
}

async function kickUser(clientId: string) {
    try {
        const res = await apiFetch(`/ws/room/${roomId}/kick`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ client_id: clientId }),
        });
        if (!res.ok) throw new Error(await parseApiError(res, "Failed to kick user"));
    } catch (err: unknown) {
        toast.error(err instanceof Error ? err.message : "Failed to kick user");
    }
}

onMounted(() => {
    const url = buildWsUrl();

    // clear PIN from history before async work so it's never left in state
    if (window.history.state?.pin) {
        const { pin: _, ...rest } = window.history.state;
        void _;
        history.replaceState(rest, "");
    }

    chat.connect(url);

    apiFetch(`/ws/room/${roomId}`).then(async (res) => {
        if (!res.ok) return;
        const data = await res.json();
        roomStore.setRoom(data.id, data.name);
        isCreator.value = data.is_creator ?? false;
        chat.seedRemaining(data.expires_at, data.ttl * 60);
        chat.seedOnlineCount(data.clients);
    }).catch(() => {});
});

onBeforeUnmount(() => {
    if (deleteTimer) clearTimeout(deleteTimer);
});
</script>
