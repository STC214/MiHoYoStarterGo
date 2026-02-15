# MiHoYoStarterGo 程序逻辑实现路径总览（按功能模块）

## 1. 总体架构与调用主链

### 1.1 启动链路
1. 入口 `main.go` 启动时先读取配置（窗口尺寸、主题、账号、游戏路径等）。
2. 初始化 Wails 应用并绑定 `App`（`app.go`）。
3. 关闭窗口前回写窗口位置与尺寸到 `config.json`。

> 主链：`main.go -> NewApp() -> wails.Run(...) -> app.startup(ctx)`。

### 1.2 分层职责
- **`app.go`（桥接层）**：对前端暴露可调用方法，不承载复杂业务。
- **`app_logic/*`（用例层）**：组装业务流程（账号增删、监控启动、设置保存等）。
- **`logic/*`（基础能力层）**：配置存储、加解密、OCR、截图、进程控制、注册表读写等。

---

## 2. 功能模块与具体实现路径

## 2.1 配置与持久化模块（`logic/config.go`, `app_logic/config.go`）

### 功能点
- 读取/写入 `config.json`。
- 保存主题、游戏路径、窗口尺寸等配置。
- 导出账号备份。

### 实现路径
1. 前端调用 `App.GetSettings/SaveTheme/SaveGamePaths/ExportBackup`。
2. `app.go` 转发到 `app_logic/config.go`。
3. `app_logic` 调用 `logic.LoadConfig/SaveConfig/ExportPlaintextBackup`。
4. `logic/config.go` 执行 JSON 序列化与文件读写。

### 关键机制
- 首次运行若配置不存在，返回内置默认配置。
- `GamePaths` 做了 `nil` 保护，避免前端 map 访问异常。

---

## 2.2 账号管理模块（`app_logic/account.go`, `logic/config.go`）

### 功能点
- 新增账号、删除账号、查看明文密码。

### 新增账号路径
1. `App.AddAccount(...)`。
2. `app_logic.AddAccount(...)` 调用 `logic.EncryptString(pwd)`。
3. 构建 `logic.Account`（ID、别名、账号、加密密码、游戏 ID、首次登录标记等）。
4. 回写 `config.json`。

### 删除账号路径
1. `App.DeleteAccount(id)`。
2. `app_logic.DeleteAccount(id)` 过滤账号数组。
3. 保存配置。

### 密码处理机制
- 使用 AES-GCM。
- 密钥由设备指纹（CPU + 主板序列号）经 MD5 + 盐衍生。

---

## 2.3 设备指纹与加密模块（`logic/device.go`, `logic/config.go`）

### 功能点
- 生成设备绑定密钥，确保密码密文不可跨机直接复用。

### 实现路径
1. `GetDeviceFingerprint()` 通过 WMI 查询 CPU ID 与主板序列号。
2. 拼接后计算 MD5，输出设备指纹。
3. `getSecretKey()` 再对指纹加盐 MD5 并转 32 字节 hex key。
4. `EncryptString/DecryptString` 使用 AES-GCM 执行加解密。

---

## 2.4 游戏进程控制模块（`logic/process.go`, `app_logic/monitor.go`）

### 功能点
- 判断游戏是否运行。
- 按游戏类型启动进程。
- 杀死游戏进程（预留能力）。

### 启动路径
1. `App.StartGameExecution(gameID)`。
2. `app_logic.StartGame(gameID)` 从配置读取游戏 exe 路径。
3. 调用 `logic.StartProcess(path)` 启动。
4. Windows 下通过 `cmd /C start` + `CREATE_NO_WINDOW` 隐藏黑框。

### 运行检测路径
- `logic.IsGameRunning(gameID)`：游戏 ID -> 目标进程名映射 -> gopsutil 遍历进程匹配。

---

## 2.5 环境预注入模块（`app_logic/env_patch.go`, `logic/registry.go`）

### 功能点
- 在“非首次登录 + 有 Token + 游戏未运行”条件下，预写注册表 Token，减少重复登录。

### 实现路径
1. `App.PrepareAccountEnvironment(acc)`。
2. `app_logic.HandleEnvPatch(acc)` 做前置条件检查。
3. 命中条件时将 hex Token 解码为字节。
4. 调用 `logic.WriteToken(gameID, tokenBytes)` 写入对应注册表项。

### 注册表寻址机制
- 先按游戏 ID 找到注册表路径与 key 前缀。
- 再通过 `findActualKeyName` 动态查找真实键名（兼容后缀变化）。

---

## 2.6 自动化监控主流程（`app_logic/monitor.go`, `logic/automation.go`）

### 功能点
- 自动识别登录界面并填充账号密码。
- 识别“进入游戏成功”后提取 Token 并落库。
- 支持暂停与手动停止。

### 完整流程路径
1. `App.ForceStartMonitor(acc)`：重置 `IsPaused/ShouldCancel`，异步启动监控。
2. `app_logic.RunMonitor(...)`：先解密账号密码，再进入核心监控。
3. `logic.StartAutomationMonitor(...)` 启动 ticker 循环（300ms）：
   - 检查 `cancel`（立即退出）。
   - 检查 `pause`（跳过本轮）。
   - 校验目标游戏进程是否存在。
   - 根据 `gameID` 匹配窗口标题并置前。
   - 截图 -> OCR -> 文本点位。
   - 若判定为“登录界面”（画面 B），执行输入与点击序列。
   - 若判定为“进入游戏”（画面 A），提取 Token，更新账号状态并发事件给前端。
4. 结束时清理临时截图文件。

### 控制接口
- `App.TogglePause()`：切换暂停。
- `App.StopMonitor()`：设置取消标记。

---

## 2.7 OCR 与图像识别模块（`logic/capture.go`, `logic/ocr_engine.go`, `logic/automation.go`）

### 功能点
- 窗口截图、OCR 识别、文字框坐标提取。

### 实现路径
1. `CaptureWindow(windowName)`：WinAPI `PrintWindow` 抓取窗口内容，转 `image.RGBA`。
2. 保存临时图片 `temp_*.png`。
3. `RecognizeWithPos(imagePath)`：调用本地 `PaddleOCR-json.exe`。
4. 解析 OCR JSON 为 `[]TextPoint`（文本 + 坐标 + 宽高）。
5. 自动化逻辑使用文本条件 + 坐标进行鼠标键盘模拟。

### 识别判定策略
- **画面 B（登录页）**：检测“手机号/邮箱”+“同意/已阅读”。
- **画面 A（成功页）**：窗口底部检测“进入游戏/点击进入”且无中文干扰项。

---

## 2.8 令牌提取与账号状态闭环（`logic/automation.go`, `logic/account_logic.go`, `logic/registry.go`）

### 功能点
- 登录成功后抓取注册表 Token，写回账号信息，关闭首次登录标记。

### 闭环路径（实际运行）
- `StartAutomationMonitor -> finalizeAccountStorage -> ReadToken -> LoadConfig -> 更新 Token/IsFirstLogin -> SaveConfig`。

### 备用路径
- `FinalizeAccountData(gameID, username)` 提供同类能力（当前主流程里未直接调用）。

---

## 3. 风险与优化建议（按功能模块映射）

1. **监控控制并发风险（监控模块）**
   - `pause/cancel` 由多个 goroutine 读写，建议改 `atomic.Bool` 或 channel。

2. **错误处理缺口（配置/账号/OCR/流程编排）**
   - 多处忽略 error，建议统一返回错误码并透传到前端。

3. **明文备份风险（配置模块）**
   - `ExportPlaintextBackup` 默认导出明文密码，建议增加口令加密或强提醒。

4. **OCR 数据健壮性（OCR 模块）**
   - 解析 `d.Box` 前需校验长度与边界，避免异常输出导致 panic。

5. **平台隔离（系统能力模块）**
   - Windows 依赖建议使用 `go:build windows` 与非 Windows stub 分层。

---

## 4. 一句话总结

该程序本质上是一个 **“前端触发 -> Go 用例编排 -> Windows 系统能力执行 -> OCR 驱动自动化 -> 注册表 Token 回写”** 的闭环桌面自动化系统；核心价值在本地离线、系统集成深、账号切换链路完整。
