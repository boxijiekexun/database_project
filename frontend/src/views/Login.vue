<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import request from '../api/request'

const router = useRouter()
const loading = ref(false)
const form = reactive({
  username: 'student01',
  password: '123456'
})

function homeByRole(role) {
  if (role === 'STUDENT' || role === 'TEACHER') return '/user'
  if (role === 'MERCHANT') return '/merchant'
  if (role === 'ADMIN') return '/admin'
  return '/login'
}

async function submit() {
  loading.value = true
  try {
    const data = await request.post('/api/auth/login', form)
    localStorage.setItem('token', data.token)
    localStorage.setItem('role', data.user.role)
    router.push(homeByRole(data.user.role))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="login-wrap">
    <div class="card">
      <h2>登录</h2>
      <p class="tip">示例账号：student01 / 123456</p>
      <form @submit.prevent="submit">
        <label>
          用户名
          <input v-model="form.username" required />
        </label>
        <label>
          密码
          <input v-model="form.password" type="password" required />
        </label>
        <button :disabled="loading" type="submit">{{ loading ? '登录中...' : '登录' }}</button>
      </form>
    </div>
  </section>
</template>

<style scoped>
.login-wrap {
  min-height: calc(100vh - 100px);
  display: grid;
  place-items: center;
}

.card {
  width: 320px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 10px 30px rgba(20, 53, 92, 0.12);
  padding: 20px;
}

h2 {
  margin: 0 0 8px;
}

.tip {
  margin-top: 0;
  color: #57708c;
  font-size: 13px;
}

form {
  display: grid;
  gap: 12px;
}

label {
  display: grid;
  gap: 6px;
  font-size: 14px;
}

input {
  border: 1px solid #cad7e3;
  border-radius: 6px;
  padding: 8px 10px;
}

button {
  background: #1f6fb2;
  color: #fff;
  font-weight: 700;
}

button:disabled {
  opacity: 0.7;
}
</style>
