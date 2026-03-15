<template>
    <div class="bg-background text-foreground h-dvh overflow-hidden dark">
        <div class="max-w-2xl mx-auto p-6 flex flex-col h-full min-h-0">

            <!-- Header -->
            <div class="flex justify-between items-center mb-6 shrink-0">
                <h1 class="text-2xl font-bold">TempChat</h1>
                <LoadingButton @click="logout" variant="destructive" :loading="loading" loading-text="Logging Out...">
                    <LogOut class="size-4" /> Log Out
                </LoadingButton>
            </div>

            <!-- Create Room -->
            <Card class="mb-6 shrink-0">
                <CardHeader>
                    <CardTitle class="text-lg">New Room</CardTitle>
                </CardHeader>
                <CardContent>
                    <form @submit.prevent="createRoom" class="space-y-3">
                        <div class="flex gap-2">
                            <Input
                                v-model="newRoomName"
                                placeholder="Room name"
                                required
                                class="flex-1"
                            />
                            <select
                                v-model="newRoomTTL"
                                class="bg-input border border-border rounded-md px-3 py-2 text-sm"
                            >
                                <option :value="15">15 min</option>
                                <option :value="30">30 min</option>
                                <option :value="60">1 hour</option>
                                <option :value="360">6 hours</option>
                                <option :value="1440">24 hours</option>
                            </select>
                        </div>
                        <div class="flex items-center gap-3">
                            <label class="flex items-center gap-2 text-sm cursor-pointer select-none">
                                <input
                                    type="checkbox"
                                    v-model="enablePIN"
                                    class="rounded"
                                />
                                PIN protection
                            </label>
                            <Input
                                v-if="enablePIN"
                                v-model="newRoomPIN"
                                type="text"
                                inputmode="numeric"
                                maxlength="4"
                                placeholder="4-digit PIN"
                                class="w-32"
                            />
                        </div>
                        <Button type="submit" class="w-full">Create</Button>
                    </form>
                </CardContent>
            </Card>

            <!-- Rooms -->
            <Card class="flex-1 min-h-0 flex flex-col m-0">
                <CardHeader class="shrink-0">
                    <CardTitle class="text-lg">Active Rooms</CardTitle>
                </CardHeader>
                <CardContent class="flex-1 min-h-0 flex flex-col space-y-2 overflow-y-auto pt-0.5">
                    <div
                        v-for="room in roomStore.rooms"
                        :key="room.id"
                        class="transition-transform hover:scale-[1.02] hover:shadow-lg rounded-md"
                    >
                        <Card class="p-4">
                            <div class="flex flex-row justify-between items-center">
                                <div class="flex-1 min-w-0">
                                    <div class="flex items-center gap-2">
                                        <p class="font-medium truncate">{{ room.name }}</p>
                                        <Lock v-if="room.has_pin" class="size-3 text-muted-foreground shrink-0" />
                                    </div>
                                    <p class="text-xs text-muted-foreground mt-0.5">
                                        {{ room.clients }} online · {{ formatExpiry(room.expires_at) }}
                                    </p>
                                </div>
                                <div class="flex items-center gap-2 ml-4 shrink-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        @click="copyLink(room.id)"
                                    >
                                        <Link2 class="size-3" />
                                    </Button>
                                    <Button
                                        variant="default"
                                        size="sm"
                                        @click="handleJoin(room)"
                                    >
                                        Join
                                    </Button>
                                </div>
                            </div>

                            <!-- Inline PIN form for protected rooms -->
                            <div v-if="pinTarget === room.id" class="mt-3 flex gap-2">
                                <Input
                                    v-model="enteredPIN"
                                    type="text"
                                    inputmode="numeric"
                                    maxlength="4"
                                    placeholder="Enter PIN"
                                    class="flex-1"
                                    @keydown.enter="confirmJoin(room)"
                                    autofocus
                                />
                                <Button size="sm" @click="confirmJoin(room)">Go</Button>
                                <Button size="sm" variant="ghost" @click="pinTarget = null">✕</Button>
                            </div>
                        </Card>
                    </div>

                    <p v-if="roomStore.rooms.length === 0 && !loadError" class="text-sm text-muted-foreground text-center py-4">
                        No active rooms. Create one above.
                    </p>

                    <Alert v-if="loadError" variant="destructive" class="mt-4">
                        <p>{{ loadError }}</p>
                    </Alert>
                </CardContent>
            </Card>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { apiFetch } from "@/services/api";
import { useRoomStore } from "@/stores/room";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Alert } from "@/components/ui/alert";
import { toast } from "vue-sonner";
import { LogOut, Lock, Link2 } from "lucide-vue-next";
import LoadingButton from "@/components/LoadingButton.vue";

interface Room {
    id: string;
    name: string;
    ttl: number;
    expires_at: string;
    has_pin: boolean;
    clients: number;
}

const newRoomName = ref("");
const newRoomTTL = ref(60);
const enablePIN = ref(false);
const newRoomPIN = ref("");
const loadError = ref("");
const loading = ref(false);
const pinTarget = ref<string | null>(null);
const enteredPIN = ref("");

const router = useRouter();
const auth = useAuthStore();
const roomStore = useRoomStore();

function formatExpiry(expiresAt: string): string {
    const diff = new Date(expiresAt).getTime() - Date.now();
    if (diff <= 0) return "expired";
    const h = Math.floor(diff / 3600000);
    const m = Math.floor((diff % 3600000) / 60000);
    if (h > 0) return `${h}h ${m}m left`;
    return `${m}m left`;
}

async function fetchRooms() {
    try {
        const res = await apiFetch("/ws/getRooms");
        if (!res.ok) {
            if (res.status === 401) {
                await auth.logout();
                router.push("/login");
                return;
            }
            throw new Error("Failed to load rooms");
        }
        roomStore.rooms = await res.json();
    } catch (err: unknown) {
        loadError.value = err instanceof Error ? err.message : String(err);
    }
}

async function createRoom() {
    const pin = enablePIN.value ? newRoomPIN.value.trim() : undefined;
    if (enablePIN.value && (!pin || pin.length !== 4 || !/^\d{4}$/.test(pin))) {
        toast.error("PIN must be exactly 4 digits");
        return;
    }

    try {
        const body: Record<string, unknown> = { name: newRoomName.value, ttl: newRoomTTL.value };
        if (pin) body.pin = pin;

        const res = await apiFetch("/ws/createRoom", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        });

        if (!res.ok) {
            const errData = await res.json();
            throw new Error(errData.error || "Failed to create room");
        }

        const room: Room = await res.json();
        newRoomName.value = "";
        newRoomPIN.value = "";
        enablePIN.value = false;

        toast.success(`Room "${room.name}" created`, {
            description: `Link copied to clipboard`,
            duration: 3000,
        });

        await navigator.clipboard.writeText(`${window.location.origin}/join/${room.id}`);
        fetchRooms();
    } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : String(err);
        toast.error("Failed to create room", { description: msg });
    }
}

function handleJoin(room: Room) {
    if (room.has_pin) {
        if (pinTarget.value === room.id) {
            pinTarget.value = null;
        } else {
            pinTarget.value = room.id;
            enteredPIN.value = "";
        }
        return;
    }
    roomStore.setRoom(room.id, room.name);
    router.push({ name: "Room", params: { roomId: room.id } });
}

function confirmJoin(room: Room) {
    if (!/^\d{4}$/.test(enteredPIN.value)) {
        toast.error("PIN must be exactly 4 digits");
        return;
    }
    const pin = enteredPIN.value;
    pinTarget.value = null;
    enteredPIN.value = "";
    roomStore.setRoom(room.id, room.name);
    router.push({ name: "Room", params: { roomId: room.id }, state: { pin } });
}

async function copyLink(roomId: string) {
    await navigator.clipboard.writeText(`${window.location.origin}/join/${roomId}`);
    toast.success("Link copied");
}

function logout() {
    loading.value = true;
    setTimeout(() => {
        auth.logout();
        router.push("/login");
        loading.value = false;
    }, 1000);
}

onMounted(fetchRooms);
</script>
