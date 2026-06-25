<script setup lang="ts">
import type { ServiceLink } from '../types';
import '../styles/services.css';

defineProps<{ service: ServiceLink }>();
</script>

<template>
  <a :href="service.url" target="_blank" rel="noopener" class="card service-card">
    <div class="service-card-name">{{ service.name }}</div>
    <div class="service-card-desc">{{ service.description }}</div>
    <span class="service-card-category">{{ service.category }}</span>
  </a>
</template>