<template>
    <div class="flex flex-col h-full">
        <!-- Header -->
        <div class="flex justify-between items-center mb-2 shrink-0 pb-2 border-b-2 border-border">
            <div class="min-w-0 border-l-4 border-primary pl-3">
                <h1 class="text-lg font-black tracking-tight truncate">{{ roomName }}</h1>
                <p v-if="subtitle" class="text-xs text-muted-foreground font-mono">{{ subtitle }}</p>
            </div>
            <div class="flex items-center gap-1.5 ml-2 shrink-0">
                <slot name="header-actions" />
                <button
                    @click="showUsers = !showUsers"
                    :class="[
                        'px-2 py-1 text-xs font-bold border-2 neo-btn',
                        showUsers
                            ? 'bg-primary text-primary-foreground border-primary'
                            : 'border-border hover:border-primary/80',
                    ]"
                    aria-label="Toggle online users"
                >{{ onlineUsers.length }} online</button>
                <slot name="header-end" />
            </div>
        </div>

        <slot name="management" />

        <!-- Connection status -->
        <div
            v-if="connectionStatus === 'connecting' && onlineUsers.length > 0"
            class="mb-2 shrink-0 border-l-4 border-yellow-500 pl-2 py-0.5 text-xs text-yellow-500 font-mono"
        >reconnecting...</div>
        <div
            v-else-if="connectionStatus === 'disconnected'"
            class="mb-2 shrink-0 border-l-4 border-destructive pl-2 py-0.5 text-xs text-destructive font-mono flex items-center gap-2"
        >
            connection lost
            <button
                @click="retryFn"
                class="font-bold border border-destructive px-1.5 py-0.5 hover:bg-destructive hover:text-white neo-btn text-[10px]"
            >retry</button>
        </div>

        <!-- Online users panel -->
        <div v-if="showUsers" class="mb-2 shrink-0 border-2 border-border p-2 bg-card">
            <p v-if="onlineUsers.length === 0" class="text-xs text-muted-foreground font-mono">no presence yet</p>
            <div class="flex flex-wrap gap-1.5">
                <span
                    v-for="u in onlineUsers"
                    :key="u.id"
                    class="flex items-center gap-1 px-2 py-0.5 text-xs font-bold border-2 border-border bg-muted"
                >
                    {{ u.username }}<span v-if="u.is_guest" class="text-muted-foreground font-normal">*</span>
                    <button
                        v-if="isCreator && u.id !== currentUserId"
                        @click="$emit('kick', u.id)"
                        class="text-muted-foreground hover:text-destructive ml-0.5"
                        aria-label="Kick user"
                    >
                        <X class="size-3" />
                    </button>
                </span>
            </div>
        </div>

        <!-- Messages -->
        <div class="flex-1 overflow-y-auto border-2 border-border p-3 mb-2 space-y-1.5 bg-card">
            <ChatMessageCard
                v-for="(msg, index) in messages"
                :key="index"
                :message="msg"
                :current-user="username"
            />
            <div ref="bottomEl"></div>
        </div>

        <!-- Countdown -->
        <div v-if="remaining > 0" class="mb-1.5 shrink-0 flex items-center gap-2">
            <span
                class="text-xs font-mono font-bold tabular-nums shrink-0"
                :class="remaining < 60 ? 'text-destructive' : 'text-muted-foreground'"
            >{{ formattedRemaining }}</span>
            <div class="flex-1 h-1 bg-muted overflow-hidden">
                <div
                    class="h-full transition-all duration-1000 ease-linear"
                    :class="remaining < 60 ? 'bg-destructive animate-pulse' : 'bg-primary'"
                    :style="{ width: totalRemaining ? `${(remaining / totalRemaining) * 100}%` : '100%' }"
                ></div>
            </div>
        </div>

        <!-- Input -->
        <form @submit.prevent="handleSend" class="flex gap-0 shrink-0">
            <input
                v-model="newMessage"
                ref="inputEl"
                placeholder="type a message..."
                autocomplete="off"
                @focus="onInputFocus"
                class="flex-1 bg-input border-2 border-border border-r-0 px-3 py-2 text-base focus:outline-none focus:border-primary font-mono disabled:opacity-40"
                :disabled="connectionStatus !== 'connected'"
            />
            <button
                type="submit"
                :disabled="!newMessage.trim() || connectionStatus !== 'connected'"
                class="px-4 py-2 text-sm font-black bg-primary text-primary-foreground border-2 border-primary neo-btn disabled:opacity-40 disabled:cursor-not-allowed shrink-0"
            >↑</button>
        </form>
    </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from "vue";
import type { ChatMessage, ClientInfo, ConnectionStatus } from "@/composables/useChatRoom";

import { X } from "lucide-vue-next";
import ChatMessageCard from "@/components/ChatMessageCard.vue";

const props = defineProps<{
    roomName: string;
    subtitle?: string;
    messages: ChatMessage[];
    onlineUsers: ClientInfo[];
    remaining: number;
    totalRemaining: number;
    connectionStatus: ConnectionStatus;
    formattedRemaining: string;
    username: string;
    isCreator?: boolean;
    currentUserId?: string;
    sendFn: (content: string) => boolean;
    retryFn: () => void;
}>();

defineEmits<{
    kick: [clientId: string];
}>();

const showUsers = ref(false);
const newMessage = ref("");
const bottomEl = ref<HTMLElement | null>(null);
const inputEl = ref<HTMLInputElement | null>(null);

function handleSend() {
    if (!newMessage.value.trim()) return;
    if (props.sendFn(newMessage.value)) {
        newMessage.value = "";
    }
}

watch(
    () => props.messages.length,
    () => scrollToBottom(),
);

function scrollToBottom() {
    nextTick(() => bottomEl.value?.scrollIntoView({ behavior: "smooth" }));
}

function onInputFocus() {
    setTimeout(() => {
        nextTick(() => bottomEl.value?.scrollIntoView({ behavior: "instant" }));
    }, 400);
}

onMounted(() => {
    if (!window.matchMedia("(hover: none)").matches) {
        inputEl.value?.focus();
    }
});
</script>
