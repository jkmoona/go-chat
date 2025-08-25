<template>
    <div
        class="flex flex-col items-center justify-center min-h-screen bg-neutral-900"
    >
        <div class="w-96 p-8">
            <h2 class="text-3xl font-bold mb-6 text-white text-center">
                Register
            </h2>
            <form @submit.prevent="register" class="w-full">
                <div class="mb-4">
                    <label
                        for="username"
                        class="block text-sm font-medium text-neutral-300"
                        >Username</label
                    >
                    <input
                        v-model="username"
                        type="text"
                        id="username"
                        required
                        class="mt-1 p-0.5 block w-full bg-neutral-800 border-neutral-600 text-white rounded-md shadow-sm focus:border-blue-500"
                    />
                </div>
                <div class="mb-4">
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
                        class="mt-1 p-0.5 block w-full bg-neutral-800 border-neutral-600 text-white rounded-md shadow-sm focus:border-blue-500"
                    />
                </div>
                <button
                    type="submit"
                    class="w-full bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition duration-200 cursor-pointer"
                    :disabled="loading"
                >
                    {{ loading ? "Registering..." : "Register" }}
                </button>
            </form>
            <p class="mt-4 text-sm text-neutral-400 text-center">
                Already have an account?
                <a href="/login" class="text-blue-400 hover:text-blue-300"
                    >Login</a
                >
            </p>
            <p v-if="error" class="mt-4 text-red-400 text-center">
                {{ error }}
            </p>
            <p v-if="success" class="mt-4 text-green-400 text-center">
                {{ success }}
            </p>
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
const success = ref("");
const loading = ref(false);
const auth = useAuthStore();
const router = useRouter();

const register = async () => {
    error.value = "";
    success.value = "";
    loading.value = true;
    try {
        await auth.register(username.value, password.value);

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
