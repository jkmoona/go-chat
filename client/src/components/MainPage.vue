<template>
    <div class="max-w-2xl mx-auto p-6">
        <h1 class="text-2xl font-bold mb-6">Chat Rooms</h1>

        <!-- Create Room Form -->
        <div class="mb-8">
            <h2 class="text-lg font-semibold mb-2">Create a Room</h2>
            <form @submit.prevent="createRoom" class="flex gap-2">
                <input
                    v-model="newid"
                    placeholder="Room ID"
                    class="flex-1 border rounded p-2"
                    required
                />
                <input
                    v-model="newRoomName"
                    placeholder="Room Name"
                    class="flex-1 border rounded p-2"
                    required
                />
                <button class="bg-green-600 text-white px-4 py-2 rounded">
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

        <!-- Room List -->
        <div>
            <h2 class="text-lg font-semibold mb-2">Available Rooms</h2>
            <ul class="space-y-2">
                <li
                    v-for="room in rooms"
                    :key="room.id"
                    class="flex justify-between items-center border p-3 rounded"
                >
                    <div>
                        <p class="font-medium">{{ room.name }}</p>
                        <p class="text-sm text-gray-600">ID: {{ room.id }}</p>
                    </div>
                    <button
                        @click="joinRoom(room.id)"
                        class="bg-blue-600 text-white px-3 py-1 rounded"
                    >
                        Join
                    </button>
                </li>
            </ul>
            <p v-if="loadError" class="text-red-600 mt-4">{{ loadError }}</p>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { apiFetch } from "../services/api";

const rooms = ref([]);
const newid = ref("");
const newRoomName = ref("");
const createError = ref("");
const createSuccess = ref("");
const loadError = ref("");
const router = useRouter();
const auth = useAuthStore();

async function fetchRooms() {
    try {
        const res = await apiFetch("http://localhost:8080/ws/getRooms");
        if (!res.ok) throw new Error("Could not load rooms");
        rooms.value = await res.json();
    } catch (err) {
        loadError.value = err.message;
    }
}

async function createRoom() {
    createError.value = "";
    createSuccess.value = "";

    try {
        const res = await apiFetch("http://localhost:8080/ws/createRoom", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                id: newid.value,
                name: newRoomName.value,
            }),
        });

        if (!res.ok) {
            const err = await res.json();
            throw new Error(err.error || "Failed to create room");
        }

        createSuccess.value = `Room "${newRoomName.value}" created successfully!`;
        newid.value = "";
        newRoomName.value = "";
        fetchRooms(); // refresh room list
    } catch (err) {
        createError.value = err.error;
    }
}

function joinRoom(id) {
    router.push(`/room/${id}`);
}

onMounted(fetchRooms);
</script>
