package main

import (
	"context"
	"fmt"
	"time"

	// 這裡的路徑必須與 go.mod 中的 module 名一致
	"MiHoYoStarterGo/logic"
)

// App struct
type App struct {
	ctx          context.Context
	IsPaused     bool // 用於全局控制自動化的暫停狀態
	ShouldCancel bool // [新增] 用於全局控制自動化的取消信號
}

// NewApp 創建 App 實例
func NewApp() *App {
	return &App{
		IsPaused:     false,
		ShouldCancel: false,
	}
}

// startup 在 Wails 啟動時運行
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- 調試與狀態管理 ---

// TogglePause 供前端調用：切換暫停/繼續狀態
func (a *App) TogglePause() string {
	a.IsPaused = !a.IsPaused
	if a.IsPaused {
		fmt.Println(">> [App] 自動化監控已手動暫停")
		return "已暫停"
	}
	fmt.Println(">> [App] 自動化監控已恢復運行")
	return "運行中"
}

// StopMonitor [新增] 供前端調用：手動終止自動化流程並回收資源
func (a *App) StopMonitor() string {
	a.ShouldCancel = true
	fmt.Println(">> [App] 用戶手動觸發了終止信號，正在關閉監控協程...")
	return "SUCCESS"
}

// --- 配置與主題管理 ---

// GetSettings 獲取初始化配置
func (a *App) GetSettings() (*logic.ConfigData, error) {
	return logic.LoadConfig()
}

// SaveTheme 切換並保存主題
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

// --- 賬號管理邏輯 ---

// AddAccount 添加新賬號
func (a *App) AddAccount(alias, username, password, gameID string) string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED: 無法加載配置文件"
	}

	encPwd, err := logic.EncryptString(password)
	if err != nil {
		return "FAILED: 加密失敗"
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
		return "FAILED: 保存失敗"
	}
	return "SUCCESS"
}

// DeleteAccount [新增] 根據 ID 刪除賬號
func (a *App) DeleteAccount(id string) string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED: 無法讀取配置"
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
		return "FAILED: 未找到該賬號"
	}

	cfg.Accounts = newAccounts
	if err := logic.SaveConfig(cfg); err != nil {
		return "FAILED: 保存失敗"
	}
	fmt.Printf(">> [App] 賬號 ID %s 已成功刪除\n", id)
	return "SUCCESS"
}

// GetPlaintext 解密密文
func (a *App) GetPlaintext(encryptedText string) string {
	decrypted, err := logic.DecryptString(encryptedText)
	if err != nil {
		return "解密失敗"
	}
	return decrypted
}

// ExportBackup 導出明文 JSON 備份
func (a *App) ExportBackup() string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED: 無法加載配置"
	}
	fileName, err := logic.ExportPlaintextBackup(cfg)
	if err != nil {
		return "FAILED: " + err.Error()
	}
	return "SUCCESS: 已導出至 " + fileName
}

// --- 核心切換與自動化調度 ---

// RequestSwitch 核心業務邏輯
func (a *App) RequestSwitch(acc logic.Account) string {
	// 1. 進程檢查：如果遊戲正在運行，返回衝突狀態碼
	if logic.IsGameRunning(acc.GameID) {
		fmt.Printf(">> [App] 檢測到遊戲 %s 正在運行，觸發衝突處理\n", acc.GameID)
		return "RUNNING_CONFLICT"
	}

	// 2. 遊戲未運行，執行正常啟動監控
	return a.ForceStartMonitor(acc)
}

// ForceStartMonitor 強制開啟監控
func (a *App) ForceStartMonitor(acc logic.Account) string {
	// 解密密碼
	realPwd, err := logic.DecryptString(acc.Password)
	if err != nil {
		return "FAILED: 賬號解密失敗。"
	}

	// [重要] 重置信號狀態
	a.IsPaused = false
	a.ShouldCancel = false

	// 啟動自動化監控協程
	// 這裡傳遞了 a.ctx (用於事件), &a.IsPaused (用於暫停), &a.ShouldCancel (用於終止)
	go logic.StartAutomationMonitor(
		a.ctx,
		acc.GameID,
		acc.Username,
		realPwd,
		acc.IsFirstLogin,
		&a.IsPaused,
		&a.ShouldCancel,
	)

	fmt.Printf(">> [App] 賬號 %s 的自動化監控已啟動\n", acc.Alias)
	return "SUCCESS"
}
