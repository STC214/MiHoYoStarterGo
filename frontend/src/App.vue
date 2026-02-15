<template>
  <div :class="['container', settings.theme]">
    <MenuBar
      :showStatus="showStatusModal"
      :pauseStatus="pauseStatus"
      @toggleTheme="toggleTheme"
      @handleExport="handleExport"
      @openPath="modalType = 'path'"
      @openAdd="openAddModal"
    />

    <div class="tab-bar">
      <div
        v-for="game in games"
        :key="game.id"
        :class="['tab', activeTab === game.id ? 'active' : '']"
        @click="activeTab = game.id"
      >
        {{ game.name }}
      </div>
    </div>

    <main class="content-area">
      <div v-if="filteredAccounts.length > 0" class="account-grid">
        <AccountCard
          v-for="acc in filteredAccounts"
          :key="acc.id"
          :account="acc"
          @copy="copyToClipboard"
          @togglePass="togglePassword"
          @run="handleRunRequest"
          @delete="handleDelete"
        />
      </div>
      <div v-else class="empty-state">暂无账号，请点击上方“添加账号”</div>
    </main>

    <Modals
      :isOpen="isAnyModalOpen"
      :activeType="modalType"
      :newAcc="newAcc"
      :statusTip="statusTip"
      :pauseStatus="pauseStatus"
      :runContext="runContext"
      :games="games"
      :gamePaths="settings.game_paths"
      @close="closeModal"
      @confirmAdd="handleAdd"
      @runAction="handleRunAction"
      @togglePause="togglePause"
      @stopMonitor="handleStopMonitor"
      @updatePath="updatePathValue"
      @browse="handleBrowse"
      @savePaths="handleSavePaths"
      @cancelStop="modalType = 'status'"
      @confirmStop="confirmStopMonitor"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import MenuBar from './components/MenuBar.vue'
import AccountCard from './components/AccountCard.vue'
import Modals from './components/Modals.vue'

import {
  GetSettings,
  SaveTheme,
  SaveGamePaths,
  SelectGameFile,
  AddAccount,
  DeleteAccount,
  GetPlaintext,
  ExportBackup,
  ExecuteLoginAction,
  IsGameRunning,
  StopMonitor,
  TogglePauseMonitor,
  GetMonitorStatus
} from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime'

const games = [
  { id: 'GenshinCN', name: '原神 (国服)' },
  { id: 'StarRailCN', name: '崩坏：星穹铁道' },
  { id: 'ZZZCN', name: '绝区零' }
]

const activeTab = ref('GenshinCN')
const modalType = ref('')
const statusTip = ref('正在启动...')
const pauseStatus = ref('RUNNING')
const showStatusModal = ref(false)
const selectedAccount = ref(null)
const runContext = reactive({
  isRunning: false,
  alias: ''
})

const settings = reactive({
  theme: 'theme-darcula',
  accounts: [],
  game_paths: {}
})

const newAcc = reactive({ alias: '', username: '', password: '' })

const filteredAccounts = computed(() => {
  return settings.accounts.filter(acc => (acc.game_id === activeTab.value || acc.GameID === activeTab.value))
})

const isAnyModalOpen = computed(() => modalType.value !== '')

onMounted(async () => {
  await loadAll()
  pauseStatus.value = await GetMonitorStatus()

  EventsOn('monitor_status', tip => {
    statusTip.value = tip
  })

  EventsOn('monitor_finished', () => {
    showStatusModal.value = false
    modalType.value = ''
    pauseStatus.value = 'RUNNING'
    loadAll()
  })
})

const loadAll = async () => {
  const data = await GetSettings()
  settings.theme = data.theme || 'theme-darcula'
  settings.accounts = (data.accounts || []).map(acc => ({
    ...acc,
    id: acc.id || acc.ID,
    game_id: acc.game_id || acc.GameID,
    alias: acc.alias || acc.Alias,
    username: acc.username || acc.Username,
    password: acc.password || acc.Password,
    token: acc.token || acc.Token,
    device_fingerprint: acc.device_fingerprint || acc.DeviceFingerprint,
    is_first_login: acc.is_first_login ?? acc.IsFirstLogin,
    showPlain: false,
    plainText: ''
  }))
  settings.game_paths = data.game_paths || {}
}

const openAddModal = () => {
  newAcc.alias = ''
  newAcc.username = ''
  newAcc.password = ''
  modalType.value = 'add'
}

const handleAdd = async () => {
  const res = await AddAccount(newAcc.alias, newAcc.username, newAcc.password, activeTab.value)
  if (res === 'SUCCESS') {
    modalType.value = ''
    await loadAll()
  } else {
    alert('保存失败: ' + res)
  }
}

const handleRunRequest = async acc => {
  selectedAccount.value = acc
  runContext.alias = acc.alias
  runContext.isRunning = await IsGameRunning(acc.game_id)
  modalType.value = 'runAction'
}

const handleRunAction = async action => {
  if (!selectedAccount.value || action === 'cancel') {
    modalType.value = ''
    return
  }

  const res = await ExecuteLoginAction(selectedAccount.value, action)
  if (res === 'PATH_NOT_FOUND') {
    modalType.value = 'path'
    return
  }
  if (res === 'START_FAILED') {
    alert('启动游戏失败，请检查路径和权限')
    return
  }
  if (res === 'GAME_NOT_RUNNING') {
    alert('目标游戏进程当前未运行，请改用“换号并启动游戏”或“手动启动游戏”')
    return
  }
  if (res === 'GAME_RUNNING') {
    alert('目标游戏进程已在运行，请改用“切换账号并自动重启”或“手动切换到登录界面”')
    return
  }
  if (res.startsWith('PATCH_FAILED:')) {
    alert('写入账号环境失败：' + res)
    return
  }
  if (res !== 'STARTED') {
    alert('执行失败：' + res)
    return
  }

  statusTip.value = '流程已启动，正在等待识别登录界面...'
  showStatusModal.value = true
  modalType.value = 'status'
}

const togglePause = async () => {
  await TogglePauseMonitor()
  pauseStatus.value = await GetMonitorStatus()
}

const handleStopMonitor = () => {
  modalType.value = 'stopConfirm'
}

const confirmStopMonitor = async () => {
  await StopMonitor()
  modalType.value = ''
  showStatusModal.value = false
}

const closeModal = () => {
  modalType.value = ''
}

const updatePathValue = ({ id, val }) => {
  settings.game_paths[id] = val
}

const handleBrowse = async id => {
  const path = await SelectGameFile()
  if (path) settings.game_paths[id] = path
}

const handleSavePaths = async () => {
  await SaveGamePaths(settings.game_paths)
  modalType.value = ''
  alert('已保存')
}

const toggleTheme = async () => {
  settings.theme = settings.theme === 'theme-darcula' ? 'theme-monokai' : 'theme-darcula'
  await SaveTheme(settings.theme)
}

const togglePassword = async acc => {
  if (!acc.showPlain) {
    acc.plainText = await GetPlaintext(acc.password)
    acc.showPlain = true
  } else {
    acc.showPlain = false
  }
}

const handleDelete = async acc => {
  if (confirm(`确定删除 ${acc.alias} 吗？`)) {
    await DeleteAccount(acc.id)
    await loadAll()
  }
}

const handleExport = async () => {
  alert('备份已导出: ' + (await ExportBackup()))
}

const copyToClipboard = text => {
  navigator.clipboard.writeText(text)
}
</script>

<style>
:root {
  --bg-app: #1e1e1e;
  --bg-card: #252526;
  --primary: #007acc;
  --border: #3c3c3c;
  --text: #cccccc;
  --text-dim: #888888;
}

.theme-monokai {
  --bg-app: #272822;
  --bg-card: #3e3d32;
  --primary: #a6e22e;
  --border: #49483e;
  --text: #f8f8f2;
  --text-dim: #75715e;
}

body {
  margin: 0;
  font-family: 'Segoe UI', sans-serif;
  background: var(--bg-app);
  color: var(--text);
  overflow: hidden;
}

.container {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.tab-bar {
  display: flex;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
}

.tab {
  padding: 10px 20px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  font-size: 14px;
}

.tab.active {
  border-bottom-color: var(--primary);
  color: var(--primary);
  font-weight: bold;
}

.content-area {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.account-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 15px;
}

.empty-state {
  text-align: center;
  margin-top: 100px;
  color: var(--text-dim);
}
</style>
