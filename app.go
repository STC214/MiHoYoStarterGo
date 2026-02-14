package main

import (
	"context"
	"fmt"
	"time"

	// 这里的路径必须与 go.mod 中的 module 名一致
	"MiHoYoStarterGo/logic"
)

// App struct
type App struct {
	ctx          context.Context
	IsPaused     bool // 用于全局控制自动化的暂停状态
	ShouldCancel bool // [新增] 用于全局控制自动化的取消信号
}

// NewApp 创建 App 实例
func NewApp() *App {
	return &App{
		IsPaused:     false,
		ShouldCancel: false,
	}
}

// startup 在 Wails 启动时运行
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- 调试与状态管理 ---

// TogglePause 供前端调用：切换暂停/继续状态
func (a *App) TogglePause() string {
	a.IsPaused = !a.IsPaused
	if a.IsPaused {
		fmt.Println(">> [App] 自动化监控已手动暂停")
		return "已暂停"
	}
	fmt.Println(">> [App] 自动化监控已恢复运行")
	return "运行中"
}

// StopMonitor [新增] 供前端调用：手动终止自动化流程并回收资源
func (a *App) StopMonitor() string {
	a.ShouldCancel = true
	fmt.Println(">> [App] 用户手动触发了终止信号，正在关闭监控协程...")
	return "SUCCESS"
}

// --- 配置与主题管理 ---

// GetSettings 获取初始化配置
func (a *App) GetSettings() (*logic.ConfigData, error) {
	return logic.LoadConfig()
}

// SaveTheme 切换并保存主题
func (a *App) SaveTheme(themeName string) string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "Error: " + err.Error()
	}
	cfg.Theme = themeName
	if err := logic.SaveConfig(cfg); err != nil {
		return "Error: " + err.Error()
	}
	return "Success"
}

// --- 账号管理逻辑 ---

// AddAccount 添加新账号
func (a *App) AddAccount(alias, username, password, gameID string) string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED: 无法加载配置文件"
	}

	encPwd, err := logic.EncryptString(password)
	if err != nil {
		return "FAILED: 加密失败"
	}

	newAcc := logic.Account{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Alias:        alias,
		Username:     username,
		Password:     encPwd,
		GameID:       gameID,
		IsFirstLogin: true,
		CreateTime:   time.Now().Unix(),
	}

	cfg.Accounts = append(cfg.Accounts, newAcc)
	if err := logic.SaveConfig(cfg); err != nil {
		return "FAILED: 保存失败"
	}
	return "SUCCESS"
}

// DeleteAccount [新增] 根据 ID 删除账号
func (a *App) DeleteAccount(id string) string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED: 无法读取配置"
	}

	var newAccounts []logic.Account
	found := false
	for _, acc := range cfg.Accounts {
		if acc.ID != id {
			newAccounts = append(newAccounts, acc)
		} else {
			found = true
		}
	}

	if !found {
		return "FAILED: 未找到该账号"
	}

	cfg.Accounts = newAccounts
	if err := logic.SaveConfig(cfg); err != nil {
		return "FAILED: 保存失败"
	}
	fmt.Printf(">> [App] 账号 ID %s 已成功删除\n", id)
	return "SUCCESS"
}

// GetPlaintext 解密密文
func (a *App) GetPlaintext(encryptedText string) string {
	decrypted, err := logic.DecryptString(encryptedText)
	if err != nil {
		return "解密失败"
	}
	return decrypted
}

// ExportBackup 导出明文 JSON 备份
func (a *App) ExportBackup() string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED: 无法加载配置"
	}
	fileName, err := logic.ExportPlaintextBackup(cfg)
	if err != nil {
		return "FAILED: " + err.Error()
	}
	return "SUCCESS: 已导出至 " + fileName
}

// --- 核心切换与自动化调度 ---

// RequestSwitch 核心业务逻辑
func (a *App) RequestSwitch(acc logic.Account) string {
	// 1. 进程检查：如果游戏正在运行，返回冲突状态码
	if logic.IsGameRunning(acc.GameID) {
		fmt.Printf(">> [App] 检测到游戏 %s 正在运行，触发冲突处理\n", acc.GameID)
		return "RUNNING_CONFLICT"
	}

	// 2. 游戏未运行，执行正常启动监控
	return a.ForceStartMonitor(acc)
}

// ForceStartMonitor 强制开启监控
func (a *App) ForceStartMonitor(acc logic.Account) string {
	// 解密密码
	realPwd, err := logic.DecryptString(acc.Password)
	if err != nil {
		return "FAILED: 账号解密失败。"
	}

	// [重要] 重置信号状态
	a.IsPaused = false
	a.ShouldCancel = false

	// 启动自动化监控协程
	// 这里传递了 a.ctx (用于事件), &a.IsPaused (用于暂停), &a.ShouldCancel (用于终止)
	go logic.StartAutomationMonitor(
		a.ctx,
		acc.GameID,
		acc.Username,
		realPwd,
		acc.IsFirstLogin,
		&a.IsPaused,
		&a.ShouldCancel,
	)

	fmt.Printf(">> [App] 账号 %s 的自动化监控已启动\n", acc.Alias)
	return "SUCCESS"
}
