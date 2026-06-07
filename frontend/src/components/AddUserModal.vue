<script setup>
defineProps({
  form: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['submit', 'close'])
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content glass">
      <div class="modal-header">
        <h2>Add New User</h2>
        <button class="btn-close" @click="emit('close')">&times;</button>
      </div>

      <form class="modal-body" @submit.prevent="emit('submit')">
        <div class="form-group">
          <label>Username</label>
          <input v-model="form.username" type="text" required placeholder="e.g. jdoe" />
        </div>

        <div class="form-group">
          <label>Email</label>
          <input v-model="form.email" type="email" required placeholder="e.g. jdoe@erp.com" />
        </div>

        <div class="form-group">
          <label>Role</label>
          <select v-model="form.role" class="form-select">
            <option value="SuperAdmin">SuperAdmin</option>
            <option value="Admin">Admin</option>
            <option value="User">User</option>
          </select>
        </div>

        <div class="form-group">
          <label>Temporary Password</label>
          <input v-model="form.password" type="password" required placeholder="Enter password" />
        </div>

        <div class="modal-footer">
          <button type="button" class="btn btn-outline" @click="emit('close')">Cancel</button>
          <button type="submit" class="btn btn-primary">Create User</button>
        </div>
      </form>
    </div>
  </div>
</template>
