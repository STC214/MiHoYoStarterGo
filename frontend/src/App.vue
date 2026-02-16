<template>
  <div :class="['container', settings.theme]">
    <MenuBar
      :monitorActive="monitorActive"
      :pauseStatus="pauseStatus"
      @toggleTheme="toggleTheme"
      @handleExport="handleExport"
      @openStatus="openStatusModal"
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
        <div
          v-for="acc in filteredAccounts"
          :key="acc.id"
          class="account-draggable"
          :class="{ 'drag-over': dragState.overId === acc.id && dragState.sourceId !== acc.id }"
          draggable="true"
          @dragstart="onCardDragStart(acc, $event)"
          @dragover="onCardDragOver(acc, $event)"
          @drop="onCardDrop(acc, $event)"
          @dragend="onCardDragEnd"
        >
          <AccountCard
            :account="acc"
            @copy="copyToClipboard"
            @togglePass="togglePassword"
            @run="handleRunRequest"
            @edit="openEditModal"
            @delete="handleDelete"
          />
        </div>
      </div>
      <div v-else class="empty-state">暂无账号，请点击上方“添加账号”</div>
    </main>

    <Modals
      :isOpen="isAnyModalOpen"
      :activeType="modalType"
      :newAcc="newAcc"
      :statusTip="statusTip"
      :messageText="messageText"
      :pauseStatus="pauseStatus"
      :runContext="runContext"
      :games="games"
      :gamePaths="settings.game_paths"
      @close="closeModal"
      @confirmAdd="handleAdd"
      @confirmEdit="handleEdit"
      @runAction="handleRunAction"
      @captureDebug="handleCaptureDebug"
      @togglePause="togglePause"
      @stopMonitor="handleStopMonitor"
      @updatePath="updatePathValue"
      @browse="handleBrowse"
      @savePaths="handleSavePaths"
      @cancelStop="modalType = 'status'"
      @confirmStop="confirmStopMonitor"
      @confirmDelete="confirmDelete"
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
  SaveAccountsOrder,
  SelectGameFile,
  AddAccount,
  DeleteAccount,
  UpdateAccount,
  GetPlaintext,
  ExportBackup,
  CaptureDebugWindow,
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
const messageText = ref('')
const pauseStatus = ref('RUNNING')
const monitorActive = ref(false)
const selectedAccount = ref(null)
const pendingDeleteAccount = ref(null)
const editingAccountId = ref('')
const runContext = reactive({
  isRunning: false,
  alias: ''
})
const dragState = reactive({
  sourceId: '',
  overId: '',
  saving: false
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
    monitorActive.value = false
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
  editingAccountId.value = ''
  newAcc.alias = ''
  newAcc.username = ''
  newAcc.password = ''
  modalType.value = 'add'
}

const openEditModal = acc => {
  editingAccountId.value = acc.id
  newAcc.alias = acc.alias || ''
  newAcc.username = acc.username || ''
  newAcc.password = ''
  modalType.value = 'edit'
}

const handleAdd = async () => {
  const res = await AddAccount(newAcc.alias, newAcc.username, newAcc.password, activeTab.value)
  if (res === 'SUCCESS') {
    modalType.value = ''
    await loadAll()
  } else {
    showMessage('保存失败: ' + res)
  }
}

const handleEdit = async () => {
  if (!editingAccountId.value) {
    showMessage('编辑目标无效，请重试')
    return
  }
  const res = await UpdateAccount(editingAccountId.value, newAcc.alias, newAcc.username, newAcc.password)
  if (res === 'SUCCESS') {
    editingAccountId.value = ''
    modalType.value = ''
    await loadAll()
  } else {
    showMessage('修改失败: ' + res)
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
    showMessage('启动游戏失败，请检查路径和权限')
    return
  }
  if (res === 'GAME_NOT_RUNNING') {
    showMessage('目标游戏进程当前未运行，请改用“换号并启动游戏”或“手动启动游戏”')
    return
  }
  if (res === 'GAME_RUNNING') {
    showMessage('目标游戏进程已在运行，请改用“切换账号并自动重启”或“手动切换到登录界面”')
    return
  }
  if (res.startsWith('PATCH_FAILED:')) {
    showMessage('写入账号环境失败：' + res)
    return
  }
  if (res !== 'STARTED') {
    showMessage('执行失败：' + res)
    return
  }

  statusTip.value = '流程已启动，正在等待识别登录界面...'
  monitorActive.value = true
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
  monitorActive.value = false
  modalType.value = ''
}

const openStatusModal = () => {
  if (!monitorActive.value) {
    showMessage('当前无进行中的登录流程')
    return
  }
  modalType.value = 'status'
}

const closeModal = () => {
  pendingDeleteAccount.value = null
  editingAccountId.value = ''
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
  messageText.value = '游戏执行文件路径已保存'
  modalType.value = 'message'
}

const toggleTheme = async () => {
  settings.theme = settings.theme === 'theme-darcula' ? 'theme-monokai' : 'theme-darcula'
  await SaveTheme(settings.theme)
}

const showMessage = msg => {
  messageText.value = msg
  modalType.value = 'message'
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
  pendingDeleteAccount.value = acc
  messageText.value = `确定删除账号「${acc.alias}」吗？`
  modalType.value = 'deleteConfirm'
}

const confirmDelete = async () => {
  if (!pendingDeleteAccount.value) {
    modalType.value = ''
    return
  }
  await DeleteAccount(pendingDeleteAccount.value.id)
  pendingDeleteAccount.value = null
  modalType.value = ''
  await loadAll()
}

const handleExport = async () => {
  showMessage('备份已导出: ' + (await ExportBackup()))
}

const handleCaptureDebug = async () => {
  const gameID = selectedAccount.value?.game_id || activeTab.value
  const fileName = await CaptureDebugWindow(gameID)
  if (fileName === 'FAILED_CAPTURE') {
    showMessage('截图失败：未找到目标游戏窗口，请先打开对应游戏窗口')
    return
  }
  if (fileName === 'FAILED_WRITE') {
    showMessage('截图失败：无法写入项目根目录')
    return
  }
  showMessage('调试截图已保存到项目根目录：' + fileName)
}

const copyToClipboard = text => {
  navigator.clipboard.writeText(text)
}

const getAccGameID = acc => acc.game_id || acc.GameID

const reorderActiveTabAccounts = (sourceID, targetID) => {
  if (!sourceID || !targetID || sourceID === targetID) {
    return false
  }

  const source = settings.accounts.find(acc => acc.id === sourceID)
  const target = settings.accounts.find(acc => acc.id === targetID)
  if (!source || !target) {
    return false
  }
  if (getAccGameID(source) !== activeTab.value || getAccGameID(target) !== activeTab.value) {
    return false
  }

  const tabAccounts = settings.accounts.filter(acc => getAccGameID(acc) === activeTab.value)
  const from = tabAccounts.findIndex(acc => acc.id === sourceID)
  const to = tabAccounts.findIndex(acc => acc.id === targetID)
  if (from < 0 || to < 0 || from === to) {
    return false
  }

  const reordered = [...tabAccounts]
  const [moved] = reordered.splice(from, 1)
  reordered.splice(to, 0, moved)

  let p = 0
  settings.accounts = settings.accounts.map(acc => {
    if (getAccGameID(acc) !== activeTab.value) {
      return acc
    }
    const replacement = reordered[p]
    p += 1
    return replacement
  })
  return true
}

const saveAccountOrder = async () => {
  const order = settings.accounts.map(acc => acc.id)
  return SaveAccountsOrder(order)
}

const resetDragState = () => {
  dragState.sourceId = ''
  dragState.overId = ''
}

const onCardDragStart = (acc, event) => {
  dragState.sourceId = acc.id
  dragState.overId = acc.id
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', acc.id)
  }
}

const onCardDragOver = (acc, event) => {
  event.preventDefault()
  if (dragState.sourceId && dragState.sourceId !== acc.id) {
    dragState.overId = acc.id
  }
}

const onCardDrop = async (acc, event) => {
  event.preventDefault()
  if (!dragState.sourceId || dragState.saving) {
    resetDragState()
    return
  }

  const snapshot = [...settings.accounts]
  const moved = reorderActiveTabAccounts(dragState.sourceId, acc.id)
  resetDragState()
  if (!moved) {
    return
  }

  dragState.saving = true
  const res = await saveAccountOrder()
  dragState.saving = false
  if (res !== 'SUCCESS') {
    settings.accounts = snapshot
    showMessage('排序保存失败: ' + res)
  }
}

const onCardDragEnd = () => {
  resetDragState()
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
  justify-content: center;
  align-items: stretch;
  gap: 10px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  width: 100%;
  box-sizing: border-box;
  padding: 8px 12px 0;
  margin: 0 auto;
}

.tab {
  padding: 10px 14px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  font-size: 14px;
  text-align: center;
  box-sizing: border-box;
  flex: 1 1 0;
  min-width: 0;
  max-width: 280px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media (max-width: 860px) {
  .tab-bar {
    gap: 8px;
    padding: 8px 10px 0;
  }

  .tab {
    font-size: 13px;
    padding: 9px 10px;
    max-width: none;
  }
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
  scrollbar-width: thin;
  scrollbar-color: var(--primary) var(--bg-card);
}

.content-area::-webkit-scrollbar {
  width: 10px;
}

.content-area::-webkit-scrollbar-track {
  background: var(--bg-card);
  border-left: 1px solid var(--border);
}

.content-area::-webkit-scrollbar-thumb {
  background: var(--primary);
  border: 2px solid var(--bg-card);
  border-radius: 999px;
}

.content-area::-webkit-scrollbar-thumb:hover {
  filter: brightness(1.12);
}

.account-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 15px;
}

.account-draggable {
  cursor: grab;
}

.account-draggable.drag-over {
  outline: 2px dashed var(--primary);
  outline-offset: 4px;
  border-radius: 8px;
}

.empty-state {
  text-align: center;
  margin-top: 100px;
  color: var(--text-dim);
}
</style>

