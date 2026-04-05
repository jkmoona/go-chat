<template>
    <div class="flex justify-center items-center min-h-screen bg-background dark px-4">
        <div class="w-full max-w-sm">
            <!-- Brand -->
            <div class="mb-8">
                <h1 class="text-4xl font-black tracking-tight">TempChat</h1>
                <div class="h-1.5 w-12 bg-primary mt-2"></div>
                <p class="text-xs text-muted-foreground mt-2 font-mono">rooms that don't stick around</p>
            </div>

            <form @submit.prevent="register" autocomplete="off">
                <div class="border-2 border-border p-6 neo-card">
                    <h2 class="text-lg font-black mb-5 tracking-tight">Sign Up</h2>

                    <div class="space-y-4">
                        <div>
                            <label class="block text-[11px] font-black uppercase tracking-widest mb-1.5 text-muted-foreground">Username</label>
                            <input
                                type="text"
                                v-model="username"
                                placeholder="yourname"
                                minlength="3"
                                maxlength="30"
                                pattern="[a-zA-Z0-9]+"
                                required
                                class="w-full bg-input border-2 border-border px-3 py-2 text-sm focus:outline-none focus:border-primary font-mono"
                            />
                            <p class="text-[11px] text-muted-foreground mt-1 font-mono">letters and numbers, 3–30 chars</p>
                            <p v-if="usernameError" class="text-xs text-destructive mt-1 font-mono">{{ usernameError }}</p>
                        </div>
                        <div>
                            <label class="block text-[11px] font-black uppercase tracking-widest mb-1.5 text-muted-foreground">Password</label>
                            <input
                                type="password"
                                v-model="password"
                                minlength="6"
                                required
                                class="w-full bg-input border-2 border-border px-3 py-2 text-sm focus:outline-none focus:border-primary"
                            />
                            <p v-if="passwordError" class="text-xs text-destructive mt-1 font-mono">{{ passwordError }}</p>
                        </div>
                        <button
                            type="submit"
                            :disabled="loading"
                            class="w-full py-2.5 text-sm font-black bg-primary text-primary-foreground border-2 border-primary neo-btn disabled:opacity-60"
                        >{{ loading ? "Creating account..." : "Sign Up →" }}</button>
                    </div>

                    <p class="mt-5 text-xs text-center text-muted-foreground font-mono">
                        have an account?
                        <RouterLink to="/login" class="text-primary font-black">sign in</RouterLink>
                    </p>
                </div>
            </form>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";

const username = ref("");
const password = ref("");
const usernameError = ref("");
const passwordError = ref("");
const loading = ref(false);
const auth = useAuthStore();
const router = useRouter();

function validate(): boolean {
    usernameError.value = "";
    passwordError.value = "";

    if (username.value.length < 3) {
        usernameError.value = "Username must be at least 3 characters";
    } else if (username.value.length > 30) {
        usernameError.value = "Username must be at most 30 characters";
    } else if (!/^[a-zA-Z0-9]+$/.test(username.value)) {
        usernameError.value = "Username must contain only letters and numbers";
    }

    if (password.value.length < 6) {
        passwordError.value = "Password must be at least 6 characters";
    }

    return !usernameError.value && !passwordError.value;
}

const register = async () => {
    if (!validate()) return;
    loading.value = true;
    try {
        await auth.register(username.value, password.value);
        toast.success("Registration successful!", { description: "Redirecting to login...", duration: 1500 });
        await router.push("/login");
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.error("Registration failed", { description: message });
    } finally {
        loading.value = false;
    }
};
</script>
