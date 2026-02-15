<template>
  <Transition name="fade">
    <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')">
      
      <div v-if="activeType === 'add'" class="modal-content">
        <h3>添加新賬號</h3>
        <div class="form-group">
          <label>別名</label>
          <input 
            :value="newAcc.alias" 
            @input="updateNewAcc('alias', $event.target.value)" 
            placeholder="如：大號" 
          />
        </div>
        <div class="form-group">
          <label>遊戲賬號</label>
          <input 
            :value="newAcc.username" 
            @input="updateNewAcc('username', $event.target.value)" 
            placeholder="手機號/郵箱" 
          />
        </div>
        <div class="form-group">
          <label>遊戲密碼</label>
          <input 
            :value="newAcc.password" 
            @input="updateNewAcc('password', $event.target.value)" 
            type="password" 
            placeholder="請輸入密碼" 
          />
        </div>
        <div class="modal-actions">
          <button @click="$emit('close')">取消</button>
          <button class="btn-primary" @click="$emit('confirmAdd')">確認添加</button>
        </div>
      </div>

      <div v-if="activeType === 'status'" class="modal-content status-modal">
        <div class="loader"></div>
        <div class="status-box">
          <p class="status-text">{{ statusTip }}</p>
        </div>
        <div class="modal-actions">
          <button class="btn-warning" @click="$emit('togglePause')">
            {{ pauseStatus === '已暫停' ? '繼續監控' : '暫停監控' }}
          </button>
          <button class="btn-danger" @click="$emit('stopMonitor')">停止並退出</button>
        </div>
      </div>

      <div v-if="activeType === 'path'" class="modal-content">
        <h3>遊戲執行文件路徑 (.exe)</h3>
        <div v-for="game in games" :key="game.id" class="form-group">
          <label>{{ game.name }}</label>
          <div class="path-input-group">
            <input 
              :value="gamePaths[game.id]" 
              @input="$emit('updatePath', { id: game.id, val: $event.target.value })" 
              placeholder="請選擇或填寫路徑..." 
            />
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
const props = defineProps({
  isOpen: Boolean,
  activeType: String, 
  newAcc: Object,
  statusTip: String,
  pauseStatus: String,
  games: Array,
  gamePaths: Object
})

const emit = defineEmits([
  'close', 
  'confirmAdd', 
  'togglePause', 
  'stopMonitor', 
  'updatePath', 
  'browse', 
  'savePaths', 
  'cancelStop', 
  'confirmStop'
])

// 輔助函數：解決 Props 只讀問題，手動通知父組件修改對象
const updateNewAcc = (key, value) => {
  // 原有逻辑保留：
  const updated = { ...props.newAcc, [key]: value };
  // 雖然你的 App.vue 目前是直接傳入響應式對象 newAcc，
  // 但在組件內直接修改對象屬性是不推薦的。這裡我們直接在組件內修改傳入的引用，
  // 或是讓 App.vue 使用 v-model:newAcc。
  // 為了兼容你目前的 App.vue 寫法，這裡直接賦值：
  props.newAcc[key] = value;
}
</script>

<style scoped>
/* 确保引用了基础样式 */
@import "./Modals.css";

/* 動畫效果 */
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.modal-desc {
  color: var(--text-dim);
  font-size: 14px;
  margin: 15px 0;
  line-height: 1.5;
}

/* 按钮居中微调 */
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.btn-warning { background: #f39c12; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
.btn-danger { background: #e74c3c; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
.btn-primary { background: var(--primary); color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
</style>