<script setup lang="ts">
import { useRoute } from 'vue-router'
import {
  DataBoard,
  Ticket,
  List,
  Calendar,
  Box,
  ShoppingBag,
  CircleCheck,
  Wallet,
  PieChart,
  Setting
} from '@element-plus/icons-vue'

const props = defineProps<{
  isCollapse: boolean
}>()

const route = useRoute()
</script>

<template>
  <el-menu
    :default-active="route.path"
    class="admin-menu"
    :collapse="isCollapse"
    :collapse-transition="false"
    router
  >
    <el-menu-item index="/dashboard">
      <el-icon><DataBoard /></el-icon>
      <template #title>看板</template>
    </el-menu-item>

    <el-sub-menu index="/activities">
      <template #title>
        <el-icon><Ticket /></el-icon>
        <span>活动</span>
      </template>
      <el-menu-item index="/activities/list">
        <template #title>活动列表</template>
      </el-menu-item>
      <el-menu-item index="/activities/create">
        <template #title>新建活动</template>
      </el-menu-item>
    </el-sub-menu>

    <el-menu-item index="/sessions">
      <el-icon><Calendar /></el-icon>
      <template #title>场次票档</template>
    </el-menu-item>

    <el-menu-item index="/inventory">
      <el-icon><Box /></el-icon>
      <template #title>库存</template>
    </el-menu-item>

    <el-menu-item index="/orders">
      <el-icon><ShoppingBag /></el-icon>
      <template #title>订单</template>
    </el-menu-item>

    <el-menu-item index="/risk">
      <el-icon><CircleCheck /></el-icon>
      <template #title>风控</template>
    </el-menu-item>

    <el-menu-item index="/finance">
      <el-icon><Wallet /></el-icon>
      <template #title>财务</template>
    </el-menu-item>

    <el-menu-item index="/reports">
      <el-icon><PieChart /></el-icon>
      <template #title>报表</template>
    </el-menu-item>

    <el-menu-item index="/settings">
      <el-icon><Setting /></el-icon>
      <template #title>系统设置</template>
    </el-menu-item>
  </el-menu>
</template>

<style scoped>
.admin-menu {
  height: 100%;
  border-right: none;
  font-weight: 500;
  background-color: transparent; /* Allow sidebar surface background to show */
  --el-menu-bg-color: transparent;
  --el-menu-hover-bg-color: var(--el-color-primary-light-9);
}

html.dark .admin-menu {
  --el-menu-hover-bg-color: var(--el-color-primary-light-9);
}

.admin-menu:not(.el-menu--collapse) {
  width: var(--sidebar-width);
}

.el-menu-item.is-active {
  background-color: var(--el-color-primary-light-9);
  border-right: 3px solid var(--el-color-primary);
}
</style>
