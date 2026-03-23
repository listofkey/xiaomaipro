<script setup lang="ts">
import AppSidebar from './components/AppSidebar.vue';
import AppHeader from './components/AppHeader.vue';
import { ref } from 'vue';

const isCollapse = ref(false);

const handleCollapse = (collapse: boolean) => {
  isCollapse.value = collapse;
};
</script>

<template>
  <el-container class="admin-layout">
    <!-- Top Header -->
    <el-header height="var(--header-height)" class="app-header-wrapper">
      <AppHeader @toggle-sidebar="handleCollapse" :is-collapse="isCollapse" />
    </el-header>

    <el-container class="admin-body">
      <!-- Left Sidebar -->
      <el-aside :width="isCollapse ? 'var(--sidebar-width-collapsed)' : 'var(--sidebar-width)'" class="app-sidebar-wrapper">
        <AppSidebar :is-collapse="isCollapse" />
      </el-aside>

      <!-- Main Content Area -->
      <el-main class="admin-main">
        <div class="content-wrapper">
          <router-view v-slot="{ Component }">
            <transition name="el-fade-in-linear" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-layout {
  height: 100vh;
  width: 100vw;
  background-color: var(--el-bg-color-page);
}

.app-header-wrapper {
  padding: 0;
  background-color: var(--bg-header);
  border-bottom: 1px solid var(--border-color);
  backdrop-filter: blur(20px);
  z-index: 100;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.admin-body {
  height: calc(100vh - var(--header-height));
  overflow: hidden;
}

.app-sidebar-wrapper {
  background-color: var(--bg-surface);
  border-right: 1px solid var(--border-color);
  transition: width var(--transition-normal);
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
}

.admin-main {
  background-color: transparent;
  padding: 24px;
  overflow-y: auto;
  position: relative;
}

.content-wrapper {
  max-width: 1600px;
  margin: 0 auto;
}
</style>
