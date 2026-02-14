<template>
  <div :class="['container', settings.theme]">
    <header class="title-bar" style="--wails-draggable:drag">
      <span>米哈游启动器增强版 (Go + Vue3)</span>
      <div v-if="showStatusModal" class="status-indicator-mini">
        监控中: <span :class="pauseStatus === '运行中' ? 'text-green' : 'text-red'">{{ pauseStatus }}</span>
      </div>
    </header>

    <nav class="menu-bar">
      <div class="menu-item" @click="toggleTheme">🎨 切换主题</div>
      <div class="menu-item" @click="handleExport">📥 导出备份</div>
      <div class="menu-item" @click="openAddModal">➕ 添加账号</div>
    </nav>

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
      <div class="account-grid">
        <div v-for="acc in filteredAccounts" :key="acc.id" class="account-card">
          <div class="info">
            <div class="alias">{{ acc.alias }}</div>
            <div class="details" @click="copyToClipboard(acc.username)">
              👤 {{ acc.username }}
            </div>
            <div class="details" @click="togglePassword(acc)">
              🔑 {{ acc.showPlain ? acc.plainText : '••••••••' }}
            </div>
            <div class="token-tag" :class="acc.token ? 'has-token' : 'no-token'">
              {{ acc.token ? '✅ 已保存 Token' : '❌ 待首次登录' }}
            </div>
          </div>
          
          <div class="card-actions">
            <button class="btn-primary" @click="runSwitch(acc)">切换并登录</button>
            <button class="btn-delete-mini" @click="handleDelete(acc)" title="删除账号">🗑️</button>
          </div>
        </div>
      </div>
    </main>

    <div v-if="showAddModal" class="modal-overlay">
      <div class="modal-content">
        <h3>添加新账号</h3>
        <div class="form-group">
          <label>别名</label>
          <input v-model="newAcc.alias" placeholder="如：大号" />
        </div>
        <div class="form-group">
          <label>游戏账号</label>
          <input v-model="newAcc.username" placeholder="手机号/邮箱" />
        </div>
        <div class="form-group">
          <label>游戏密码</label>
          <input v-model="newAcc.password" type="password" placeholder="请输入密码" />
        </div>
        <div class="modal-actions">
          <button @click="showAddModal = false">取消</button>
          <button class="btn-primary" @click="handleAdd">确认添加</button>
        </div>
      </div>
    </div>

    <div v-if="showConflictModal" class="modal-overlay">
      <div class="modal-content conflict-modal">
        <h3>⚠️ 检测到游戏正在运行</h3>
        <p class="status-tip">若游戏已在登录界面，可点击直接监控。</p>
        <div class="modal-actions full-width">
          <button class="btn-primary" @click="handleDirectMonitor">
            直接开始监控
          </button>
          <button class="btn-secondary" @click="showConflictModal = false">
            取消
          </button>
        </div>
      </div>
    </div>

    <div v-if="showStatusModal" class="modal-overlay">
      <div class="modal-content status-modal">
        <div class="loader"></div>
        <h3>自动化监控中</h3>
        <p class="status-tip">正在等待识别游戏画面...</p>
        <div class="status-box">
          当前状态：<span :class="pauseStatus === '运行中' ? 'text-green' : 'text-red'">{{ pauseStatus }}</span>
        </div>
        <div class="modal-actions full-width">
          <button class="btn-secondary" @click="handleTogglePause">
            {{ pauseStatus === '运行中' ? '⏸️ 暂停' : '▶️ 继续' }}
          </button>
          
          <button class="btn-danger" @click="handleStopMonitor">
            🛑 停止并取消监控
          </button>
          
          <button @click="showStatusModal = false">隐藏视窗</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import * as AppSync from '../wailsjs/go/main/App' 
import { EventsOn } from '../wailsjs/runtime/runtime'

// 从模块中解构方法，新增 DeleteAccount
const { 
  GetSettings, SaveTheme, AddAccount, GetPlaintext, 
  ExportBackup, RequestSwitch, ForceStartMonitor, TogglePause,
  StopMonitor = () => { console.warn("StopMonitor 方法未在后端定义") },
  DeleteAccount = () => { console.warn("DeleteAccount 方法未在后端定义") }
} = AppSync

const settings = reactive({ theme: 'theme-darcula', accounts: [] })
const activeTab = ref('GenshinCN')
const games = [
  { id: 'GenshinCN', name: '原神' },
  { id: 'StarRailCN', name: '星穹铁道' },
  { id: 'ZZZCN', name: '绝区零' }
]

const showAddModal = ref(false)
const showStatusModal = ref(false)
const showConflictModal = ref(false)
const pauseStatus = ref('运行中')
const pendingAcc = ref(null)

const newAcc = reactive({ alias: '', username: '', password: '' })

const filteredAccounts = computed(() => {
  // 对应后端逻辑：根据 game_id 分类显示
  return settings.accounts ? settings.accounts.filter(a => a.game_id === activeTab.value) : []
})

const loadAll = async () => {
  try {
    const cfg = await GetSettings()
    settings.theme = cfg.theme
    settings.accounts = cfg.accounts || []
  } catch (e) {
    console.error("无法加载设置:", e)
  }
}

const toggleTheme = async () => {
  const next = settings.theme === 'theme-darcula' ? 'theme-monokai' : 'theme-darcula'
  await SaveTheme(next)
  settings.theme = next
}

const runSwitch = async (acc) => {
  const res = await RequestSwitch(acc)
  if (res === 'RUNNING_CONFLICT') {
    pendingAcc.value = acc
    showConflictModal.value = true
  } else if (res === 'SUCCESS') {
    pauseStatus.value = '运行中'
    showStatusModal.value = true
  }
}

const handleDirectMonitor = async () => {
  showConflictModal.value = false
  const res = await ForceStartMonitor(pendingAcc.value)
  if (res === 'SUCCESS') {
    showStatusModal.value = true
  }
}

const handleTogglePause = async () => {
  pauseStatus.value = await TogglePause()
}

const handleStopMonitor = async () => {
  if (confirm("确定要停止监控吗？这将释放 OCR 资源并停止自动化流程。")) {
    await StopMonitor()
    showStatusModal.value = false
    pendingAcc.value = null
  }
}

// 新增：处理账号删除逻辑
const handleDelete = async (acc) => {
  if (confirm(`确定要删除账号 [${acc.alias}] 吗？此操作不可恢复。`)) {
    const res = await DeleteAccount(acc.id)
    if (res === 'SUCCESS') {
      await loadAll() // 删除成功后刷新列表
    } else {
      alert("删除失败: " + res)
    }
  }
}

const handleAdd = async () => {
  const res = await AddAccount(newAcc.alias, newAcc.username, newAcc.password, activeTab.value)
  if (res === 'SUCCESS') {
    showAddModal.value = false
    loadAll()
  }
}

const togglePassword = async (acc) => {
  if (!acc.showPlain) {
    acc.plainText = await GetPlaintext(acc.password)
    acc.showPlain = true
  } else {
    acc.showPlain = false
  }
}

const handleExport = async () => {
  const res = await ExportBackup()
  alert(res)
}

const copyToClipboard = (t) => { navigator.clipboard.writeText(t) }
const openAddModal = () => {
  newAcc.alias = ''; newAcc.username = ''; newAcc.password = '';
  showAddModal.value = true;
}

onMounted(() => {
  loadAll()
  EventsOn("monitor_finished", (status) => {
    showStatusModal.value = false
    pendingAcc.value = null
    loadAll()
  })
})
</script>

<style scoped>
.container { width: 100vw; height: 100vh; display: flex; flex-direction: column; background: var(--bg-app); color: var(--text); overflow: hidden; }
.theme-darcula { --bg-app: #1e1e1e; --bg-card: #2b2b2b; --border: #3c3f41; --text: #a9b7c6; --text-dim: #888; --primary: #4caf50; }
.theme-monokai { --bg-app: #272822; --bg-card: #3e3d32; --border: #49483e; --text: #f8f8f2; --text-dim: #75715e; --primary: #a6e22e; }

.title-bar { height: 40px; background: var(--bg-card); display: flex; align-items: center; padding: 0 15px; border-bottom: 1px solid var(--border); }
.status-indicator-mini { font-size: 11px; margin-left: auto; background: rgba(0,0,0,0.3); padding: 2px 8px; border-radius: 10px; }

.menu-bar { display: flex; background: var(--bg-card); padding: 5px 10px; gap: 10px; }
.menu-item { padding: 5px 12px; border-radius: 4px; cursor: pointer; font-size: 13px; }
.menu-item:hover { background: var(--bg-app); }

.tab-bar { display: flex; border-bottom: 1px solid var(--border); background: var(--bg-card); }
.tab { padding: 10px 25px; cursor: pointer; font-size: 14px; }
.tab.active { border-bottom: 2px solid var(--primary); color: var(--primary); }

.content-area { flex: 1; padding: 20px; overflow-y: auto; }
.account-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 15px; }
.account-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 8px; padding: 15px; display: flex; justify-content: space-between; align-items: center; }

.card-actions { display: flex; flex-direction: column; gap: 8px; align-items: flex-end; }
.btn-delete-mini { background: transparent; border: 1px solid #444; color: #777; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; transition: all 0.2s; }
.btn-delete-mini:hover { border-color: #f44336; color: #f44336; background: rgba(244, 67, 54, 0.1); }

.token-tag { font-size: 10px; padding: 2px 6px; border-radius: 4px; margin-top: 5px; display: inline-block; }
.has-token { background: rgba(76, 175, 80, 0.15); color: #4caf50; }
.no-token { background: rgba(244, 67, 54, 0.15); color: #f44336; }

.modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.7); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-content { background: var(--bg-card); padding: 25px; border-radius: 10px; width: 380px; border: 1px solid var(--border); }
.status-modal { text-align: center; }
.status-box { background: var(--bg-app); padding: 12px; border-radius: 6px; margin: 15px 0; border: 1px dashed var(--border); }

.loader { border: 3px solid var(--border); border-top: 3px solid var(--primary); border-radius: 50%; width: 30px; height: 30px; animation: spin 1s linear infinite; margin: 0 auto 10px; }
@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }

button { padding: 8px 16px; border-radius: 4px; border: none; cursor: pointer; font-size: 13px; }
.btn-primary { background: var(--primary); color: #000; font-weight: bold; }
.btn-secondary { background: #555; color: #fff; }
.btn-danger { background: #d32f2f; color: white; font-weight: bold; }

.text-green { color: #4caf50; }
.text-red { color: #f44336; }

.form-group { margin-bottom: 15px; text-align: left; display: flex; flex-direction: column; }
.form-group label { font-size: 12px; margin-bottom: 5px; color: var(--text-dim); }
.form-group input { padding: 10px; background: var(--bg-app); border: 1px solid var(--border); color: var(--text); border-radius: 4px; }
.modal-actions { margin-top: 20px; display: flex; justify-content: flex-end; gap: 10px; }
</style>