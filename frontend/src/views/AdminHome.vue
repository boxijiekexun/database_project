<script setup>
import { computed, onMounted, ref } from 'vue'
import request from '../api/request'

const orders = ref([])
const currentOrder = ref(null)
const loading = ref(false)
const importFile = ref(null)
const importSubmitting = ref(false)
const importResult = ref(null)
const users = ref([])
const currentUser = ref(null)
const userLoading = ref(false)
const userDetailVisible = ref(false)
const resetPasswordResult = ref(null)
const userFilters = ref({ keyword: '', role: '', status: '' })
const userPagination = ref({ page: 1, pageSize: 10, total: 0 })

const importTemplate = 'username,password,email,phone,role,real_name,student_no,college,class_name\nstudent02,123456,student02@example.com,13000000010,STUDENT,张三,20230001,计算机学院,计科2301\nteacher02,,teacher02@example.com,13000000011,TEACHER,李老师,,信息工程学院,'

const statusTextMap = {
  PENDING_PAYMENT: '待支付',
  PAID: '已支付',
  ACCEPTED: '商家已接单',
  PREPARING: '制作中',
  DELIVERING: '配送中',
  COMPLETED: '已完成',
  CANCELLED: '已取消'
}

const timelineOrder = ['PENDING_PAYMENT', 'PAID', 'ACCEPTED', 'PREPARING', 'DELIVERING', 'COMPLETED']

async function loadOrders() {
  orders.value = await request.get('/api/admin/orders')
}

function buildUserQuery() {
  const params = new URLSearchParams()
  if (userFilters.value.keyword.trim()) params.set('keyword', userFilters.value.keyword.trim())
  if (userFilters.value.role) params.set('role', userFilters.value.role)
  if (userFilters.value.status) params.set('status', userFilters.value.status)
  params.set('page', String(userPagination.value.page))
  params.set('page_size', String(userPagination.value.pageSize))
  return params.toString()
}

async function loadUsers() {
  const data = await request.get(`/api/admin/users?${buildUserQuery()}`)
  users.value = data.items || []
  userPagination.value = {
    page: data.page || userPagination.value.page,
    pageSize: data.page_size || userPagination.value.pageSize,
    total: data.total || 0
  }
}

async function searchUsers() {
  userPagination.value = { ...userPagination.value, page: 1 }
  await loadUsers()
}

async function changeUserStatus(user, status) {
  await request.patch(`/api/admin/users/${user.id}/status`, { status })
  await loadUsers()
  if (currentUser.value?.id === user.id) {
    currentUser.value = await request.get(`/api/admin/users/${user.id}`)
  }
}

async function changeUserRole(user, role) {
  await request.patch(`/api/admin/users/${user.id}/role`, { role })
  await loadUsers()
  if (currentUser.value?.id === user.id) {
    currentUser.value = await request.get(`/api/admin/users/${user.id}`)
  }
}

async function resetUserPassword(user) {
  const newPassword = window.prompt('请输入新密码，留空则默认 123456', '')
  const data = await request.post(`/api/admin/users/${user.id}/reset-password`, { new_password: newPassword || '' })
  resetPasswordResult.value = data
  window.alert(`已重置 ${data.username} 的密码，新密码：${data.new_password}`)
}

async function viewUserDetail(userId) {
  currentUser.value = await request.get(`/api/admin/users/${userId}`)
  userDetailVisible.value = true
}

function closeUserDetail() {
  userDetailVisible.value = false
}

function userRoleOptions() {
  return ['STUDENT', 'TEACHER', 'MERCHANT']
}

function userStatusOptions() {
  return ['ACTIVE', 'DISABLED', 'PENDING']
}

async function prevUserPage() {
  if (userPagination.value.page <= 1) return
  userPagination.value = { ...userPagination.value, page: userPagination.value.page - 1 }
  await loadUsers()
}

async function nextUserPage() {
  const totalPages = Math.max(1, Math.ceil(userPagination.value.total / userPagination.value.pageSize))
  if (userPagination.value.page >= totalPages) return
  userPagination.value = { ...userPagination.value, page: userPagination.value.page + 1 }
  await loadUsers()
}

async function viewDetail(orderId) {
  currentOrder.value = await request.get(`/api/admin/orders/${orderId}`)
}

function onImportFileChange(event) {
  const [file] = event.target.files || []
  importFile.value = file || null
}

async function uploadUsers() {
  if (!importFile.value) {
    window.alert('请先选择 CSV 文件')
    return
  }

  const formData = new FormData()
  formData.append('file', importFile.value)

  importSubmitting.value = true
  try {
    const data = await request.post('/api/admin/users/import', formData)
    importResult.value = data
    importFile.value = null
    const input = document.querySelector('#user-import-file')
    if (input) {
      input.value = ''
    }
  } finally {
    importSubmitting.value = false
  }
}

async function copyTemplate() {
  await navigator.clipboard.writeText(importTemplate)
  window.alert('CSV 模板已复制')
}

function statusText(status) {
  return statusTextMap[status] || status
}

function timelineClass(status, node) {
  if (status === 'CANCELLED') return 'pending'
  const currentIndex = timelineOrder.indexOf(status)
  const nodeIndex = timelineOrder.indexOf(node)
  if (nodeIndex < currentIndex) return 'done'
  if (nodeIndex === currentIndex) return 'active'
  return 'pending'
}

const orderCount = computed(() => orders.value.length)
const completedOrderCount = computed(() => orders.value.filter((o) => o.status === 'COMPLETED').length)
const totalSales = computed(() => {
  return orders.value.reduce((sum, o) => sum + Number(o.total_amount || 0), 0)
})
const pendingOrderCount = computed(() => {
  return orders.value.filter((o) => !['COMPLETED', 'CANCELLED'].includes(o.status)).length
})

onMounted(async () => {
  loading.value = true
  try {
    userLoading.value = true
    await Promise.all([loadOrders(), loadUsers()])
  } finally {
    loading.value = false
    userLoading.value = false
  }
})
</script>

<template>
  <section class="grid" v-if="!loading">
    <article class="panel wide import-panel">
      <div class="panel-head">
        <div>
          <h3>管理员批量导入用户</h3>
          <p class="subtle">上传 CSV 文件即可批量创建学生或教师账号，默认密码为空时为 123456。</p>
        </div>
        <button type="button" class="secondary-btn" @click="copyTemplate">复制模板</button>
      </div>

      <div class="template-box">
        <label>CSV 表头模板</label>
        <pre>{{ importTemplate }}</pre>
      </div>

      <div class="import-controls">
        <input id="user-import-file" type="file" accept=".csv,text/csv" @change="onImportFileChange" />
        <button type="button" :disabled="importSubmitting" @click="uploadUsers">
          {{ importSubmitting ? '导入中...' : '上传并导入' }}
        </button>
      </div>

      <div v-if="importResult" class="result-box">
        <div class="result-summary">
          <span>总行数：{{ importResult.total }}</span>
          <span>成功：{{ importResult.success_count }}</span>
          <span>失败：{{ importResult.failed_count }}</span>
        </div>
        <div v-if="importResult.errors && importResult.errors.length > 0" class="error-list">
          <h4>失败明细</h4>
          <div v-for="error in importResult.errors" :key="`${error.line}-${error.username}-${error.reason}`" class="error-item">
            <span>第 {{ error.line }} 行</span>
            <span>{{ error.username || '未填写用户名' }}</span>
            <span>{{ error.reason }}</span>
          </div>
        </div>
      </div>
    </article>

    <article class="panel wide">
      <div class="panel-head">
        <div>
          <h3>用户管理</h3>
          <p class="subtle">可按关键词、角色和状态筛选，支持修改状态、角色和重置密码。</p>
        </div>
        <button type="button" class="secondary-btn" @click="searchUsers">刷新筛选结果</button>
      </div>

      <div class="user-filters">
        <input v-model="userFilters.keyword" placeholder="username / email / phone" />
        <select v-model="userFilters.role">
          <option value="">全部角色</option>
          <option value="STUDENT">STUDENT</option>
          <option value="TEACHER">TEACHER</option>
          <option value="MERCHANT">MERCHANT</option>
          <option value="ADMIN">ADMIN</option>
        </select>
        <select v-model="userFilters.status">
          <option value="">全部状态</option>
          <option value="ACTIVE">ACTIVE</option>
          <option value="DISABLED">DISABLED</option>
          <option value="PENDING">PENDING</option>
        </select>
        <button type="button" @click="searchUsers">查询</button>
      </div>

      <div v-if="users.length === 0" class="empty-text">暂无用户</div>
      <div v-else class="user-table-wrap">
        <table class="table user-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>用户名</th>
              <th>邮箱</th>
              <th>手机号</th>
              <th>角色</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>{{ user.id }}</td>
              <td>
                <strong>{{ user.username }}</strong>
                <div class="muted tiny" v-if="user.profile && (user.profile.real_name || user.profile.student_no)">
                  {{ user.profile.real_name || '未填姓名' }} · {{ user.profile.student_no || '未填学号' }}
                </div>
              </td>
              <td>{{ user.email || '—' }}</td>
              <td>{{ user.phone || '—' }}</td>
              <td>
                <select :value="user.role" @change="changeUserRole(user, $event.target.value)">
                  <option v-for="role in userRoleOptions()" :key="role" :value="role">{{ role }}</option>
                </select>
              </td>
              <td>
                <select :value="user.status" @change="changeUserStatus(user, $event.target.value)">
                  <option v-for="status in userStatusOptions()" :key="status" :value="status">{{ status }}</option>
                </select>
              </td>
              <td>{{ user.created_at || '—' }}</td>
              <td>
                <div class="row-actions">
                  <button type="button" @click="viewUserDetail(user.id)">详情</button>
                  <button type="button" @click="resetUserPassword(user)">重置密码</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <div class="pagination-row">
          <button type="button" @click="prevUserPage">上一页</button>
          <span>第 {{ userPagination.page }} 页 / 共 {{ Math.max(1, Math.ceil(userPagination.total / userPagination.pageSize)) }} 页，{{ userPagination.total }} 条</span>
          <button type="button" @click="nextUserPage">下一页</button>
        </div>
      </div>

      <div v-if="resetPasswordResult" class="result-box">
        <strong>最近一次重置结果</strong>
        <p>用户名：{{ resetPasswordResult.username }}</p>
        <p>新密码：{{ resetPasswordResult.new_password }}</p>
      </div>
    </article>

    <article class="panel wide">
      <h3>订单统计</h3>
      <div class="stats-grid">
        <div class="stat-card">
          <p class="label">订单总数</p>
          <p class="value">{{ orderCount }}</p>
        </div>
        <div class="stat-card">
          <p class="label">已完成订单</p>
          <p class="value">{{ completedOrderCount }}</p>
        </div>
        <div class="stat-card">
          <p class="label">总销售额</p>
          <p class="value">￥{{ totalSales.toFixed(2) }}</p>
        </div>
        <div class="stat-card">
          <p class="label">待处理订单</p>
          <p class="value">{{ pendingOrderCount }}</p>
        </div>
      </div>
    </article>

    <article class="panel">
      <h3>全局订单列表</h3>
      <div v-if="orders.length === 0">暂无订单</div>
      <div v-for="order in orders" :key="order.id" class="order-card">
        <div>
          <strong>订单#{{ order.id }}</strong>
          <p>用户：{{ order.user_name }} | 店铺：{{ order.store_name }}</p>
          <p>金额：￥{{ order.total_amount }} | 状态：{{ statusText(order.status) }}</p>
        </div>
        <button @click="viewDetail(order.id)">查看详情</button>
      </div>
    </article>

    <article class="panel">
      <h3>订单详情</h3>
      <div v-if="!currentOrder">请选择一个订单</div>
      <div v-else>
        <p><strong>订单号：</strong>{{ currentOrder.order_no }}</p>
        <p><strong>店铺名：</strong>{{ currentOrder.store_name }}</p>
        <p><strong>用户名：</strong>{{ currentOrder.user_name }}</p>
        <p><strong>状态：</strong>{{ statusText(currentOrder.status) }}</p>
        <p><strong>总金额：</strong>￥{{ currentOrder.total_amount }}</p>
        <p><strong>备注：</strong>{{ currentOrder.remark || '无' }}</p>
        <p><strong>创建时间：</strong>{{ currentOrder.created_at }}</p>

        <h4>订单状态时间线</h4>
        <p v-if="currentOrder.status === 'CANCELLED'" class="cancel-text">订单已取消</p>
        <div v-else class="timeline">
          <div v-for="node in timelineOrder" :key="node" class="timeline-item" :class="timelineClass(currentOrder.status, node)">
            <span class="dot"></span>
            <span>{{ statusText(node) }}</span>
          </div>
        </div>

        <h4>明细</h4>
        <table class="table">
          <thead>
            <tr>
              <th>菜品名</th>
              <th>单价</th>
              <th>数量</th>
              <th>小计</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in currentOrder.items || []" :key="item.id || `${item.dish_id}-${item.dish_name}`">
              <td>{{ item.dish_name }}</td>
              <td>￥{{ item.dish_price }}</td>
              <td>{{ item.quantity }}</td>
              <td>￥{{ item.subtotal }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <div v-if="userDetailVisible" class="modal-mask" @click.self="closeUserDetail">
      <section class="modal">
        <header class="modal-head">
          <h3>用户详情</h3>
          <button @click="closeUserDetail">关闭</button>
        </header>
        <div v-if="currentUser" class="detail-grid user-detail-grid">
          <p><strong>ID：</strong>{{ currentUser.id }}</p>
          <p><strong>用户名：</strong>{{ currentUser.username }}</p>
          <p><strong>邮箱：</strong>{{ currentUser.email || '无' }}</p>
          <p><strong>手机号：</strong>{{ currentUser.phone || '无' }}</p>
          <p><strong>角色：</strong>{{ currentUser.role }}</p>
          <p><strong>状态：</strong>{{ currentUser.status }}</p>
          <p><strong>创建时间：</strong>{{ currentUser.created_at || '无' }}</p>
          <p><strong>更新时间：</strong>{{ currentUser.updated_at || '无' }}</p>
          <p><strong>姓名：</strong>{{ currentUser.profile?.real_name || '无' }}</p>
          <p><strong>学号：</strong>{{ currentUser.profile?.student_no || '无' }}</p>
          <p><strong>学院：</strong>{{ currentUser.profile?.college || '无' }}</p>
          <p><strong>班级：</strong>{{ currentUser.profile?.class_name || '无' }}</p>
        </div>
      </section>
    </div>
  </section>
  <p v-else>加载中...</p>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.wide {
  grid-column: 1 / -1;
}

.panel {
  background: #fff;
  border-radius: 10px;
  padding: 12px;
  box-shadow: 0 6px 24px rgba(24, 61, 99, 0.08);
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.panel-head h3 {
  margin: 0;
}

.subtle {
  margin: 6px 0 0;
  color: #6c7f92;
  font-size: 13px;
}

.secondary-btn {
  background: #eef5fb;
  color: #1f6fb2;
  border: 1px solid #c9d9e8;
}

.secondary-btn:hover {
  background: #e4f0fb;
}

.user-filters {
  display: grid;
  grid-template-columns: 1.6fr repeat(2, minmax(140px, 1fr)) auto;
  gap: 10px;
  margin-bottom: 12px;
}

.user-filters input,
.user-filters select,
.user-table select {
  border: 1px solid #cad7e3;
  border-radius: 6px;
  padding: 8px;
  background: #fff;
}

.user-table-wrap {
  overflow-x: auto;
}

.user-table {
  min-width: 100%;
}

.tiny {
  font-size: 12px;
}

.row-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.pagination-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  color: #51657b;
}

.result-box {
  margin-top: 12px;
  border-top: 1px solid #e6edf4;
  padding-top: 12px;
}

.empty-text {
  color: #607489;
}

.user-detail-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.panel-head h3 {
  margin: 0;
}

.subtle {
  margin: 6px 0 0;
  color: #6c7f92;
  font-size: 13px;
}

.template-box {
  border: 1px dashed #c8d7e6;
  border-radius: 8px;
  padding: 10px;
  background: #f7fbff;
}

.template-box label {
  display: inline-block;
  margin-bottom: 8px;
  font-weight: 700;
  color: #385069;
}

.template-box pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.6;
  color: #213547;
}

.import-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  margin-top: 12px;
}

.import-controls input[type='file'] {
  padding: 8px;
  border: 1px solid #dbe5ee;
  border-radius: 8px;
  background: #fff;
}

.secondary-btn {
  background: #eef5fb;
  color: #1f6fb2;
  border: 1px solid #c9d9e8;
}

.secondary-btn:hover {
  background: #e4f0fb;
}

.result-box {
  margin-top: 12px;
  border-top: 1px solid #e6edf4;
  padding-top: 12px;
}

.result-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: #32485d;
  font-weight: 700;
}

.error-list {
  margin-top: 12px;
}

.error-list h4 {
  margin: 0 0 8px;
}

.error-item {
  display: grid;
  grid-template-columns: 90px 140px 1fr;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid #f1d1d1;
  border-radius: 8px;
  background: #fff7f7;
  margin-bottom: 8px;
  color: #9a4040;
}

.order-card {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid #e0e8f0;
  border-radius: 8px;
  padding: 8px;
  margin-bottom: 8px;
}

.order-card p {
  margin: 4px 0 0;
  font-size: 13px;
  color: #5f748c;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(140px, 1fr));
  gap: 10px;
}

.stat-card {
  border: 1px solid #dbe5ee;
  border-radius: 8px;
  padding: 10px;
  background: #f7fbff;
}

.label {
  margin: 0;
  color: #607489;
  font-size: 13px;
}

.value {
  margin: 6px 0 0;
  font-size: 20px;
  font-weight: 700;
}

.timeline {
  display: grid;
  gap: 8px;
  margin: 8px 0;
}

.timeline-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #7b8fa3;
}

.timeline-item .dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid #c9d5df;
}

.timeline-item.done {
  color: #198754;
}

.timeline-item.done .dot {
  border-color: #198754;
  background: #198754;
}

.timeline-item.active {
  color: #1f6fb2;
  font-weight: 700;
}

.timeline-item.active .dot {
  border-color: #1f6fb2;
  background: #1f6fb2;
}

.cancel-text {
  color: #a63b3b;
  font-weight: 700;
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th,
.table td {
  border: 1px solid #dbe5ee;
  padding: 8px;
  text-align: left;
}

button {
  background: #1f6fb2;
  color: #fff;
}

@media (max-width: 900px) {
  .grid {
    grid-template-columns: 1fr;
  }

  .stats-grid {
    grid-template-columns: repeat(2, minmax(120px, 1fr));
  }
}
</style>
