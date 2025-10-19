<template>
  <div
    v-if="message.type === 'system'"
    class="w-full text-center text-sm italic text-muted-foreground my-1"
  >
    {{ message.content }}
  </div>

  <div
    v-else
    :class="[
      'max-w-xs break-words rounded px-3 py-2 my-1',
      isOwnMessage ? 'bg-user text-primary-foreground ml-auto' : 'bg-other text-secondary-foreground mr-auto',
    ]"
  >
    <span v-if="!isOwnMessage" class="font-semibold mr-1">{{ message.username }}:</span>
    {{ message.content }}
  </div>
</template>

<script setup lang="ts">
interface ChatMessage {
  type: string;
  username?: string;
  content: string;
}

const props = defineProps<{
  message: ChatMessage;
  currentUser: string;
}>();

const isOwnMessage = props.message.username === props.currentUser && props.message.type !== 'system';
</script>
