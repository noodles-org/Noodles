<script setup lang="ts">
import {useAuthStore} from '../stores/auth';
import {useThemeStore} from '../stores/theme';
import {useRoute} from 'vue-router';
import '../styles/login.css';

const auth = useAuthStore();
const theme = useThemeStore();
const route = useRoute();
const error = route.query.error as string | undefined;

const msgs: Record<string, string> = {
  oauth_denied: 'Authentication was denied by the provider.',
  invalid_state: 'Invalid session state. Please try again.',
  auth_failed: 'Authentication failed. Please try again.',
  no_code: 'No authorization code received.',
  not_authorized: 'Your account does not have access. Contact an administrator.',
  email_unverified: 'Your account has no verified email address. Access cannot be granted.',
  requests_closed: 'Access requests are temporarily closed. Please contact an administrator.',
};
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>Cluster Dashboard</h1>
      <p>Sign in with your GitHub organization or Google account</p>
      <div v-if="error" class="login-error">
        {{ msgs[error] || 'An error occurred.' }}
      </div>
      <div class="login-providers">
        <button class="btn login-provider" @click="auth.login('github')">
          <img v-if="theme.dark" src="../assets/GitHub_Invertocat_White.svg" alt="Github logo" width="18" height="18" class="service-card-icon" />
          <img v-else src="../assets/GitHub_Invertocat_Black.svg" alt="Github logo" width="18" height="18" class="service-card-icon" />
          Sign in with GitHub
        </button>
        <button class="btn login-provider" @click="auth.login('google')">
          <img src="../assets/Google_G.svg" alt="Google logo" width="18" height="18" aria-hidden="true"
               class="service-card-icon"/>
          Sign in with Google
        </button>
      </div>
    </div>
  </div>
</template>