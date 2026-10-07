<script setup lang="ts">
import {computed, onMounted, ref} from 'vue';
import {useFilesStore} from '../stores/files';
import {useAuthStore} from '../stores/auth';
import type {FileEntry} from '../types';
import '../styles/files.css';

const store = useFilesStore();
const auth = useAuthStore();

const selected = ref<Set<string>>(new Set());
const busy = ref(false);
const uploadInput = ref<HTMLInputElement | null>(null);
const preview = ref<FileEntry | null>(null);

const IMAGE_EXTS = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'avif', 'bmp'];

function isImage(entry: FileEntry): boolean {
  if (entry.isDir) return false;
  const dot = entry.name.lastIndexOf('.');
  if (dot === -1) return false;
  return IMAGE_EXTS.includes(entry.name.slice(dot + 1).toLowerCase());
}

function imageUrl(entry: FileEntry): string {
  return `/api/files/download?path=${encodeURIComponent(entry.path)}`;
}

const brokenThumbs = ref<Set<string>>(new Set());

function onThumbError(entry: FileEntry) {
  const next = new Set(brokenThumbs.value);
  next.add(entry.path);
  brokenThumbs.value = next;
}

function openPreview(entry: FileEntry) {
  preview.value = entry;
}

function closePreview() {
  preview.value = null;
}

onMounted(() => store.list(''));

const crumbs = computed(() => {
  const parts = store.path ? store.path.split('/') : [];
  const items = [{label: 'Data', path: ''}];
  let acc = '';
  for (const p of parts) {
    acc = acc ? `${acc}/${p}` : p;
    items.push({label: p, path: acc});
  }
  return items;
});

const allSelected = computed(
    () => store.entries.length > 0 && selected.value.size === store.entries.length,
);

const selectedFileCount = computed(() => {
  let n = 0;
  for (const e of store.entries) {
    if (!e.isDir && selected.value.has(e.path)) n++;
  }
  return n;
});

const moveOpen = ref(false);
const movePath = ref('');
const moveDirs = ref<FileEntry[]>([]);
const moveItems = ref<string[]>([]);
const moveLoading = ref(false);

const moveCrumbs = computed(() => {
  const parts = movePath.value ? movePath.value.split('/') : [];
  const items = [{label: 'Data', path: ''}];
  let acc = '';
  for (const p of parts) {
    acc = acc ? `${acc}/${p}` : p;
    items.push({label: p, path: acc});
  }
  return items;
});

async function loadMoveDirs(target: string) {
  moveLoading.value = true;
  try {
    const all = await store.listDir(target);
    moveDirs.value = all.filter((e) => e.isDir);
    movePath.value = target;
  } catch {
    store.error = 'Failed to load folders';
  } finally {
    moveLoading.value = false;
  }
}

function moveSelected() {
  const targets = store.entries
      .filter((e) => !e.isDir && selected.value.has(e.path))
      .map((e) => e.path);
  if (!targets.length) return;
  moveItems.value = targets;
  moveOpen.value = true;
  loadMoveDirs('');
}

function closeMove() {
  moveOpen.value = false;
  moveItems.value = [];
}

async function confirmMove() {
  const dest = movePath.value;
  let existing: Set<string>;
  try {
    const destEntries = await store.listDir(dest);
    existing = new Set(destEntries.map((e) => e.name));
  } catch {
    store.error = 'Failed to inspect destination';
    return;
  }

  busy.value = true;
  try {
    for (const from of moveItems.value) {
      const name = from.slice(from.lastIndexOf('/') + 1);
      const to = dest ? `${dest}/${name}` : name;
      if (to === from) continue;
      if (existing.has(name) && !confirm(`"${name}" already exists in the destination. Overwrite?`)) {
        break;
      }
      await store.move(from, dest);
    }
  } catch {
    store.error = 'Move failed';
  } finally {
    busy.value = false;
    closeMove();
    selected.value = new Set();
    await store.list();
  }
}

function navigate(target: string) {
  selected.value = new Set();
  store.list(target);
}

function openEntry(entry: FileEntry) {
  if (entry.isDir) navigate(entry.path);
}

function toggle(path: string) {
  const next = new Set(selected.value);
  if (next.has(path)) next.delete(path);
  else next.add(path);
  selected.value = next;
}

function toggleAll() {
  if (allSelected.value) {
    selected.value = new Set();
  } else {
    selected.value = new Set(store.entries.map((e) => e.path));
  }
}

function formatSize(entry: FileEntry): string {
  if (entry.isDir) return '—';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = entry.size;
  let i = 0;
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024;
    i++;
  }
  return `${i === 0 ? size : size.toFixed(1)} ${units[i]}`;
}

function formatDate(iso: string): string {
  if (!iso) return '';
  return new Date(iso).toLocaleString();
}

async function run(fn: () => Promise<void>, confirmMsg?: string) {
  if (confirmMsg && !confirm(confirmMsg)) return;
  busy.value = true;
  try {
    await fn();
  } catch {
    store.error = 'Action failed';
  } finally {
    busy.value = false;
  }
}

function deleteSelected() {
  const targets = [...selected.value];
  if (!targets.length) return;
  run(async () => {
    await store.removeMany(targets);
    selected.value = new Set();
  }, `Delete ${targets.length} item(s)? This cannot be undone.`);
}

function deleteOne(entry: FileEntry) {
  run(() => store.remove(entry.path), `Delete "${entry.name}"?`);
}

function renameEntry(entry: FileEntry) {
  const name = prompt('New name:', entry.name);
  if (!name || name === entry.name) return;
  run(() => store.rename(entry, name));
}

function newFolder() {
  const name = prompt('New folder name:');
  if (!name) return;
  run(() => store.mkdir(name));
}

function triggerUpload() {
  uploadInput.value?.click();
}

function onFileChosen(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  run(() => store.upload(file)).finally(() => {
    input.value = '';
  });
}
</script>

<template>
  <div class="container">
    <div class="files-header">
      <h1 class="page-title">Foundry Files</h1>
      <div class="files-toolbar">
        <button class="btn" :disabled="store.loading || busy" @click="navigate(store.path)">Refresh</button>
        <template v-if="auth.canMutate">
          <button class="btn" :disabled="busy" @click="newFolder">New folder</button>
          <button class="btn btn-primary" :disabled="busy" @click="triggerUpload">Upload</button>
          <button
              class="btn"
              :disabled="busy || !selectedFileCount"
              @click="moveSelected"
          >Move ({{ selectedFileCount }})
          </button>
          <button
              class="btn btn-danger"
              :disabled="busy || !selected.size"
              @click="deleteSelected"
          >Delete ({{ selected.size }})
          </button>
          <input ref="uploadInput" type="file" class="files-upload-input" @change="onFileChosen"/>
        </template>
      </div>
    </div>

    <nav class="files-breadcrumb">
      <template v-for="(c, i) in crumbs" :key="c.path">
        <span v-if="i > 0" class="files-crumb-sep">/</span>
        <button
            class="files-crumb"
            :class="{ active: i === crumbs.length - 1 }"
            @click="navigate(c.path)"
        >{{ c.label }}</button>
      </template>
    </nav>

    <div v-if="store.error" class="error-msg">{{ store.error }}</div>
    <div v-if="store.loading && !store.entries.length" class="loading">Loading files…</div>
    <div v-else-if="!store.entries.length" class="empty">This folder is empty.</div>

    <table v-else class="card files-table">
      <thead>
      <tr>
        <th v-if="auth.canMutate" class="files-col-check">
          <input type="checkbox" :checked="allSelected" @change="toggleAll"/>
        </th>
        <th>Name</th>
        <th class="files-col-size">Size</th>
        <th class="files-col-date">Modified</th>
        <th class="files-col-actions"></th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="entry in store.entries" :key="entry.path">
        <td v-if="auth.canMutate" class="files-col-check">
          <input
              type="checkbox"
              :checked="selected.has(entry.path)"
              @change="toggle(entry.path)"
          />
        </td>
        <td class="files-name">
          <button v-if="entry.isDir" class="files-link" @click="openEntry(entry)">
            📁 {{ entry.name }}
          </button>
          <span v-else class="files-file">
            <button
                v-if="isImage(entry) && !brokenThumbs.has(entry.path)"
                class="files-thumb-btn"
                @click="openPreview(entry)"
            >
              <img
                  class="files-thumb"
                  :src="imageUrl(entry)"
                  :alt="entry.name"
                  loading="lazy"
                  @error="onThumbError(entry)"
              />
            </button>
            <span v-else class="files-icon">📄</span>
            {{ entry.name }}
          </span>
        </td>
        <td class="files-col-size">{{ formatSize(entry) }}</td>
        <td class="files-col-date">{{ formatDate(entry.modTime) }}</td>
        <td class="files-col-actions">
          <div class="files-actions">
            <button v-if="!entry.isDir" class="btn btn-sm" @click="store.download(entry)">Download</button>
            <template v-if="auth.canMutate">
              <button class="btn btn-sm" :disabled="busy" @click="renameEntry(entry)">Rename</button>
              <button class="btn btn-sm btn-danger" :disabled="busy" @click="deleteOne(entry)">Delete</button>
            </template>
          </div>
        </td>
      </tr>
      </tbody>
    </table>

    <div v-if="preview" class="files-preview-overlay" @click="closePreview">
      <div class="files-preview-box" @click.stop>
        <div class="files-preview-header">
          <span class="files-preview-name">{{ preview.name }}</span>
          <button class="btn btn-sm" @click="closePreview">Close</button>
        </div>
        <img class="files-preview-img" :src="imageUrl(preview)" :alt="preview.name"/>
      </div>
    </div>

    <div v-if="moveOpen" class="files-preview-overlay" @click="closeMove">
      <div class="files-move-box" @click.stop>
        <div class="files-preview-header">
          <span class="files-preview-name">Move {{ moveItems.length }} item(s)</span>
          <button class="btn btn-sm" @click="closeMove">Cancel</button>
        </div>

        <nav class="files-breadcrumb">
          <template v-for="(c, i) in moveCrumbs" :key="c.path">
            <span v-if="i > 0" class="files-crumb-sep">/</span>
            <button
                class="files-crumb"
                :class="{ active: i === moveCrumbs.length - 1 }"
                @click="loadMoveDirs(c.path)"
            >{{ c.label }}</button>
          </template>
        </nav>

        <div class="files-move-list">
          <div v-if="moveLoading" class="loading">Loading…</div>
          <div v-else-if="!moveDirs.length" class="empty">No subfolders here.</div>
          <button
              v-for="d in moveDirs"
              :key="d.path"
              class="files-move-dir"
              @click="loadMoveDirs(d.path)"
          >📁 {{ d.name }}</button>
        </div>

        <div class="files-preview-header">
          <span class="files-move-dest">Destination: Data{{ movePath ? '/' + movePath : '' }}</span>
          <button class="btn btn-primary btn-sm" :disabled="busy" @click="confirmMove">Move here</button>
        </div>
      </div>
    </div>
  </div>
</template>
