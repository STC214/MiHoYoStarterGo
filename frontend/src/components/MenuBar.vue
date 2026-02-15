<template>
  <nav class="menu-bar">
    <div class="menu-item" @click="$emit('toggleTheme')">🎨 切換主題</div>
    <div class="menu-item" @click="$emit('handleExport')">📥 導出備份</div>
    <div class="menu-item" @click="$emit('openPath')">⚙️ 遊戲路徑</div>
    <div class="menu-item" @click="$emit('openAdd')">➕ 添加賬號</div>
    
    <div v-if="showStatus" class="status-indicator-mini" style="margin-left: auto;">
      監諾中: <span :class="pauseStatus === '運行中' ? 'text-green' : 'text-red'">{{ pauseStatus }}</span>
    </div>
  </nav>
</template>

<script setup>
defineProps({
  showStatus: Boolean,
  pauseStatus: String
})
defineEmits(['toggleTheme', 'handleExport', 'openPath', 'openAdd'])
</script>

<style scoped>
.menu-bar { display: flex; background: var(--bg-card); padding: 5px 10px; gap: 10px; align-items: center; }
.menu-item { padding: 5px 12px; border-radius: 4px; cursor: pointer; font-size: 13px; }
.menu-item:hover { background: var(--bg-app); }
.status-indicator-mini { font-size: 11px; background: rgba(0,0,0,0.3); padding: 4px 10px; border-radius: 10px; }
.text-green { color: #4caf50; }
.text-red { color: #f44336; }
</style>