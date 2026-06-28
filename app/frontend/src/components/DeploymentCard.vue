<script setup lang="ts">
import {ref} from 'vue';
import type {DeploymentInfo} from '../types';
import {useDeploymentsStore} from '../stores/deployments';
import {useAuthStore} from '../stores/auth';
import '../styles/deployments.css';

const _ = defineProps<{ deployment: DeploymentInfo }>();
const store = useDeploymentsStore();
const auth = useAuthStore();
const busy = ref(false);

function cls(status: string): string {
  const m: Record<string, string> = {
    Healthy: 'healthy',
    Progressing: 'progressing',
    Degraded: 'degraded',
    Suspended: 'suspended',
  };
  return m[status] || 'unknown';
}

async function act(fn: () => Promise<void>, confirmMsg?: string) {
  if (confirmMsg && !confirm(confirmMsg)) return;
  busy.value = true;
  try {
    await fn();
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="card deployment-card">
    <div :class="['dep-dot', `dot-${cls(deployment.healthStatus)}`]"/>
    <div class="dep-info">
      <div class="dep-name">{{ deployment.name }}</div>
      <div class="dep-meta">
        <span>{{ deployment.namespace }}</span>
        <span>{{ deployment.readyReplicas }}/{{ deployment.replicas }} ready</span>
        <span>{{ deployment.image.split('/').pop() }}</span>
      </div>
    </div>
    <div class="dep-badges">
      <span :class="['badge', `badge-${cls(deployment.healthStatus)}`]">
        {{ deployment.healthStatus }}
      </span>
      <span
          v-if="deployment.syncStatus !== 'Unknown'"
          :class="['badge', deployment.syncStatus === 'Synced' ? 'badge-synced' : 'badge-outofsync']"
      >{{ deployment.syncStatus }}</span>
    </div>
    <div class="dep-actions" v-if="auth.isAdmin">
      <button
          class="btn btn-sm"
          :disabled="busy || deployment.paused"
          @click="act(() => store.restart(deployment.namespace, deployment.name), `Restart ${deployment.name}?`)"
      >Restart
      </button>
      <button
          v-if="!deployment.paused"
          class="btn btn-sm btn-warning"
          :disabled="busy"
          @click="act(() => store.pause(deployment.namespace, deployment.name), `Pause ${deployment.name}? Scales to 0.`)"
      >Pause
      </button>
      <button
          v-else
          class="btn btn-sm btn-success"
          :disabled="busy"
          @click="act(() => store.resume(deployment.namespace, deployment.name))"
      >Resume
      </button>
    </div>
  </div>
</template>