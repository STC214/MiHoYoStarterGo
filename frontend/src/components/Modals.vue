<template>
  <Transition name="fade">
    <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')">
      <div v-if="activeType === 'add'" class="modal-content">
        <h3>添加新账号</h3>
        <div class="form-group">
          <label>别名</label>
          <input
            :value="newAcc.alias"
            @input="updateNewAcc('alias', $event.target.value)"
            placeholder="如：大号"
          />
        </div>
        <div class="form-group">
          <label>游戏账号</label>
          <input
            :value="newAcc.username"
            @input="updateNewAcc('username', $event.target.value)"
            placeholder="手机号/邮箱"
          />
        </div>
        <div class="form-group">
          <label>游戏密码</label>
          <input
            :value="newAcc.password"
            @input="updateNewAcc('password', $event.target.value)"
            type="password"
            placeholder="请输入密码"
          />
        </div>
        <div class="modal-actions">
          <button @click="$emit('close')">取消</button>
          <button class="btn-primary" @click="$emit('confirmAdd')">确认添加</button>
        </div>
      </div>

      <div v-if="activeType === 'runAction'" class="modal-content">
        <h3>切换账号：{{ runContext.alias }}</h3>
        <p class="modal-desc">{{ runContext.isRunning ? '检测到目标游戏进程正在运行，请选择处理方式。' : '检测到目标游戏进程未运行，请选择处理方式。' }}</p>

        <div class="modal-actions full-width">
          <template v-if="runContext.isRunning">
            <button class="btn-primary" @click="$emit('runAction', 'running_restart')">切换账号并自动重启</button>
            <button class="btn-warning" @click="$emit('runAction', 'running_manual')">手动切换到登录界面</button>
          </template>
          <template v-else>
            <button class="btn-primary" @click="$emit('runAction', 'stopped_auto_start')">换号并启动游戏</button>
            <button class="btn-warning" @click="$emit('runAction', 'stopped_manual_wait')">手动启动游戏</button>
          </template>
          <button class="btn-danger" @click="$emit('runAction', 'cancel')">取消</button>
        </div>
      </div>

      <div v-if="activeType === 'status'" class="modal-content status-modal">
        <div class="loader"></div>
        <div class="status-box">
          <p class="status-text">{{ statusTip }}</p>
        </div>
        <div class="modal-actions">
          <button class="btn-warning" @click="$emit('togglePause')">
            {{ pauseStatus === 'PAUSED' ? '继续监控' : '暂停监控' }}
          </button>
          <button class="btn-danger" @click="$emit('stopMonitor')">停止并退出</button>
        </div>
      </div>

      <div v-if="activeType === 'path'" class="modal-content">
        <h3>游戏执行文件路径 (.exe)</h3>
        <div v-for="game in games" :key="game.id" class="form-group">
          <label>{{ game.name }}</label>
          <div class="path-input-group">
            <input
              :value="gamePaths[game.id]"
              @input="$emit('updatePath', { id: game.id, val: $event.target.value })"
              placeholder="请选择或填写路径..."
            />
            <button class="btn-browse" @click="$emit('browse', game.id)">浏览</button>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="$emit('close')">取消</button>
          <button class="btn-primary" @click="$emit('savePaths')">保存路径</button>
        </div>
      </div>

      <div v-if="activeType === 'stopConfirm'" class="modal-content">
        <h3>确认停止？</h3>
        <p class="modal-desc">确定要停止监控吗？这将释放 OCR 资源并停止自动化流程。</p>
        <div class="modal-actions">
          <button @click="$emit('cancelStop')">继续监控</button>
          <button class="btn-danger" @click="$emit('confirmStop')">确认停止</button>
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
  runContext: Object,
  games: Array,
  gamePaths: Object
})

defineEmits([
  'close',
  'confirmAdd',
  'runAction',
  'togglePause',
  'stopMonitor',
  'updatePath',
  'browse',
  'savePaths',
  'cancelStop',
  'confirmStop'
])

const updateNewAcc = (key, value) => {
  props.newAcc[key] = value
}
</script>

<style scoped>
@import './Modals.css';

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.modal-desc {
  color: var(--text-dim);
  font-size: 14px;
  margin: 15px 0;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
  flex-wrap: wrap;
}

.btn-warning {
  background: #f39c12;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

.btn-danger {
  background: #e74c3c;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

.btn-primary {
  background: var(--primary);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}
</style>
