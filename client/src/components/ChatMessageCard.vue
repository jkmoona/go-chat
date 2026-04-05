<template>
    <div
        v-if="message.type === 'system'"
        class="text-center text-xs text-muted-foreground py-1 font-mono"
    >— {{ message.content }} —</div>

    <div
        v-else
        :class="[
            'max-w-[78%] break-words px-3 py-2 my-1.5 border-2',
            isOwnMessage
                ? 'bg-user border-primary shadow-[2px_2px_0_rgba(0,0,0,0.35)] ml-auto'
                : 'bg-other border-foreground/25 shadow-[2px_2px_0_rgba(255,255,255,0.1)] mr-auto',
        ]"
    >
        <span v-if="!isOwnMessage" class="block text-[11px] font-black mb-0.5 opacity-70">{{ message.username }}</span>
        <span class="text-sm">{{ message.content }}</span>
        <span v-if="message.timestamp" class="block text-[10px] opacity-40 mt-0.5 font-mono">{{ formatTime(message.timestamp) }}</span>
    </div>
</template>

<script setup lang="ts">
interface ChatMessage {
    type: string;
    username?: string;
    content: string;
    timestamp?: number;
}

const props = defineProps<{
    message: ChatMessage;
    currentUser: string;
}>();

const isOwnMessage = props.message.username === props.currentUser && props.message.type !== "system";

function formatTime(ts: number): string {
    return new Date(ts).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}
</script>
