<template>
    <div class="flex flex-col items-center justify-center min-h-screen bg-neutral-900">
        <div class="w-96 p-8">
            <h1 class="text-3xl font-bold mb-6 text-white text-center">Login</h1>
            <form @submit.prevent="login" class="w-full">
                <div class="mb-4">
                    <label
                        for="username"
                        class="block text-sm font-medium text-neutral-300"
                        >Username</label
                    >
                    <input
                        v-model="username"
                        type="username"
                        id="username"
                        required
                        class="mt-1 p-0.5 block w-full bg-neutral-800 border-neutral-600 text-white rounded-md shadow-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                    />
                </div>
                <div class="mb-6">
                    <label
                        for="password"
                        class="block text-sm font-medium text-neutral-300"
                        >Password</label
                    >
                    <input
                        v-model="password"
                        type="password"
                        id="password"
                        required
                        class="mt-1 p-0.5 block w-full bg-neutral-800 border-neutral-600 text-white rounded-md shadow-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
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
            <p class="mt-4 text-sm text-neutral-400 text-center">
                Don't have an account?
                <a href="/register" class="text-blue-400 hover:text-blue-300"
                    >Register</a
                >
            </p>
            <p v-if="error" class="mt-4 text-red-400 text-center">{{ error }}</p>
        </div>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";

const username = ref("");
const password = ref("");
const error = ref("");
const loading = ref(false);
const auth = useAuthStore();
const router = useRouter();

const login = async () => {
    error.value = "";
    loading.value = true;
    try {
        await auth.login(username.value, password.value);
        router.push("/");
    } catch (err) {
        error.value = err.message;
    } finally {
        loading.value = false;
    }
};
</script>
