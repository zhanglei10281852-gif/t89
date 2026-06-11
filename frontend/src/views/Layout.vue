<template>
  <a-layout style="min-height: 100vh">
    <a-layout-sider v-model:collapsed="collapsed" collapsible theme="dark">
      <div class="logo">
        <span v-if="!collapsed">婚庆管理系统</span>
        <span v-else>婚庆</span>
      </div>
      <a-menu theme="dark" mode="inline" :selected-keys="[selectedKey]" @click="handleMenuClick">
        <a-menu-item key="dashboard">
          <template #icon><DashboardOutlined /></template>
          <span>工作台</span>
        </a-menu-item>
        <a-sub-menu key="customers">
          <template #title>
            <span><UserOutlined /><span>客户管理</span></span>
          </template>
          <a-menu-item key="customers">客户列表</a-menu-item>
          <a-menu-item key="customers/funnel">转化漏斗</a-menu-item>
        </a-sub-menu>
        <a-menu-item key="packages">
          <template #icon><GiftOutlined /></template>
          <span>套餐管理</span>
        </a-menu-item>
        <a-menu-item key="quotes">
          <template #icon><FileTextOutlined /></template>
          <span>报价方案</span>
        </a-menu-item>
        <a-menu-item key="schedule">
          <template #icon><CalendarOutlined /></template>
          <span>档期日历</span>
        </a-menu-item>
        <a-menu-item key="contracts">
          <template #icon><FormOutlined /></template>
          <span>签约管理</span>
        </a-menu-item>
        <a-menu-item key="stats">
          <template #icon><BarChartOutlined /></template>
          <span>统计报表</span>
        </a-menu-item>
        <a-menu-item v-if="userStore.isPlanner" key="planner">
          <template #icon><UserSwitchOutlined /></template>
          <span>我的工作台</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="header">
        <div class="header-title">{{ currentTitle }}</div>
        <div class="header-right">
          <a-dropdown>
            <a-space>
              <a-avatar>{{ userStore.userInfo?.name?.charAt(0) }}</a-avatar>
              <span>{{ userStore.userInfo?.name }}</span>
              <DownOutlined />
            </a-space>
            <template #overlay>
              <a-menu @click="handleUserMenu">
                <a-menu-item key="logout">
                  <LogoutOutlined /> 退出登录
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      <a-layout-content class="content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../store/user'
import {
  DashboardOutlined,
  UserOutlined,
  GiftOutlined,
  FileTextOutlined,
  CalendarOutlined,
  FormOutlined,
  BarChartOutlined,
  UserSwitchOutlined,
  DownOutlined,
  LogoutOutlined
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const collapsed = ref(false)
const selectedKey = computed(() => route.path.replace('/', '') || 'dashboard')
const currentTitle = computed(() => route.meta.title || '婚庆管理系统')

function handleMenuClick({ key }) {
  router.push('/' + key)
}

function handleUserMenu({ key }) {
  if (key === 'logout') {
    userStore.logout()
    message.success('已退出登录')
    router.push('/login')
  }
}
</script>

<style scoped>
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.1);
}
.header {
  background: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}
.header-title {
  font-size: 18px;
  font-weight: 600;
}
.content {
  margin: 24px;
  background: white;
  padding: 24px;
  border-radius: 8px;
  min-height: calc(100vh - 112px);
}
</style>
