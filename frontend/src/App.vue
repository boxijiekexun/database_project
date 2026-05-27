<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { clearAuth } from './api/request'

const route = useRoute()
const router = useRouter()

const role = computed(() => localStorage.getItem('role') || '')
const isLoginPage = computed(() => route.path === '/login')

function logout() {
  clearAuth()
  router.push('/login')
}
</script>

<template>
  <div class="app">
    <header class="topbar">
      <h1>校园外卖管理系统</h1>
      <div v-if="!isLoginPage" class="topbar-right">
        <span class="role">当前角色：{{ role }}</span>
        <button @click="logout">退出登录</button>
      </div>
    </header>
    <main class="page">
      <section class="content-wrap">
        <router-view />
      </section>
    </main>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  background: linear-gradient(135deg, #f6fbff 0%, #eef5ff 100%);
  color: #1e2a38;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background: #1f6fb2;
  color: #fff;
}

.topbar h1 {
  margin: 0;
  font-size: 22px;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.role {
  font-size: 14px;
}

button {
  border: 0;
  border-radius: 6px;
  padding: 8px 12px;
  cursor: pointer;
}

.page {
  padding: 16px;
}

.content-wrap {
  max-width: 1200px;
  margin: 0 auto;
}

@media (max-width: 700px) {
  .topbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>
