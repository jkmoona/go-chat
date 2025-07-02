<template>
    <div class="max-w-md mx-auto mt-10 p-6 bg-white rounded-lg shadow-md">
        <h2 class="text-2xl font-bold mb-6">Register</h2>
        <form @submit.prevent="register">
            <div class="mb-4">
                <label
                    for="username"
                    class="block text-sm font-medium text-gray-700"
                    >Username</label
                >
                <input
                    v-model="username"
                    type="text"
                    id="username"
                    required
                    class="mt-1 block w-full border-gray-300 rounded-md shadow-sm"
                />
            </div>
            <div class="mb-4">
                <label
                    for="email"
                    class="block text-sm font-medium text-gray-700"
                    >Email</label
                >
                <input
                    v-model="email"
                    type="email"
                    id="email"
                    required
                    class="mt-1 block w-full border-gray-300 rounded-md shadow-sm"
                />
            </div>
            <div class="mb-4">
                <label
                    for="password"
                    class="block text-sm font-medium text-gray-700"
                    >Password</label
                >
                <input
                    v-model="password"
                    type="password"
                    id="password"
                    required
                    class="mt-1 block w-full border-gray-300 rounded-md shadow-sm"
                />
            </div>
            <button
                type="submit"
                class="w-full bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
            >
                Register
            </button>
        </form>
        <p class="mt-4 text-sm text-gray-600">
            Already have an account?
            <a href="/login" class="text-blue-600 hover:underline">Login</a>
        </p>
        <p v-if="error" class="mt-4 text-red-600">{{ error }}</p>
        <p v-if="success" class="mt-4 text-green-600">{{ success }}</p>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";
const username = ref("");
const email = ref("");
const password = ref("");
const error = ref("");
const success = ref("");
const loading = ref(false);
const auth = useAuthStore();
const router = useRouter();

const register = async () => {
    error.value = "";
    success.value = "";
    loading.value = true;
    try {
        await auth.register(username.value, email.value, password.value);
        success.value = "Registration successful! Redirecting to login...";
        setTimeout(() => {
            router.push("/login");
        }, 1200);
    } catch (err) {
        error.value = err.message;
    } finally {
        loading.value = false;
    }
};
</script>
