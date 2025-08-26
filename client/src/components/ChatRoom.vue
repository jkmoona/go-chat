<template>
    <div class="flex flex-col h-dvh bg-neutral-900 text-neutral-200">
        <div class="flex flex-col h-full w-full max-w-2xl mx-auto p-4">
            <div class="flex justify-between items-center mb-4">
                <h1 class="text-xl font-bold">
                    {{ roomStore.currentRoom.name }} :
                    {{ roomStore.currentRoom.id }}
                </h1>
                <button
                    @click="leaveRoom"
                    class="bg-red-600 px-3 py-1 rounded cursor-pointer"
                >
                    Leave
                </button>
            </div>

            <div
                class="flex-1 bg-neutral-900 border rounded overflow-y-auto p-3 mb-4 space-y-1"
            >
                <div
                    v-for="(msg, index) in messages"
                    :key="index"
                    :class="[
                        'flex',
                        msg.type === 'system'
                            ? 'justify-center'
                            : msg.username === username
                            ? 'justify-end'
                            : 'justify-start',
                    ]"
                >
                    <div
                        :class="[
                            'rounded max-w-xs break-words',
                            msg.type === 'system'
                                ? 'text-neutral-300 text-sm italic px-1 py-1'
                                : msg.username === username
                                ? 'bg-blue-500 px-3 py-2'
                                : 'bg-neutral-300 text-neutral-900 px-3 py-2',
                        ]"
                    >
                        <span
                            v-if="
                                msg.type !== 'system' &&
                                msg.username !== username
                            "
                            class="font-semibold mr-2"
                            >{{ msg.username }}:</span
                        >
                        {{ msg.content }}
                    </div>
                </div>
                <div ref="bottomEl"></div>
            </div>

            <form @submit.prevent="send" class="flex gap-2">
                <input
                    v-model="newMessage"
                    ref="inputEl"
                    placeholder="Type a message..."
                    class="flex-1 border rounded p-2 focus:outline-none"
                />
                <button
                    :disabled="!newMessage.trim()"
                    class="bg-blue-600 px-4 py-2 rounded cursor-pointer"
                >
                    Send
                </button>
            </form>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { useRoomStore } from "../stores/room";
import { WS_URL } from "../services/constants";

const route = useRoute();
const router = useRouter();
const roomStore = useRoomStore();
const roomId = route.params.roomId;
const newMessage = ref("");
const messages = ref([]);
const bottomEl = ref(null);
const inputEl = ref(null);
let socket;

const auth = useAuthStore();
const username = auth.user?.username || "guest";
const userId = auth.user?.id || "guest";

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

    socket.onopen = () => console.log("✅ WebSocket connected");
    socket.onmessage = async (event) => {
        let msg;
        try {
            msg = JSON.parse(event.data);
        } catch {
            msg = { type: "system", content: msg.content };
        }
        messages.value.push(msg);
        await nextTick();
        scrollToBottom();
    };
    socket.onerror = (err) => console.error("WebSocket error:", err);
    socket.onclose = async () => {
        console.log("❌ WebSocket closed");
    };
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

    const msg = {
        type: "chat",
        username,
        content: newMessage.value,
        roomId,
    };

    try {
        socket.send(JSON.stringify(msg));
        messages.value.push({
            type: "chat",
            username,
            content: newMessage.value,
        });
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
    router.push("/");
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
