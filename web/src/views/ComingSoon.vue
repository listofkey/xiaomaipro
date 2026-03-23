<script setup lang="ts">
// Dummy generic component for other views
import { Coffee } from '@element-plus/icons-vue'
</script>

<template>
  <div class="page-container glass-panel">
    <el-empty
      :image-size="200"
      description="This module space is reserved for functionality integration."
    >
      <template #image>
         <el-icon :size="100" class="primary-gradient pulse"><Coffee /></el-icon>
      </template>
      <h2>{{ $route.name }} - Development in Progress</h2>
    </el-empty>
  </div>
</template>

<style scoped>
.page-container {
  height: calc(100vh - 120px);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-color);
  border-radius: 20px;
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
}

.pulse {
  animation: pulse-op 2s infinite;
}

@keyframes pulse-op {
  0% { opacity: 0.7; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.1); }
  100% { opacity: 0.7; transform: scale(1); }
}

h2 {
  font-weight: 600;
  background: linear-gradient(135deg, var(--el-text-color-primary) 0%, var(--el-text-color-secondary) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 12px;
}
</style>
