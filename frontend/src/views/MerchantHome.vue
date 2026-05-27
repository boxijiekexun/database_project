<script setup>
import { onMounted, reactive, ref } from 'vue'
import request from '../api/request'

const store = ref(null)
const categories = ref([])
const dishes = ref([])
const orders = ref([])
const loading = ref(false)
const detailVisible = ref(false)
const currentOrder = ref(null)

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

const dishForm = reactive({
  category_id: '',
  name: '',
  description: '',
  price: '',
  stock: ''
})
const aiDescriptionLoading = ref(false)
const aiPromptLoading = ref(false)
const aiDescriptionResult = ref(null)
const aiPromptResult = ref(null)

const nextOrderStatusMap = {
  PAID: 'ACCEPTED',
  ACCEPTED: 'PREPARING',
  PREPARING: 'DELIVERING',
  DELIVERING: 'COMPLETED'
}

const nextActionTextMap = {
  PAID: '接单',
  ACCEPTED: '开始制作',
  PREPARING: '开始配送',
  DELIVERING: '完成订单'
}

async function loadBase() {
  store.value = await request.get('/api/merchant/store')
  categories.value = await request.get('/api/merchant/categories')
  dishes.value = await request.get('/api/merchant/dishes')
  orders.value = await request.get('/api/merchant/orders')
}

async function createDish() {
  await request.post('/api/merchant/dishes', {
    category_id: Number(dishForm.category_id),
    name: dishForm.name,
    description: dishForm.description,
    price: Number(dishForm.price),
    stock: Number(dishForm.stock),
    status: 'ON_SALE'
  })
  dishForm.name = ''
  dishForm.description = ''
  dishForm.price = ''
  dishForm.stock = ''
  await loadDishes()
}

async function generateDishDescription() {
  if (!dishForm.name.trim()) {
    window.alert('请先输入菜品名')
    return
  }
  aiDescriptionLoading.value = true
  try {
    const data = await request.post('/api/ai/dish-description', {
      dish_name: dishForm.name,
      ingredients: dishForm.description,
      style: '校园食堂'
    })
    aiDescriptionResult.value = data
    if (data?.description) {
      dishForm.description = data.description
    }
  } finally {
    aiDescriptionLoading.value = false
  }
}

async function generateImagePrompt() {
  if (!dishForm.name.trim()) {
    window.alert('请先输入菜品名')
    return
  }
  aiPromptLoading.value = true
  try {
    aiPromptResult.value = await request.post('/api/ai/dish-image-prompt', {
      dish_name: dishForm.name,
      description: dishForm.description,
      style: '真实食物摄影'
    })
  } finally {
    aiPromptLoading.value = false
  }
}

async function loadDishes() {
  dishes.value = await request.get('/api/merchant/dishes')
}

async function toggleDishStatus(dish) {
  const status = dish.status === 'ON_SALE' ? 'OFF_SALE' : 'ON_SALE'
  await request.patch(`/api/merchant/dishes/${dish.id}/status`, { status })
  await loadDishes()
}

async function moveOrder(order) {
  const next = nextOrderStatusMap[order.status]
  if (!next) return
  await request.patch(`/api/merchant/orders/${order.id}/status`, { status: next })
  window.alert('订单状态更新成功')
  await loadOrders()
}

async function loadOrders() {
  orders.value = await request.get('/api/merchant/orders')
}

function nextStatus(order) {
  return nextOrderStatusMap[order.status] || ''
}

function nextActionText(order) {
  return nextActionTextMap[order.status] || ''
}

function statusText(status) {
  return statusTextMap[status] || status
}

async function viewOrderDetail(order) {
  if (Array.isArray(order.items) && order.items.length > 0) {
    currentOrder.value = order
  } else {
    currentOrder.value = await request.get(`/api/merchant/orders/${order.id}`)
  }
  detailVisible.value = true
}

function closeDetail() {
  detailVisible.value = false
}

function timelineClass(status, node) {
  if (status === 'CANCELLED') return 'pending'
  const currentIndex = timelineOrder.indexOf(status)
  const nodeIndex = timelineOrder.indexOf(node)
  if (nodeIndex < currentIndex) return 'done'
  if (nodeIndex === currentIndex) return 'active'
  return 'pending'
}

onMounted(async () => {
  loading.value = true
  try {
    await loadBase()
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section v-if="!loading" class="grid">
    <article class="panel">
      <h3>我的店铺</h3>
      <div v-if="store">
        <p><strong>{{ store.name }}</strong></p>
        <p>状态：{{ store.status }}</p>
        <p>地址：{{ store.address }}</p>
      </div>
    </article>

    <article class="panel">
      <h3>分类</h3>
      <ul>
        <li v-for="c in categories" :key="c.id">{{ c.id }} - {{ c.name }}</li>
      </ul>
    </article>

    <article class="panel wide">
      <h3>新增菜品</h3>
      <form class="form" @submit.prevent="createDish">
        <select v-model="dishForm.category_id" required>
          <option value="" disabled>选择分类</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <input v-model="dishForm.name" placeholder="菜品名" required />
        <input v-model="dishForm.description" placeholder="描述" />
        <input v-model="dishForm.price" type="number" min="0" step="0.01" placeholder="价格" required />
        <input v-model="dishForm.stock" type="number" min="0" placeholder="库存" required />
        <div class="ai-actions">
          <button type="button" @click="generateDishDescription" :disabled="aiDescriptionLoading">
            {{ aiDescriptionLoading ? '生成中...' : 'AI 生成菜品文案' }}
          </button>
          <button type="button" @click="generateImagePrompt" :disabled="aiPromptLoading">
            {{ aiPromptLoading ? '生成中...' : 'AI 生成图片提示词' }}
          </button>
        </div>
        <button type="submit">新增菜品</button>
      </form>
      <div v-if="aiDescriptionResult || aiPromptResult" class="ai-result-box">
        <div v-if="aiDescriptionResult" class="ai-result-card">
          <strong>菜品文案</strong>
          <p>{{ aiDescriptionResult.title }}</p>
          <p>{{ aiDescriptionResult.description }}</p>
          <p class="muted">{{ aiDescriptionResult.recommend_reason }}</p>
          <p class="muted" v-if="aiDescriptionResult.message">{{ aiDescriptionResult.message }}</p>
        </div>
        <div v-if="aiPromptResult" class="ai-result-card">
          <strong>图片提示词</strong>
          <p>{{ aiPromptResult.prompt }}</p>
          <p class="muted">Negative: {{ aiPromptResult.negative_prompt }}</p>
          <p class="muted">{{ aiPromptResult.note }}</p>
          <p class="muted" v-if="aiPromptResult.message">{{ aiPromptResult.message }}</p>
        </div>
      </div>
    </article>

    <article class="panel wide">
      <h3>菜品列表</h3>
      <div class="item-card" v-for="dish in dishes" :key="dish.id">
        <div>
          <strong>{{ dish.name }}</strong>
          <p>￥{{ dish.price }} | 库存 {{ dish.stock }} | {{ dish.status }}</p>
        </div>
        <button @click="toggleDishStatus(dish)">{{ dish.status === 'ON_SALE' ? '下架' : '上架' }}</button>
      </div>
    </article>

    <article class="panel wide">
      <h3>订单管理</h3>
      <div class="item-card" v-for="order in orders" :key="order.id">
        <div>
          <strong>订单#{{ order.id }}</strong>
          <p>用户：{{ order.user_name }} | 金额：￥{{ order.total_amount }} | 状态：{{ statusText(order.status) }}</p>
        </div>
        <div class="actions">
          <button @click="viewOrderDetail(order)">查看详情</button>
          <button v-if="nextStatus(order)" @click="moveOrder(order)">{{ nextActionText(order) }}</button>
        </div>
      </div>
    </article>

    <div v-if="detailVisible" class="modal-mask" @click.self="closeDetail">
      <section class="modal">
        <header class="modal-head">
          <h3>订单详情</h3>
          <button @click="closeDetail">关闭</button>
        </header>
        <div v-if="currentOrder" class="detail-grid">
          <p><strong>订单号：</strong>{{ currentOrder.order_no }}</p>
          <p><strong>店铺名：</strong>{{ currentOrder.store_name }}</p>
          <p><strong>用户名：</strong>{{ currentOrder.user_name }}</p>
          <p><strong>总金额：</strong>￥{{ currentOrder.total_amount }}</p>
          <p><strong>状态：</strong>{{ statusText(currentOrder.status) }}</p>
          <p><strong>备注：</strong>{{ currentOrder.remark || '无' }}</p>
          <p><strong>创建时间：</strong>{{ currentOrder.created_at }}</p>
        </div>

        <div class="timeline-wrap">
          <h4>订单状态时间线</h4>
          <p v-if="currentOrder?.status === 'CANCELLED'" class="cancel-text">订单已取消</p>
          <div v-else class="timeline">
            <div v-for="node in timelineOrder" :key="node" class="timeline-item" :class="timelineClass(currentOrder.status, node)">
              <span class="dot"></span>
              <span>{{ statusText(node) }}</span>
            </div>
          </div>
        </div>

        <h4>订单明细</h4>
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
            <tr v-for="item in currentOrder?.items || []" :key="item.id || `${item.dish_id}-${item.dish_name}`">
              <td>{{ item.dish_name }}</td>
              <td>￥{{ item.dish_price }}</td>
              <td>{{ item.quantity }}</td>
              <td>￥{{ item.subtotal }}</td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>
  </section>
  <p v-else>加载中...</p>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(280px, 1fr));
  gap: 12px;
}

.panel {
  background: #fff;
  border-radius: 10px;
  padding: 12px;
  box-shadow: 0 6px 24px rgba(24, 61, 99, 0.08);
}

.wide {
  grid-column: 1 / -1;
}

.form {
  display: grid;
  grid-template-columns: repeat(6, minmax(120px, 1fr));
  gap: 8px;
}

.ai-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  grid-column: 1 / -1;
}

.ai-result-box {
  display: grid;
  gap: 10px;
  margin-top: 10px;
}

.ai-result-card {
  border: 1px dashed #c8d7e6;
  border-radius: 8px;
  padding: 10px;
  background: #f7fbff;
}

.muted {
  color: #607489;
}

input,
select {
  border: 1px solid #cad7e3;
  border-radius: 6px;
  padding: 8px;
}

.item-card {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid #e0e8f0;
  border-radius: 8px;
  padding: 8px;
  margin-bottom: 8px;
}

.actions {
  display: grid;
  gap: 8px;
  align-content: start;
  justify-items: end;
}

.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: grid;
  place-items: center;
  z-index: 1000;
  padding: 12px;
}

.modal {
  background: #fff;
  border-radius: 10px;
  padding: 12px;
  width: min(900px, 100%);
  max-height: 90vh;
  overflow: auto;
}

.modal-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px 12px;
}

.timeline-wrap {
  margin: 10px 0;
}

.timeline {
  display: grid;
  gap: 8px;
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

.item-card p {
  margin: 4px 0 0;
  font-size: 13px;
  color: #5f748c;
}

button {
  background: #1f6fb2;
  color: #fff;
}

button:disabled {
  opacity: 0.55;
}

@media (max-width: 900px) {
  .grid {
    grid-template-columns: 1fr;
  }

  .form {
    grid-template-columns: 1fr;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
