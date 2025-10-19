<template>
    <div
        class="flex justify-center items-center min-h-screen bg-background dark"
    >
        <form @submit.prevent="login" autocomplete="off">
            <Card class="w-96">
                <CardHeader>
                    <CardTitle class="text-2xl"> Login </CardTitle>
                    <CardDescription>
                        Enter your username to login to your account
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
                            loadingText="Logging in..."
                        >
                            Login
                        </LoadingButton>
                    </div>
                    <div class="mt-4 text-center text-sm">
                        Don't have an account?
                        <RouterLink to="/register" class="underline">
                            Sign up
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
const loading = ref(false);
const auth = useAuthStore();
const router = useRouter();

const login = async () => {
    error.value = "";
    loading.value = true;
    try {
        await auth.login(username.value, password.value);

        toast.success("Login successfull!", {
            description: "Redirecting to the dashboard...",
            duration: 1500,
        });

        setTimeout(() => {
            router.push("/");
            loading.value = false;
        }, 1000);
    } catch (err: unknown) {
        if (err instanceof Error) {
            error.value = err.message;
        } else {
            error.value = String(err);
        }

        toast.error("Login failed", {
            description: error.value,
        });
        loading.value = false;
    }
};
</script>
