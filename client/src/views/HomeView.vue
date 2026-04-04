<template>
    <div class="bg-background text-foreground h-dvh overflow-hidden dark">
        <div class="max-w-2xl mx-auto px-4 py-4 sm:px-6 sm:py-6 flex flex-col h-full min-h-0">

            <!-- Header -->
            <div class="flex justify-between items-center mb-6 shrink-0 pb-4 border-b-2 border-border">
                <div>
                    <h1 class="text-2xl font-black tracking-tight">TempChat</h1>
                    <div class="h-1 w-10 bg-primary mt-1"></div>
                </div>
                <LoadingButton
                    @click="logout"
                    :loading="loading"
                    loading-text="..."
                    class="px-3 py-1.5 text-xs font-bold border-2 border-destructive text-destructive neo-btn hover:bg-destructive hover:text-white bg-transparent"
                    variant="ghost"
                >Log Out</LoadingButton>
            </div>

            <!-- Create Room -->
            <div class="mb-6 shrink-0 border-2 border-border p-4 neo-card">
                <p class="text-xs font-black uppercase tracking-widest text-muted-foreground mb-3">New Room</p>
                <form @submit.prevent="createRoom" class="space-y-3">
                    <Input
                        v-model="newRoomName"
                        placeholder="Room name"
                        required
                        class="border-2 border-border focus:border-primary"
                    />
                    <!-- TTL pills -->
                    <div class="flex flex-wrap gap-1.5">
                        <button
                            v-for="opt in newRoomTTLOptions"
                            :key="opt.value"
                            type="button"
                            @click="newRoomTTL = opt.value"
                            :class="[
                                'px-2.5 py-1 text-xs font-bold border-2 neo-btn',
                                newRoomTTL === opt.value
                                    ? 'bg-primary text-primary-foreground border-primary'
                                    : 'border-border hover:border-primary/80',
                            ]"
                        >{{ opt.label }}</button>
                    </div>
                    <!-- PIN -->
                    <div class="flex items-center gap-3">
                        <label class="flex items-center gap-2 text-sm cursor-pointer select-none font-bold">
                            <input
                                type="checkbox"
                                v-model="enablePIN"
                                class="size-4 accent-primary"
                            />
                            PIN lock
                        </label>
                        <div v-if="enablePIN" class="flex flex-col">
                            <Input
                                v-model="newRoomPIN"
                                type="text"
                                inputmode="numeric"
                                maxlength="4"
                                placeholder="4 digits"
                                class="w-28 border-2 border-border focus:border-primary font-mono"
                                @input="pinError = ''"
                            />
                            <p v-if="pinError" class="text-xs text-destructive mt-1 font-mono">{{ pinError }}</p>
                        </div>
                    </div>
                    <button
                        type="submit"
                        class="w-full py-2 text-sm font-black bg-primary text-primary-foreground border-2 border-primary neo-btn"
                    >Create Room →</button>
                </form>
            </div>

            <!-- Rooms -->
            <div class="flex-1 min-h-0 flex flex-col border-2 border-border">
                <div class="shrink-0 px-4 pt-3 pb-2 border-b-2 border-border">
                    <p class="text-xs font-black uppercase tracking-widest text-muted-foreground">Active Rooms</p>
                </div>
                <div class="flex-1 min-h-0 overflow-y-auto p-3 space-y-2">

                    <!-- Your rooms -->
                    <template v-if="myRooms.length > 0">
                        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground px-1">Your Rooms</p>
                        <div
                            v-for="room in myRooms"
                            :key="room.id"
                            class="border-2 border-border neo-card p-3"
                        >
                            <div class="flex flex-row justify-between items-start gap-2">
                                <div class="min-w-0">
                                    <div class="flex items-center gap-2">
                                        <p class="font-black truncate">{{ room.name }}</p>
                                        <span v-if="room.has_pin" class="text-[10px] font-mono text-muted-foreground border border-border px-1">PIN</span>
                                    </div>
                                    <p class="text-xs text-muted-foreground font-mono mt-0.5">
                                        {{ room.clients }} online · {{ formatExpiry(room.expires_at) }}
                                    </p>
                                </div>
                                <div class="flex items-center gap-1.5 shrink-0">
                                    <button
                                        @click="copyLink(room.id)"
                                        class="px-2 py-1 text-xs font-bold border-2 border-border neo-btn hover:border-primary/80"
                                        aria-label="Copy room link"
                                    ><Link2 class="size-3" /></button>
                                    <button
                                        @click="handleJoin(room)"
                                        class="px-2.5 py-1 text-xs font-black bg-primary text-primary-foreground border-2 border-primary neo-btn"
                                    >Join</button>
                                </div>
                            </div>
                            <!-- Management controls -->
                            <div class="flex items-center gap-1.5 mt-3 pt-3 border-t-2 border-border flex-wrap">
                                <span class="text-[10px] font-mono text-muted-foreground">extend:</span>
                                <button
                                    v-for="opt in extendTTLOptions"
                                    :key="opt.value"
                                    type="button"
                                    @click="extendSelections[room.id] = extendSelections[room.id] === opt.value ? '' : opt.value"
                                    :class="[
                                        'px-2 py-0.5 text-[11px] font-bold border-2 neo-btn',
                                        extendSelections[room.id] === opt.value
                                            ? 'bg-primary text-primary-foreground border-primary'
                                            : 'border-border hover:border-primary/80',
                                    ]"
                                >{{ opt.label }}</button>
                                <button
                                    :disabled="!extendSelections[room.id]"
                                    @click="extendRoom(room)"
                                    class="px-2 py-0.5 text-[11px] font-bold border-2 border-border neo-btn hover:bg-primary hover:text-primary-foreground hover:border-primary disabled:opacity-30 disabled:cursor-not-allowed"
                                >Extend</button>
                                <button
                                    @click="handleDeleteRoom(room)"
                                    :class="[
                                        'ml-auto px-2 py-0.5 text-[11px] font-bold border-2 neo-btn',
                                        deleteConfirming[room.id]
                                            ? 'border-destructive text-destructive bg-destructive/10 animate-pulse'
                                            : 'border-border hover:border-destructive hover:text-destructive',
                                    ]"
                                >{{ deleteConfirming[room.id] ? "Sure?" : "Delete" }}</button>
                            </div>
                        </div>
                    </template>

                    <!-- Other rooms -->
                    <p v-if="myRooms.length > 0 && otherRooms.length > 0" class="text-[10px] font-black uppercase tracking-widest text-muted-foreground px-1 pt-2">Other Rooms</p>
                    <div
                        v-for="room in otherRooms"
                        :key="room.id"
                        class="border-2 border-border p-3"
                    >
                        <div class="flex flex-row justify-between items-start gap-2">
                            <div class="min-w-0">
                                <div class="flex items-center gap-2">
                                    <p class="font-black truncate">{{ room.name }}</p>
                                    <span v-if="room.has_pin" class="text-[10px] font-mono text-muted-foreground border border-border px-1">PIN</span>
                                </div>
                                <p class="text-xs text-muted-foreground font-mono mt-0.5">
                                    {{ room.clients }} online · {{ formatExpiry(room.expires_at) }}
                                </p>
                            </div>
                            <div class="flex items-center gap-1.5 shrink-0">
                                <button
                                    @click="copyLink(room.id)"
                                    class="px-2 py-1 text-xs font-bold border-2 border-border neo-btn hover:border-primary/80"
                                    aria-label="Copy room link"
                                ><Link2 class="size-3" /></button>
                                <button
                                    @click="handleJoin(room)"
                                    class="px-2.5 py-1 text-xs font-black bg-primary text-primary-foreground border-2 border-primary neo-btn"
                                >Join</button>
                            </div>
                        </div>
                        <!-- Inline PIN form -->
                        <div v-if="pinTarget === room.id" class="mt-3 pt-3 border-t-2 border-border flex gap-2">
                            <Input
                                v-model="enteredPIN"
                                type="text"
                                inputmode="numeric"
                                maxlength="4"
                                placeholder="Enter PIN"
                                class="flex-1 border-2 border-border focus:border-primary font-mono"
                                @keydown.enter="confirmJoin(room)"
                                autofocus
                            />
                            <button @click="confirmJoin(room)" class="px-3 py-1 text-xs font-black bg-primary text-primary-foreground border-2 border-primary neo-btn">Go</button>
                            <button @click="pinTarget = null" class="px-2 py-1 text-xs font-bold border-2 border-border neo-btn hover:border-destructive hover:text-destructive">✕</button>
                        </div>
                    </div>

                    <p v-if="roomStore.rooms.length === 0 && !loadError" class="text-sm text-muted-foreground text-center py-6 font-mono">
                        no active rooms — create one above
                    </p>

                    <div v-if="loadError" class="border-2 border-destructive p-3 mt-2">
                        <p class="text-sm text-destructive font-mono">{{ loadError }}</p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { apiFetch } from "@/services/api";
import { parseApiError } from "@/utils/parseError";
import { useRoomStore, type Room } from "@/stores/room";

import { Input } from "@/components/ui/input";
import { toast } from "vue-sonner";
import { Link2 } from "lucide-vue-next";
import LoadingButton from "@/components/LoadingButton.vue";

const newRoomTTLOptions = [
    { value: 15, label: "15m" },
    { value: 30, label: "30m" },
    { value: 60, label: "1h" },
    { value: 360, label: "6h" },
    { value: 1440, label: "24h" },
];

const extendTTLOptions = [
    { value: "15", label: "+15m" },
    { value: "30", label: "+30m" },
    { value: "60", label: "+1h" },
    { value: "360", label: "+6h" },
];

const extendSelections = ref<Record<string, string>>({});
const deleteConfirming = ref<Record<string, boolean>>({});
const deleteTimers: Record<string, ReturnType<typeof setTimeout>> = {};

const newRoomName = ref("");
const newRoomTTL = ref(60);
const enablePIN = ref(false);
const newRoomPIN = ref("");
const pinError = ref("");
const loadError = ref("");
const loading = ref(false);
const pinTarget = ref<string | null>(null);
const enteredPIN = ref("");

const router = useRouter();
const auth = useAuthStore();
const roomStore = useRoomStore();

let refreshInterval: ReturnType<typeof setInterval> | null = null;

const myRooms = computed(() => roomStore.rooms.filter((r) => r.is_creator));
const otherRooms = computed(() => roomStore.rooms.filter((r) => !r.is_creator));

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
            throw new Error(await parseApiError(res, "Failed to load rooms"));
        }
        roomStore.rooms = await res.json();
    } catch (err: unknown) {
        loadError.value = err instanceof Error ? err.message : String(err);
    }
}

async function createRoom() {
    const pin = enablePIN.value ? newRoomPIN.value.trim() : undefined;
    if (enablePIN.value && (!pin || !/^\d{4}$/.test(pin))) {
        pinError.value = "PIN must be exactly 4 digits";
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

        if (!res.ok) throw new Error(await parseApiError(res, "Failed to create room"));

        const room: Room = await res.json();
        newRoomName.value = "";
        newRoomPIN.value = "";
        enablePIN.value = false;
        pinError.value = "";

        try {
            await navigator.clipboard.writeText(`${window.location.origin}/join/${room.id}`);
            toast.success(`Room "${room.name}" created`, {
                description: "Link copied to clipboard",
                duration: 3000,
            });
        } catch {
            toast.success(`Room "${room.name}" created`);
        }
        fetchRooms();
    } catch (err: unknown) {
        toast.error(err instanceof Error ? err.message : "Failed to create room");
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

async function extendRoom(room: Room) {
    const ttl = Number(extendSelections.value[room.id]);
    if (!ttl) return;

    try {
        const res = await apiFetch(`/ws/room/${room.id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ ttl }),
        });
        if (!res.ok) throw new Error(await parseApiError(res, "Failed to extend room"));
        extendSelections.value[room.id] = "";
        toast.success("Room extended");
        fetchRooms();
    } catch (err: unknown) {
        toast.error(err instanceof Error ? err.message : "Failed to extend room");
    }
}

function handleDeleteRoom(room: Room) {
    if (!deleteConfirming.value[room.id]) {
        deleteConfirming.value[room.id] = true;
        deleteTimers[room.id] = setTimeout(() => {
            deleteConfirming.value[room.id] = false;
        }, 3000);
    } else {
        if (deleteTimers[room.id]) clearTimeout(deleteTimers[room.id]);
        deleteRoom(room);
    }
}

async function deleteRoom(room: Room) {
    try {
        const res = await apiFetch(`/ws/room/${room.id}`, { method: "DELETE" });
        if (!res.ok) throw new Error(await parseApiError(res, "Failed to delete room"));
        toast.success("Room deleted");
        fetchRooms();
    } catch (err: unknown) {
        toast.error(err instanceof Error ? err.message : "Failed to delete room");
        deleteConfirming.value[room.id] = false;
    }
}

async function copyLink(roomId: string) {
    try {
        await navigator.clipboard.writeText(`${window.location.origin}/join/${roomId}`);
        toast.success("Link copied");
    } catch {
        toast.error("Failed to copy link");
    }
}

async function logout() {
    loading.value = true;
    try {
        await auth.logout();
    } finally {
        router.push("/login");
        loading.value = false;
    }
}

onMounted(() => {
    fetchRooms();
    refreshInterval = setInterval(fetchRooms, 15_000);
});

onBeforeUnmount(() => {
    if (refreshInterval) clearInterval(refreshInterval);
    Object.values(deleteTimers).forEach(clearTimeout);
});
</script>
