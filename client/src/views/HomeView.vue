<template>
    <div class="bg-background text-foreground h-dvh overflow-hidden dark">
        <div class="max-w-2xl mx-auto p-6 flex flex-col h-full min-h-0">

            <!-- Header -->
            <div class="flex justify-between items-center mb-6 shrink-0">
                <h1 class="text-2xl font-bold">Chat Rooms</h1>
                 <LoadingButton @click="logout" variant="destructive" :loading="loading" loading-text="Logging Out...">
                    <LogOut class="size-4" /> Log Out
                </LoadingButton>
            </div>

            <!-- Create Room -->
            <Card class="mb-6 shrink-0">
                <CardHeader>
                    <CardTitle class="text-lg">Create a Room</CardTitle>
                </CardHeader>
                <CardContent>
                    <form @submit.prevent="createRoom" class="flex gap-2">
                        <Input
                            v-model="newRoomName"
                            placeholder="Room Name"
                            required
                            class="flex-1"
                        />
                        <Button> Create </Button>
                    </form>
                </CardContent>
            </Card>

            <!-- Rooms -->
            <Card class="flex-1 min-h-0 flex flex-col">
                <CardHeader class="shrink-0">
                    <CardTitle class="text-lg">Available Rooms</CardTitle>
                </CardHeader>
                <CardContent class="flex-1 min-h-0 grid gap-2 overflow-y-auto pt-0.5">
                    <div
                        v-for="room in roomStore.rooms"
                        :key="room.id"
                        class="transition-transform hover:scale-[1.02] hover:shadow-lg rounded-md"
                    >
                        <Card class="flex flex-row justify-between items-center p-8">
                            <div>
                                <p class="font-medium">{{ room.name }}</p>
                                <p class="text-sm text-muted-foreground">
                                    ID: {{ room.id }}
                                </p>
                            </div>
                            <Button
                                variant="default"
                                @click="joinRoom(room.id, room.name)"
                            >
                                Join
                            </Button>
                        </Card>
                    </div>

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
import { LogOut } from "lucide-vue-next";
import LoadingButton from "@/components/LoadingButton.vue";

const newRoomName = ref("");
const createError = ref("");
const createSuccess = ref("");
const loadError = ref("");
const router = useRouter();
const auth = useAuthStore();
const roomStore = useRoomStore();
const loading = ref(false)

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
    createError.value = "";
    createSuccess.value = "";

    try {
        const res = await apiFetch("/ws/createRoom", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name: newRoomName.value }),
        });

        if (!res.ok) {
            const errData = await res.json();
            throw new Error(errData.error || "Failed to create room");
        }

        toast.success("Room created successfully!", {
            duration: 1000,
        });
        
        newRoomName.value = "";
        fetchRooms();
    } catch (err: unknown) {
        createError.value = err instanceof Error ? err.message : String(err);
        toast.success("Failed to create room", {
            description: createError.value,
            duration: 1000,
        });
    }
}

function joinRoom(roomId: string, roomName: string) {
    roomStore.setRoom(roomId, roomName);
    router.push(`/room/${roomId}`);
}

function logout() {
    loading.value = true
    setTimeout(() => {
        auth.logout();
        router.push("/login");
        loading.value = false
    }, 1000)
}

onMounted(fetchRooms);
</script>
