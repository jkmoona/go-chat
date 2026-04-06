<script setup lang="ts">
import { onMounted, onBeforeUnmount } from 'vue';
import { RouterView } from 'vue-router';
import { Toaster } from '@/components/ui/sonner'
import 'vue-sonner/style.css'

function updateAppHeight() {
    const h = window.visualViewport?.height ?? window.innerHeight;
    document.documentElement.style.setProperty('--app-height', `${h}px`);
}

onMounted(() => {
    updateAppHeight();
    window.visualViewport?.addEventListener('resize', updateAppHeight);
});

onBeforeUnmount(() => {
    window.visualViewport?.removeEventListener('resize', updateAppHeight);
});
</script>

<template>  
    <RouterView />
    <Toaster position="top-right" richColors theme="dark" />
</template>

