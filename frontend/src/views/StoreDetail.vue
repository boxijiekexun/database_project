<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import request from '../api/request'

const props = defineProps({
  storeId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['cart-changed', 'review-changed'])

const storeInfo = ref(null)
const categories = ref([])
const dishes = ref([])
const reviews = ref([])
const loading = ref(false)
const cartMap = ref({})

const categoryFilter = ref('')

const statusTextMap = {
  ON_SALE: '上架中',
  OFF_SALE: '已下架'
}

const selectedStoreId = computed(() => props.storeId ? Number(props.storeId) : null)

async function loadStore() {
  if (!selectedStoreId.value) {
    storeInfo.value = null
    categories.value = []
    dishes.value = []
    reviews.value = []
    return
  }
  const [store, categoryList, reviewList] = await Promise.all([
    request.get(`/api/stores/${selectedStoreId.value}`),
    request.get(`/api/stores/${selectedStoreId.value}/categories`),
    request.get(`/api/stores/${selectedStoreId.value}/reviews`)
  ])
  storeInfo.value = store
  categories.value = categoryList
  reviews.value = reviewList
  categoryFilter.value = ''
  await loadDishes()
}

async function loadDishes() {
  if (!selectedStoreId.value) {
    dishes.value = []
    return
  }
  const params = new URLSearchParams()
  if (categoryFilter.value) {
    params.set('category_id', categoryFilter.value)
  }
  const query = params.toString() ? `?${params.toString()}` : ''
  dishes.value = await request.get(`/api/stores/${selectedStoreId.value}/dishes${query}`)
}

function syncCartMap(items) {
  const next = {}
  for (const item of items) {
    next[item.dish_id] = {
      cartItemId: item.id,
      quantity: item.quantity
    }
  }
  cartMap.value = next
}

function dishCartState(dishId) {
  return cartMap.value[dishId] || { cartItemId: null, quantity: 0 }
}

async function addToCart(dish) {
  await request.post('/api/cart/items', { dish_id: dish.id, quantity: 1 })
  emit('cart-changed')
}

async function incrementCart(dish) {
  const state = dishCartState(dish.id)
  await request.put(`/api/cart/items/${state.cartItemId}`, { quantity: state.quantity + 1 })
  emit('cart-changed')
}

async function decrementCart(dish) {
  const state = dishCartState(dish.id)
  if (state.quantity <= 1) {
    await request.delete(`/api/cart/items/${state.cartItemId}`)
  } else {
    await request.put(`/api/cart/items/${state.cartItemId}`, { quantity: state.quantity - 1 })
  }
  emit('cart-changed')
}

function statusText(status) {
  return statusTextMap[status] || status
}

watch(() => props.storeId, async () => {
  loading.value = true
  try {
    await loadStore()
  } finally {
    loading.value = false
  }
})

watch(categoryFilter, async () => {
  await loadDishes()
})

onMounted(async () => {
  if (selectedStoreId.value) {
    loading.value = true
    try {
      await loadStore()
    } finally {
      loading.value = false
    }
  }
})

defineExpose({
  syncCartMap,
  refreshReviews: loadStore,
  reloadReviews: async () => {
    if (!selectedStoreId.value) return
    reviews.value = await request.get(`/api/stores/${selectedStoreId.value}/reviews`)
  }
})
</script>

<template>
  <section v-if="selectedStoreId" class="detail-shell">
    <header v-if="storeInfo" class="store-banner">
      <div>
        <h3>{{ storeInfo.name }}</h3>
        <p>{{ storeInfo.description || '暂无简介' }}</p>
      </div>
      <div class="banner-meta">
        <span>电话：{{ storeInfo.phone || '暂无' }}</span>
        <span>地址：{{ storeInfo.address || '暂无' }}</span>
        <span>状态：{{ storeInfo.status }}</span>
        <span>评分：{{ storeInfo.rating }}</span>
      </div>
    </header>

    <div class="content-grid" v-if="!loading">
      <section class="panel">
        <div class="panel-head">
          <h4>点餐区</h4>
          <select v-model="categoryFilter">
            <option value="">全部分类</option>
            <option v-for="category in categories" :key="category.id" :value="category.id">{{ category.name }}</option>
          </select>
        </div>

        <div v-if="dishes.length === 0" class="empty">暂无菜品</div>
        <div v-for="dish in dishes" :key="dish.id" class="dish-card">
          <div>
            <strong>{{ dish.name }}</strong>
            <p>￥{{ dish.price }} | 库存 {{ dish.stock }} | {{ statusText(dish.status) }}</p>
          </div>
          <div class="counter-area">
            <button v-if="dishCartState(dish.id).quantity === 0" @click="addToCart(dish)" :disabled="dish.status !== 'ON_SALE'">加入购物车</button>
            <div v-else class="counter">
              <button @click="decrementCart(dish)">-</button>
              <span>{{ dishCartState(dish.id).quantity }}</span>
              <button @click="incrementCart(dish)" :disabled="dish.status !== 'ON_SALE'">+</button>
            </div>
          </div>
        </div>
      </section>

      <section class="panel">
        <h4>点评区</h4>
        <div v-if="reviews.length === 0" class="empty">暂无评论</div>
        <div v-for="review in reviews" :key="review.id" class="review-card">
          <div class="review-head">
            <strong>{{ review.username || `用户${review.user_id}` }}</strong>
            <span>评分 {{ review.rating }}</span>
          </div>
          <p>{{ review.content || '无评论内容' }}</p>
          <small>{{ review.created_at }}</small>
        </div>
      </section>
    </div>

    <p v-else class="loading-text">加载中...</p>
  </section>
</template>

<style scoped>
.detail-shell {
  background: #fff;
  border: 1px solid #dbe5ee;
  border-radius: 12px;
  padding: 14px;
  box-shadow: 0 6px 24px rgba(24, 61, 99, 0.08);
}

.store-banner {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 12px;
  border: 1px solid #dbe5ee;
  border-radius: 10px;
  background: linear-gradient(135deg, #f6fbff 0%, #eef5ff 100%);
  margin-bottom: 12px;
}

.store-banner h3 {
  margin: 0 0 6px;
}

.store-banner p {
  margin: 0;
  color: #587189;
}

.banner-meta {
  display: grid;
  gap: 6px;
  min-width: 220px;
  color: #40576f;
  font-size: 14px;
}

.content-grid {
  display: grid;
  grid-template-columns: 1.3fr 1fr;
  gap: 12px;
}

.panel {
  border: 1px solid #dbe5ee;
  border-radius: 10px;
  padding: 12px;
  background: #fff;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.panel-head h4 {
  margin: 0;
}

.panel-head select {
  border: 1px solid #cad7e3;
  border-radius: 6px;
  padding: 6px 10px;
}

.dish-card,
.review-card {
  border: 1px solid #e3ebf3;
  border-radius: 8px;
  padding: 10px;
  margin-bottom: 10px;
}

.dish-card {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.dish-card p,
.review-card p {
  margin: 4px 0 0;
  color: #5f748c;
  font-size: 13px;
}

.counter-area {
  min-width: 140px;
  display: flex;
  justify-content: flex-end;
}

.counter {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  border: 1px solid #cad7e3;
  border-radius: 999px;
  padding: 4px 8px;
}

.counter button,
.counter-area > button {
  background: #1f6fb2;
  color: #fff;
}

.counter button {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  padding: 0;
}

.review-head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.review-head span {
  color: #1f6fb2;
}

.empty,
.loading-text {
  color: #607489;
}

@media (max-width: 900px) {
  .store-banner,
  .content-grid {
    grid-template-columns: 1fr;
    display: grid;
  }

  .counter-area {
    justify-content: flex-start;
  }
}
</style>
