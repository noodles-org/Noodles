<script setup lang="ts">
import {ref, onMounted, onUnmounted} from 'vue';
import {useAuthStore} from '../stores/auth';
import {useThemeStore} from '../stores/theme';
import {useRoute} from 'vue-router';
import '../styles/nav.css';

const auth = useAuthStore();
const theme = useThemeStore();
const route = useRoute();

const profileOpen = ref(false);
const profileRef = ref<HTMLElement | null>(null);

const navItems = [
  {path: '/services', label: 'Services'},
  {path: '/deployments', label: 'Deployments'},
  {path: '/docs', label: 'Docs'},
];

function toggleProfile() {
  profileOpen.value = !profileOpen.value;
}

function closeProfile(e: MouseEvent) {
  if (profileRef.value && !profileRef.value.contains(e.target as Node)) {
    profileOpen.value = false;
  }
}

onMounted(() => document.addEventListener('click', closeProfile));
onUnmounted(() => document.removeEventListener('click', closeProfile));
</script>

<template>
  <nav class="nav">
    <span class="nav-brand">Noodles Dashboard</span>
    <ul class="nav-links">
      <li v-for="item in navItems" :key="item.path">
        <router-link
            :to="item.path"
            class="nav-link"
            :class="{ active: route.path.startsWith(item.path) }"
        >{{ item.label }}
        </router-link>
      </li>
    </ul>
    <div class="nav-spacer"/>

    <!-- Desktop: inline user info -->
    <button class="nav-theme-btn nav-desktop-only" @click="theme.toggle()" :title="theme.dark ? 'Switch to light mode' : 'Switch to dark mode'">
      <img v-if="theme.dark" src="../assets/light_mode_fill.svg" alt="Switch to light mode" width="18" height="18" class="nav-icon" />
      <img v-else src="../assets/dark_mode_fill.svg" alt="Switch to dark mode" width="18" height="18" class="nav-icon" />
    </button>
    <div class="nav-user nav-desktop-only" v-if="auth.user">
      <span>{{ auth.user.email }}</span>
      <span class="nav-role">{{ auth.user.role }}</span>
      <button class="nav-logout" @click="auth.logout()">Sign out</button>
    </div>

    <!-- Mobile: profile button + dropdown -->
    <div class="nav-profile-wrap nav-mobile-only" ref="profileRef">
      <button class="nav-profile-btn" @click="toggleProfile">
        <img src="../assets/account.svg" alt="Profile" width="20" height="20" class="nav-icon" />
      </button>
      <div class="nav-profile-menu" v-if="profileOpen">
        <div class="nav-profile-info" v-if="auth.user">
          <span class="nav-profile-email">{{ auth.user.email }}</span>
          <span class="nav-role">{{ auth.user.role }}</span>
        </div>
        <button class="nav-profile-item" @click="theme.toggle()">
          <img v-if="theme.dark" src="../assets/light_mode_fill.svg" alt="Switch to light mode" width="16" height="16" class="nav-icon" />
          <img v-else src="../assets/dark_mode_fill.svg" alt="Switch to dark mode" width="16" height="16" class="nav-icon" />
          {{ theme.dark ? 'Light mode' : 'Dark mode' }}
        </button>
        <button class="nav-profile-item nav-profile-signout" @click="auth.logout()">Sign out</button>
      </div>
    </div>
  </nav>
</template>
