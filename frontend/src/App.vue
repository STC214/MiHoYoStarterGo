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
      :newAcc="newAcc"
      :statusTip="statusTip"
      :pauseStatus="pauseStatus"
      :games="games"
      :gamePaths="settings.game_paths"
      @close="closeModal"
      @confirmAdd="handleAdd"
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

// 導入 Wails 自動生成的 JS
import { 
  GetSettings, SaveTheme, SaveGamePaths, SelectGameFile, 
  AddAccount, DeleteAccount, GetPlaintext, ExportBackup,
  PrepareAccountEnvironment, StartGame, StartMonitor, 
  StopMonitor, TogglePauseMonitor, GetMonitorStatus 
} from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime'

// 狀態定義
const games = [
  { id: 'GenshinCN', name: '原神 (國服)' },
  { id: 'StarRailCN', name: '崩壞：星穹鐵道' },
  { id: 'ZZZCN', name: '絕區零' }
]
const activeTab = ref('GenshinCN')
const modalType = ref('')
const statusTip = ref('正在啟動...')
const pauseStatus = ref('運行中')
const showStatusModal = ref(false)

const settings = reactive({
  theme: 'theme-darcula',
  accounts: [],
  game_paths: {}
})

const newAcc = reactive({ alias: '', username: '', password: '' })

// 計算屬性
// 修复：确保过滤时同时匹配 game_id 或 GameID (兼容 Go 导出)
const filteredAccounts = computed(() => {
  return settings.accounts.filter(acc => (acc.game_id === activeTab.value || acc.GameID === activeTab.value))
})
const isAnyModalOpen = computed(() => modalType.value !== '')

// 邏輯處理
onMounted(async () => {
  await loadAll()
  EventsOn("monitor_status", (tip) => { statusTip.value = tip })
})

const loadAll = async () => {
  const data = await GetSettings()
  settings.theme = data.theme
  // 修复：将 Go 的大写字段映射到 JS 的小写字段，确保 UI 显示正常
  settings.accounts = (data.accounts || []).map(acc => ({ 
    ...acc, 
    game_id: acc.game_id || acc.GameID,
    alias: acc.alias || acc.Alias,
    username: acc.username || acc.Username,
    showPlain: false, 
    plainText: '' 
  }))
  settings.game_paths = data.game_paths || {}
}

const openAddModal = () => {
  newAcc.alias = ''; newAcc.username = ''; newAcc.password = ''; modalType.value = 'add'
}

const handleAdd = async () => {
  const res = await AddAccount(newAcc.alias, newAcc.username, newAcc.password, activeTab.value)
  if (res === 'SUCCESS') { 
    modalType.value = ''; 
    await loadAll() 
  } else {
    alert("保存失敗: " + res)
  }
}

const handleRunRequest = async (acc) => {
  const patchRes = await PrepareAccountEnvironment(acc)
  if (patchRes === 'GAME_RUNNING') return alert("遊戲正在運行，請先關閉")
  
  const startRes = await StartGame(acc.game_id)
  if (startRes === 'PATH_NOT_FOUND') { modalType.value = 'path'; return }

  if (acc.is_first_login || !acc.token) {
    modalType.value = 'status'; showStatusModal.value = true
    await StartMonitor(acc)
  }
}

const togglePause = async () => {
  await TogglePauseMonitor()
  pauseStatus.value = await GetMonitorStatus()
}

const handleStopMonitor = () => { modalType.value = 'stopConfirm' }
const confirmStopMonitor = async () => {
  await StopMonitor(); modalType.value = ''; showStatusModal.value = false
}

const closeModal = () => { modalType.value = '' }
const updatePathValue = ({id, val}) => { settings.game_paths[id] = val }
const handleBrowse = async (id) => {
  const path = await SelectGameFile()
  if (path) settings.game_paths[id] = path
}

const handleSavePaths = async () => {
  await SaveGamePaths(settings.game_paths); modalType.value = ''; alert("已保存")
}

const toggleTheme = async () => {
  settings.theme = settings.theme === 'theme-darcula' ? 'theme-monokai' : 'theme-darcula'
  await SaveTheme(settings.theme)
}

const togglePassword = async (acc) => {
  if (!acc.showPlain) {
    acc.plainText = await GetPlaintext(acc.password); acc.showPlain = true
  } else { acc.showPlain = false }
}

const handleDelete = async (acc) => {
  if (confirm(`確定刪除 ${acc.alias} 嗎？`)) {
    await DeleteAccount(acc.id); await loadAll()
  }
}

const handleExport = async () => { alert("備份已導出: " + await ExportBackup()) }
const copyToClipboard = (text) => { navigator.clipboard.writeText(text) }
</script>

<style>
/* 全局基礎佈局與 CSS 變量 */
:root {
  --bg-app: #1e1e1e; --bg-card: #252526; --primary: #007acc; --border: #3c3c3c; --text: #cccccc; --text-dim: #888888;
}
.theme-monokai {
  --bg-app: #272822; --bg-card: #3e3d32; --primary: #a6e22e; --border: #49483e; --text: #f8f8f2; --text-dim: #75715e;
}

body { margin: 0; font-family: 'Segoe UI', sans-serif; background: var(--bg-app); color: var(--text); overflow: hidden; }
.container { width: 100vw; height: 100vh; display: flex; flex-direction: column; }

.tab-bar { display: flex; background: var(--bg-card); border-bottom: 1px solid var(--border); }
.tab { padding: 10px 20px; cursor: pointer; border-bottom: 2px solid transparent; font-size: 14px; }
.tab.active { border-bottom-color: var(--primary); color: var(--primary); font-weight: bold; }

.content-area { flex: 1; padding: 20px; overflow-y: auto; }
.account-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 15px; }
.empty-state { text-align: center; margin-top: 100px; color: var(--text-dim); }

/* 修复：弹窗居中核心样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: var(--bg-card);
  padding: 25px;
  border-radius: 8px;
  min-width: 350px;
  max-width: 90%;
  border: 1px solid var(--border);
  box-shadow: 0 10px 30px rgba(0,0,0,0.5);
}

.text-green { color: #4caf50; }
.text-red { color: #f44336; }
</style>