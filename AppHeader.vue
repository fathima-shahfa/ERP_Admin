<script setup>
import { ref } from 'vue'

defineProps({
  currentUser: {
    type: Object,
    required: true
  },
  query: {
    type: String,
    default: ''
  },
  results: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:query', 'select-result', 'logout'])
const searchInput = ref(null)

defineExpose({
  focus: () => searchInput.value?.focus(),
  blur: () => searchInput.value?.blur()
})
</script>

<template>
  <header class="header glass">
    <div class="search input-group global-search-container">
      <i class="icon">S</i>
      <input
        ref="searchInput"
        type="text"
        :value="query"
        placeholder="Global search (Ctrl+K)..."
        @input="emit('update:query', $event.target.value)"
      />

      <div v-if="query && results.length > 0" class="search-dropdown glass-card">
        <div
          v-for="(result, index) in results"
          :key="index"
          class="search-result-item"
          @click="emit('select-result', result)"
        >
          <div class="result-type">{{ result.type }}</div>
          <div class="result-info">
            <h4>{{ result.title }}</h4>
            <span>{{ result.subtitle }}</span>
          </div>
        </div>
      </div>

      <div v-else-if="query" class="search-dropdown glass-card p-3 text-secondary text-center">
        No results found for "{{ query }}"
      </div>
    </div>

    <div class="user-profile">
      <div class="profile-info">
        <span>{{ currentUser.username }}</span>
        <button class="btn-text logout-btn" @click="emit('logout')">Logout</button>
      </div>
      <div class="avatar">{{ currentUser.initials }}</div>
    </div>
  </header>
</template>
