<script setup lang="ts">
import { 
  DataAnalysis, 
  TrendCharts, 
  UserFilled, 
  TopRight, 
  BottomRight 
} from '@element-plus/icons-vue'
</script>

<template>
  <div class="dashboard-container">
    <div class="page-header">
      <div>
        <h1 class="page-title text-gradient">数据看板 Dashboard</h1>
        <p class="page-subtitle">Welcome back, Super Admin. Here's what's happening today.</p>
      </div>
      <div>
        <el-button type="primary" size="large" class="shadow-glow" round>
          Generate Report
        </el-button>
      </div>
    </div>

    <!-- Stats Grid -->
    <el-row :gutter="24" class="mb-8">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card ghost-card">
          <div class="stat-content">
            <div class="stat-icon-wrap bg-cyan-alpha">
              <el-icon :size="28" class="text-cyan"><TrendCharts /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">Total Tickets Sold</div>
              <div class="stat-value">24,592</div>
              <div class="stat-trend trend-up">
                <el-icon><TopRight /></el-icon>
                <span>12.5% vs last week</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card ghost-card">
          <div class="stat-content">
            <div class="stat-icon-wrap bg-pink-alpha">
              <el-icon :size="28" class="text-pink"><UserFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">Active Users</div>
              <div class="stat-value">8,340</div>
              <div class="stat-trend trend-up">
                <el-icon><TopRight /></el-icon>
                <span>4.2% vs last week</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="stat-card ghost-card">
          <div class="stat-content">
            <div class="stat-icon-wrap bg-primary-alpha">
              <el-icon :size="28" class="text-primary"><DataAnalysis /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">System Load</div>
              <div class="stat-value">34%</div>
              <div class="stat-trend trend-down">
                <el-icon><BottomRight /></el-icon>
                <span>Stable</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts / Content Area -->
    <el-row :gutter="24">
      <el-col :span="16">
        <el-card shadow="hover" class="chart-card ghost-card">
          <template #header>
            <div class="card-header flex-between">
              <span class="card-title">Revenue Overview</span>
              <el-select model-value="This Week" style="width: 140px">
                <el-option label="This Week" value="week" />
                <el-option label="This Month" value="month" />
                <el-option label="This Year" value="year" />
              </el-select>
            </div>
          </template>
          <div class="chart-placeholder flex-center">
            <el-empty description="Interactive Chart Area" />
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card shadow="hover" class="chart-card ghost-card">
          <template #header>
            <span class="card-title">Recent Activities</span>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="i in 5"
              :key="i"
              :type="i === 1 ? 'primary' : 'info'"
              :size="i === 1 ? 'large' : 'normal'"
              :hollow="i !== 1"
              :timestamp="`${i * 2} mins ago`"
            >
              New order #{{ 1000 + i }} placed
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  letter-spacing: -0.5px;
}

.page-subtitle {
  color: var(--text-secondary);
  font-size: 1rem;
}

.mb-8 {
  margin-bottom: 24px;
}

/* Glass panel style for El Cards */
.ghost-card {
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  --el-card-padding: 24px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.stat-icon-wrap {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bg-cyan-alpha { background: rgba(0, 229, 255, 0.1); }
.text-cyan { color: var(--accent-cyan); }
.bg-pink-alpha { background: rgba(255, 42, 133, 0.1); }
.text-pink { color: var(--accent-pink); }
.bg-primary-alpha { background: rgba(var(--el-color-primary-rgb), 0.1); }
.text-primary { color: var(--el-color-primary); }

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
  font-weight: 500;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  margin-bottom: 8px;
}

.stat-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.85rem;
  font-weight: 600;
}

.trend-up { color: var(--el-color-success); }
.trend-down { color: var(--text-muted); }

.chart-card {
  height: 100%;
}

.card-title {
  font-size: 1.1rem;
  font-weight: 600;
}

.chart-placeholder {
  min-height: 300px;
  border: 1px dashed var(--border-color);
  border-radius: 8px;
  background: rgba(128,128,128,0.02);
}
</style>
