<template>
    <div class="flex flex-col h-dvh bg-background text-foreground dark">
        <div class="flex flex-col h-full w-full max-w-2xl mx-auto p-4">
            <div class="flex justify-between items-center mb-4">
                <h1 class="text-xl font-bold">
                    {{ roomStore.currentRoom.name }} :
                    {{ roomStore.currentRoom.id }}
                </h1>
                <LoadingButton
                    @click="leaveRoom"
                    variant="destructive"
                    loading-text="Leaving..."
                >
                    <LogOut class="size-4" /> Leave
                </LoadingButton>
            </div>

            <Card
                class="flex-1 gap-1 overflow-y-auto p-3 mb-4 space-y-2 bg-[oklch(0.22_0.006_286)] rounded-lg"
            >
                <ChatMessageCard
                    v-for="(msg, index) in messages"
                    :key="index"
                    :message="msg"
                    :current-user="username"
                />
                <div ref="bottomEl"></div>
            </Card>

            <form @submit.prevent="send" class="flex gap-2">
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
import { ref, onMounted, onBeforeUnmount, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useRoomStore } from "@/stores/room";
import { WS_URL } from "@/services/constants";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { LogOut } from "lucide-vue-next";

import LoadingButton from "@/components/LoadingButton.vue";
import ChatMessageCard from "@/components/ChatMessageCard.vue";

const route = useRoute();
const router = useRouter();
const roomStore = useRoomStore();
const roomId = route.params.roomId;
const auth = useAuthStore();
const loading = ref(false);

const username = auth.user?.username || "guest";
const userId = auth.user?.id || "guest";

interface ChatMessage {
    type: string;
    username?: string;
    content: string;
    roomId?: string;
}

const newMessage = ref<string>("");
const messages = ref<ChatMessage[]>([]);
const bottomEl = ref<HTMLElement | null>(null);
const inputEl = ref<HTMLInputElement | null>(null);
let socket: WebSocket | null = null;

function scrollToBottom() {
    bottomEl.value?.scrollIntoView({ behavior: "smooth" });
}

async function connectSocket() {
    await auth.tryRefresh();

    if (socket) {
        socket.close();
        socket = null;
    }

    const wsUrl = `${WS_URL}/ws/joinRoom/${roomId}?userId=${userId}&username=${username}`;
    socket = new WebSocket(wsUrl);

    socket.onopen = () => console.log("WebSocket connected");
    socket.onmessage = async (event: MessageEvent) => {
        let msg: ChatMessage;
        try {
            msg = JSON.parse(event.data);
        } catch {
            msg = { type: "system", content: event.data };
        }
        messages.value.push(msg);
        if (messages.value.length > 500) messages.value.shift();
        await nextTick();
        scrollToBottom();
    };
    socket.onerror = (err: Event) => console.error("WebSocket error:", err);
    socket.onclose = () => console.log("WebSocket closed");
}

async function send() {
    if (!newMessage.value.trim()) return;

    if (!socket || socket.readyState !== WebSocket.OPEN) {
        messages.value.push({
            type: "system",
            content: "Connection lost. Trying to reconnect...",
        });
        connectSocket();
        return;
    }

    const msg: ChatMessage = {
        type: "chat",
        username,
        content: newMessage.value,
        roomId: roomId as string,
    };

    try {
        socket.send(JSON.stringify(msg));
        messages.value.push({ ...msg });
        await nextTick();
        scrollToBottom();
        newMessage.value = "";
    } catch (err) {
        messages.value.push({
            type: "system",
            content: "Failed to send message. Please try again.",
        });
        console.error("Failed to send message:", err);
    }

    newMessage.value = "";
}

const leaveRoom = () => {
    if (socket) {
        socket.close();
        socket = null;
    }
    loading.value = true;
    setTimeout(() => {
        router.push("/");
        loading.value = false;
    }, 1000);
};

onMounted(() => {
    connectSocket();
    inputEl.value?.focus();
});

onBeforeUnmount(() => {
    if (socket) {
        socket.close();
        socket = null;
    }
});
</script>
