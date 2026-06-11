<template>
  <div>
    <div class="editor-header">
      <div>
        <h2 style="margin: 0">{{ customer?.groom_name }} & {{ customer?.bride_name }} 报价方案</h2>
        <a-tag v-if="customer" :color="getStatusColor(customer.status)">{{ getStatusText(customer.status) }}</a-tag>
      </div>
      <div class="header-actions">
        <a-space>
          <a-radio-group v-model:value="quoteType">
            <a-radio-button value="package">选择套餐</a-radio-button>
            <a-radio-button value="custom">自定义组合</a-radio-button>
          </a-radio-group>
          <a-button type="primary" @click="saveQuote" :loading="saving">保存方案</a-button>
        </a-space>
      </div>
    </div>

    <a-row :gutter="24">
      <a-col :span="16">
        <a-card v-if="quoteType === 'package'" title="选择套餐">
          <a-row :gutter="16">
            <a-col :span="12" v-for="pkg in packages" :key="pkg.id">
              <a-card hoverable :class="{ 'selected': selectedPackageId === pkg.id }" @click="selectPackage(pkg.id)">
                <div class="pkg-header">
                  <span class="pkg-name">{{ pkg.name }}</span>
                  <CheckOutlined v-if="selectedPackageId === pkg.id" class="check-icon" />
                </div>
                <div class="pkg-price">¥{{ pkg.total_price.toLocaleString() }}</div>
                <div class="pkg-desc">{{ pkg.description }}</div>
              </a-card>
            </a-col>
          </a-row>
        </a-card>

        <a-card v-if="quoteType === 'custom'" title="服务项列表（可拖拽调整顺序）">
          <div class="service-list">
            <div
              v-for="item in serviceItems"
              :key="item.id"
              class="service-item"
              draggable="true"
              @dragstart="onDragStart($event, item)"
              @dragover.prevent
              @drop="onDrop($event, item)"
            >
              <div class="item-left">
                <MenuOutlined class="drag-handle" />
                <div>
                  <div class="item-name">{{ item.name }}</div>
                  <div class="item-range">价格范围: ¥{{ item.base_price_min.toLocaleString() }} - ¥{{ item.base_price_max.toLocaleString() }}/{{ item.unit }}</div>
                </div>
              </div>
              <a-button type="primary" size="small" @click="addItem(item)">
                <PlusOutlined /> 添加
              </a-button>
            </div>
          </div>
        </a-card>

        <a-card title="已选服务项" style="margin-top: 16px">
          <a-table
            :columns="itemColumns"
            :data-source="selectedItems"
            :pagination="false"
            row-key="_id"
            @change="updateTotal"
          >
            <template #bodyCell="{ column, record, index }">
              <template v-if="column.key === 'name'">
                {{ record.service_item?.name || record.name }}
              </template>
              <template v-else-if="column.key === 'specification'">
                <a-input v-model:value="record.specification" placeholder="规格描述" size="small" />
              </template>
              <template v-else-if="column.key === 'quantity'">
                <a-input-number v-model:value="record.quantity" :min="1" size="small" @change="updateSubtotal(record)" />
              </template>
              <template v-else-if="column.key === 'unit_price'">
                <a-input-number v-model:value="record.unit_price" :min="0" size="small" @change="updateSubtotal(record)" />
              </template>
              <template v-else-if="column.key === 'subtotal'">
                ¥{{ record.subtotal?.toLocaleString() || 0 }}
              </template>
              <template v-else-if="column.key === 'action'">
                <a-button type="text" danger size="small" @click="removeItem(index)">
                  <DeleteOutlined />
                </a-button>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>

      <a-col :span="8">
        <a-card title="价格汇总" class="price-card">
          <div class="price-row">
            <span>服务项小计:</span>
            <span>¥{{ itemsTotal.toLocaleString() }}</span>
          </div>
          <a-divider />
          <div class="price-row total">
            <span>方案总价:</span>
            <span class="total-price">¥{{ totalPrice.toLocaleString() }}</span>
          </div>
          <div class="price-row" v-if="customer">
            <span>客户预算:</span>
            <span>¥{{ ((customer.budget_min || 0) * 10000).toLocaleString() }} - ¥{{ ((customer.budget_max || 0) * 10000).toLocaleString() }}</span>
          </div>
          <a-alert
            v-if="customer && customer.budget_max && totalPrice > customer.budget_max * 10000"
            type="warning"
            :message="`报价已超出预算 ¥${(totalPrice - customer.budget_max * 10000).toLocaleString()}`"
            style="margin-top: 12px"
          />
        </a-card>

        <a-card title="已创建版本" style="margin-top: 16px">
          <a-timeline>
            <a-timeline-item v-for="q in existingQuotes" :key="q.id">
              <a-tag :color="q.is_confirmed ? 'green' : 'blue'">{{ q.version }}</a-tag>
              <span>¥{{ q.total_price.toLocaleString() }}</span>
              <a-tag v-if="q.is_confirmed" color="green">已确认</a-tag>
            </a-timeline-item>
          </a-timeline>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import { PlusOutlined, CheckOutlined, DeleteOutlined, MenuOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const route = useRoute()
const router = useRouter()

const customerId = route.params.customerId
const quoteId = route.query.quote_id

const customer = ref(null)
const packages = ref([])
const serviceItems = ref([])
const existingQuotes = ref([])
const quoteType = ref('package')
const selectedPackageId = ref(null)
const selectedItems = ref([])
const saving = ref(false)
let dragItem = null
let itemIdCounter = 0

const itemColumns = [
  { title: '服务项目', key: 'name', width: 100 },
  { title: '规格描述', key: 'specification', width: 200 },
  { title: '数量', key: 'quantity', width: 100 },
  { title: '单价', key: 'unit_price', width: 120 },
  { title: '小计', key: 'subtotal', width: 120 },
  { title: '操作', key: 'action', width: 60 }
]

const itemsTotal = computed(() => {
  return selectedItems.value.reduce((sum, item) => sum + (item.subtotal || 0), 0)
})

const totalPrice = computed(() => {
  if (quoteType.value === 'package' && selectedPackageId.value) {
    const pkg = packages.value.find(p => p.id === selectedPackageId.value)
    return pkg?.total_price || 0
  }
  return itemsTotal.value
})

function getStatusColor(status) {
  const colors = { consulting: 'blue', signed: 'green', preparing: 'orange', completed: 'purple', lost: 'red' }
  return colors[status] || 'default'
}

function getStatusText(status) {
  const texts = { consulting: '咨询中', signed: '已签约', preparing: '筹备中', completed: '已完成', lost: '已流失' }
  return texts[status] || status
}

function selectPackage(id) {
  selectedPackageId.value = id
  const pkg = packages.value.find(p => p.id === id)
  if (pkg) {
    selectedItems.value = pkg.items.map(item => ({
      ...item,
      _id: ++itemIdCounter,
      name: item.service_item?.name,
      service_item: item.service_item,
      unit_price: item.price / item.quantity,
      subtotal: item.price
    }))
  }
}

function addItem(item) {
  selectedItems.value.push({
    _id: ++itemIdCounter,
    service_item_id: item.id,
    name: item.name,
    service_item: item,
    specification: '',
    quantity: 1,
    unit_price: item.base_price_max,
    subtotal: item.base_price_max
  })
  updateTotal()
}

function removeItem(index) {
  selectedItems.value.splice(index, 1)
  updateTotal()
}

function updateSubtotal(record) {
  record.subtotal = record.quantity * record.unit_price
  updateTotal()
}

function updateTotal() {}

function onDragStart(e, item) {
  dragItem = item
}

function onDrop(e, targetItem) {
  if (!dragItem || dragItem.id === targetItem.id) return
  const fromIndex = serviceItems.value.findIndex(i => i.id === dragItem.id)
  const toIndex = serviceItems.value.findIndex(i => i.id === targetItem.id)
  serviceItems.value.splice(fromIndex, 1)
  serviceItems.value.splice(toIndex, 0, dragItem)
  dragItem = null
}

async function saveQuote() {
  if (quoteType.value === 'package' && !selectedPackageId.value) {
    message.error('请选择套餐')
    return
  }
  if (quoteType.value === 'custom' && selectedItems.value.length === 0) {
    message.error('请添加至少一个服务项')
    return
  }

  saving.value = true
  try {
    const items = selectedItems.value.map(item => ({
      service_item_id: item.service_item_id,
      specification: item.specification,
      quantity: item.quantity,
      unit_price: item.unit_price,
      subtotal: item.subtotal
    }))

    await api.quotes.create({
      customer_id: parseInt(customerId),
      package_id: quoteType.value === 'package' ? selectedPackageId.value : null,
      is_custom: quoteType.value === 'custom',
      items
    })

    message.success('方案保存成功')
    router.push('/quotes')
  } finally {
    saving.value = false
  }
}

async function loadData() {
  const [customerRes, packagesRes, serviceRes, quotesRes] = await Promise.all([
    api.customers.get(customerId),
    api.packages.list(),
    api.serviceItems.list(),
    api.quotes.list({ customer_id: customerId })
  ])
  
  customer.value = customerRes.data
  packages.value = packagesRes.data || []
  serviceItems.value = serviceRes.data || []
  existingQuotes.value = quotesRes.data || []

  if (quoteId) {
    const quoteRes = await api.quotes.get(quoteId)
    const quote = quoteRes.data
    if (quote) {
      quoteType.value = quote.is_custom ? 'custom' : 'package'
      if (!quote.is_custom && quote.package_id) {
        selectPackage(quote.package_id)
      } else if (quote.items) {
        selectedItems.value = quote.items.map(item => ({
          ...item,
          _id: ++itemIdCounter,
          name: item.service_item?.name
        }))
      }
    }
  }
}

onMounted(loadData)
</script>

<style scoped>
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}
.header-actions {
  display: flex;
  align-items: center;
}
.pkg-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.pkg-name {
  font-weight: 600;
  font-size: 16px;
}
.pkg-price {
  color: #f5222d;
  font-size: 20px;
  font-weight: bold;
  margin: 8px 0;
}
.pkg-desc {
  color: #666;
  font-size: 12px;
}
.selected {
  border: 2px solid #1890ff !important;
  background: #e6f7ff;
}
.check-icon {
  color: #1890ff;
  font-size: 20px;
}
.service-list {
  max-height: 400px;
  overflow-y: auto;
}
.service-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  margin-bottom: 8px;
  border: 1px solid #f0f0f0;
  border-radius: 4px;
  cursor: move;
  background: white;
  transition: all 0.3s;
}
.service-item:hover {
  border-color: #1890ff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.15);
}
.item-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.drag-handle {
  color: #ccc;
  cursor: grab;
}
.item-name {
  font-weight: 600;
  color: #333;
}
.item-range {
  font-size: 12px;
  color: #999;
}
.price-card {
  position: sticky;
  top: 16px;
}
.price-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}
.price-row.total {
  font-size: 18px;
  font-weight: 600;
}
.total-price {
  color: #f5222d;
  font-size: 28px;
  font-weight: bold;
}
</style>
