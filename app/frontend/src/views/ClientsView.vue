<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {useClientsStore} from '../stores/clients';
import type {Role} from '../types';
import '../styles/clients.css';

const store = useClientsStore();
const roles = ref<Record<string, Role>>({});
const busy = ref<string | null>(null);

onMounted(() => store.fetchClients());

function roleFor(email: string): Role {
  return roles.value[email] ?? 'client';
}

async function act(email: string, action: 'approve' | 'reject') {
  busy.value = email;
  try {
    if (action === 'approve') await store.approve(email, roleFor(email));
    else await store.reject(email);
  } catch {
    store.error = `Failed to ${action} ${email}`;
  } finally {
    busy.value = null;
  }
}
</script>

<template>
  <div class="container">
    <div class="clients-header">
      <h1 class="page-title">Clients</h1>
      <button class="btn" :disabled="store.loading" @click="store.fetchClients()">Refresh</button>
    </div>

    <div v-if="store.error" class="error-msg">{{ store.error }}</div>
    <div v-if="store.loading && !store.pending.length && !store.approved.length" class="loading">Loading clients…</div>

    <template v-else>
      <section class="clients-section">
        <h2 class="page-title">Pending requests</h2>
        <div v-if="!store.pending.length" class="empty">No pending access requests.</div>
        <div v-else class="clients-list">
          <div v-for="c in store.pending" :key="c.email" class="card client-card">
            <div class="client-info">
              <div class="client-email">{{ c.email }}</div>
              <div class="client-meta">
                <span v-if="c.name">{{ c.name }}</span>
                <span v-if="c.createdAt">Requested {{ c.createdAt }}</span>
              </div>
            </div>
            <div class="client-actions">
              <select
                  class="client-role-select"
                  :value="roleFor(c.email)"
                  @change="roles[c.email] = ($event.target as HTMLSelectElement).value as Role"
              >
                <option value="client">client</option>
                <option value="client_admin">client_admin</option>
              </select>
              <button class="btn btn-sm btn-success" :disabled="busy === c.email" @click="act(c.email, 'approve')">
                Approve
              </button>
              <button class="btn btn-sm btn-danger" :disabled="busy === c.email" @click="act(c.email, 'reject')">
                Reject
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="clients-section">
        <h2 class="page-title">Approved clients</h2>
        <div v-if="!store.approved.length" class="empty">No approved clients yet.</div>
        <div v-else class="clients-list">
          <div v-for="c in store.approved" :key="c.email" class="card client-card">
            <div class="client-info">
              <div class="client-email">{{ c.email }}</div>
              <div class="client-meta">
                <span v-if="c.name">{{ c.name }}</span>
                <span v-if="c.createdAt">Approved {{ c.createdAt }}</span>
              </div>
            </div>
            <div class="client-actions">
              <span class="client-role">{{ c.role }}</span>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
