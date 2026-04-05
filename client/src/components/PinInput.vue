<template>
    <div class="flex gap-2">
        <input
            v-for="i in 4"
            :key="i"
            ref="inputs"
            type="text"
            inputmode="numeric"
            maxlength="1"
            :value="digits[i - 1]"
            @input="onInput(i - 1, $event as InputEvent)"
            @keydown="onKeydown(i - 1, $event)"
            @paste.prevent="onPaste($event)"
            class="w-10 h-10 text-center font-mono text-base font-bold bg-input border-2 border-border focus:outline-none focus:border-primary"
        />
    </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

const props = defineProps<{ modelValue: string }>();
const emit = defineEmits<{
    "update:modelValue": [value: string];
    complete: [value: string];
}>();

const inputs = ref<HTMLInputElement[]>([]);
const digits = ref(["", "", "", ""]);

watch(
    () => props.modelValue,
    (val) => {
        for (let i = 0; i < 4; i++) {
            digits.value[i] = val[i] ?? "";
        }
    },
);

function onInput(index: number, e: InputEvent) {
    const input = e.target as HTMLInputElement;
    const char = input.value.replace(/\D/g, "").slice(-1);
    input.value = char;
    digits.value[index] = char;
    const joined = digits.value.join("");
    emit("update:modelValue", joined);
    if (char && index < 3) {
        inputs.value[index + 1]?.focus();
    }
    if (joined.length === 4 && digits.value.every((d) => d !== "")) {
        emit("complete", joined);
    }
}

function onKeydown(index: number, e: KeyboardEvent) {
    if (e.key === "Backspace") {
        if (digits.value[index]) {
            digits.value[index] = "";
            emit("update:modelValue", digits.value.join(""));
        } else if (index > 0) {
            inputs.value[index - 1]?.focus();
        }
    }
}

function onPaste(e: ClipboardEvent) {
    const text = (e.clipboardData?.getData("text") ?? "").replace(/\D/g, "").slice(0, 4);
    for (let i = 0; i < 4; i++) {
        digits.value[i] = text[i] ?? "";
    }
    const joined = digits.value.join("");
    emit("update:modelValue", joined);
    inputs.value[Math.min(text.length, 3)]?.focus();
    if (joined.length === 4 && digits.value.every((d) => d !== "")) {
        emit("complete", joined);
    }
}
</script>
