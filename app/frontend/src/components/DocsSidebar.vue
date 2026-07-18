<script setup lang="ts">
import type {DocTocSection} from '../types';
import '../styles/docs.css';

defineProps<{ sections: DocTocSection[]; activePath: string }>();
const emit = defineEmits<{ select: [path: string] }>();
</script>

<template>
  <aside class="docs-sidebar">
    <div v-for="section in sections" :key="section.title" class="docs-section">
      <div class="docs-section-title">{{ section.title }}</div>
      <ul class="docs-items">
        <li v-for="item in section.items" :key="item.path">
          <a
              class="docs-item"
              :class="{ active: activePath === item.path }"
              @click.prevent="emit('select', item.path)"
          >{{ item.title }}</a>
        </li>
      </ul>
    </div>
  </aside>
</template>