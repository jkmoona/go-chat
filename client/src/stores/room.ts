import { defineStore } from "pinia";

interface Room {
    id: string;
    name: string;
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
