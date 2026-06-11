<template>
  <div>
    <a-row :gutter="16">
      <a-col :span="6">
        <a-card class="stat-card">
          <a-statistic title="今日咨询" :value="stats.todayConsult" :value-style="{ color: '#1890ff' }" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="stat-card">
          <a-statistic title="本月签约" :value="stats.monthSign" :value-style="{ color: '#52c41a' }" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="stat-card">
          <a-statistic title="本月营业额" :value="stats.monthRevenue" precision="2" suffix="万" :value-style="{ color: '#fa8c16' }" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card class="stat-card">
          <a-statistic title="转化率" :value="stats.conversionRate" precision="2" suffix="%" :value-style="{ color: '#722ed1' }" />
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="12">
        <a-card title="月度签约趋势">
          <v-chart class="chart" :option="monthlyChart" autoresize />
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="客户转化漏斗">
          <v-chart class="chart" :option="funnelChart" autoresize />
        </a-card>
      </a-col>
    </a-row>

    <a-card title="最近客户" style="margin-top: 16px">
      <a-table :columns="columns" :data-source="recentCustomers" :pagination="false">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="getStatusColor(record.status)">{{ getStatusText(record.status) }}</a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-button type="link" size="small" @click="goToQuote(record)">生成报价</a-button>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, FunnelChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import dayjs from 'dayjs'

use([CanvasRenderer, BarChart, FunnelChart, GridComponent, TooltipComponent, LegendComponent])

const router = useRouter()

const stats = ref({
  todayConsult: 0,
  monthSign: 0,
  monthRevenue: 0,
  conversionRate: 0
})

const recentCustomers = ref([])
const monthlyChart = ref({})
const funnelChart = ref({})

const columns = [
  { title: '新人', dataIndex: 'name', key: 'name' },
  { title: '电话', dataIndex: 'phone', key: 'phone' },
  { title: '预算', dataIndex: 'budget', key: 'budget' },
  { title: '状态', key: 'status' },
  { title: '操作', key: 'action' }
]

function getStatusColor(status) {
  const colors = { consulting: 'blue', signed: 'green', preparing: 'orange', completed: 'purple', lost: 'red' }
  return colors[status] || 'default'
}

function getStatusText(status) {
  const texts = { consulting: '咨询中', signed: '已签约', preparing: '筹备中', completed: '已完成', lost: '已流失' }
  return texts[status] || status
}

function goToQuote(customer) {
  router.push(`/quotes/editor/${customer.id}`)
}

async function loadData() {
  const year = dayjs().format('YYYY')
  const month = dayjs().format('YYYY-MM')
  
  const [monthlyRes, funnelRes, customersRes, conversionRes] = await Promise.all([
    api.stats.monthly(year),
    api.customers.funnel(),
    api.customers.list(),
    api.stats.conversion()
  ])

  stats.value.todayConsult = customersRes.data?.filter(c => 
    dayjs(c.created_at).format('YYYY-MM-DD') === dayjs().format('YYYY-MM-DD')
  ).length || 0
  
  stats.value.monthSign = monthlyRes.data?.find(d => d.month === month)?.count || 0
  stats.value.monthRevenue = ((monthlyRes.data?.find(d => d.month === month)?.amount || 0) / 10000)
  stats.value.conversionRate = conversionRes.conversion_rate || 0

  recentCustomers.value = customersRes.data?.slice(0, 5).map(c => ({
    ...c,
    name: c.groom_name + ' & ' + c.bride_name,
    budget: `${c.budget_min || 0}-${c.budget_max || 0}万`
  })) || []

  monthlyChart.value = {
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: monthlyRes.data?.map(d => d.month?.slice(5) + '月') || [] },
    yAxis: { type: 'value' },
    series: [{
      name: '签约数',
      type: 'bar',
      data: monthlyRes.data?.map(d => d.count) || [],
      itemStyle: { color: '#1890ff' }
    }]
  }

  funnelChart.value = {
    tooltip: { trigger: 'item', formatter: '{b}: {c}' },
    series: [{
      name: '转化漏斗',
      type: 'funnel',
      left: '10%',
      width: '80%',
      label: { position: 'inside', formatter: '{b}\n{c}人' },
      data: funnelRes.data?.map(d => ({ value: d.count, name: d.label })) || []
    }]
  }
}

onMounted(loadData)
</script>

<style scoped>
.stat-card {
  text-align: center;
}
.chart {
  height: 300px;
}
</style>
