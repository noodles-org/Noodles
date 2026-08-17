<script setup lang="ts">
import {useAuthStore} from './stores/auth';
import NavBar from './components/NavBar.vue';

const auth = useAuthStore();
</script>

<template>
  <NavBar v-if="auth.isAuthenticated && !auth.isPending"/>
  <router-view/>
</template>