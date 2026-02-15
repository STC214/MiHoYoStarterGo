package app_logic

import (
	"MiHoYoStarterGo/logic"
	"context"
)

// RunMonitor 负责具体的业务流：解密并启动自动化
func RunMonitor(ctx context.Context, acc logic.Account, pause, cancel *bool) {
	// 从底层 logic 获取解密后的明文密码
	pwd, _ := logic.DecryptString(acc.Password)

	// 调用底层 logic/automation.go 的核心监控函数
	logic.StartAutomationMonitor(
		ctx,
		acc.GameID,
		acc.Username,
		pwd,
		acc.IsFirstLogin,
		pause,
		cancel,
	)
}

// StartGame 启动游戏的业务封装
func StartGame(gameID string) string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED_LOAD_CONFIG"
	}

	if path, ok := cfg.GamePaths[gameID]; ok && path != "" {
		// 调用底层 logic/process.go 执行启动
		if err := logic.StartProcess(path); err == nil {
			return "SUCCESS"
		}
	}
	return "PATH_NOT_FOUND"
}
