<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

import { logout } from '@/api/auth'
import { useUserStore } from '@/store/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

function readKeyword(value: unknown) {
  return typeof value === 'string' ? value : ''
}

const searchKeyword = ref(readKeyword(route.query.keyword))
const isDark = ref(localStorage.getItem('theme') === 'dark')

const menuKeys = new Set(['home', 'concert', 'festival', 'drama', 'sports', 'exhibition'])

if (isDark.value) {
  document.documentElement.classList.add('dark')
}

const activeMenu = computed(() => {
  const path = route.path.split('/')[1] || 'home'
  return menuKeys.has(path) ? path : ''
})

watch(
  () => route.query.keyword,
  (value) => {
    searchKeyword.value = readKeyword(value)
  },
)

function toggleDark() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function handleSelect(key: string) {
  router.push(`/${key}`)
}

function resolveSearchPath() {
  const currentPath = route.path
  if (
    currentPath === '/concert' ||
    currentPath === '/festival' ||
    currentPath === '/drama' ||
    currentPath === '/sports' ||
    currentPath === '/exhibition'
  ) {
    return currentPath
  }
  return '/home'
}

function handleSearch() {
  const keyword = searchKeyword.value.trim()
  router.push({
    path: resolveSearchPath(),
    query: keyword ? { keyword } : {},
  })
}

function handleLogin() {
  router.push('/login')
}

async function handleLogout() {
  try {
    if (userStore.token) {
      await logout()
    }
  } finally {
    userStore.clearAuth()
    ElMessage.success('已退出登录')
    router.push('/login')
  }
}
</script>

<template>
  <el-container class="layout-container">
    <el-header height="64px" class="header">
      <div class="logo" @click="router.push('/home')">Ticket Pro</div>

      <el-menu
        mode="horizontal"
        :default-active="activeMenu"
        @select="handleSelect"
        class="nav-menu"
        :ellipsis="false"
      >
        <el-menu-item index="home">首页</el-menu-item>
        <el-menu-item index="concert">演唱会</el-menu-item>
        <el-menu-item index="festival">音乐节</el-menu-item>
        <el-menu-item index="drama">话剧</el-menu-item>
        <el-menu-item index="sports">体育赛事</el-menu-item>
        <el-menu-item index="exhibition">展览</el-menu-item>
      </el-menu>

      <div class="search-box">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索活动 / 艺人 / 场馆"
          clearable
          @keyup.enter="handleSearch"
        >
          <template #append>
            <el-button @click="handleSearch">
              <el-icon><Search /></el-icon>
            </el-button>
          </template>
        </el-input>
      </div>

      <div class="user-actions">
        <el-button circle class="theme-toggle" @click="toggleDark">
          <el-icon v-if="isDark"><Sunny /></el-icon>
          <el-icon v-else><Moon /></el-icon>
        </el-button>

        <template v-if="!userStore.token">
          <el-button type="primary" @click="handleLogin">登录 / 注册</el-button>
        </template>
        <template v-else>
          <el-dropdown>
            <span class="avatar-wrapper">
              <el-avatar
                :size="32"
                :src="userStore.userInfo.avatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'"
              />
              <span class="nickname">{{ userStore.userInfo.nickname || 'User' }}</span>
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/user/profile')">个人中心</el-dropdown-item>
                <el-dropdown-item @click="router.push('/order/list')">我的订单</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </div>
    </el-header>

    <el-main class="main-content">
      <RouterView />
    </el-main>

    <el-footer class="footer">
      <div class="footer-content">
        <div class="footer-section">
          <h4>关于我们</h4>
          <p>Ticket Pro 致力于为你提供便捷、可靠的活动购票与信息服务。</p>
        </div>
        <div class="footer-section">
          <h4>联系我们</h4>
          <p>客服热线：400-123-4567</p>
          <p>Email：support@ticketpro.com</p>
        </div>
      </div>
      <div class="footer-bottom">&copy; 2026 Ticket Pro. All rights reserved.</div>
    </el-footer>
  </el-container>
</template>

<style scoped>
.layout-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 0 20px;
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--el-bg-color);
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.08);
}

.logo {
  color: #2563eb;
  font-size: 24px;
  font-weight: 800;
  cursor: pointer;
  letter-spacing: 0.02em;
}

.nav-menu {
  flex: 1;
  border-bottom: none;
  justify-content: center;
}

.nav-menu :deep(.el-menu--horizontal.el-menu) {
  border-bottom: none;
}

.nav-menu :deep(.el-menu-item) {
  font-size: 15px;
  font-weight: 500;
  padding: 0 20px;
}

.search-box {
  width: 320px;
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.theme-toggle {
  font-size: 18px;
}

.avatar-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: var(--el-text-color-primary);
}

.nickname {
  font-size: 14px;
}

.main-content {
  flex: 1;
  padding: 0;
  background: var(--el-bg-color-page);
}

.footer {
  height: auto;
  padding: 38px 0 20px;
  color: #fff;
  background: #0f172a;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}

.footer-section h4 {
  margin: 0 0 12px;
  font-size: 16px;
}

.footer-section p {
  margin: 0 0 8px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 14px;
  line-height: 1.7;
}

.footer-bottom {
  margin-top: 20px;
  text-align: center;
  color: rgba(255, 255, 255, 0.54);
  font-size: 13px;
}

@media (max-width: 900px) {
  .header {
    height: auto;
    padding: 12px;
    flex-wrap: wrap;
  }

  .nav-menu {
    order: 3;
    width: 100%;
    justify-content: flex-start;
  }

  .search-box {
    flex: 1;
    min-width: 220px;
  }

  .footer-content {
    flex-direction: column;
    padding: 0 16px 20px;
  }
}
</style>
