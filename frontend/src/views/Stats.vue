<template>
  <div>
    <div style="margin-bottom: 16px">
      <a-select v-model:value="currentYear" style="width: 120px" @change="loadAllStats">
        <a-select-option v-for="y in yearOptions" :key="y" :value="y">{{ y }}年</a-select-option>
      </a-select>
    </div>

    <a-row :gutter="16">
      <a-col :span="6">
        <a-card>
          <a-statistic title="总签约数" :value="summary.totalContracts" :value-style="{ color: '#1890ff' }" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="总营业额" :value="summary.totalRevenue" precision="2" suffix="万" :value-style="{ color: '#52c41a' }" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="客户转化率" :value="summary.conversionRate" precision="2" suffix="%" :value-style="{ color: '#fa8c16' }" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="好日子需求比" :value="summary.luckyRatio" precision="2" suffix="倍" :value-style="{ color: '#722ed1' }" />
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="12">
        <a-card title="月度签约趋势">
          <v-chart class="chart" :option="monthlyBarOption" autoresize />
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="月度营业额趋势">
          <v-chart class="chart" :option="monthlyLineOption" autoresize />
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="12">
        <a-card title="策划师档期负荷率">
          <v-chart class="chart" :option="plannerLoadOption" autoresize />
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="好日子vs普通日签约对比">
          <v-chart class="chart" :option="luckyCompareOption" autoresize />
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="12">
        <a-card title="套餐销量排行">
          <a-table :columns="packageRankColumns" :data-source="packageRank" :pagination="false" size="small">
            <template #bodyCell="{ column, record, index }">
              <template v-if="column.key === 'rank'">
                <a-tag :color="index === 0 ? 'gold' : index === 1 ? 'silver' : index === 2 ? 'orange' : 'default'">
                  {{ index + 1 }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'total_amount'">
                ¥{{ record.total_amount?.toLocaleString() || 0 }}
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="服务项销量排行">
          <a-table :columns="serviceRankColumns" :data-source="serviceRank" :pagination="false" size="small">
            <template #bodyCell="{ column, record, index }">
              <template v-if="column.key === 'rank'">
                <a-tag :color="index === 0 ? 'gold' : index === 1 ? 'silver' : index === 2 ? 'orange' : 'default'">
                  {{ index + 1 }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'total_amount'">
                ¥{{ record.total_amount?.toLocaleString() || 0 }}
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from '../api'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import dayjs from 'dayjs'

use([CanvasRenderer, BarChart, LineChart, PieChart, GridComponent, TooltipComponent, LegendComponent])

const currentYear = ref(dayjs().year())
const monthlyData = ref([])
const plannerLoad = ref([])
const packageRank = ref([])
const serviceRank = ref([])
const luckyComparison = ref(null)
const conversion = ref(null)

const yearOptions = computed(() => {
  const current = dayjs().year()
  return [current - 2, current - 1, current, current + 1]
})

const summary = computed(() => ({
  totalContracts: monthlyData.value?.reduce((sum, d) => sum + (d.count || 0), 0) || 0,
  totalRevenue: ((monthlyData.value?.reduce((sum, d) => sum + (d.amount || 0), 0) || 0) / 10000),
  conversionRate: conversion.value?.conversion_rate || 0,
  luckyRatio: luckyComparison.value?.demand_ratio || 0
}))

const packageRankColumns = [
  { title: '排名', key: 'rank', width: 60 },
  { title: '套餐名称', dataIndex: 'package_name' },
  { title: '销量', dataIndex: 'sales_count', width: 80 },
  { title: '总收入', key: 'total_amount', width: 120 }
]

const serviceRankColumns = [
  { title: '排名', key: 'rank', width: 60 },
  { title: '服务项目', dataIndex: 'item_name' },
  { title: '销量', dataIndex: 'quantity', width: 80 },
  { title: '总收入', key: 'total_amount', width: 120 }
]

const monthlyBarOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: monthlyData.value?.map(d => d.month?.slice(5) + '月') || [] },
  yAxis: { type: 'value', name: '签约数' },
  series: [{
    name: '签约数',
    type: 'bar',
    data: monthlyData.value?.map(d => d.count) || [],
    itemStyle: { color: '#1890ff' },
    label: { show: true, position: 'top' }
  }]
}))

const monthlyLineOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: monthlyData.value?.map(d => d.month?.slice(5) + '月') || [] },
  yAxis: { type: 'value', name: '万元' },
  series: [{
    name: '营业额',
    type: 'line',
    smooth: true,
    data: monthlyData.value?.map(d => ((d.amount || 0) / 10000).toFixed(2)) || [],
    itemStyle: { color: '#52c41a' },
    areaStyle: { opacity: 0.3 },
    label: { show: true, formatter: '{c}万' }
  }]
}))

const plannerLoadOption = computed(() => ({
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  xAxis: { type: 'value', name: '负荷率(%)' },
  yAxis: { type: 'category', data: plannerLoad.value?.map(d => d.planner_name) || [] },
  series: [{
    name: '负荷率',
    type: 'bar',
    data: plannerLoad.value?.map(d => ({
      value: d.load_rate?.toFixed(1),
      itemStyle: {
        color: d.load_rate > 80 ? '#f5222d' : d.load_rate > 50 ? '#fa8c16' : '#52c41a'
      }
    })) || [],
    label: { show: true, position: 'right', formatter: '{c}%' }
  }]
}))

const luckyCompareOption = computed(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0 },
  series: [{
    name: '好日子vs普通日',
    type: 'pie',
    radius: ['40%', '70%'],
    data: [
      { 
        name: '好日子签约', 
        value: luckyComparison.value?.lucky_days?.total_bookings || 0,
        itemStyle: { color: '#faad14' }
      },
      { 
        name: '普通日签约', 
        value: luckyComparison.value?.normal_days?.total_bookings || 0,
        itemStyle: { color: '#1890ff' }
      }
    ],
    label: {
      formatter: '{b}: {c}单 ({d}%)'
    }
  }]
}))

async function loadAllStats() {
  const [monthlyRes, conversionRes, plannerRes, pkgRankRes, svcRankRes, luckyRes] = await Promise.all([
    api.stats.monthly(currentYear.value),
    api.stats.conversion(),
    api.stats.plannerLoad(currentYear.value),
    api.stats.packageRank(),
    api.stats.serviceRank(),
    api.stats.luckyComparison(currentYear.value)
  ])
  
  monthlyData.value = monthlyRes.data?.data || []
  conversion.value = conversionRes
  plannerLoad.value = plannerRes.data?.data || []
  packageRank.value = pkgRankRes.data || []
  serviceRank.value = svcRankRes.data || []
  luckyComparison.value = luckyRes
}

onMounted(loadAllStats)
</script>

<style scoped>
.chart {
  height: 350px;
}
</style>
