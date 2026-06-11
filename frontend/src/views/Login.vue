<template>
  <div class="login-container">
    <a-card class="login-card" title="婚庆策划管理系统">
      <a-form :model="form" @finish="handleLogin">
        <a-form-item name="username" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model:value="form.username" placeholder="用户名" size="large">
            <template #prefix><UserOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item name="password" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password v-model:value="form.password" placeholder="密码" size="large">
            <template #prefix><LockOutlined /></template>
          </a-input-password>
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" block size="large" :loading="loading">
            登录
          </a-button>
        </a-form-item>
      </a-form>
      <div class="tips">
        <p>管理员: admin / wedding@2024</p>
        <p>策划师: planner1~planner5 / p123</p>
      </div>
    </a-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const router = useRouter()
const userStore = useUserStore()

const form = reactive({
  username: '',
  password: ''
})
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    await userStore.login(form.username, form.password)
    message.success('登录成功')
    router.push('/')
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}
.tips {
  text-align: center;
  color: #999;
  font-size: 12px;
  margin-top: 16px;
}
.tips p {
  margin: 4px 0;
}
</style>
