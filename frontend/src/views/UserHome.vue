<script setup>
import { nextTick, onMounted, ref, watch } from 'vue'
import request from '../api/request'
import StoreDetail from './StoreDetail.vue'

const stores = ref([])
const areas = ref([])
const canteens = ref([])
const cartItems = ref([])
const orders = ref([])
const selectedStoreId = ref(null)
const selectedAreaId = ref(null)
const selectedCanteenId = ref(null)
const currentCanteen = ref(null)
const recommendedDishes = ref([])
const loading = ref(false)
const detailVisible = ref(false)
const currentOrder = ref(null)
const reviewedOrderMap = ref({})
const storeDetailRef = ref(null)

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

async function loadAreas() {
  areas.value = await request.get('/api/dining-areas')
}

async function loadCanteens(areaId) {
  if (!areaId) {
    canteens.value = []
    return
  }
  canteens.value = await request.get(`/api/dining-areas/${areaId}/canteens`)
}

async function loadStores() {
  if (!selectedCanteenId.value) {
    stores.value = []
    return
  }
  stores.value = await request.get(`/api/canteens/${selectedCanteenId.value}/stores`)
}

async function loadCurrentCanteen() {
  if (!selectedCanteenId.value) {
    currentCanteen.value = null
    return
  }
  currentCanteen.value = await request.get(`/api/canteens/${selectedCanteenId.value}`)
}

async function loadRecommendedDishes() {
  recommendedDishes.value = await request.get('/api/ai/recommend/dishes')
}

async function loadCart() {
  cartItems.value = await request.get('/api/cart/items')
  syncCartMap()
}

async function loadOrders() {
  orders.value = await request.get('/api/orders')
}

async function selectArea(areaId) {
  selectedAreaId.value = areaId
  await loadCanteens(areaId)
  const firstCanteen = canteens.value[0]
  if (firstCanteen) {
    await selectCanteen(firstCanteen.id)
  } else {
    selectedCanteenId.value = null
    currentCanteen.value = null
    stores.value = []
    selectedStoreId.value = null
    syncCartMap()
  }
}

async function selectCanteen(canteenId) {
  selectedCanteenId.value = canteenId
  await loadCurrentCanteen()
  await loadStores()
  selectedStoreId.value = stores.value[0]?.id || null
  await nextTick()
  syncCartMap()
}

function syncCartMap() {
  if (storeDetailRef.value) {
    storeDetailRef.value.syncCartMap(cartItems.value)
  }
}

async function handleStoreSelect(storeId) {
  selectedStoreId.value = storeId
  await nextTick()
  syncCartMap()
}

async function jumpToDishStore(dish) {
  const storeInfo = await request.get(`/api/stores/${dish.store_id}`)
  if (storeInfo?.canteen_id) {
    const canteenId = Number(storeInfo.canteen_id)
    const canteenInfo = await request.get(`/api/canteens/${canteenId}`)
    selectedAreaId.value = canteenInfo.area_id
    await loadCanteens(canteenInfo.area_id)
    await selectCanteen(canteenId)
  }
  selectedStoreId.value = dish.store_id
  await nextTick()
  syncCartMap()
}

async function addRecommendedDish(dish) {
  await request.post('/api/cart/items', { dish_id: dish.dish_id, quantity: 1 })
  await handleCartChanged()
  window.alert(`已将 ${dish.dish_name} 加入购物车`)
}

async function handleCartChanged() {
  await loadCart()
  await loadOrders()
}

async function createOrder() {
  await request.post('/api/orders', { remark: '前端演示下单' })
  await Promise.all([loadCart(), loadOrders()])
  window.alert('下单成功，已刷新购物车和订单列表')
}

async function payOrder(orderId) {
  await request.post(`/api/orders/${orderId}/pay`)
  await loadOrders()
  window.alert('支付成功')
}

async function reviewOrder(order) {
  const rating = Number(window.prompt('请输入评分(1-5)', '5'))
  if (!rating || rating < 1 || rating > 5) {
    window.alert('评分必须在1-5之间')
    return
  }
  const content = window.prompt('请输入评价内容', '味道不错，配送及时') || ''
  await request.post(`/api/orders/${order.id}/review`, {
    rating,
    content
  })
  reviewedOrderMap.value = { ...reviewedOrderMap.value, [order.id]: true }
  await loadOrders()
  if (selectedStoreId.value && Number(selectedStoreId.value) === Number(order.store_id)) {
    await storeDetailRef.value?.reloadReviews?.()
  }
  window.alert('评价成功，评论区已刷新')
}

function canPay(order) {
  return order.status === 'PENDING_PAYMENT'
}

function canReview(order) {
  return order.status === 'COMPLETED' && !reviewedOrderMap.value[order.id]
}

function statusText(status) {
  return statusTextMap[status] || status
}

async function viewOrderDetail(order) {
  if (Array.isArray(order.items) && order.items.length > 0) {
    currentOrder.value = order
  } else {
    currentOrder.value = await request.get(`/api/orders/${order.id}`)
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
    await Promise.all([loadAreas(), loadCart(), loadOrders(), loadRecommendedDishes()])
    if (areas.value.length > 0) {
      await selectArea(areas.value[0].id)
    }
  } finally {
    loading.value = false
  }
})

watch(cartItems, () => {
  syncCartMap()
}, { deep: true })
</script>

<template>
  <section v-if="!loading" class="dashboard-shell">
    <article class="panel dining-panel">
      <div class="dining-row">
        <div class="dining-group">
          <strong>食堂区域</strong>
          <div class="chip-row">
            <button
              v-for="area in areas"
              :key="area.id"
              class="chip-btn"
              :class="{ active: selectedAreaId === area.id }"
              @click="selectArea(area.id)"
            >
              {{ area.name }}
            </button>
          </div>
        </div>

        <div class="dining-group">
          <strong>食堂</strong>
          <div class="chip-row">
            <button
              v-for="canteen in canteens"
              :key="canteen.id"
              class="chip-btn"
              :class="{ active: selectedCanteenId === canteen.id }"
              @click="selectCanteen(canteen.id)"
            >
              {{ canteen.name }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="currentCanteen" class="canteen-summary">
        <div>
          <h3>{{ currentCanteen.name }}</h3>
          <p>{{ currentCanteen.description }}</p>
        </div>
        <div class="specialties-box">
          <strong>特色美食</strong>
          <p>{{ currentCanteen.specialties }}</p>
        </div>
      </div>
    </article>

    <section class="page-grid">
    <aside class="panel sidebar">
      <h3>店铺列表</h3>
      <div v-if="stores.length === 0" class="empty">暂无店铺</div>
      <button
        v-for="store in stores"
        :key="store.id"
        class="store-row"
        :class="{ active: selectedStoreId === store.id }"
        @click="handleStoreSelect(store.id)"
      >
        <span class="store-name">{{ store.name }}</span>
        <span class="store-meta">{{ store.status }} · {{ store.rating }}</span>
      </button>
    </aside>

    <main class="main-area">
      <StoreDetail
        ref="storeDetailRef"
        :store-id="selectedStoreId"
        @cart-changed="handleCartChanged"
      />

      <article class="panel recommend-panel">
        <div class="panel-title-row">
          <h3>AI 推荐菜品</h3>
          <span class="subtle">按订单统计和销量综合排序</span>
        </div>
        <div v-if="recommendedDishes.length === 0" class="empty">暂无推荐数据</div>
        <div v-for="dish in recommendedDishes" :key="dish.dish_id" class="recommend-card">
          <div>
            <strong>{{ dish.dish_name }}</strong>
            <p>{{ dish.store_name }} · ￥{{ dish.price }}</p>
            <p class="reason">{{ dish.reason }}</p>
          </div>
          <div class="recommend-actions">
            <button @click="jumpToDishStore(dish)">去查看</button>
            <button @click="addRecommendedDish(dish)">加入购物车</button>
          </div>
        </div>
      </article>

      <section class="lower-grid">
        <article class="panel">
          <h3>购物车</h3>
          <div v-if="cartItems.length === 0" class="empty">购物车为空</div>
          <div v-for="item in cartItems" :key="item.id" class="cart-card">
            <div>
              <strong>{{ item.dish_name }}</strong>
              <p>数量 {{ item.quantity }} | 小计 ￥{{ item.subtotal }}</p>
            </div>
          </div>
          <button class="primary-btn" @click="createOrder" :disabled="cartItems.length === 0">创建订单</button>
        </article>

        <article class="panel wide-orders">
          <h3>我的订单</h3>
          <div v-if="orders.length === 0" class="empty">暂无订单</div>
          <div v-for="order in orders" :key="order.id" class="order-card">
            <div>
              <strong>订单#{{ order.id }}</strong>
              <p>店铺：{{ order.store_name }} | 金额：￥{{ order.total_amount }} | 状态：{{ statusText(order.status) }}</p>
              <p>创建时间：{{ order.created_at }}</p>
            </div>
            <div class="actions">
              <button @click="viewOrderDetail(order)">查看详情</button>
              <button @click="payOrder(order.id)" :disabled="!canPay(order)">支付</button>
              <button v-if="canReview(order)" @click="reviewOrder(order)">评价</button>
              <span v-else-if="order.status === 'COMPLETED'" class="badge done">已评价或待确认</span>
            </div>
          </div>
        </article>
      </section>
    </main>

    <div v-if="detailVisible" class="modal-mask" @click.self="closeDetail">
      <section class="modal">
        <header class="modal-head">
          <h3>订单详情</h3>
          <button @click="closeDetail">关闭</button>
        </header>
        <div v-if="currentOrder" class="detail-grid">
          <p><strong>订单号：</strong>{{ currentOrder.order_no }}</p>
          <p><strong>店铺名：</strong>{{ currentOrder.store_name }}</p>
          <p><strong>用户名：</strong>{{ currentOrder.user_name || '当前用户' }}</p>
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
  </section>
  <p v-else>加载中...</p>
</template>

<style scoped>
.page-grid {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 12px;
  align-items: start;
}

.dashboard-shell {
  display: grid;
  gap: 12px;
}

.dining-panel {
  display: grid;
  gap: 12px;
}

.dining-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.dining-group strong {
  display: block;
  margin-bottom: 8px;
  color: #26425f;
}

.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip-btn {
  background: #eef5fb;
  color: #1f6fb2;
  border: 1px solid #c9d9e8;
}

.chip-btn.active {
  background: #1f6fb2;
  color: #fff;
}

.canteen-summary {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  border: 1px dashed #c8d7e6;
  border-radius: 10px;
  padding: 12px;
  background: #f7fbff;
}

.canteen-summary h3 {
  margin: 0 0 6px;
}

.canteen-summary p {
  margin: 0;
  color: #587189;
}

.specialties-box {
  min-width: 220px;
}

.panel-title-row {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: baseline;
  margin-bottom: 10px;
}

.panel-title-row h3 {
  margin: 0;
}

.subtle {
  color: #667b90;
  font-size: 13px;
}

.recommend-panel {
  margin-bottom: 12px;
}

.recommend-card {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 10px;
  border: 1px solid #e3ebf3;
  border-radius: 8px;
  margin-bottom: 8px;
}

.recommend-card p {
  margin: 4px 0 0;
  color: #5f748c;
  font-size: 13px;
}

.reason {
  color: #294860;
}

.recommend-actions {
  display: grid;
  gap: 8px;
  align-content: start;
  justify-items: end;
}

.panel {
  background: #fff;
  border-radius: 10px;
  padding: 12px;
  border: 1px solid #dbe5ee;
  box-shadow: 0 6px 24px rgba(24, 61, 99, 0.08);
}

.sidebar {
  position: sticky;
  top: 16px;
}

.store-row {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  text-align: left;
  margin-bottom: 8px;
  background: #f3f8fd;
  border: 1px solid #dbe5ee;
}

.store-row.active {
  border-color: #1f6fb2;
  background: #eaf4ff;
}

.store-name {
  font-weight: 700;
}

.store-meta {
  font-size: 12px;
  color: #63778c;
}

.main-area {
  display: grid;
  gap: 12px;
}

.lower-grid {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: 12px;
}

.wide-orders {
  min-width: 0;
}

.cart-card,
.order-card {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid #e0e8f0;
  border-radius: 8px;
  padding: 8px;
  margin-bottom: 8px;
}

.order-card p,
.cart-card p {
  margin: 4px 0 0;
  font-size: 13px;
  color: #5f748c;
}

.actions {
  display: grid;
  gap: 6px;
  align-content: start;
  justify-items: end;
}

.primary-btn,
.actions button,
.store-row,
button {
  background: #1f6fb2;
  color: #fff;
}

button:disabled {
  opacity: 0.55;
}

.empty {
  color: #607489;
}

.badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 999px;
  font-size: 12px;
}

.badge.done {
  background: #e7f7ec;
  color: #198754;
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

@media (max-width: 1100px) {
  .dining-row,
  .page-grid,
  .lower-grid,
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .sidebar {
    position: static;
  }
}
</style>
