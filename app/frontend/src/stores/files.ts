import {ref} from 'vue';
import {defineStore} from 'pinia';
import api from '../api/client';
import type {FileEntry} from '../types';

function joinPath(base: string, name: string): string {
    return base ? `${base}/${name}` : name;
}

function parentPath(path: string): string {
    const idx = path.lastIndexOf('/');
    return idx === -1 ? '' : path.slice(0, idx);
}

export const useFilesStore = defineStore('files', () => {
    const entries = ref<FileEntry[]>([]);
    const path = ref('');
    const loading = ref(false);
    const error = ref<string | null>(null);

    async function list(target = path.value) {
        loading.value = true;
        error.value = null;
        try {
            const {data} = await api.get('/files', {params: {path: target}});
            entries.value = data ?? [];
            path.value = target;
        } catch {
            error.value = 'Failed to load files';
        } finally {
            loading.value = false;
        }
    }

    function open(entry: FileEntry) {
        if (entry.isDir) return list(entry.path);
    }

    function goTo(target: string) {
        return list(target);
    }

    function goUp() {
        if (path.value) return list(parentPath(path.value));
    }

    function download(entry: FileEntry) {
        window.open(`/api/files/download?path=${encodeURIComponent(entry.path)}`, '_blank');
    }

    async function remove(target: string) {
        await api.delete('/files', {params: {path: target}});
        await list();
    }

    async function removeMany(targets: string[]) {
        for (const t of targets) {
            await api.delete('/files', {params: {path: t}});
        }
        await list();
    }

    async function upload(file: File) {
        const form = new FormData();
        form.append('file', file);
        await api.post('/files/upload', form, {params: {path: path.value}});
        await list();
    }

    async function mkdir(name: string) {
        await api.post('/files/mkdir', null, {params: {path: joinPath(path.value, name)}});
        await list();
    }

    async function rename(entry: FileEntry, newName: string) {
        const to = joinPath(parentPath(entry.path), newName);
        await api.post('/files/rename', {from: entry.path, to});
        await list();
    }

    async function listDir(target: string): Promise<FileEntry[]> {
        const {data} = await api.get('/files', {params: {path: target}});
        return data ?? [];
    }

    async function move(from: string, destDir: string) {
        const name = from.slice(from.lastIndexOf('/') + 1);
        const to = joinPath(destDir, name);
        await api.post('/files/rename', {from, to});
    }

    return {
        entries,
        path,
        loading,
        error,
        list,
        open,
        goTo,
        goUp,
        download,
        remove,
        removeMany,
        upload,
        mkdir,
        rename,
        listDir,
        move,
    };
});
