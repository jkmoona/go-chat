<template>
  <Button
    v-bind="$attrs"
    :disabled="loading || $attrs.disabled"
    class="inline-flex items-center justify-center"
  >
    <template v-if="loading">
      <Loader2 class="mr-2 h-4 w-4 animate-spin" />
      <span>{{ loadingText }}</span>
    </template>
    <template v-else>
      <slot />
    </template>
  </Button>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-vue-next";

defineProps<{
  loading?: boolean;
  loadingText?: string;
}>();
</script>