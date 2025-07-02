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
                        v-if="msg.type !== 'system' && msg.username !== username"
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
                class="bg-blue-600 text-white px-4 py-2 rounded"
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

function connectSocket() {
    const wsUrl = `ws://localhost:8080/ws/joinRoom/${roomId}`;
    socket = new WebSocket(wsUrl);

    socket.onopen = () => console.log("✅ WebSocket connected");
    socket.onmessage = async (event) => {
        let msg;
        try {
            msg = JSON.parse(event.data);
        } catch {
            msg = { type: "system", content: event.data };
        }
        messages.value.push(msg);
        await nextTick();
        scrollToBottom();
    };
    socket.onerror = (err) => console.error("WebSocket error:", err);
    socket.onclose = () => console.log("❌ WebSocket closed");
}

function send() {
    if (!newMessage.value.trim()) return;
    const msg = {
        type: "chat",
        username,
        content: newMessage.value,
        roomId,
    };
    socket.send(JSON.stringify(msg));
    newMessage.value = "";
}

onMounted(() => {
    connectSocket();
    inputEl.value?.focus();
});

onBeforeUnmount(() => {
    socket?.close();
});
</script>
