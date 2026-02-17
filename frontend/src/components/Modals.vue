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

      <div v-if="activeType === 'edit'" class="modal-content">
        <h3>编辑账号</h3>
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
          <label>游戏密码（留空表示不修改）</label>
          <input
            :value="newAcc.password"
            @input="updateNewAcc('password', $event.target.value)"
            type="password"
            placeholder="不修改可留空"
          />
        </div>
        <div class="modal-actions">
          <button @click="$emit('close')">取消</button>
          <button class="btn-primary" @click="$emit('confirmEdit')">保存修改</button>
        </div>
      </div>

      <div v-if="activeType === 'runAction'" class="modal-content">
        <h3>切换账号：{{ runContext.alias }}</h3>
        <p class="modal-desc">{{ runContext.isRunning ? '检测到目标游戏进程正在运行，请选择处理方式。' : '检测到目标游戏进程未运行，请选择处理方式。' }}</p>

        <div class="modal-actions full-width">
          <button
            v-if="runContext.gameID === 'ZZZCN'"
            class="btn-zzz-calibrate"
            @click="$emit('openZZZCalibrate')"
          >
            打开绝区零点位标定
          </button>
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

      <div v-if="activeType === 'zzzCalibrate'" class="modal-content zzz-calibrate-modal">
        <h3>绝区零点位标定</h3>
        <p class="modal-desc">
          仅用于绝区零。点击“开始记录”后，按提示将鼠标移动到目标位置，系统会倒计时自动读取坐标。
        </p>
        <div class="zzz-focus-tip">
          <span class="zzz-focus-tag">当前应放置鼠标</span>
          <strong>{{ currentStepText }}</strong>
        </div>
        <div class="zzz-step-grid">
          <div
            v-for="(stepName, idx) in zzzStepNames"
            :key="stepName"
            :class="[
              'zzz-step-item',
              zzzCalibrate.step === idx + 1 ? 'active' : '',
              zzzCalibrate.step > idx + 1 ? 'done' : ''
            ]"
          >
            <span class="step-index">步骤 {{ idx + 1 }}</span>
            <span class="step-name">{{ stepName }}</span>
          </div>
        </div>
        <div class="zzz-calibrate-panel">
          <div class="zzz-calibrate-header">
            <span>步骤：{{ zzzCalibrate.step }}/{{ zzzCalibrate.total }}</span>
            <span class="phase-tag">{{ phaseText }}</span>
          </div>
          <p class="zzz-calibrate-text">{{ zzzCalibrate.text }}</p>
          <p v-if="zzzCalibrate.label" class="zzz-calibrate-label">当前目标：{{ zzzCalibrate.label }}</p>
          <p v-if="zzzCalibrate.phase === 'captured'" class="zzz-calibrate-point">
            已记录坐标：X={{ zzzCalibrate.x }}, Y={{ zzzCalibrate.y }}
          </p>
        </div>
        <div class="modal-actions">
          <button @click="$emit('close')" :disabled="zzzCalibrate.running">关闭</button>
          <button class="btn-primary" @click="$emit('startZZZCalibrate')" :disabled="zzzCalibrate.running">
            {{ zzzCalibrate.running ? '记录中...' : '开始记录' }}
          </button>
        </div>
      </div>

      <div v-if="activeType === 'status'" class="modal-content status-modal">
        <div class="loader"></div>
        <div class="status-box">
          <p class="status-text">{{ statusTip }}</p>
        </div>
        <div class="modal-actions status-actions">
          <button class="btn-primary btn-wide" @click="$emit('captureDebug')">保存调试截图</button>
          <button class="btn-warning btn-wide" @click="$emit('togglePause')">
            {{ pauseStatus === 'PAUSED' ? '继续监控' : '暂停监控' }}
          </button>
          <button class="btn-danger btn-wide" @click="$emit('stopMonitor')">停止并退出</button>
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

      <div v-if="activeType === 'message'" class="modal-content">
        <h3>提示</h3>
        <p class="modal-desc">{{ messageText }}</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="$emit('close')">确定</button>
        </div>
      </div>

      <div v-if="activeType === 'deleteConfirm'" class="modal-content">
        <h3>确认删除？</h3>
        <p class="modal-desc">{{ messageText }}</p>
        <div class="modal-actions">
          <button @click="$emit('close')">取消</button>
          <button class="btn-danger" @click="$emit('confirmDelete')">确认删除</button>
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
import { computed } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  activeType: String,
  newAcc: Object,
  statusTip: String,
  messageText: String,
  pauseStatus: String,
  runContext: Object,
  zzzCalibrate: Object,
  games: Array,
  gamePaths: Object
})

defineEmits([
  'close',
  'confirmAdd',
  'confirmEdit',
  'runAction',
  'openZZZCalibrate',
  'startZZZCalibrate',
  'captureDebug',
  'togglePause',
  'stopMonitor',
  'updatePath',
  'browse',
  'savePaths',
  'cancelStop',
  'confirmStop',
  'confirmDelete'
])

const updateNewAcc = (key, value) => {
  props.newAcc[key] = value
}

const zzzStepNames = ['账号输入框', '密码输入框', '同意协议勾选点', '进入游戏点击点']

const currentStepText = computed(() => {
  const i = Number(props.zzzCalibrate?.step || 0) - 1
  if (i >= 0 && i < zzzStepNames.length) {
    return zzzStepNames[i]
  }
  return '等待开始记录'
})

const phaseText = computed(() => {
  const p = props.zzzCalibrate?.phase || 'idle'
  if (p === 'start') return '开始'
  if (p === 'prompt') return '请放置鼠标'
  if (p === 'captured') return '已记录'
  if (p === 'done') return '完成'
  if (p === 'error') return '错误'
  return '待机'
})
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

.status-actions {
  width: 100%;
  justify-content: center;
  flex-direction: column;
  gap: 12px;
}

.btn-wide {
  width: 100%;
  text-align: center;
  padding: 10px 16px;
  font-weight: 600;
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

.btn-secondary {
  background: transparent;
  color: var(--text);
  border: 1px solid var(--border);
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

.btn-zzz-calibrate {
  background: #2f5d42;
  color: #e8fff1;
  border: 1px solid #4f8a66;
  padding: 10px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
}

.zzz-calibrate-modal {
  border-color: #4f8a66;
}

.zzz-focus-tip {
  margin: 12px 0;
  padding: 12px;
  border-radius: 8px;
  background: linear-gradient(90deg, rgba(79, 138, 102, 0.3), rgba(79, 138, 102, 0.08));
  border: 1px solid rgba(79, 138, 102, 0.75);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.zzz-focus-tag {
  font-size: 12px;
  color: #b7e7c8;
}

.zzz-focus-tip strong {
  font-size: 18px;
  color: #e9fff2;
  letter-spacing: 0.5px;
}

.zzz-step-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 10px;
}

.zzz-step-item {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: rgba(255, 255, 255, 0.02);
}

.zzz-step-item.active {
  border-color: #7fdb9e;
  background: rgba(127, 219, 158, 0.18);
  box-shadow: 0 0 0 1px rgba(127, 219, 158, 0.25) inset;
}

.zzz-step-item.done {
  border-color: #4f8a66;
  background: rgba(79, 138, 102, 0.15);
}

.step-index {
  font-size: 11px;
  color: var(--text-dim);
}

.step-name {
  font-size: 14px;
}

.zzz-calibrate-panel {
  background: rgba(47, 93, 66, 0.12);
  border: 1px solid rgba(79, 138, 102, 0.5);
  border-radius: 8px;
  padding: 12px;
}

.zzz-calibrate-header {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #bfe8cd;
}

.phase-tag {
  background: rgba(79, 138, 102, 0.35);
  border-radius: 999px;
  padding: 2px 8px;
}

.zzz-calibrate-text {
  margin: 10px 0 0;
}

.zzz-calibrate-label,
.zzz-calibrate-point {
  margin: 8px 0 0;
  color: #bfe8cd;
  font-size: 13px;
}

</style>
