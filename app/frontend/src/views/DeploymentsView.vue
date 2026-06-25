<script setup lang="ts">
import { onMounted } from 'vue';
import { useDeploymentsStore } from '../stores/deployments';
import DeploymentCard from '../components/DeploymentCard.vue';
import '../styles/deployments.css';

const store = useDeploymentsStore();
onMounted(() => store.fetchDeployments());
</script>

<template>
  <div class="container">
    <div class="dep-header">
      <h1 class="page-title">Deployments</h1>
      <button class="btn" :disabled="store.loading" @click="store.fetchDeployments()">Refresh</button>
    </div>

    <div v-if="store.error" class="error-msg">{{ store.error }}</div>
    <div v-if="store.loading && !store.deployments.length" class="loading">Loading deployments…</div>
    <div v-else-if="!store.deployments.length" class="empty">No deployments found in managed namespaces.</div>
    <div v-else class="dep-list">
      <DeploymentCard
          v-for="d in store.deployments"
          :key="`${d.namespace}/${d.name}`"
          :deployment="d"
      />
    </div>
  </div>
</template>