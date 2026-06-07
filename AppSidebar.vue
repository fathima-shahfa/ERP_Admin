<script setup>
defineProps({
  currentView: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['change-view'])

const navItems = [
  { view: 'dashboard', icon: 'DB', label: 'Dashboard' },
  { view: 'users', icon: 'US', label: 'User Management' },
  { view: 'settings', icon: 'GS', label: 'Global Settings' },
  { view: 'integrations', icon: 'IN', label: 'Integrations' }
]
</script>

<template>
  <aside class="sidebar glass">
    <div class="logo-container">
      <div class="logo"></div>
      <h2>ERP Admin</h2>
    </div>

    <nav>
      <a
        v-for="item in navItems"
        :key="item.view"
        href="#"
        :class="['nav-item', { active: currentView === item.view }]"
        @click.prevent="emit('change-view', item.view)"
      >
        <i class="icon">{{ item.icon }}</i> {{ item.label }}
      </a>
    </nav>
  </aside>
</template>
