import { createRouter, createWebHistory } from 'vue-router'
import FileExplorer from '@/views/FileExplorer.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: FileExplorer,
    },
    {
      path: '/files',
      name: 'files',
      component: FileExplorer,
    },
  ],
})

export default router
