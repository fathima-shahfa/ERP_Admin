<script setup>
defineProps({
  users: {
    type: Array,
    default: () => []
  },
  searchQuery: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:searchQuery', 'add-user', 'delete-user'])
</script>

<template>
  <div class="users-view fade-in">
    <div class="page-header">
      <h1 class="page-title">User Management</h1>
      <button class="btn btn-primary" @click="emit('add-user')">+ Add New User</button>
    </div>

    <div class="tools-bar glass-card">
      <div class="search input-group">
        <i class="icon">S</i>
        <input
          type="text"
          :value="searchQuery"
          placeholder="Search users by name, email, or role..."
          @input="emit('update:searchQuery', $event.target.value)"
        />
      </div>
    </div>

    <div class="glass-card table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Username</th>
            <th>Email</th>
            <th>Role</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="users.length === 0">
            <td colspan="5" class="text-center">No users found matching "{{ searchQuery }}"</td>
          </tr>
          <tr v-for="user in users" v-else :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.username }}</td>
            <td>{{ user.email }}</td>
            <td><span class="badge">{{ user.role }}</span></td>
            <td class="action-cell">
              <button class="btn-text" @click="alert('Edit user coming soon!')">Edit</button>
              <button class="btn-text text-danger" @click="emit('delete-user', user.id, user.username)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
