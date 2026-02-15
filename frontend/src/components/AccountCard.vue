<template>
  <div class="account-card">
    <div class="info">
      <div class="alias">{{ account.alias }}</div>
      <div class="details" @click="$emit('copy', account.username)">
        👤 {{ account.username }}
      </div>
      <div class="details" @click="$emit('togglePass', account)">
        🔑 {{ account.showPlain ? account.plainText : '••••••••' }}
      </div>
      <div class="token-tag" :class="account.token ? 'has-token' : 'no-token'">
        {{ account.token ? '✅ 已保存 Token' : '❌ 待首次登錄' }}
      </div>
    </div>
    
    <div class="card-actions">
      <button class="btn-primary" @click="$emit('run', account)">切換並登錄</button>
      <button class="btn-delete-mini" @click="$emit('delete', account)" title="刪除賬號">
        <span class="icon">🗑️</span>
      </button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  account: Object
})
defineEmits(['copy', 'togglePass', 'run', 'delete'])
</script>

<style scoped>
.account-card { 
  background: var(--bg-card); 
  border: 1px solid var(--border); 
  border-radius: 8px; 
  padding: 15px; 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
}

.info .alias { font-weight: bold; margin-bottom: 5px; }
.info .details { font-size: 12px; color: var(--text-dim); cursor: pointer; margin-top: 3px; }

.token-tag { font-size: 10px; padding: 2px 6px; border-radius: 4px; margin-top: 5px; display: inline-block; }
.has-token { background: rgba(76, 175, 80, 0.15); color: #4caf50; }
.no-token { background: rgba(244, 67, 54, 0.15); color: #f44336; }

/* 佈局修正 */
.card-actions {
  display: flex;
  align-items: stretch; /* 關鍵：讓子元素在交叉軸（高度）上自動拉伸填滿 */
  gap: 10px; /* 替代 margin-right，更現代的寫法 */
}

.btn-primary { 
  background: var(--primary); 
  color: #ffffff; 
  font-weight: bold; 
  padding: 8px 16px; 
  border-radius: 4px; 
  border: 1px solid var(--border); 
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s ease;
  line-height: 1.2; /* 固定行高確保基準一致 */
}

.btn-primary:hover {
  opacity: 0.9;
  border-color: #ffffff; 
}

.btn-delete-mini { 
  background: transparent; 
  border: 1px solid var(--border); 
  color: #777; 
  padding: 0 10px; /* 橫向內邊距，高度由 stretch 決定 */
  border-radius: 4px; 
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.btn-delete-mini .icon {
  font-size: 14px;
  line-height: 1;
}

.btn-delete-mini:hover {
  border-color: #f44336;
  color: #f44336;
}
</style>