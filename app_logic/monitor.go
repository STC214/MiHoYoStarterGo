package app_logic

import (
	"MiHoYoStarterGo/logic"
	"context"
)

// RunMonitor 负责具体的业务流：解密并启动自动化监控流程
func RunMonitor(ctx context.Context, acc logic.Account, pause, cancel *bool) {
	// 1. 从底层 logic 获取解密後的明文密码
	pwd, err := logic.DecryptString(acc.Password)
	if err != nil {
		// 如果解密失败，可以在此处记录日志或通过 runtime 发送前端通知
		return
	}

	// 2. 呼叫底层 logic/automation.go 的核心监控函数
	// 该函数应在内部使用 goroutine 运行，避免阻塞 Wails 主线程
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

// StartGame 启动游戏的业务封装
func StartGame(gameID string) string {
	// 1. 加载配置获取路径
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "FAILED_LOAD_CONFIG"
	}

	// 2. 检查对应游戏的路径是否存在
	path, ok := cfg.GamePaths[gameID]
	if !ok || path == "" {
		return "PATH_NOT_FOUND"
	}

	// 3. 呼叫底层 logic/process.go 执行进程启动
	if err := logic.StartProcess(path); err != nil {
		return "START_FAILED"
	}

	return "SUCCESS"
}
