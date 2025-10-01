<template>
    <div class="bg-neutral-900 text-neutral-200 min-h-screen">
        <div class="max-w-2xl mx-auto p-6">
            <div class="flex justify-between items-center mb-6">
                <h1 class="text-2xl font-bold">Chat Rooms</h1>
                <button
                    @click="logout"
                    class="bg-red-600 px-3 py-1 rounded cursor-pointer"
                >
                    Log Out
                </button>
            </div>
            <div class="mb-8">
                <h2 class="text-lg font-semibold mb-2">Create a Room</h2>
                <form @submit.prevent="createRoom" class="flex gap-2">
                    <input
                        v-model="newRoomName"
                        placeholder="Room Name"
                        class="flex-1 border rounded p-2"
                        required
                    />
                    <button
                        class="bg-green-600 px-4 py-2 rounded cursor-pointer"
                    >
                        Create
                    </button>
                </form>
                <p v-if="createError" class="text-red-600 mt-2">
                    {{ createError }}
                </p>
                <p v-if="createSuccess" class="text-green-600 mt-2">
                    {{ createSuccess }}
                </p>
            </div>
            <div>
                <h2 class="text-lg font-semibold mb-2">Available Rooms</h2>
                <ul class="space-y-2">
                    <li
                        v-for="room in roomStore.rooms"
                        :key="room.id"
                        class="flex justify-between items-center border p-3 rounded"
                    >
                        <div>
                            <p class="font-medium">{{ room.name }}</p>
                            <p class="text-sm text-neutral-400">ID: {{ room.id }}</p>
                        </div>
                        <button
                            @click="joinRoom(room.id, room.name)"
                            class="bg-blue-600 px-3 py-1 rounded cursor-pointer"
                        >
                            Join
                        </button>
                    </li>
                </ul>
                <p v-if="loadError" class="text-red-600 mt-4">{{ loadError }}</p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { apiFetch } from "../services/api";
import { useRoomStore } from "../stores/room";

const newRoomName = ref<string>("");
const createError = ref<string>("");
const createSuccess = ref<string>("");
const loadError = ref<string>("");
const router = useRouter();
const auth = useAuthStore();
const roomStore = useRoomStore();

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
        const rooms = await res.json();
        roomStore.rooms = rooms;
    } catch (err: unknown) {
        if (err instanceof Error) {
            loadError.value = err.message;
        } else {
            loadError.value = String(err);
        }
    }
}

async function createRoom() {
    createError.value = "";
    createSuccess.value = "";

    try {
        const res = await apiFetch("/ws/createRoom", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                name: newRoomName.value,
            }),
        });

        if (!res.ok) {
            const errData = await res.json();
            throw new Error(errData.error || "Failed to create room");
        }

        createSuccess.value = `Room "${newRoomName.value}" created successfully!`;
        newRoomName.value = "";
        fetchRooms();
    } catch (err: unknown) {
        if (err instanceof Error) {
            createError.value = err.message;
        } else {
            createError.value = String(err);
        }
    }
}

function joinRoom(roomId: string, roomName: string) {
    roomStore.setRoom(roomId, roomName);
    router.push(`/room/${roomId}`);
}

function logout() {
    auth.logout();
    router.push("/login");
}

onMounted(fetchRooms);
</script>
