import { defineStore } from "pinia";

export interface Room {
    id: string;
    name: string;
    ttl: number;
    expires_at: string;
    has_pin: boolean;
    clients: number;
    is_creator: boolean;
}

export const useRoomStore = defineStore("room", {
    state: () => ({
        currentRoom: {
            id: null as string | null,
            name: null as string | null,
        },
        rooms: [] as Room[],
    }),
    actions: {
        setRoom(id: string, name: string) {
            this.currentRoom = { id, name };
        },
    },
});
