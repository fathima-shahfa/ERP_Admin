<script setup>
defineProps({
  modules: {
    type: Array,
    default: () => []
  },
  users: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  activities: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['configure', 'activity-click'])
</script>

<template>
  <div class="dashboard fade-in">
    <h1 class="page-title">System Overview</h1>

    <div class="stats-grid">
      <div class="stat-card glass-card">
        <h3>Total Modules</h3>
        <div class="value">{{ modules.length }}</div>
      </div>
      <div class="stat-card glass-card">
        <h3>Active Users</h3>
        <div class="value">{{ users.length || 0 }}</div>
      </div>
      <div class="stat-card glass-card">
        <h3>System Status</h3>
        <div class="value success">Online</div>
      </div>
    </div>

    <div class="dashboard-split">
      <div class="modules-section">
        <h2 class="section-title">Connected Modules</h2>
        <div v-if="loading" class="loading">Loading modules...</div>
        <div v-else class="modules-grid">
          <div v-for="mod in modules" :key="mod.id" class="module-card glass-card">
            <div class="module-header">
              <span :class="['status-indicator', mod.status]"></span>
              <h3>{{ mod.name }}</h3>
            </div>
            <p class="module-meta">Group: {{ mod.group }}</p>
            <div class="module-actions">
              <button class="btn btn-outline" @click="emit('configure', mod.name)">Configure</button>
            </div>
          </div>
        </div>
      </div>

      <div class="activity-section">
        <h2 class="section-title">Recent Activity</h2>
        <div class="glass-card activity-card">
          <div
            v-for="activity in activities"
            :key="activity.id"
            class="activity-item"
            @click="emit('activity-click', activity)"
          >
            <div :class="['activity-dot', activity.type]"></div>
            <div class="activity-content">
              <h4>{{ activity.action }}</h4>
              <p>{{ activity.detail }}</p>
              <span class="activity-time">{{ activity.time }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
