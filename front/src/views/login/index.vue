<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

import { login, register } from '@/api/auth'
import { useUserStore } from '@/store/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const isLoginMode = ref(true)

const form = reactive({
  account: '',
  password: '',
  confirmPassword: '',
  nickname: '',
})

const accountPlaceholder = computed(() => '请输入手机号或邮箱')

const accountValidator = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  const trimmed = value.trim()
  const phonePattern = /^1[3-9]\d{9}$/
  const emailPattern = /^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/

  if (!trimmed) {
    callback(new Error('请输入手机号或邮箱'))
    return
  }
  if (!phonePattern.test(trimmed) && !emailPattern.test(trimmed)) {
    callback(new Error('请输入正确的手机号或邮箱'))
    return
  }

  callback()
}

const confirmPasswordValidator = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (isLoginMode.value) {
    callback()
    return
  }
  if (!value) {
    callback(new Error('请再次输入密码'))
    return
  }
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
    return
  }

  callback()
}

const rules: FormRules<typeof form> = {
  account: [{ validator: accountValidator, trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度为 6 到 20 位', trigger: 'blur' },
  ],
  confirmPassword: [{ validator: confirmPasswordValidator, trigger: 'blur' }],
  nickname: [{ max: 20, message: '昵称不能超过 20 个字符', trigger: 'blur' }],
}

function normalizeUserInfo(payload: {
  id: string
  phone: string
  email: string
  nickname: string
  avatar: string
  status: number
  isVerified: boolean
  realName?: string
  createdAt?: string
}) {
  return {
    id: payload.id,
    phone: payload.phone,
    email: payload.email,
    nickname: payload.nickname,
    avatar: payload.avatar,
    status: payload.status,
    isVerified: payload.isVerified,
    isRealName: payload.isVerified,
    realName: payload.realName || '',
    createdAt: payload.createdAt || '',
  }
}

function resetMode(mode: 'login' | 'register') {
  isLoginMode.value = mode === 'login'
  form.password = ''
  form.confirmPassword = ''
  if (mode === 'login') {
    form.nickname = ''
  }
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  if (!formRef.value) {
    return
  }

  await formRef.value.validate()

  const account = form.account.trim()
  const nickname = form.nickname.trim()
  const isEmail = account.includes('@')

  loading.value = true
  try {
    const result = isLoginMode.value
      ? await login({
          account,
          password: form.password,
        })
      : await register({
          phone: isEmail ? undefined : account,
          email: isEmail ? account : undefined,
          password: form.password,
          nickname: nickname || undefined,
        })

    userStore.setAuth({
      accessToken: result.accessToken,
      refreshToken: result.refreshToken,
      userInfo: normalizeUserInfo(result.userInfo),
    })

    ElMessage.success(isLoginMode.value ? '登录成功' : '注册成功')

    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.push(redirect)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-shell">
      <section class="login-brand">
        <p class="eyebrow">Xiaomai Pro</p>
        <h1>更稳的票务账户体系</h1>
        <p class="summary">
          登录后即可同步订单、购票人、地址与实名状态，认证接口已接入网关与 user-rpc。
        </p>
      </section>

      <section class="login-panel">
        <div class="panel-header">
          <h2>{{ isLoginMode ? '账号登录' : '创建账号' }}</h2>
          <p>{{ isLoginMode ? '使用手机号或邮箱登录' : '使用手机号或邮箱完成注册' }}</p>
        </div>

        <el-form ref="formRef" :model="form" :rules="rules" size="large" label-position="top">
          <el-form-item label="账号" prop="account">
            <el-input
              v-model="form.account"
              :placeholder="accountPlaceholder"
              autocomplete="username"
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <el-form-item v-if="!isLoginMode" label="昵称" prop="nickname">
            <el-input
              v-model="form.nickname"
              placeholder="请输入昵称，留空则由系统生成"
              maxlength="20"
              show-word-limit
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              show-password
              autocomplete="current-password"
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <el-form-item v-if="!isLoginMode" label="确认密码" prop="confirmPassword">
            <el-input
              v-model="form.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              show-password
              autocomplete="new-password"
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <el-button type="primary" class="submit-button" :loading="loading" @click="handleSubmit">
            {{ isLoginMode ? '登录' : '注册并登录' }}
          </el-button>

          <div class="mode-switch">
            <span v-if="isLoginMode">
              还没有账号？
              <el-link type="primary" :underline="false" @click="resetMode('register')">立即注册</el-link>
            </span>
            <span v-else>
              已有账号？
              <el-link type="primary" :underline="false" @click="resetMode('login')">返回登录</el-link>
            </span>
          </div>
        </el-form>
      </section>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background:
    radial-gradient(circle at top left, rgba(242, 214, 165, 0.45), transparent 32%),
    radial-gradient(circle at bottom right, rgba(64, 158, 255, 0.2), transparent 30%),
    linear-gradient(135deg, #fbf7ef 0%, #f1efe9 50%, #eef3f8 100%);
}

.login-shell {
  width: min(1080px, 100%);
  min-height: 620px;
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  overflow: hidden;
  border: 1px solid rgba(30, 41, 59, 0.08);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 24px 80px rgba(46, 62, 79, 0.18);
  backdrop-filter: blur(14px);
}

.login-brand {
  padding: 72px 64px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  color: #1f2937;
  background:
    linear-gradient(160deg, rgba(250, 204, 21, 0.16), transparent 48%),
    linear-gradient(180deg, rgba(15, 23, 42, 0.03), rgba(15, 23, 42, 0));
}

.eyebrow {
  margin: 0 0 18px;
  color: #8c6a20;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.login-brand h1 {
  margin: 0;
  font-size: 50px;
  line-height: 1.06;
  letter-spacing: -0.04em;
}

.summary {
  max-width: 420px;
  margin: 24px 0 0;
  color: #475569;
  font-size: 17px;
  line-height: 1.8;
}

.login-panel {
  padding: 56px 48px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.panel-header {
  margin-bottom: 28px;
}

.panel-header h2 {
  margin: 0;
  font-size: 30px;
  color: #111827;
}

.panel-header p {
  margin: 12px 0 0;
  color: #6b7280;
  font-size: 14px;
}

.submit-button {
  width: 100%;
  height: 48px;
  margin-top: 8px;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 600;
}

.mode-switch {
  margin-top: 22px;
  color: #6b7280;
  text-align: center;
  font-size: 14px;
}

@media (max-width: 900px) {
  .login-shell {
    grid-template-columns: 1fr;
  }

  .login-brand {
    padding: 48px 32px 20px;
  }

  .login-brand h1 {
    font-size: 36px;
  }

  .summary {
    max-width: none;
  }

  .login-panel {
    padding: 28px 32px 40px;
  }
}

@media (max-width: 520px) {
  .login-page {
    padding: 12px;
  }

  .login-brand,
  .login-panel {
    padding-left: 20px;
    padding-right: 20px;
  }
}
</style>
