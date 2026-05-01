<script setup lang="ts">
import { computed } from 'vue'
import type { MetaObject } from '@/types/file'

const props = defineProps<{
  files: MetaObject[]
}>()

const getFileIcon = (fileType: string): string => {
  const type = fileType.toLowerCase()
  if (type.includes('image') || type.includes('png') || type.includes('jpg') || type.includes('jpeg')) return '🖼️'
  if (type.includes('pdf')) return '📄'
  if (type.includes('video')) return '🎥'
  if (type.includes('audio')) return '🎵'
  if (type.includes('zip') || type.includes('archive')) return '📦'
  if (type.includes('text')) return '📝'
  return '📄'
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}
</script>

<template>
  <div class="file-grid">
    <div
      v-for="file in files"
      :key="file.id"
      class="file-card"
      :class="{ 'deleted': file.deleteFlag }"
    >
      <div class="file-icon">{{ getFileIcon(file.fileType) }}</div>
      <div class="file-name">{{ file.fileName }}</div>
      <div class="file-type">{{ file.fileType }}</div>
      <div class="file-version">v{{ file.version }}</div>
      <div v-if="file.deleteFlag" class="deleted-badge">Deleted</div>
    </div>
    <div v-if="files.length === 0" class="empty-state">
      <p>No files found</p>
    </div>
  </div>
</template>

<style scoped>
.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 20px;
  padding: 20px;
}

.file-card {
  background: #fff;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  min-height: 180px;
}

.file-card:hover {
  border-color: #4CAF50;
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.file-card.deleted {
  opacity: 0.5;
  border-color: #f44336;
}

.file-icon {
  font-size: 48px;
  margin-bottom: 8px;
}

.file-name {
  font-weight: 600;
  font-size: 14px;
  text-align: center;
  word-break: break-word;
  max-width: 100%;
  color: #333;
}

.file-type {
  font-size: 11px;
  color: #666;
  text-align: center;
}

.file-version {
  font-size: 10px;
  color: #999;
  background: #f5f5f5;
  padding: 2px 8px;
  border-radius: 10px;
}

.deleted-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  background: #f44336;
  color: white;
  font-size: 10px;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 600;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
  color: #999;
  font-size: 18px;
}

.empty-state p {
  margin: 0;
}
</style>
