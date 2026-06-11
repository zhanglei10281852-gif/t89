<template>
  <div>
    <div style="margin-bottom: 16px">
      <a-space>
        <a-select v-model:value="currentYear" style="width: 120px" @change="loadCalendar">
          <a-select-option v-for="y in yearOptions" :key="y" :value="y">{{ y }}年</a-select-option>
        </a-select>
        <a-date-picker v-model:value="checkDate" mode="date" placeholder="查询某天档期" @change="checkAvailability" style="width: 200px" />
        <a-button @click="loadCalendar">刷新日历</a-button>
      </a-space>
    </div>

    <a-row :gutter="16">
      <a-col :span="18">
        <a-card title="全年档期日历">
          <div class="calendar-legend">
            <a-space>
              <span class="legend-item"><span class="legend-color level-0"></span>空闲</span>
              <span class="legend-item"><span class="legend-color level-1"></span>1单</span>
              <span class="legend-item"><span class="legend-color level-2"></span>2单</span>
              <span class="legend-item"><span class="legend-color level-3"></span>3单</span>
              <span class="legend-item"><span class="legend-color level-4"></span>4单+</span>
              <span class="legend-item"><span class="lucky-star">★</span>好日子</span>
            </a-space>
          </div>
          <div v-for="(monthData, monthIndex) in calendarMonths" :key="monthIndex" class="month-block">
            <div class="month-title">{{ monthNames[monthIndex] }}</div>
            <div class="week-header">
              <span v-for="w in weekDays" :key="w">{{ w }}</span>
            </div>
            <div class="days-grid">
              <div
                v-for="day in monthData"
                :key="day.date"
                :class="['day-cell', `busy-${day.busy_level}`, { 'lucky-day': day.is_lucky, 'other-month': !day.in_month, 'today': day.is_today }]"
                @click="selectDate(day)"
              >
                <div class="day-number">{{ day.day }}</div>
                <span v-if="day.is_lucky" class="lucky-star" title="好日子">★</span>
                <div v-if="day.count > 0" class="day-count">{{ day.count }}单</div>
              </div>
            </div>
          </div>
        </a-card>
      </a-col>

      <a-col :span="6">
        <a-card v-if="selectedDate" :title="`${selectedDate.date} 档期查询`" class="sticky-card">
          <a-table :columns="availabilityColumns" :data-source="availabilityData" :pagination="false" size="small">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <a-tag :color="record.status === '空闲' ? 'green' : 'red'">{{ record.status }}</a-tag>
              </template>
            </template>
          </a-table>
          <a-alert v-if="availabilityData.filter(d => d.status === '已占').length > 0" type="info" show-icon style="margin-top: 12px">
            <template #message>冲突提示</template>
            <template #description>
              当天已有 {{ availabilityData.filter(d => d.status === '已占').length }} 位策划师档期被占用
            </template>
          </a-alert>
        </a-card>

        <a-card title="2026年好日子列表" style="margin-top: 16px">
          <div class="lucky-list">
            <div v-for="ld in luckyDays" :key="ld.id" class="lucky-item">
              <span class="lucky-star">★</span>
              <span>{{ formatDate(ld.date) }}</span>
              <span class="lucky-remark">{{ ld.remark }}</span>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from '../api'
import dayjs from 'dayjs'

const currentYear = ref(dayjs().year())
const checkDate = ref(null)
const selectedDate = ref(null)
const calendarData = ref([])
const availabilityData = ref([])
const luckyDays = ref([])

const yearOptions = computed(() => {
  const current = dayjs().year()
  return [current - 1, current, current + 1]
})

const monthNames = ['一月', '二月', '三月', '四月', '五月', '六月', '七月', '八月', '九月', '十月', '十一月', '十二月']
const weekDays = ['日', '一', '二', '三', '四', '五', '六']

const availabilityColumns = [
  { title: '策划师', dataIndex: 'planner_name' },
  { title: '擅长风格', dataIndex: 'style' },
  { title: '状态', key: 'status' }
]

const calendarMonths = computed(() => {
  const months = []
  for (let m = 0; m < 12; m++) {
    const monthDays = []
    const firstDay = dayjs(`${currentYear.value}-${String(m + 1).padStart(2, '0')}-01`)
    const startOffset = firstDay.day()
    const daysInMonth = firstDay.daysInMonth()
    
    for (let i = 0; i < startOffset; i++) {
      const prevDay = firstDay.subtract(startOffset - i, 'day')
      const dateStr = prevDay.format('YYYY-MM-DD')
      const data = calendarData.value.find(d => d.date === dateStr)
      monthDays.push({
        date: dateStr,
        day: prevDay.date(),
        in_month: false,
        is_lucky: data?.is_lucky || false,
        lucky_remark: data?.lucky_remark || '',
        count: data?.count || 0,
        busy_level: data?.busy_level || 0,
        is_today: dateStr === dayjs().format('YYYY-MM-DD')
      })
    }
    
    for (let d = 1; d <= daysInMonth; d++) {
      const dateStr = `${currentYear.value}-${String(m + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
      const data = calendarData.value.find(dd => dd.date === dateStr)
      monthDays.push({
        date: dateStr,
        day: d,
        in_month: true,
        is_lucky: data?.is_lucky || false,
        lucky_remark: data?.lucky_remark || '',
        count: data?.count || 0,
        busy_level: data?.busy_level || 0,
        is_today: dateStr === dayjs().format('YYYY-MM-DD')
      })
    }
    
    while (monthDays.length < 42) {
      const lastDate = dayjs(monthDays[monthDays.length - 1].date).add(1, 'day')
      const dateStr = lastDate.format('YYYY-MM-DD')
      const data = calendarData.value.find(dd => dd.date === dateStr)
      monthDays.push({
        date: dateStr,
        day: lastDate.date(),
        in_month: false,
        is_lucky: data?.is_lucky || false,
        lucky_remark: data?.lucky_remark || '',
        count: data?.count || 0,
        busy_level: data?.busy_level || 0,
        is_today: dateStr === dayjs().format('YYYY-MM-DD')
      })
    }
    
    months.push(monthDays)
  }
  return months
})

function formatDate(date) {
  return dayjs(date).format('YYYY-MM-DD')
}

async function loadCalendar() {
  const [calRes, luckyRes] = await Promise.all([
    api.schedules.calendar(currentYear.value),
    api.schedules.luckyDays()
  ])
  calendarData.value = calRes.data?.data || []
  luckyDays.value = luckyRes.data || []
}

async function checkAvailability(date) {
  if (!date) return
  const dateStr = date.format('YYYY-MM-DD')
  const res = await api.schedules.availability(dateStr)
  availabilityData.value = res.data || []
  selectedDate.value = { date: dateStr }
}

function selectDate(day) {
  checkDate.value = dayjs(day.date)
  checkAvailability(dayjs(day.date))
}

onMounted(() => {
  loadCalendar()
})
</script>

<style scoped>
.calendar-legend {
  margin-bottom: 16px;
  padding: 8px;
  background: #fafafa;
  border-radius: 4px;
}
.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}
.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 2px;
  display: inline-block;
}
.level-0 { background: #f0f0f0; }
.level-1 { background: #bae7ff; }
.level-2 { background: #69c0ff; }
.level-3 { background: #1890ff; }
.level-4 { background: #0050b3; }
.lucky-star {
  color: #faad14;
  font-weight: bold;
}
.month-block {
  margin-bottom: 24px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
}
.month-title {
  background: #1890ff;
  color: white;
  padding: 8px 12px;
  font-weight: 600;
}
.week-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  background: #fafafa;
  text-align: center;
  padding: 4px 0;
  font-weight: 600;
  font-size: 12px;
  color: #666;
}
.days-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
}
.day-cell {
  min-height: 70px;
  padding: 4px;
  border: 1px solid #f0f0f0;
  cursor: pointer;
  position: relative;
  transition: all 0.2s;
}
.day-cell:hover {
  box-shadow: inset 0 0 0 2px #1890ff;
}
.other-month {
  background: #fafafa;
  opacity: 0.5;
}
.today {
  box-shadow: inset 0 0 0 2px #faad14;
}
.busy-0 { background: #fff; }
.busy-1 { background: #e6f7ff; }
.busy-2 { background: #91d5ff; }
.busy-3 { background: #40a9ff; color: white; }
.busy-4 { background: #096dd9; color: white; }
.day-number {
  font-weight: 600;
}
.day-count {
  font-size: 11px;
  margin-top: 4px;
}
.lucky-day .day-number {
  color: #faad14;
}
.lucky-day .lucky-star {
  position: absolute;
  top: 2px;
  right: 4px;
  font-size: 14px;
}
.sticky-card {
  position: sticky;
  top: 16px;
}
.lucky-list {
  max-height: 300px;
  overflow-y: auto;
}
.lucky-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  border-bottom: 1px solid #f0f0f0;
  font-size: 13px;
}
.lucky-remark {
  color: #999;
  margin-left: auto;
  font-size: 12px;
}
</style>
