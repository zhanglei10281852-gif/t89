<template>
  <div>
    <a-card class="welcome-card">
      <div class="welcome-content">
        <a-avatar :size="64">{{ userStore.userInfo?.name?.charAt(0) }}</a-avatar>
        <div class="welcome-text">
          <h2>欢迎回来，{{ userStore.userInfo?.name }}</h2>
          <p>您的擅长风格: <a-tag color="blue">{{ userStore.userInfo?.style }}</a-tag></p>
        </div>
      </div>
    </a-card>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="8">
        <a-card>
          <a-statistic title="本月档期" :value="monthlyScheduleCount" :value-style="{ color: '#1890ff' }" />
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card>
          <a-statistic title="我的待办客户" :value="todoCustomers.length" :value-style="{ color: '#fa8c16' }" />
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card>
          <a-statistic title="年度负荷率" :value="annualLoadRate" precision="1" suffix="%" :value-style="{ color: '#52c41a' }" />
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="12">
        <a-card title="我的待办客户">
          <a-table :columns="todoColumns" :data-source="todoCustomers" :pagination="false" size="small">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                {{ record.groom_name }} & {{ record.bride_name }}
              </template>
              <template v-else-if="column.key === 'expected_date'">
                {{ formatDate(record.expected_date) }}
              </template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="getStatusColor(record.status)">{{ getStatusText(record.status) }}</a-tag>
              </template>
              <template v-else-if="column.key === 'action'">
                <a-button type="link" size="small" @click="viewCustomer(record)">查看详情</a-button>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="本月日程">
          <div class="calendar-nav">
            <a-button-group>
              <a-button @click="prevMonth"><LeftOutlined /></a-button>
              <a-button>{{ currentMonth.format('YYYY年MM月') }}</a-button>
              <a-button @click="nextMonth"><RightOutlined /></a-button>
            </a-button-group>
          </div>
          <div class="mini-calendar">
            <div class="week-row">
              <div v-for="w in weekDays" :key="w" class="week-cell">{{ w }}</div>
            </div>
            <div class="days-grid">
              <div
                v-for="day in calendarDays"
                :key="day.date"
                :class="['day-cell', { 'has-schedule': day.hasSchedule, 'other-month': !day.inMonth, 'today': day.isToday }]"
              >
                <span class="day-num">{{ day.day }}</span>
                <div v-if="day.hasSchedule" class="schedule-dot"></div>
              </div>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <a-card title="我的档期列表" style="margin-top: 16px">
      <a-table :columns="scheduleColumns" :data-source="mySchedules" :loading="loading">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'wedding_date'">
            {{ formatDate(record.wedding_date) }}
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { useUserStore } from '../store/user'
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const currentMonth = ref(dayjs())
const mySchedules = ref([])
const allContracts = ref([])
const allCustomers = ref([])

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

const todoColumns = [
  { title: '新人', key: 'name' },
  { title: '电话', dataIndex: 'phone' },
  { title: '预计日期', key: 'expected_date', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '操作', key: 'action', width: 100 }
]

const scheduleColumns = [
  { title: '客户', dataIndex: 'customer_name' },
  { title: '服务类型', dataIndex: 'service_type' },
  { title: '婚礼日期', key: 'wedding_date', width: 150 },
  { title: '创建时间', dataIndex: 'created_at', width: 170 }
]

const todoCustomers = computed(() => {
  return allCustomers.value?.filter(c => 
    c.status === 'preparing' || c.status === 'signed'
  ).slice(0, 10) || []
})

const monthlyScheduleCount = computed(() => {
  const monthStr = currentMonth.value.format('YYYY-MM')
  return mySchedules.value?.filter(s => 
    dayjs(s.wedding_date).format('YYYY-MM') === monthStr
  ).length || 0
})

const annualLoadRate = computed(() => {
  const year = currentMonth.value.year()
  const yearStr = String(year)
  const yearSchedules = mySchedules.value?.filter(s => 
    dayjs(s.wedding_date).format('YYYY') === yearStr
  ).length || 0
  const maxPossible = 52
  return Math.min((yearSchedules / maxPossible) * 100, 100)
})

const calendarDays = computed(() => {
  const days = []
  const firstDay = currentMonth.value.startOf('month')
  const startOffset = firstDay.day()
  const daysInMonth = currentMonth.value.daysInMonth()
  
  const scheduleDates = new Set(
    mySchedules.value?.map(s => dayjs(s.wedding_date).format('YYYY-MM-DD')) || []
  )
  
  for (let i = 0; i < startOffset; i++) {
    const d = firstDay.subtract(startOffset - i, 'day')
    days.push({
      date: d.format('YYYY-MM-DD'),
      day: d.date(),
      inMonth: false,
      hasSchedule: scheduleDates.has(d.format('YYYY-MM-DD')),
      isToday: d.format('YYYY-MM-DD') === dayjs().format('YYYY-MM-DD')
    })
  }
  
  for (let i = 1; i <= daysInMonth; i++) {
    const d = currentMonth.value.date(i)
    days.push({
      date: d.format('YYYY-MM-DD'),
      day: i,
      inMonth: true,
      hasSchedule: scheduleDates.has(d.format('YYYY-MM-DD')),
      isToday: d.format('YYYY-MM-DD') === dayjs().format('YYYY-MM-DD')
    })
  }
  
  while (days.length < 42) {
    const last = days[days.length - 1]
    const d = dayjs(last.date).add(1, 'day')
    days.push({
      date: d.format('YYYY-MM-DD'),
      day: d.date(),
      inMonth: false,
      hasSchedule: scheduleDates.has(d.format('YYYY-MM-DD')),
      isToday: d.format('YYYY-MM-DD') === dayjs().format('YYYY-MM-DD')
    })
  }
  
  return days
})

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

function prevMonth() {
  currentMonth.value = currentMonth.value.subtract(1, 'month')
}

function nextMonth() {
  currentMonth.value = currentMonth.value.add(1, 'month')
}

function viewCustomer(customer) {
  router.push(`/quotes/editor/${customer.id}`)
}

async function loadData() {
  loading.value = true
  try {
    const monthStr = currentMonth.value.format('YYYY-MM')
    const [schedulesRes, contractsRes, customersRes] = await Promise.all([
      api.schedules.mySchedule(monthStr),
      api.contracts.list(),
      api.customers.list()
    ])
    mySchedules.value = schedulesRes || []
    allContracts.value = contractsRes || []
    allCustomers.value = customersRes || []
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.welcome-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  margin-bottom: 16px;
}
.welcome-card :deep(.ant-card-body) {
  padding: 24px;
}
.welcome-content {
  display: flex;
  align-items: center;
  gap: 16px;
}
.welcome-text h2 {
  margin: 0 0 8px 0;
  color: white;
}
.welcome-text p {
  margin: 0;
  color: rgba(255, 255, 255, 0.8);
}
.calendar-nav {
  text-align: center;
  margin-bottom: 12px;
}
.mini-calendar {
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
}
.week-row {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  background: #fafafa;
  text-align: center;
  padding: 8px 0;
  font-size: 12px;
  font-weight: 600;
  color: #666;
}
.days-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
}
.day-cell {
  height: 50px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1px solid #f0f0f0;
  font-size: 14px;
  position: relative;
}
.other-month {
  color: #d9d9d9;
}
.today {
  background: #e6f7ff;
  font-weight: bold;
}
.day-num {
  font-weight: 500;
}
.schedule-dot {
  width: 6px;
  height: 6px;
  background: #f5222d;
  border-radius: 50%;
  margin-top: 2px;
}
.has-schedule {
  background: #fff2e8;
}
</style>
