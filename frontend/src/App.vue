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
      <div v-for="game in games" :key="game.id"
        :class="['tab', activeTab === game.id ? 'active' : '']"
        @click="activeTab = game.id">
        {{ game.name }}
      </div>
    </div>

    <main class="content-area">
      <div v-if="filteredAccounts.length > 0" class="account-grid">
        <AccountCard 
          v-for="acc in filteredAccounts" :key="acc.id" 
          :account="acc"
          @copy="copyToClipboard"
          @togglePass="togglePassword"
          @run="handleRunRequest"
          @delete="handleDelete"
        />
      </div>
      <div v-else class="empty-state">暫無賬號，請點擊上方「添加賬號」</div>
    </main>

    <Modals 
      :isOpen="isAnyModalOpen"
      :activeType="modalType"
      :statusTip="statusTip"
      :pauseStatus="pauseStatus"
      :games="games"
      :gamePaths="settings.game_paths"
      v-model:newAcc="newAcc"
      @close="closeModal"
      @confirmAdd="handleAdd"
      @togglePause="handleTogglePause"
      @execStart="execStartGame"
      @directMonitor="handleDirectMonitor"
      @execRestart="execRestartGame"
      @updatePath="updatePathValue"
      @browse="handleBrowse"
      @savePaths="handleSavePaths"
      @reqStopConfirm="modalType = 'stopConfirm'"
      @cancelStop="modalType = 'status'"
      @confirmStop="confirmStopMonitor"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { EventsOn } from '../wailsjs/runtime/runtime'
import * as AppSync from '../wailsjs/go/main/App'

import MenuBar from './components/MenuBar.vue'
import AccountCard from './components/AccountCard.vue'
import Modals from './components/Modals.vue'

const { 
  GetSettings, SaveTheme, AddAccount, GetPlaintext, ExportBackup, 
  ForceStartMonitor, TogglePause, StopMonitor, DeleteAccount, 
  SaveGamePaths, IsGameRunning, StartGameExecution, KillGameProcess, SelectGameFile 
} = AppSync

const settings = reactive({ theme: 'theme-darcula', accounts: [], game_paths: {} })
const activeTab = ref('GenshinCN')
const games = [{ id: 'GenshinCN', name: '原神' }, { id: 'StarRailCN', name: '星穹鐵道' }, { id: 'ZZZCN', name: '絕區零' }]
const modalType = ref('') 
const showStatusModal = computed(() => modalType.value === 'status' || modalType.value === 'stopConfirm')
const isAnyModalOpen = computed(() => modalType.value !== '')

const pauseStatus = ref('運行中')
const statusTip = ref('正在等待識別遊戲畫面...')
const pendingAcc = ref(null)
const newAcc = reactive({ alias: '', username: '', password: '' })

const filteredAccounts = computed(() => settings.accounts?.filter(a => a.game_id === activeTab.value) || [])

const loadAll = async () => {
  const cfg = await GetSettings()
  settings.theme = cfg.theme
  settings.accounts = cfg.accounts || []
  settings.game_paths = cfg.game_paths || { GenshinCN: '', StarRailCN: '', ZZZCN: '' }
}

const handleRunRequest = async (acc) => {
  pendingAcc.value = acc
  const isRunning = await IsGameRunning(acc.game_id)
  modalType.value = isRunning ? 'conflict' : 'launch'
}

const handleDirectMonitor = async () => {
  statusTip.value = "正在接管當前遊戲窗口..."
  modalType.value = 'status'
  if (await ForceStartMonitor(pendingAcc.value) !== 'SUCCESS') {
    modalType.value = ''
  }
}

const execStartGame = async () => {
  const path = settings.game_paths[pendingAcc.value.game_id]
  if (!path) return alert("請先設置遊戲路徑")
  statusTip.value = "正在啟動遊戲..."
  modalType.value = 'status'
  if (await StartGameExecution(pendingAcc.value.game_id) === 'SUCCESS') {
    await ForceStartMonitor(pendingAcc.value)
  }
}

const execRestartGame = async () => {
  statusTip.value = "正在重啟遊戲..."
  modalType.value = 'status'
  if (await KillGameProcess(pendingAcc.value.game_id) === 'SUCCESS') {
    await execStartGame()
  }
}

const handleTogglePause = async () => { pauseStatus.value = await TogglePause() }
const confirmStopMonitor = async () => { await StopMonitor(); modalType.value = ''; pendingAcc.value = null }
const closeModal = () => { modalType.value = '' }

const handleAdd = async () => {
  if (await AddAccount(newAcc.alias, newAcc.username, newAcc.password, activeTab.value) === 'SUCCESS') {
    modalType.value = ''; await loadAll()
  }
}

const updatePathValue = ({id, val}) => { settings.game_paths[id] = val }
const handleBrowse = async (id) => {
  const path = await SelectGameFile()
  if (path) settings.game_paths[id] = path
}

const handleSavePaths = async () => { await SaveGamePaths(settings.game_paths); modalType.value = ''; alert("已保存") }
const toggleTheme = async () => {
  settings.theme = settings.theme === 'theme-darcula' ? 'theme-monokai' : 'theme-darcula'
  await SaveTheme(settings.theme)
}

const togglePassword = async (acc) => {
  if (!acc.showPlain) { acc.plainText = await GetPlaintext(acc.password); acc.showPlain = true }
  else acc.showPlain = false
}

const handleDelete = async (acc) => {
  if (confirm(`確定刪除 ${acc.alias} 嗎？`)) {
    await DeleteAccount(acc.id); await loadAll()
  }
}

const handleExport = async () => alert(await ExportBackup())
const copyToClipboard = (t) => navigator.clipboard.writeText(t)
const openAddModal = () => { newAcc.alias = ''; newAcc.username = ''; newAcc.password = ''; modalType.value = 'add' }

onMounted(() => {
  loadAll()
  EventsOn("monitor_finished", () => { modalType.value = ''; loadAll() })
})
</script>