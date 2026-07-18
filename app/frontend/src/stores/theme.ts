import {ref} from 'vue';
import {defineStore} from 'pinia';

export const useThemeStore = defineStore('theme', () => {
    const dark = ref(localStorage.getItem('theme') === 'dark');

    function apply() {
        document.documentElement.setAttribute('data-theme', dark.value ? 'dark' : 'light');
    }

    function toggle() {
        dark.value = !dark.value;
        localStorage.setItem('theme', dark.value ? 'dark' : 'light');
        apply();
    }

    apply();

    return {dark, toggle};
});
