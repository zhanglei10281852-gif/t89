<template>
  <div>
    <a-row :gutter="16">
      <a-col :span="8">
        <a-card>
          <a-statistic title="总咨询客户" :value="totalCustomers" :value-style="{ color: '#1890ff' }" />
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card>
          <a-statistic title="已签约客户" :value="signedCustomers" :value-style="{ color: '#52c41a' }" />
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card>
          <a-statistic title="整体转化率" :value="conversionRate" precision="2" suffix="%" :value-style="{ color: '#fa8c16' }" />
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="12">
        <a-card title="转化漏斗图">
          <v-chart class="chart" :option="funnelOption" autoresize />
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="各阶段客户数">
          <v-chart class="chart" :option="barOption" autoresize />
        </a-card>
      </a-col>
    </a-row>

    <a-card title="各阶段明细" style="margin-top: 16px">
      <a-row :gutter="16">
        <a-col :span="24">
          <a-table :pagination="false" :data-source="funnelData">
            <template #columns>
              <a-table-column title="阶段" dataIndex="label" key="label" />
              <a-table-column title="状态" dataIndex="status" key="status" />
              <a-table-column title="客户数" dataIndex="count" key="count" />
              <a-table-column title="占比" key="ratio">
                <template #default="{ record }">
                  {{ totalCustomers > 0 ? ((record.count / totalCustomers) * 100).toFixed(2) : 0 }}%
                </template>
              </a-table-column>
              <a-table-column title="转化率" key="conversion">
                <template #default="{ record, index }">
                  <span v-if="index === 0">-</span>
                  <span v-else>
                    {{ funnelData[index - 1].count > 0 ? ((record.count / funnelData[index - 1].count) * 100).toFixed(2) : 0 }}%
                  </span>
                </template>
              </a-table-column>
            </template>
          </a-table>
        </a-col>
      </a-row>
    </a-card>

    <a-card title="客户来源分布" style="margin-top: 16px">
      <a-row :gutter="16">
        <a-col :span="12">
          <v-chart class="chart" :option="sourcePieOption" autoresize />
        </a-col>
        <a-col :span="12">
          <v-chart class="chart" :option="stylePieOption" autoresize />
        </a-col>
      </a-row>
    </a-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from '../api'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { FunnelChart, BarChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, FunnelChart, BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent])

const funnelData = ref([])
const allCustomers = ref([])

const totalCustomers = computed(() => {
  return funnelData.value.reduce((sum, item) => sum + (item.count || 0), 0)
})

const signedCustomers = computed(() => {
  const signed = funnelData.value.find(d => d.status === 'signed')
  const preparing = funnelData.value.find(d => d.status === 'preparing')
  const completed = funnelData.value.find(d => d.status === 'completed')
  return (signed?.count || 0) + (preparing?.count || 0) + (completed?.count || 0)
})

const conversionRate = computed(() => {
  const nonLost = totalCustomers.value - (funnelData.value.find(d => d.status === 'lost')?.count || 0)
  return nonLost > 0 ? (signedCustomers.value / nonLost * 100) : 0
})

const funnelOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c}人 ({d}%)' },
  series: [{
    name: '转化漏斗',
    type: 'funnel',
    left: '10%',
    top: 60,
    bottom: 60,
    width: '80%',
    min: 0,
    max: totalCustomers.value,
    sort: 'descending',
    gap: 2,
    label: { position: 'inside', formatter: '{b}\n{c}人', color: '#fff' },
    itemStyle: { borderColor: '#fff', borderWidth: 2 },
    data: funnelData.value.map(d => ({
      value: d.count,
      name: d.label,
      itemStyle: {
        color: d.status === 'consulting' ? '#1890ff' :
               d.status === 'signed' ? '#52c41a' :
               d.status === 'preparing' ? '#fa8c16' :
               d.status === 'completed' ? '#722ed1' : '#ff4d4f'
      }
    }))
  }]
}))

const barOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: funnelData.value.map(d => d.label) },
  yAxis: { type: 'value' },
  series: [{
    name: '客户数',
    type: 'bar',
    data: funnelData.value.map(d => d.count),
    itemStyle: {
      color: (params) => {
        const status = funnelData.value[params.dataIndex]?.status
        return status === 'consulting' ? '#1890ff' :
               status === 'signed' ? '#52c41a' :
               status === 'preparing' ? '#fa8c16' :
               status === 'completed' ? '#722ed1' : '#ff4d4f'
      }
    }
  }]
}))

const sourcePieOption = computed(() => {
  const sourceMap = {}
  allCustomers.value.forEach(c => {
    sourceMap[c.source] = (sourceMap[c.source] || 0) + 1
  })
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    title: { text: '客户来源分布', left: 'center' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      data: Object.entries(sourceMap).map(([name, value]) => ({ name: name || '未知', value }))
    }]
  }
})

const stylePieOption = computed(() => {
  const styleMap = {}
  allCustomers.value.forEach(c => {
    styleMap[c.style_preference] = (styleMap[c.style_preference] || 0) + 1
  })
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    title: { text: '风格偏好分布', left: 'center' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      data: Object.entries(styleMap).map(([name, value]) => ({ name: name || '未知', value }))
    }]
  }
})

async function loadData() {
  const [funnelRes, customersRes] = await Promise.all([
    api.customers.funnel(),
    api.customers.list()
  ])
  funnelData.value = funnelRes.data || []
  allCustomers.value = customersRes.data || []
}

onMounted(loadData)
</script>

<style scoped>
.chart {
  height: 350px;
}
</style>
