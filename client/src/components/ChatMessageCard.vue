<template>
    <div
        v-if="message.type === 'system'"
        class="text-center text-xs text-muted-foreground py-1 font-mono"
    >— {{ message.content }} —</div>

    <div
        v-else
        :class="[
            'max-w-[78%] break-words px-2.5 py-1.5 my-1 border-2',
            isOwnMessage
                ? 'bg-user border-primary ml-auto'
                : 'bg-other border-foreground/25 mr-auto',
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
