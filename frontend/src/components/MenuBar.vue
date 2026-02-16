<template>
  <nav class="menu-bar">
    <div class="menu-item" @click="$emit('toggleTheme')">
      <div class="menu-text">切换主题</div>
      <div class="menu-icon">🎨</div>
    </div>

    <div class="menu-item" @click="$emit('handleExport')">
      <div class="menu-text">导出备份</div>
      <div class="menu-icon">📦</div>
    </div>

    <div class="menu-item" @click="$emit('openPath')">
      <div class="menu-text">游戏路径</div>
      <div class="menu-icon">⚙️</div>
    </div>

    <div class="menu-item" @click="$emit('openAdd')">
      <div class="menu-text">添加账号</div>
      <div class="menu-icon">➕</div>
    </div>

    <div class="status-indicator-mini" @click="$emit('openStatus')">
      <div class="menu-text">状态监视：</div>
      <div :class="monitorActive ? (pauseStatus === 'RUNNING' ? 'text-green' : 'text-red') : 'text-gray'" class="status-line">
        {{ monitorActive ? (pauseStatus === 'RUNNING' ? '运行中' : '已暂停') : '未运行' }}
      </div>
    </div>
  </nav>
</template>

<script setup>
defineProps({
  monitorActive: Boolean,
  pauseStatus: String
})

defineEmits(['toggleTheme', 'handleExport', 'openPath', 'openAdd', 'openStatus'])
</script>

<style scoped>
.menu-bar {
  display: flex;
  align-items: stretch;
  background: var(--bg-card);
  padding: 6px 10px;
  height: auto;
  min-height: 46px;
  overflow: visible;
  box-sizing: border-box;
  gap: 0;
}

.menu-item,
.status-indicator-mini {
  min-width: max-content;
  padding: 6px 0.55em;
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  text-align: center;
  box-sizing: border-box;
  position: relative;
}

.status-indicator-mini {
  margin-left: auto;
}

.menu-item + .menu-item {
  margin-left: 1em;
}

.menu-item + .menu-item::before {
  content: "";
  position: absolute;
  left: -0.5em;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 1px;
  height: 62%;
  background: var(--border);
}

.menu-item:hover,
.status-indicator-mini:hover {
  background: var(--bg-app);
}

.menu-text {
  font-size: 12px;
  line-height: 1.1;
  white-space: nowrap;
}

.menu-icon {
  font-size: 14px;
  line-height: 1;
}

.status-line {
  font-size: 12px;
  line-height: 1.1;
  font-weight: 600;
  white-space: nowrap;
}

.text-green {
  color: #4caf50;
}

.text-red {
  color: #f44336;
}

.text-gray {
  color: #a0a0a0;
}
</style>
