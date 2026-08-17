<template>
  <div>
    <div style="margin-bottom: 16px">
      <a-space>
        <a-select v-model:value="filterStatus" placeholder="状态筛选" style="width: 150px" allow-clear>
          <a-select-option value="consulting">咨询中</a-select-option>
          <a-select-option value="signed">已签约</a-select-option>
          <a-select-option value="preparing">筹备中</a-select-option>
          <a-select-option value="completed">已完成</a-select-option>
          <a-select-option value="lost">已流失</a-select-option>
        </a-select>
        <a-input-search v-model:value="searchText" placeholder="搜索新人姓名/电话" style="width: 250px" @search="loadData" />
        <a-button type="primary" @click="showModal = true">
          <PlusOutlined /> 新增客户
        </a-button>
      </a-space>
    </div>

    <a-table :columns="columns" :data-source="customers" :loading="loading" row-key="id">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <div>{{ record.groom_name }} & {{ record.bride_name }}</div>
        </template>
        <template v-else-if="column.key === 'expected_date'">
          {{ formatDate(record.expected_date) }}
        </template>
        <template v-else-if="column.key === 'budget'">
          {{ record.budget_min || 0 }}-{{ record.budget_max || 0 }}万
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record.status)">{{ getStatusText(record.status) }}</a-tag>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="editCustomer(record)">编辑</a-button>
            <a-button type="link" size="small" @click="goToQuote(record)">报价</a-button>
            <a-dropdown>
              <template #overlay>
                <a-menu @click="({ key }) => changeStatus(record, key)">
                  <a-menu-item key="consulting">咨询中</a-menu-item>
                  <a-menu-item key="signed">已签约</a-menu-item>
                  <a-menu-item key="preparing">筹备中</a-menu-item>
                  <a-menu-item key="completed">已完成</a-menu-item>
                  <a-menu-item key="lost">已流失</a-menu-item>
                </a-menu>
              </template>
              <a-button type="link" size="small">状态 <DownOutlined /></a-button>
            </a-dropdown>
            <a-popconfirm title="确定删除?" ok-text="确定" cancel-text="取消" @confirm="deleteCustomer(record.id)">
              <a-button type="link" size="small" danger>删除</a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal v-model:open="showModal" :title="editId ? '编辑客户' : '新增客户'" @ok="handleSubmit" :confirm-loading="submitting">
      <a-form :model="form" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="新郎姓名" required>
              <a-input v-model:value="form.groom_name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="新娘姓名" required>
              <a-input v-model:value="form.bride_name" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="联系电话" required>
              <a-input v-model:value="form.phone" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="婚礼预计日期">
              <a-date-picker v-model:value="form.expected_date" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="预算最低(万)">
              <a-input-number v-model:value="form.budget_min" :min="0" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="预算最高(万)">
              <a-input-number v-model:value="form.budget_max" :min="0" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="桌数预估">
              <a-input-number v-model:value="form.table_count" :min="0" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="婚礼风格偏好">
              <a-select v-model:value="form.style_preference" style="width: 100%">
                <a-select-option value="中式">中式</a-select-option>
                <a-select-option value="西式">西式</a-select-option>
                <a-select-option value="户外">户外</a-select-option>
                <a-select-option value="极简">极简</a-select-option>
                <a-select-option value="奢华">奢华</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="客户来源">
              <a-select v-model:value="form.source" style="width: 100%">
                <a-select-option value="朋友推荐">朋友推荐</a-select-option>
                <a-select-option value="网络">网络</a-select-option>
                <a-select-option value="到店">到店</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="婚宴酒店名称">
          <a-input v-model:value="form.hotel_name" />
        </a-form-item>
        <a-form-item label="备注">
          <a-textarea v-model:value="form.remark" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { PlusOutlined, DownOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'

const router = useRouter()

const customers = ref([])
const loading = ref(false)
const submitting = ref(false)
const showModal = ref(false)
const editId = ref(null)
const filterStatus = ref(null)
const searchText = ref('')

const form = reactive({
  groom_name: '',
  bride_name: '',
  phone: '',
  expected_date: null,
  budget_min: null,
  budget_max: null,
  style_preference: '',
  source: '',
  hotel_name: '',
  table_count: null,
  remark: ''
})

const columns = [
  { title: '新人', key: 'name' },
  { title: '电话', dataIndex: 'phone' },
  { title: '预计日期', key: 'expected_date' },
  { title: '预算', key: 'budget' },
  { title: '风格', dataIndex: 'style_preference' },
  { title: '来源', dataIndex: 'source' },
  { title: '状态', key: 'status' },
  { title: '操作', key: 'action', width: 250 }
]

function getStatusColor(status) {
  const colors = { consulting: 'blue', signed: 'green', preparing: 'orange', completed: 'purple', lost: 'red' }
  return colors[status] || 'default'
}

function getStatusText(status) {
  const texts = { consulting: '咨询中', signed: '已签约', preparing: '筹备中', completed: '已完成', lost: '已流失' }
  return texts[status] || status
}

function formatDate(date) {
  return date ? dayjs(date).format('YYYY-MM-DD') : '-'
}

function resetForm() {
  Object.assign(form, {
    groom_name: '',
    bride_name: '',
    phone: '',
    expected_date: null,
    budget_min: null,
    budget_max: null,
    style_preference: '',
    source: '',
    hotel_name: '',
    table_count: null,
    remark: ''
  })
  editId.value = null
}

function editCustomer(record) {
  editId.value = record.id
  Object.assign(form, {
    ...record,
    expected_date: record.expected_date ? dayjs(record.expected_date) : null
  })
  showModal.value = true
}

function goToQuote(record) {
  router.push(`/quotes/editor/${record.id}`)
}

async function changeStatus(record, status) {
  await api.customers.updateStatus(record.id, status)
  message.success('状态已更新')
  loadData()
}

async function deleteCustomer(id) {
  await api.customers.delete(id)
  message.success('删除成功')
  loadData()
}

async function handleSubmit() {
  if (!form.groom_name || !form.bride_name || !form.phone) {
    message.error('请填写必填项')
    return
  }
  
  submitting.value = true
  try {
    const data = {
      ...form,
      expected_date: form.expected_date ? form.expected_date.format('YYYY-MM-DD') : null
    }
    
    if (editId.value) {
      await api.customers.update(editId.value, data)
      message.success('更新成功')
    } else {
      await api.customers.create(data)
      message.success('创建成功')
    }
    showModal.value = false
    resetForm()
    loadData()
  } finally {
    submitting.value = false
  }
}

async function loadData() {
  loading.value = true
  try {
    const params = {}
    if (filterStatus.value) params.status = filterStatus.value
    const res = await api.customers.list(params)
    let list = res || []
    if (searchText.value) {
      const text = searchText.value.toLowerCase()
      list = list.filter(c => 
        c.groom_name.toLowerCase().includes(text) ||
        c.bride_name.toLowerCase().includes(text) ||
        c.phone.includes(text)
      )
    }
    customers.value = list
  } finally {
    loading.value = false
  }
}

watch(filterStatus, loadData)
onMounted(loadData)
</script>
