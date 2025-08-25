<template>
    <div class="max-w-2xl mx-auto p-4">
        <h1 class="text-xl font-bold mb-4">Room: {{ roomId }}</h1>

        <div
            class="border rounded h-64 overflow-y-auto p-3 bg-gray-50 mb-4 flex flex-col gap-2"
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
                        'px-3 py-2 rounded max-w-xs break-words',
                        msg.type === 'system'
                            ? 'bg-yellow-100 text-gray-700 italic'
                            : msg.username === username
                            ? 'bg-blue-500 text-white'
                            : 'bg-gray-200 text-gray-900',
                    ]"
                >
                    <span
                        v-if="
                            msg.type !== 'system' && msg.username !== username
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
                class="flex-1 border rounded p-2"
            />
            <button
                :disabled="!newMessage.trim()"
                class="bg-blue-600 text-white px-4 py-2 rounded cursor-pointer"
            >
                Send
            </button>
        </form>
    </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from "vue";
import { useRoute } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { WS_URL } from "../services/constants";

const route = useRoute();
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
