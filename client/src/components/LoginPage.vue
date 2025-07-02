<template>
    <div class="flex flex-col items-center justify-center h-screen bg-neutral-900">
        <h1 class="text-3xl font-bold mb-6">Login</h1>
        <form @submit.prevent="login" class="w-80">
            <div class="mb-4">
                <label
                    for="email"
                    class="block text-sm font-medium text-neutral-100"
                    >Email</label
                >
                <input
                    v-model="email"
                    type="email"
                    id="email"
                    required
                    class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500"
                />
            </div>
            <div class="mb-6">
                <label
                    for="password"
                    class="block text-sm font-medium text-neutral-100"
                    >Password</label
                >
                <input
                    v-model="password"
                    type="password"
                    id="password"
                    required
                    class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500"
                />
            </div>
            <button
                type="submit"
                class="w-full bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition duration-200"
                :disabled="loading"
            >
                {{ loading ? "Logging in..." : "Login" }}
            </button>
        </form>
        <p class="mt-4 text-sm text-gray-600">
            Don't have an account?
            <a href="/signup" class="text-blue-600 hover:underline"
                >Register</a
            >
        </p>
        <p v-if="error" class="mt-4 text-red-600">{{ error }}</p>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";
const email = ref("");
const password = ref("");
const error = ref("");
const loading = ref(false);
const auth = useAuthStore();
const router = useRouter();

const login = async () => {
    error.value = "";
    loading.value = true;
    try {
        await auth.login(email.value, password.value);
        router.push("/"); // Redirect to main page after login
    } catch (err) {
        error.value = err.message;
    } finally {
        loading.value = false;
    }
};
</script>

<style scoped>
/* Add any component-specific styles here */
</style>
