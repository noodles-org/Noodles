<script setup lang="ts">
import {useAuthStore} from '../stores/auth';
import {useRoute} from 'vue-router';
import '../styles/nav.css';

const auth = useAuthStore();
const route = useRoute();

const navItems = [
  {path: '/services', label: 'Services'},
  {path: '/deployments', label: 'Deployments'},
  {path: '/docs', label: 'Docs'},
];
</script>

<template>
  <nav class="nav">
    <span class="nav-brand">Cluster Dashboard</span>
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
    <div class="nav-user" v-if="auth.user">
      <span>{{ auth.user.email }}</span>
      <span class="nav-role">{{ auth.user.role }}</span>
      <button class="nav-logout" @click="auth.logout()">Sign out</button>
    </div>
  </nav>
</template>