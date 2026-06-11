<template>
  <div>
    <div style="margin-bottom: 16px">
      <a-select v-model:value="filterCustomer" placeholder="选择客户" style="width: 200px" allow-clear @change="loadData">
        <a-select-option v-for="c in customers" :key="c.id" :value="c.id">
          {{ c.groom_name }} & {{ c.bride_name }}
        </a-select-option>
      </a-select>
    </div>

    <a-table :columns="columns" :data-source="quotes" :loading="loading" row-key="id">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'customer'">
          {{ record.customer?.groom_name }} & {{ record.customer?.bride_name }}
        </template>
        <template v-else-if="column.key === 'package'">
          {{ record.package?.name || (record.is_custom ? '自定义组合' : '-') }}
        </template>
        <template v-else-if="column.key === 'total_price'">
          ¥{{ record.total_price.toLocaleString() }}
        </template>
        <template v-else-if="column.key === 'is_confirmed'">
          <a-tag :color="record.is_confirmed ? 'green' : 'orange'">
            {{ record.is_confirmed ? '已确认' : '待确认' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'created_at'">
          {{ formatDate(record.created_at) }}
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="viewDetail(record)">查看</a-button>
            <a-button type="link" size="small" @click="newVersion(record)" v-if="!record.is_confirmed">新建版本</a-button>
            <a-button type="link" size="small" @click="confirmQuote(record)" v-if="!record.is_confirmed">确认方案</a-button>
            <a-button type="link" size="small" @click="goToEditor(record)">编辑</a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal v-model:open="detailVisible" title="报价方案详情" width="700px" :footer="null">
      <div v-if="currentQuote">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="客户">{{ currentQuote.customer?.groom_name }} & {{ currentQuote.customer?.bride_name }}</a-descriptions-item>
          <a-descriptions-item label="版本">{{ currentQuote.version }}</a-descriptions-item>
          <a-descriptions-item label="套餐">{{ currentQuote.package?.name || (currentQuote.is_custom ? '自定义组合' : '-') }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="currentQuote.is_confirmed ? 'green' : 'orange'">
              {{ currentQuote.is_confirmed ? '已确认' : '待确认' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="创建时间" :span="2">{{ formatDateTime(currentQuote.created_at) }}</a-descriptions-item>
        </a-descriptions>
        <h4 style="margin-top: 16px">服务项明细</h4>
        <a-table :columns="itemColumns" :data-source="currentQuote.items" :pagination="false" size="small">
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'price'">
              ¥{{ record.unit_price.toLocaleString() }}
            </template>
            <template v-else-if="column.key === 'subtotal'">
              ¥{{ record.subtotal.toLocaleString() }}
            </template>
          </template>
        </a-table>
        <div style="text-align: right; margin-top: 16px; font-size: 18px">
          总计: <span style="color: #f5222d; font-weight: bold">¥{{ currentQuote.total_price.toLocaleString() }}</span>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { message, Modal } from 'ant-design-vue'
import dayjs from 'dayjs'

const router = useRouter()

const quotes = ref([])
const customers = ref([])
const loading = ref(false)
const filterCustomer = ref(null)
const detailVisible = ref(false)
const currentQuote = ref(null)

const columns = [
  { title: '客户', key: 'customer' },
  { title: '版本', dataIndex: 'version', width: 80 },
  { title: '套餐/类型', key: 'package' },
  { title: '总价', key: 'total_price' },
  { title: '状态', key: 'is_confirmed', width: 100 },
  { title: '创建时间', key: 'created_at', width: 170 },
  { title: '操作', key: 'action', width: 250 }
]

const itemColumns = [
  { title: '服务项目', dataIndex: 'service_item.name' },
  { title: '规格描述', dataIndex: 'specification' },
  { title: '数量', dataIndex: 'quantity', width: 80 },
  { title: '单价', key: 'price', width: 120 },
  { title: '小计', key: 'subtotal', width: 120 }
]

function formatDate(date) {
  return date ? dayjs(date).format('YYYY-MM-DD') : '-'
}

function formatDateTime(date) {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

function goToEditor(record) {
  router.push(`/quotes/editor/${record.customer_id}?quote_id=${record.id}`)
}

function viewDetail(record) {
  currentQuote.value = record
  detailVisible.value = true
}

async function newVersion(record) {
  try {
    await api.quotes.createVersion(record.id)
    message.success('已创建新版本')
    loadData()
  } catch (e) {}
}

async function confirmQuote(record) {
  Modal.confirm({
    title: '确认报价方案',
    content: `确认将 ${record.version} 版本设为最终方案？`,
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      await api.quotes.confirm(record.id)
      message.success('方案已确认')
      loadData()
    }
  })
}

async function loadData() {
  loading.value = true
  try {
    const params = filterCustomer.value ? { customer_id: filterCustomer.value } : {}
    const [quotesRes, customersRes] = await Promise.all([
      api.quotes.list(params),
      api.customers.list()
    ])
    quotes.value = quotesRes.data || []
    customers.value = customersRes.data || []
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>
