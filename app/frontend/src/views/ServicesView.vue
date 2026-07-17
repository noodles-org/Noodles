<script setup lang="ts">
import {ref, onMounted} from 'vue';
import api from '../api/client';
import type {ServiceLink} from '../types';
import ServiceCard from '../components/ServiceCard.vue';
import '../styles/services.css';

const services = ref<ServiceLink[]>([]);
const loading = ref(true);

onMounted(async () => {
  try {
    const {data} = await api.get('/services');
    services.value = data;
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="container">
    <h1 class="page-title">Services</h1>
    <div v-if="loading" class="loading">Loading…</div>
    <div v-else class="services-grid">
      <ServiceCard v-for="s in services" :key="s.name" :service="s" />
    </div>
  </div>
</template>