import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import HomeView from "../views/HomeView.vue";
import RoomView from "../views/RoomView.vue";
import LoginView from "../views/LoginView.vue";
import RegisterView from "../views/RegisterView.vue";
import { useAuthStore } from "../stores/auth";

const routes: RouteRecordRaw[] = [
    { path: "/", name: "Home", component: HomeView },
    { path: "/room/:roomId", name: "Room", component: RoomView },
    { path: "/login", name: "Login", component: LoginView },
    { path: "/register", name: "Register", component: RegisterView },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach(async (to, from, next) => {
    const auth = useAuthStore();
    const publicPages = ["/login", "/register"];
    const authRequired = !publicPages.includes(to.path);

    if (authRequired && !auth.isAuthenticated) {
        const refreshed = await auth.tryRefresh();
        if (!refreshed) return next("/login");
    }
    next();
});

export default router;
