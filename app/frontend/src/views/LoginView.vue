<script setup lang="ts">
import {useAuthStore} from '../stores/auth';
import {useRoute} from 'vue-router';
import '../styles/login.css';

const auth = useAuthStore();
const route = useRoute();
const error = route.query.error as string | undefined;

const msgs: Record<string, string> = {
  oauth_denied: 'Authentication was denied by the provider.',
  invalid_state: 'Invalid session state. Please try again.',
  auth_failed: 'Authentication failed. Please try again.',
  no_code: 'No authorization code received.',
  not_authorized:
      'Your GitHub account does not have access to this dashboard. Contact an org admin.',
};
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>Cluster Dashboard</h1>
      <p>Sign in with your GitHub organization account</p>
      <div v-if="error" class="login-error">
        {{ msgs[error] || 'An error occurred.' }}
      </div>
      <button class="btn btn-primary" @click="auth.login()">
        Sign in with SSO
      </button>
    </div>
  </div>
</template>