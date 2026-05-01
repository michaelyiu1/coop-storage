<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import FileGrid from '@/components/FileGrid.vue'
import type { MetaObject } from '@/types/file'

const route = useRoute()
const files = ref<MetaObject[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const username = ref('')

const API_BASE_URL = 'http://localhost:7678'

const fetchFiles = async (user: string) => {
  loading.value = true
  error.value = null

  try {
    const response = await fetch(`${API_BASE_URL}/read_all_meta?user=${encodeURIComponent(user)}`)

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`User "${user}" not found or has no files`)
      }
      throw new Error(`Failed to fetch files: ${response.status} ${response.statusText}`)
    }

    const data = await response.json()
    files.value = data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
    files.value = []
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  if (username.value.trim()) {
    fetchFiles(username.value.trim())
  }
}

onMounted(() => {
  const userParam = route.query.user as string
  if (userParam) {
    username.value = userParam
    fetchFiles(userParam)
  } else {
    loading.value = false
  }
})
</script>

<template>
  <div class="file-explorer">
    <header class="explorer-header">
      <h1>File Explorer</h1>
      <div class="search-bar">
        <input
          v-model="username"
          type="text"
          placeholder="Enter username"
          @keyup.enter="handleSearch"
        />
        <button @click="handleSearch">Load Files</button>
      </div>
    </header>

    <div class="explorer-content">
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>Loading files...</p>
      </div>

      <div v-else-if="error" class="error">
        <p>❌ {{ error }}</p>
      </div>

      <div v-else-if="username">
        <div class="user-info">
          <h2>Files for: {{ username }}</h2>
          <span class="file-count">{{ files.length }} file{{ files.length !== 1 ? 's' : '' }}</span>
        </div>
        <FileGrid :files="files" />
      </div>

      <div v-else class="welcome">
        <p>👆 Enter a username to view their files</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.file-explorer {
  min-height: 100vh;
  background: #f5f5f5;
}

.explorer-header {
  background: white;
  padding: 20px 40px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.explorer-header h1 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 28px;
}

.search-bar {
  display: flex;
  gap: 12px;
  max-width: 600px;
}

.search-bar input {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 0.2s;
}

.search-bar input:focus {
  outline: none;
  border-color: #4CAF50;
}

.search-bar button {
  padding: 12px 24px;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.search-bar button:hover {
  background: #45a049;
}

.search-bar button:active {
  transform: scale(0.98);
}

.explorer-content {
  padding: 20px 40px;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: #666;
}

.spinner {
  border: 4px solid #f3f3f3;
  border-top: 4px solid #4CAF50;
  border-radius: 50%;
  width: 50px;
  height: 50px;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error {
  text-align: center;
  padding: 40px 20px;
  color: #f44336;
  font-size: 16px;
  background: #ffebee;
  border-radius: 8px;
  margin: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.user-info h2 {
  margin: 0;
  color: #333;
  font-size: 24px;
}

.file-count {
  background: #e3f2fd;
  color: #1976d2;
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 600;
}

.welcome {
  text-align: center;
  padding: 80px 20px;
  color: #999;
  font-size: 20px;
}

.welcome p {
  margin: 0;
}
</style>
