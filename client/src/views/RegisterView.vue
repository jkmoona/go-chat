<template>
    <div
        class="flex justify-center items-center min-h-screen bg-background dark"
    >
        <form @submit.prevent="register" autocomplete="off">
            <Card class="w-96">
                <CardHeader>
                    <CardTitle class="text-2xl"> Register </CardTitle>
                    <CardDescription>
                        Enter your username to create an account
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div class="grid gap-4">
                        <div class="grid gap-2">
                            <Label for="username">Username</Label>
                            <Input
                                id="username"
                                type="username"
                                placeholder="guest"
                                v-model="username"
                                required
                            />
                        </div>
                        <Label for="password">Password</Label>
                        <Input
                            id="password"
                            type="password"
                            v-model="password"
                            required
                        />
                        <LoadingButton
                            type="submit"
                            class="w-full"
                            :loading="loading"
                            loadingText="Creating account..."
                            >Sign Up</LoadingButton
                        >
                    </div>
                    <div class="mt-4 text-center text-sm">
                        Already have an account?
                        <RouterLink to="/login" class="underline">
                            Sign in
                        </RouterLink>
                    </div>
                </CardContent>
            </Card>
        </form>
    </div>
</template>

<script setup lang="ts">
import LoadingButton from "@/components/LoadingButton.vue";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "vue-sonner";

import { ref } from "vue";
import { useAuthStore } from "@/stores/auth";
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

        toast.success("Registration successful! Redirecting to login...", {
            description: "You can now log in with your new account.",
            duration: 1500,
        });

        setTimeout(() => {
            router.push("/login");
            loading.value = false;
        }, 1000);
    } catch (err: unknown) {
        if (err instanceof Error) {
            error.value = err.message;
        } else {
            error.value = String(err);
        }

        toast.error("Registration failed", {
            description: error.value,
        });
        loading.value = false;
    }
};
</script>
