<template>
  <Transition name="fade">
    <div v-if="isOpen" class="modal-overlay">
      
      <div v-if="activeType === 'add'" class="modal-content">
        <h3>添加新賬號</h3>
        <div class="form-group">
          <label>別名</label>
          <input :value="newAcc.alias" @input="$emit('update:newAcc', { ...newAcc, alias: $event.target.value })" placeholder="如：大號" />
        </div>
        <div class="form-group">
          <label>遊戲賬號</label>
          <input :value="newAcc.username" @input="$emit('update:newAcc', { ...newAcc, username: $event.target.value })" placeholder="手機號/郵箱" />
        </div>
        <div class="form-group">
          <label>遊戲密碼</label>
          <input :value="newAcc.password" @input="$emit('update:newAcc', { ...newAcc, password: $event.target.value })" type="password" placeholder="請輸入密碼" />
        </div>
        <div class="modal-actions">
          <button @click="$emit('close')">取消</button>
          <button class="btn-primary" @click="$emit('confirmAdd')">確認添加</button>
        </div>
      </div>

      <div v-if="activeType === 'status'" class="modal-content status-modal">
        <div class="loader"></div>
        <h3>自動化監控中</h3>
        <p class="status-tip">{{ statusTip }}</p>
        <div class="status-box">
          當前狀態：<span :class="pauseStatus === '運行中' ? 'text-green' : 'text-red'">{{ pauseStatus }}</span>
        </div>
        <div class="modal-actions full-width">
          <button class="btn-secondary" @click="$emit('togglePause')">
            {{ pauseStatus === '運行中' ? '⏸️ 暫停' : '▶️ 繼續' }}
          </button>
          <button class="btn-danger" @click="$emit('reqStopConfirm')">🛑 停止並取消監控</button>
          <button @click="$emit('close')">隱藏視窗</button>
        </div>
      </div>

      <div v-if="activeType === 'launch'" class="modal-content">
        <h3>🎮 遊戲尚未啟動</h3>
        <p class="modal-desc">檢測到遊戲進程未運行，請選擇操作：</p>
        <div class="modal-actions full-width">
          <button class="btn-primary" @click="$emit('execStart')">🚀 啟動遊戲並自動登錄</button>
          <button class="btn-secondary" @click="$emit('directMonitor')">⏳ 僅開啟監控(手動啟動)</button>
          <button @click="$emit('close')">取消</button>
        </div>
      </div>

      <div v-if="activeType === 'conflict'" class="modal-content">
        <h3 class="text-orange">⚠️ 檢測到遊戲正在運行</h3>
        <p class="modal-desc">您可以接管當前窗口，或重啟執行全新登錄：</p>
        <div class="modal-actions full-width">
          <button class="btn-primary" @click="$emit('directMonitor')">🎯 直接開始監控</button>
          <button class="btn-danger" @click="$emit('execRestart')">🔄 關閉並重新啟動遊戲</button>
          <button @click="$emit('close')">取消</button>
        </div>
      </div>

      <div v-if="activeType === 'path'" class="modal-content path-modal">
        <h3>⚙️ 遊戲路徑設置</h3>
        <div v-for="game in games" :key="game.id" class="form-group">
          <label>{{ game.name }} (.exe 絕對路徑)</label>
          <div class="path-input-group">
            <input :value="gamePaths[game.id]" @input="$emit('updatePath', { id: game.id, val: $event.target.value })" placeholder="請選擇或填寫路徑..." />
            <button class="btn-browse" @click="$emit('browse', game.id)">瀏覽</button>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="$emit('close')">取消</button>
          <button class="btn-primary" @click="$emit('savePaths')">保存路徑</button>
        </div>
      </div>

      <div v-if="activeType === 'stopConfirm'" class="modal-content">
        <h3>確認停止？</h3>
        <p class="modal-desc">確定要停止監控嗎？這將釋放 OCR 資源並停止自動化流程。</p>
        <div class="modal-actions">
          <button @click="$emit('cancelStop')">繼續監控</button>
          <button class="btn-danger" @click="$emit('confirmStop')">確認停止</button>
        </div>
      </div>

    </div>
  </Transition>
</template>

<script setup>
defineProps({
  isOpen: Boolean,
  activeType: String, 
  newAcc: Object,
  statusTip: String,
  pauseStatus: String,
  games: Array,
  gamePaths: Object
})

defineEmits([
  'close', 'confirmAdd', 'togglePause', 'reqStopConfirm', 
  'execStart', 'directMonitor', 'execRestart', 
  'updatePath', 'browse', 'savePaths', 'cancelStop', 'confirmStop',
  'update:newAcc'
])
</script>

<style scoped>
@import "./Modals.css";
/* 補充組件內必要布局 */
.modal-desc { font-size: 14px; color: var(--text-dim); margin: 10px 0 20px; line-height: 1.5; }
.text-orange { color: #ff9800; }
</style>