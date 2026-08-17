<script setup lang="ts">
import {useAuthStore} from '../stores/auth';
import {useThemeStore} from '../stores/theme';
import '../styles/login.css';

const auth = useAuthStore();
useThemeStore();
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>Awaiting Approval</h1>
      <p>
        Thanks{{ auth.user?.name ? `, ${auth.user.name}` : '' }} — your access request has been
        received and is waiting for an administrator to approve it.
      </p>
      <p>You will be able to sign in normally once your account has been granted a role.</p>
      <button class="btn btn-primary" @click="auth.logout()">
        Sign out
      </button>
    </div>
  </div>
</template>
