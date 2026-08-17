<template>
  <div>
    <div style="margin-bottom: 16px">
      <a-button type="primary" @click="showModal = true" v-if="userStore.isAdmin">
        <PlusOutlined /> 新增套餐
      </a-button>
    </div>

    <a-row :gutter="16">
      <a-col :span="8" v-for="pkg in packages" :key="pkg.id">
        <a-card hoverable>
          <template #actions>
            <a-button type="link" @click="editPackage(pkg)" v-if="userStore.isAdmin">编辑</a-button>
            <a-popconfirm title="确定删除?" ok-text="确定" cancel-text="取消" @confirm="deletePackage(pkg.id)" v-if="userStore.isAdmin">
              <a-button type="link" danger>删除</a-button>
            </a-popconfirm>
          </template>
          <a-card-meta>
            <template #title>
              <div class="package-title">
                <span>{{ pkg.name }}</span>
                <a-tag color="gold">{{ pkg.is_active ? '启用' : '禁用' }}</a-tag>
              </div>
            </template>
            <template #description>{{ pkg.description }}</template>
          </a-card-meta>
          <div class="package-price">
            <span class="price-label">套餐价格:</span>
            <span class="price-value">¥{{ pkg.total_price.toLocaleString() }}</span>
          </div>
          <div class="package-validity" v-if="pkg.valid_from || pkg.valid_to">
            有效期: {{ formatDate(pkg.valid_from) }} 至 {{ formatDate(pkg.valid_to) }}
          </div>
          <a-divider />
          <div class="package-items">
            <div class="items-title">包含服务项:</div>
            <div v-for="item in pkg.items" :key="item.id" class="item-row">
              <span class="item-name">{{ item.service_item?.name }}</span>
              <span class="item-spec">{{ item.specification }}</span>
              <span class="item-price">¥{{ item.price.toLocaleString() }}</span>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <a-modal v-model:open="showModal" :title="editId ? '编辑套餐' : '新增套餐'" @ok="handleSubmit" :confirm-loading="submitting" width="800px">
      <a-form :model="form" layout="vertical">
        <a-form-item label="套餐名称" required>
          <a-input v-model:value="form.name" />
        </a-form-item>
        <a-form-item label="套餐描述">
          <a-textarea v-model:value="form.description" :rows="3" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="套餐总价" required>
              <a-input-number v-model:value="form.total_price" :min="0" :precision="2" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="有效期开始">
              <a-date-picker v-model:value="form.valid_from" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="有效期结束">
              <a-date-picker v-model:value="form.valid_to" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="启用状态">
          <a-switch v-model:checked="form.is_active" checked-children="启用" un-checked-children="禁用" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { api } from '../api'
import { useUserStore } from '../store/user'
import { PlusOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'

const userStore = useUserStore()

const packages = ref([])
const showModal = ref(false)
const editId = ref(null)
const submitting = ref(false)

const form = reactive({
  name: '',
  description: '',
  total_price: 0,
  valid_from: null,
  valid_to: null,
  is_active: true
})

function formatDate(date) {
  return date ? dayjs(date).format('YYYY-MM-DD') : '不限'
}

function resetForm() {
  Object.assign(form, {
    name: '',
    description: '',
    total_price: 0,
    valid_from: null,
    valid_to: null,
    is_active: true
  })
  editId.value = null
}

function editPackage(pkg) {
  editId.value = pkg.id
  Object.assign(form, {
    ...pkg,
    valid_from: pkg.valid_from ? dayjs(pkg.valid_from) : null,
    valid_to: pkg.valid_to ? dayjs(pkg.valid_to) : null
  })
  showModal.value = true
}

async function deletePackage(id) {
  await api.packages.delete(id)
  message.success('删除成功')
  loadData()
}

async function handleSubmit() {
  if (!form.name) {
    message.error('请填写套餐名称')
    return
  }
  
  submitting.value = true
  try {
    const data = {
      ...form,
      valid_from: form.valid_from ? form.valid_from.format('YYYY-MM-DD') : null,
      valid_to: form.valid_to ? form.valid_to.format('YYYY-MM-DD') : null
    }
    
    if (editId.value) {
      await api.packages.update(editId.value, data)
      message.success('更新成功')
    } else {
      await api.packages.create(data)
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
  const res = await api.packages.list()
  packages.value = res || []
}

onMounted(loadData)
</script>

<style scoped>
.package-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.package-price {
  margin-top: 12px;
  font-size: 14px;
}
.price-label {
  color: #666;
}
.price-value {
  color: #f5222d;
  font-size: 24px;
  font-weight: bold;
  margin-left: 8px;
}
.package-validity {
  margin-top: 8px;
  color: #999;
  font-size: 12px;
}
.items-title {
  font-weight: 600;
  margin-bottom: 8px;
}
.item-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px dashed #f0f0f0;
  font-size: 13px;
}
.item-name {
  color: #1890ff;
  font-weight: 500;
  min-width: 80px;
}
.item-spec {
  flex: 1;
  color: #666;
  padding: 0 12px;
}
.item-price {
  color: #fa8c16;
  font-weight: 500;
}
</style>
