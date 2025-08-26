import { defineStore } from "pinia";

export const useRoomStore = defineStore("room", {
    state: () => ({
        currentRoom: {
            id: null,
            name: null,
        },
        rooms: [],
    }),
    actions: {
        setRoom(id, name) {
            this.currentRoom = { id, name };
        },
    },
});
