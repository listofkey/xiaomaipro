<script setup lang="ts">
import { ref } from 'vue'
import { useDark, useToggle } from '@vueuse/core'
import { 
  Fold, 
  Expand, 
  Search, 
  Bell, 
  ChatDotRound, 
  User, 
  Setting, 
  SwitchButton,
  Sunny,
  Moon
} from '@element-plus/icons-vue'

const props = defineProps<{
  isCollapse: boolean
}>()

const emit = defineEmits(['toggle-sidebar'])

const searchStr = ref('')

const isDark = useDark()
const toggleDark = useToggle(isDark)

const toggleCollapse = () => {
  emit('toggle-sidebar', !props.isCollapse)
}
</script>

<template>
  <div class="header-content">
    <!-- Left Section -->
    <div class="header-left">
      <div class="logo-area flex-center">
        <!-- Reusing the css gradient background -->
        <div class="logo-icon glass-panel primary-gradient">
          <span style="color: #fff; font-weight: bold; font-size: 16px;">TX</span>
        </div>
        <h2 class="logo-text text-gradient">TickeX Admin</h2>
      </div>

      <el-icon class="collapse-icon ml-4" @click="toggleCollapse">
        <component :is="isCollapse ? Expand : Fold" />
      </el-icon>

      <el-tag
        type="success"
        effect="plain"
        round
        class="ml-4 env-badge"
      >
        <span class="pulse-dot"></span>
        Production
      </el-tag>
    </div>

    <!-- Center Section -->
    <div class="header-center">
      <el-input
        v-model="searchStr"
        placeholder="全局搜索... (Ctrl+K)"
        class="search-input"
        :prefix-icon="Search"
      >
        <template #suffix>
          <div class="hotkey">⌘K</div>
        </template>
      </el-input>
    </div>

    <!-- Right Section -->
    <div class="header-right">
      <!-- Theme Switch -->
      <el-switch
        v-model="isDark"
        inline-prompt
        :active-icon="Moon"
        :inactive-icon="Sunny"
        @change="toggleDark"
        class="theme-switch"
        style="margin-right: 16px;"
      />

      <el-badge :value="3" class="action-badge">
        <el-button circle :icon="ChatDotRound" class="action-btn" />
      </el-badge>
      
      <el-badge :value="2" type="danger" class="action-badge mx-3">
        <el-button circle :icon="Bell" class="action-btn" />
      </el-badge>

      <div class="divider mx-3"></div>

      <el-dropdown trigger="click">
        <div class="user-profile flex-center">
          <el-avatar :size="32" class="avatar-gradient">A</el-avatar>
          <div class="user-info ml-2">
            <span class="user-name">Admin User</span>
            <span class="user-role">Super Admin</span>
          </div>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item :icon="User">个人中心</el-dropdown-item>
            <el-dropdown-item :icon="Setting">偏好设置</el-dropdown-item>
            <el-dropdown-item divided :icon="SwitchButton" class="danger">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<style scoped>
.header-content {
  height: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 200px;
}

.logo-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--el-color-primary) 0%, var(--el-color-success) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(135deg, var(--el-text-color-primary) 0%, var(--el-text-color-secondary) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.collapse-icon {
  font-size: 20px;
  cursor: pointer;
  color: var(--el-text-color-regular);
  transition: color 0.3s;
}

.collapse-icon:hover {
  color: var(--el-color-primary);
}

.env-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  background: rgba(var(--el-color-success-rgb), 0.1);
}

.pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: var(--el-color-success);
  box-shadow: 0 0 8px var(--el-color-success);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.5); opacity: 0.5; }
  100% { transform: scale(1); opacity: 1; }
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.search-input {
  width: 400px;
  --el-input-bg-color: var(--bg-surface);
  --el-input-border-color: var(--border-color);
}

.hotkey {
  font-size: 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 0 4px;
  color: var(--text-muted);
  background: rgba(128,128,128, 0.1);
}

.header-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.action-badge {
  margin-left: 10px;
}

.action-btn {
  border: none !important;
  background: transparent !important;
  font-size: 18px;
  color: var(--el-text-color-regular);
}

.action-btn:hover {
  color: var(--el-color-primary);
  background-color: var(--el-fill-color-light) !important;
}

.theme-switch {
  --el-switch-on-color: var(--el-fill-color-dark);
  --el-switch-off-color: var(--el-color-primary);
}

.divider {
  width: 1px;
  height: 24px;
  background-color: var(--border-color);
  margin: 0 16px;
}

.ml-4 { margin-left: 16px; }
.ml-2 { margin-left: 8px; }
.mx-3 { margin-left: 12px; margin-right: 12px; }

.user-profile {
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background-color 0.3s;
}

.user-profile:hover {
  background-color: var(--el-fill-color-light);
}

.avatar-gradient {
  background: linear-gradient(135deg, var(--el-color-primary) 0%, var(--el-color-success) 100%);
  color: white;
  font-weight: bold;
}

.user-info {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.user-role {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

::v-deep(.danger) {
  color: var(--el-color-danger);
}
</style>
