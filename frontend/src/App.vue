<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { adminApi } from './api/adminApi'

const views = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'users', label: 'Users' },
  { id: 'roles', label: 'Roles & Permissions' },
  { id: 'modules', label: 'Modules' },
  { id: 'settings', label: 'Settings' },
  { id: 'audit', label: 'Audit Logs' }
]

const currentView = ref('dashboard')
const loading = ref(false)
const error = ref('')
const isLoggedIn = ref(false)
const sessionUser = ref(null)
const loginForm = reactive({ username: 'admin', password: 'admin123' })

const dashboard = ref({})
const users = ref([])
const roles = ref([])
const permissions = ref([])
const modules = ref([])
const settings = ref([])
const auditLogs = ref([])

const userSearch = ref('')
const auditSearch = ref('')
const selectedUser = ref(null)
const selectedRole = ref(null)
const selectedModule = ref(null)

const userForm = reactive({
  id: null,
  username: '',
  full_name: '',
  email: '',
  department: '',
  role_id: '',
  status: 'active',
  password: 'Welcome123'
})

const roleForm = reactive({
  id: null,
  name: '',
  description: '',
  permissions: []
})

const moduleForm = reactive({
  id: '',
  status: 'active',
  owner: '',
  description: ''
})

const groupedSettings = computed(() => {
  return settings.value.reduce((groups, setting) => {
    groups[setting.setting_group] ||= []
    groups[setting.setting_group].push(setting)
    return groups
  }, {})
})

const groupedPermissions = computed(() => {
  return permissions.value.reduce((groups, permission) => {
    groups[permission.module] ||= []
    groups[permission.module].push(permission)
    return groups
  }, {})
})

const filteredUsers = computed(() => {
  const query = userSearch.value.toLowerCase()
  if (!query) return users.value
  return users.value.filter(user =>
    user.username.toLowerCase().includes(query) ||
    user.email.toLowerCase().includes(query) ||
    user.full_name.toLowerCase().includes(query) ||
    user.department.toLowerCase().includes(query) ||
    (user.role?.name || '').toLowerCase().includes(query)
  )
})

const filteredAuditLogs = computed(() => {
  const query = auditSearch.value.toLowerCase()
  if (!query) return auditLogs.value
  return auditLogs.value.filter(log =>
    log.actor.toLowerCase().includes(query) ||
    log.action.toLowerCase().includes(query) ||
    log.target.toLowerCase().includes(query) ||
    log.severity.toLowerCase().includes(query)
  )
})

async function loadAll() {
  loading.value = true
  error.value = ''
  try {
    const [
      dashboardData,
      usersData,
      rolesData,
      permissionsData,
      modulesData,
      settingsData,
      auditData
    ] = await Promise.all([
      adminApi.dashboard(),
      adminApi.users(),
      adminApi.roles(),
      adminApi.permissions(),
      adminApi.modules(),
      adminApi.settings(),
      adminApi.auditLogs()
    ])

    dashboard.value = dashboardData
    users.value = usersData
    roles.value = rolesData
    permissions.value = permissionsData
    modules.value = modulesData
    settings.value = settingsData
    auditLogs.value = auditData
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

async function login() {
  error.value = ''
  try {
    const result = await adminApi.login(loginForm)
    sessionUser.value = result.user
    isLoggedIn.value = true
    await loadAll()
  } catch (err) {
    error.value = err.message
  }
}

function logout() {
  isLoggedIn.value = false
  sessionUser.value = null
  currentView.value = 'dashboard'
}

function resetUserForm() {
  selectedUser.value = null
  Object.assign(userForm, {
    id: null,
    username: '',
    full_name: '',
    email: '',
    department: '',
    role_id: roles.value[0]?.id || '',
    status: 'active',
    password: 'Welcome123'
  })
}

function editUser(user) {
  selectedUser.value = user
  Object.assign(userForm, {
    id: user.id,
    username: user.username,
    full_name: user.full_name,
    email: user.email,
    department: user.department,
    role_id: user.role_id,
    status: user.status,
    password: 'Welcome123'
  })
}

async function saveUser() {
  try {
    if (userForm.id) {
      await adminApi.updateUser(userForm.id, {
        email: userForm.email,
        full_name: userForm.full_name,
        department: userForm.department,
        role_id: Number(userForm.role_id),
        status: userForm.status
      })
    } else {
      await adminApi.createUser({
        username: userForm.username,
        email: userForm.email,
        full_name: userForm.full_name,
        department: userForm.department,
        role_id: Number(userForm.role_id),
        status: userForm.status,
        password: userForm.password
      })
    }
    resetUserForm()
    await loadAll()
  } catch (err) {
    alert(err.message)
  }
}

async function removeUser(user) {
  if (!confirm(`Delete ${user.username}?`)) return
  await adminApi.deleteUser(user.id)
  await loadAll()
}

async function resetPassword(user) {
  const password = prompt(`New password for ${user.username}`, 'Welcome123')
  if (!password) return
  await adminApi.resetPassword(user.id, password)
  alert('Password reset successfully')
}

function resetRoleForm() {
  selectedRole.value = null
  Object.assign(roleForm, { id: null, name: '', description: '', permissions: [] })
}

function editRole(role) {
  selectedRole.value = role
  Object.assign(roleForm, {
    id: role.id,
    name: role.name,
    description: role.description,
    permissions: [...role.permissions]
  })
}

function togglePermission(permissionID) {
  if (roleForm.permissions.includes(permissionID)) {
    roleForm.permissions = roleForm.permissions.filter(id => id !== permissionID)
  } else {
    roleForm.permissions.push(permissionID)
  }
}

async function saveRole() {
  try {
    const payload = {
      name: roleForm.name,
      description: roleForm.description,
      permissions: roleForm.permissions
    }
    if (roleForm.id) await adminApi.updateRole(roleForm.id, payload)
    else await adminApi.createRole(payload)
    resetRoleForm()
    await loadAll()
  } catch (err) {
    alert(err.message)
  }
}

async function removeRole(role) {
  if (!confirm(`Delete role ${role.name}?`)) return
  try {
    await adminApi.deleteRole(role.id)
    await loadAll()
  } catch (err) {
    alert(err.message)
  }
}

function editModule(module) {
  selectedModule.value = module
  Object.assign(moduleForm, {
    id: module.id,
    status: module.status,
    owner: module.owner,
    description: module.description
  })
}

async function saveModule() {
  await adminApi.updateModule(moduleForm.id, moduleForm)
  selectedModule.value = null
  await loadAll()
}

async function saveSettings() {
  await adminApi.saveSettings(settings.value)
  await loadAll()
  alert('Settings saved')
}

function formatDate(value) {
  if (!value) return 'Never'
  return new Date(value).toLocaleString()
}

onMounted(() => {
  resetUserForm()
})
</script>

<template>
  <div class="layout">
    <section v-if="!isLoggedIn" class="login-container">
      <form class="login-box glass" @submit.prevent="login">
        <div class="login-header">
          <div class="logo"></div>
          <h2>ERP Admin Module</h2>
          <p>Sign in to manage users, roles, modules, settings, and audit history.</p>
        </div>
        <div v-if="error" class="error-alert">{{ error }}</div>
        <div class="form-group">
          <label>Username</label>
          <input v-model="loginForm.username" type="text" />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input v-model="loginForm.password" type="password" />
        </div>
        <button class="btn btn-primary login-btn" type="submit">Sign In</button>
      </form>
    </section>

    <template v-else>
      <aside class="sidebar glass">
        <div class="logo-container">
          <div class="logo"></div>
          <h2>ERP Admin</h2>
        </div>
        <nav>
          <a
            v-for="view in views"
            :key="view.id"
            href="#"
            :class="['nav-item', { active: currentView === view.id }]"
            @click.prevent="currentView = view.id"
          >
            <i class="icon">{{ view.label.slice(0, 2).toUpperCase() }}</i> {{ view.label }}
          </a>
        </nav>
      </aside>

      <main class="main-content">
        <header class="header glass">
          <div>
            <strong>{{ sessionUser.full_name || sessionUser.username }}</strong>
            <span class="muted"> / {{ sessionUser.role.name }}</span>
          </div>
          <div class="header-actions">
            <button class="btn btn-outline" @click="loadAll">Refresh</button>
            <button class="btn-text logout-btn" @click="logout">Logout</button>
          </div>
        </header>

        <div v-if="error" class="error-alert">{{ error }}</div>
        <div v-if="loading" class="loading glass-card">Loading admin data...</div>

        <section v-if="currentView === 'dashboard'" class="fade-in">
          <h1 class="page-title">Admin Dashboard</h1>
          <div class="stats-grid">
            <div class="stat-card glass-card"><h3>Total Users</h3><div class="value">{{ dashboard.total_users }}</div></div>
            <div class="stat-card glass-card"><h3>Active Users</h3><div class="value success">{{ dashboard.active_users }}</div></div>
            <div class="stat-card glass-card"><h3>Suspended Users</h3><div class="value danger">{{ dashboard.suspended_users }}</div></div>
            <div class="stat-card glass-card"><h3>Roles</h3><div class="value">{{ dashboard.roles }}</div></div>
            <div class="stat-card glass-card"><h3>Active Modules</h3><div class="value success">{{ dashboard.active_modules }}</div></div>
            <div class="stat-card glass-card"><h3>Audit Events</h3><div class="value">{{ dashboard.audit_logs }}</div></div>
          </div>

          <h2 class="section-title">Recent Activity</h2>
          <div class="glass-card activity-card">
            <div v-for="log in dashboard.recent_activity" :key="log.id" class="activity-item">
              <div :class="['activity-dot', log.severity]"></div>
              <div class="activity-content">
                <h4>{{ log.action }} - {{ log.target }}</h4>
                <p>{{ log.details }}</p>
                <span class="activity-time">{{ log.actor }} / {{ formatDate(log.created_at) }}</span>
              </div>
            </div>
          </div>
        </section>

        <section v-if="currentView === 'users'" class="fade-in">
          <div class="page-header">
            <h1 class="page-title">User Management</h1>
            <button class="btn btn-outline" @click="resetUserForm">New User</button>
          </div>
          <div class="admin-grid">
            <form class="glass-card" @submit.prevent="saveUser">
              <h3>{{ userForm.id ? 'Edit User' : 'Create User' }}</h3>
              <div class="form-group"><label>Username</label><input v-model="userForm.username" :disabled="!!userForm.id" required /></div>
              <div class="form-group"><label>Full Name</label><input v-model="userForm.full_name" /></div>
              <div class="form-group"><label>Email</label><input v-model="userForm.email" type="email" required /></div>
              <div class="form-group"><label>Department</label><input v-model="userForm.department" /></div>
              <div class="form-group">
                <label>Role</label>
                <select v-model="userForm.role_id" class="form-select" required>
                  <option v-for="role in roles" :key="role.id" :value="role.id">{{ role.name }}</option>
                </select>
              </div>
              <div class="form-group">
                <label>Status</label>
                <select v-model="userForm.status" class="form-select">
                  <option value="active">Active</option>
                  <option value="suspended">Suspended</option>
                  <option value="pending">Pending</option>
                </select>
              </div>
              <div v-if="!userForm.id" class="form-group"><label>Temporary Password</label><input v-model="userForm.password" required /></div>
              <button class="btn btn-primary" type="submit">{{ userForm.id ? 'Save User' : 'Create User' }}</button>
            </form>

            <div>
              <div class="tools-bar glass-card">
                <div class="input-group"><i class="icon">S</i><input v-model="userSearch" placeholder="Search users..." /></div>
              </div>
              <div class="glass-card table-container">
                <table class="data-table">
                  <thead><tr><th>User</th><th>Role</th><th>Status</th><th>Last Login</th><th>Actions</th></tr></thead>
                  <tbody>
                    <tr v-for="user in filteredUsers" :key="user.id">
                      <td><strong>{{ user.username }}</strong><br><span class="muted">{{ user.email }}</span></td>
                      <td><span class="badge">{{ user.role?.name }}</span></td>
                      <td><span :class="['status-badge', user.status]">{{ user.status }}</span></td>
                      <td>{{ formatDate(user.last_login_at) }}</td>
                      <td class="action-cell">
                        <button class="btn-text" @click="editUser(user)">Edit</button>
                        <button class="btn-text" @click="resetPassword(user)">Reset</button>
                        <button class="btn-text text-danger" @click="removeUser(user)">Delete</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <section v-if="currentView === 'roles'" class="fade-in">
          <div class="page-header">
            <h1 class="page-title">Roles & Permissions</h1>
            <button class="btn btn-outline" @click="resetRoleForm">New Role</button>
          </div>
          <div class="admin-grid">
            <form class="glass-card" @submit.prevent="saveRole">
              <h3>{{ roleForm.id ? 'Edit Role' : 'Create Role' }}</h3>
              <div class="form-group"><label>Role Name</label><input v-model="roleForm.name" :disabled="selectedRole?.is_system" required /></div>
              <div class="form-group"><label>Description</label><input v-model="roleForm.description" /></div>
              <div v-for="(items, group) in groupedPermissions" :key="group" class="permission-group">
                <h4>{{ group }}</h4>
                <label v-for="permission in items" :key="permission.id" class="check-row">
                  <input type="checkbox" :checked="roleForm.permissions.includes(permission.id)" @change="togglePermission(permission.id)" />
                  <span>{{ permission.label }}</span>
                </label>
              </div>
              <button class="btn btn-primary" type="submit">{{ roleForm.id ? 'Save Role' : 'Create Role' }}</button>
            </form>
            <div class="roles-grid">
              <div v-for="role in roles" :key="role.id" class="glass-card">
                <div class="card-row">
                  <h3>{{ role.name }}</h3>
                  <span v-if="role.is_system" class="badge">System</span>
                </div>
                <p class="muted">{{ role.description }}</p>
                <p>{{ role.permissions.length }} permissions assigned</p>
                <div class="action-cell">
                  <button class="btn-text" @click="editRole(role)">Edit</button>
                  <button v-if="!role.is_system" class="btn-text text-danger" @click="removeRole(role)">Delete</button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section v-if="currentView === 'modules'" class="fade-in">
          <h1 class="page-title">Module Management</h1>
          <div class="modules-grid">
            <div v-for="module in modules" :key="module.id" class="glass-card module-card">
              <div class="module-header">
                <span :class="['status-indicator', module.status]"></span>
                <h3>{{ module.name }}</h3>
              </div>
              <p class="muted">Group {{ module.module_group }} / Owner: {{ module.owner }}</p>
              <p>{{ module.description }}</p>
              <div class="module-actions"><button class="btn btn-outline" @click="editModule(module)">Configure</button></div>
            </div>
          </div>
          <form v-if="selectedModule" class="glass-card mt-4" @submit.prevent="saveModule">
            <h3>Configure {{ selectedModule.name }}</h3>
            <div class="settings-grid">
              <div class="form-group"><label>Status</label><select v-model="moduleForm.status" class="form-select"><option value="active">Active</option><option value="maintenance">Maintenance</option><option value="inactive">Inactive</option></select></div>
              <div class="form-group"><label>Owner</label><input v-model="moduleForm.owner" /></div>
            </div>
            <div class="form-group"><label>Description</label><input v-model="moduleForm.description" /></div>
            <button class="btn btn-primary" type="submit">Save Module</button>
          </form>
        </section>

        <section v-if="currentView === 'settings'" class="fade-in">
          <div class="page-header">
            <h1 class="page-title">System Settings</h1>
            <button class="btn btn-primary" @click="saveSettings">Save Settings</button>
          </div>
          <div class="settings-grid">
            <div v-for="(items, group) in groupedSettings" :key="group" class="glass-card">
              <h3>{{ group }}</h3>
              <div v-for="setting in items" :key="setting.key" class="form-group">
                <label>{{ setting.label }}</label>
                <select v-if="setting.value === 'true' || setting.value === 'false'" v-model="setting.value" class="form-select">
                  <option value="true">Enabled</option>
                  <option value="false">Disabled</option>
                </select>
                <input v-else v-model="setting.value" />
              </div>
            </div>
          </div>
        </section>

        <section v-if="currentView === 'audit'" class="fade-in">
          <div class="page-header">
            <h1 class="page-title">Audit Logs</h1>
            <div class="input-group"><i class="icon">S</i><input v-model="auditSearch" placeholder="Search audit logs..." /></div>
          </div>
          <div class="glass-card table-container">
            <table class="data-table">
              <thead><tr><th>Time</th><th>Actor</th><th>Action</th><th>Target</th><th>Severity</th><th>Details</th></tr></thead>
              <tbody>
                <tr v-for="log in filteredAuditLogs" :key="log.id">
                  <td>{{ formatDate(log.created_at) }}</td>
                  <td>{{ log.actor }}</td>
                  <td>{{ log.action }}</td>
                  <td>{{ log.target }}</td>
                  <td><span :class="['status-badge', log.severity]">{{ log.severity }}</span></td>
                  <td>{{ log.details }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </main>
    </template>
  </div>
</template>
