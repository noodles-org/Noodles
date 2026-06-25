<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import api from '../api/client';
import DocsSidebar from '../components/DocsSidebar.vue';
import type { DocTocSection } from '../types';
import '../styles/docs.css';

const sections = ref<DocTocSection[]>([]);
const activePath = ref('');
const html = ref('');
const loading = ref(false);

async function loadDoc(path: string) {
  activePath.value = path;
  loading.value = true;
  try {
    const { data } = await api.get(`/docs/content?path=${encodeURIComponent(path)}`);
    html.value = DOMPurify.sanitize(marked.parse(data.content) as string);
  } catch {
    html.value = '<p>Failed to load document.</p>';
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  const { data } = await api.get('/docs/toc');
  sections.value = data.sections || [];
  if (sections.value[0]?.items[0]) {
    await loadDoc(sections.value[0].items[0].path);
  }
});
</script>

<template>
  <div class="docs-layout">
    <DocsSidebar :sections="sections" :active-path="activePath" @select="loadDoc" />
    <div class="docs-content" v-if="html" v-html="html" />
    <div class="docs-content docs-empty" v-else-if="!loading">Select a document from the sidebar.</div>
    <div class="docs-content docs-empty" v-else>Loading…</div>
  </div>
</template>