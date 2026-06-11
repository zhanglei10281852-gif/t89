<template>
  <div>
    <div style="margin-bottom: 16px">
      <a-button type="primary" @click="showModal = true">
        <PlusOutlined /> 新增签约
      </a-button>
    </div>

    <a-table :columns="columns" :data-source="contracts" :loading="loading" row-key="id">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'customer'">
          {{ record.customer?.groom_name }} & {{ record.customer?.bride_name }}
        </template>
        <template v-else-if="column.key === 'planner'">
          {{ record.planner?.name }}
        </template>
        <template v-else-if="column.key === 'wedding_date'">
          {{ formatDate(record.wedding_date) }}
        </template>
        <template v-else-if="column.key === 'sign_date'">
          {{ formatDate(record.sign_date) }}
        </template>
        <template v-else-if="column.key === 'total_amount'">
          ¥{{ record.total_amount.toLocaleString() }}
        </template>
        <template v-else-if="column.key === 'advance_payment'">
          ¥{{ record.advance_payment.toLocaleString() }}
        </template>
        <template v-else-if="column.key === 'final_payment_due'">
          {{ formatDate(record.final_payment_due) }}
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record.status)">{{ getStatusText(record.status) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'is_final_paid'">
          <a-tag :color="record.is_final_paid ? 'green' : 'orange'">
            {{ record.is_final_paid ? '已结清' : '待尾款' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="viewDetail(record)">详情</a-button>
            <a-dropdown v-if="!record.is_final_paid">
              <template #overlay>
                <a-menu @click="({ key }) => updateContract(record, key)">
                  <a-menu-item key="paid">标记尾款已付</a-menu-item>
                </a-menu>
              </template>
              <a-button type="link" size="small">更多操作</a-button>
            </a-dropdown>
            <a-dropdown>
              <template #overlay>
                <a-menu @click="({ key }) => updateContract(record, key)">
                  <a-menu-item key="preparing">筹备中</a-menu-item>
                  <a-menu-item key="completed">已完成</a-menu-item>
                </a-menu>
              </template>
              <a-button type="link" size="small">状态</a-button>
            </a-dropdown>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal v-model:open="showModal" title="新增签约" @ok="handleSubmit" :confirm-loading="submitting" width="800px">
      <a-form :model="form" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="选择客户" required>
              <a-select v-model:value="form.customer_id" style="width: 100%" @change="onCustomerChange">
                <a-select-option v-for="c in availableCustomers" :key="c.id" :value="c.id">
                  {{ c.groom_name }} & {{ c.bride_name }} (预算: {{c.budget_min}}-{{c.budget_max}}万)
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="选择报价方案" required>
              <a-select v-model:value="form.quote_id" style="width: 100%" @change="onQuoteChange">
                <a-select-option v-for="q in customerQuotes" :key="q.id" :value="q.id">
                  {{ q.version }} - ¥{{ q.total_price.toLocaleString() }} {{ q.is_confirmed ? '(已确认)' : '' }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="分配策划师" required>
              <a-select v-model:value="form.planner_id" style="width: 100%" @change="checkPlanner">
                <a-select-option v-for="p in planners" :key="p.id" :value="p.id">
                  {{ p.name }} ({{ p.style }})
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="婚礼日期" required>
              <a-date-picker v-model:value="form.wedding_date" style="width: 100%" @change="checkPlanner" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="合同总金额" required>
              <a-input-number v-model:value="form.total_amount" :min="0" :precision="2" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="预付款金额(≥30%)" required>
              <a-input-number v-model:value="form.advance_payment" :min="minAdvance" :precision="2" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="尾款支付截止">
              <a-input :value="form.final_payment_due" disabled style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-alert
          v-if="conflictInfo"
          type="warning"
          show-icon
          :message="`档期冲突: ${conflictInfo.conflicts?.map(c => c.staff_name).join(', ')} 当天已有安排`"
        >
          <template #description>
            推荐可替换: {{ conflictInfo.recommendations?.map(r => r.staff_name).join(', ') }}
          </template>
        </a-alert>
      </a-form>
    </a-modal>

    <a-modal v-model:open="detailVisible" title="签约详情" width="800px" :footer="null">
      <div v-if="currentContract">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="客户">{{ currentContract.customer?.groom_name }} & {{ currentContract.customer?.bride_name }}</a-descriptions-item>
          <a-descriptions-item label="联系电话">{{ currentContract.customer?.phone }}</a-descriptions-item>
          <a-descriptions-item label="策划师">{{ currentContract.planner?.name }} ({{ currentContract.planner?.style }})</a-descriptions-item>
          <a-descriptions-item label="婚礼日期">{{ formatDate(currentContract.wedding_date) }}</a-descriptions-item>
          <a-descriptions-item label="签约日期">{{ formatDate(currentContract.sign_date) }}</a-descriptions-item>
          <a-descriptions-item label="总金额">¥{{ currentContract.total_amount.toLocaleString() }}</a-descriptions-item>
          <a-descriptions-item label="预付款">¥{{ currentContract.advance_payment.toLocaleString() }}</a-descriptions-item>
          <a-descriptions-item label="尾款截止">{{ formatDate(currentContract.final_payment_due) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getStatusColor(currentContract.status)">{{ getStatusText(currentContract.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="尾款状态">
            <a-tag :color="currentContract.is_final_paid ? 'green' : 'orange'">
              {{ currentContract.is_final_paid ? '已结清' : '待支付' }}
            </a-tag>
          </a-descriptions-item>
        </a-descriptions>
        <h4 style="margin-top: 16px">报价明细</h4>
        <a-table :columns="quoteColumns" :data-source="currentContract.quote?.items || []" :pagination="false" size="small">
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'subtotal'">
              ¥{{ record.subtotal.toLocaleString() }}
            </template>
          </template>
        </a-table>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { api } from '../api'
import { PlusOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'

const contracts = ref([])
const customers = ref([])
const planners = ref([])
const customerQuotes = ref([])
const loading = ref(false)
const submitting = ref(false)
const showModal = ref(false)
const detailVisible = ref(false)
const currentContract = ref(null)
const conflictInfo = ref(null)

const form = reactive({
  customer_id: null,
  quote_id: null,
  planner_id: null,
  wedding_date: null,
  total_amount: 0,
  advance_payment: 0,
  final_payment_due: ''
})

const columns = [
  { title: '客户', key: 'customer' },
  { title: '策划师', key: 'planner' },
  { title: '婚礼日期', key: 'wedding_date', width: 120 },
  { title: '签约日期', key: 'sign_date', width: 120 },
  { title: '总金额', key: 'total_amount', width: 120 },
  { title: '预付款', key: 'advance_payment', width: 120 },
  { title: '尾款截止', key: 'final_payment_due', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '付款', key: 'is_final_paid', width: 100 },
  { title: '操作', key: 'action', width: 200 }
]

const quoteColumns = [
  { title: '服务项目', dataIndex: 'service_item.name' },
  { title: '规格', dataIndex: 'specification' },
  { title: '数量', dataIndex: 'quantity', width: 60 },
  { title: '小计', key: 'subtotal', width: 100 }
]

const availableCustomers = computed(() => {
  return customers.value?.filter(c => c.status === 'consulting' || c.status === 'signed') || []
})

const minAdvance = computed(() => {
  return (form.total_amount * 0.3) || 0
})

function getStatusColor(status) {
  const colors = { preparing: 'orange', completed: 'green', cancelled: 'red' }
  return colors[status] || 'default'
}

function getStatusText(status) {
  const texts = { preparing: '筹备中', completed: '已完成', cancelled: '已取消' }
  return texts[status] || status
}

function formatDate(date) {
  return date ? dayjs(date).format('YYYY-MM-DD') : '-'
}

function onCustomerChange() {
  customerQuotes.value = []
  form.quote_id = null
  form.total_amount = 0
  form.advance_payment = 0
  if (form.customer_id) {
    api.quotes.list({ customer_id: form.customer_id }).then(res => {
      customerQuotes.value = res.data?.filter(q => q.is_confirmed) || []
    })
  }
}

function onQuoteChange() {
  const quote = customerQuotes.value.find(q => q.id === form.quote_id)
  if (quote) {
    form.total_amount = quote.total_price
    form.advance_payment = Math.ceil(quote.total_price * 0.3)
  }
}

async function checkPlanner() {
  conflictInfo.value = null
  if (form.planner_id && form.wedding_date) {
    try {
      await api.schedules.checkConflict({
        staff_ids: [form.planner_id],
        wedding_date: form.wedding_date.format('YYYY-MM-DD')
      })
    } catch (e) {
      if (e.response?.status === 409) {
        conflictInfo.value = e.response.data
      }
    }
  }
  updateFinalDueDate()
}

function updateFinalDueDate() {
  if (form.wedding_date) {
    form.final_payment_due = form.wedding_date.subtract(7, 'day').format('YYYY-MM-DD')
  }
}

watch(() => form.total_amount, () => {
  if (form.advance_payment < minAdvance.value) {
    form.advance_payment = Math.ceil(minAdvance.value)
  }
})

function viewDetail(record) {
  currentContract.value = record
  detailVisible.value = true
}

async function updateContract(record, action) {
  const data = {}
  if (action === 'paid') {
    data.is_final_paid = true
  } else {
    data.status = action
  }
  await api.contracts.update(record.id, data)
  message.success('更新成功')
  loadData()
}

async function handleSubmit() {
  if (!form.customer_id || !form.quote_id || !form.planner_id || !form.wedding_date) {
    message.error('请填写完整信息')
    return
  }
  if (form.advance_payment < minAdvance.value) {
    message.error(`预付款不能低于30% (¥${minAdvance.value.toLocaleString()})`)
    return
  }
  if (conflictInfo.value) {
    message.error('所选策划师当天档期冲突，请重新选择')
    return
  }

  submitting.value = true
  try {
    await api.contracts.create({
      customer_id: form.customer_id,
      quote_id: form.quote_id,
      planner_id: form.planner_id,
      total_amount: form.total_amount,
      advance_payment: form.advance_payment,
      wedding_date: form.wedding_date.format('YYYY-MM-DD')
    })
    message.success('签约成功')
    showModal.value = false
    Object.assign(form, {
      customer_id: null, quote_id: null, planner_id: null, wedding_date: null,
      total_amount: 0, advance_payment: 0, final_payment_due: ''
    })
    conflictInfo.value = null
    loadData()
  } finally {
    submitting.value = false
  }
}

async function loadData() {
  loading.value = true
  try {
    const [contractsRes, customersRes, plannersRes] = await Promise.all([
      api.contracts.list(),
      api.customers.list(),
      api.users.planners()
    ])
    contracts.value = contractsRes.data || []
    customers.value = customersRes.data || []
    planners.value = plannersRes.data || []
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>
