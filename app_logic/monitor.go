package app_logic

import (
	"MiHoYoStarterGo/logic"
	"context"
)

// RunMonitor 負責具體的業務流：解密並啟動自動化監控流程
func RunMonitor(ctx context.Context, acc logic.Account, pause, cancel *bool) {
	// 1. 從底層 logic 獲取解密後的明文密碼
	pwd, err := logic.DecryptString(acc.Password)
	if err != nil {
		// 如果解密失敗，可以在此處記錄日誌或通過 runtime 發送前端通知
		return
	}

	// 2. 呼叫底層 logic/automation.go 的核心監控函數
	// 該函數應在內部使用 goroutine 運行，避免阻塞 Wails 主線程
	go logic.StartAutomationMonitor(
		ctx,
		acc.GameID,
		acc.Username,
		pwd,
		acc.IsFirstLogin,
		pause,
		cancel,
	)
}

// StartGame 啟動遊戲的業務封裝
func StartGame(gameID string) string {
	// 1. 加載配置獲取路徑
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED_LOAD_CONFIG"
	}

	// 2. 檢查對應遊戲的路徑是否存在
	path, ok := cfg.GamePaths[gameID]
	if !ok || path == "" {
		return "PATH_NOT_FOUND"
	}

	// 3. 呼叫底層 logic/process.go 執行進程啟動
	if err := logic.StartProcess(path); err != nil {
		return "START_FAILED"
	}

	return "SUCCESS"
}
