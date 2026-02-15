package logic

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall" // 確保導入，用於 Windows 隱藏窗口

	"github.com/shirou/gopsutil/v3/process"
)

// IsGameRunning 檢查遊戲進程是否存在
func IsGameRunning(gameType string) bool {
	// 映射遊戲名到進程名
	processMap := map[string]string{
		"GenshinCN":  "YuanShen.exe",
		"GenshinOS":  "GenshinImpact.exe",
		"StarRailCN": "StarRail.exe",
		"ZZZCN":      "ZenlessZoneZero.exe",
	}

	target, ok := processMap[gameType]
	if !ok {
		return false
	}

	pids, _ := process.Pids()
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err == nil {
			name, _ := p.Name()
			if strings.EqualFold(name, target) {
				return true
			}
		}
	}
	return false
}

// KillGameProcess 殺死指定的遊戲進程 (核對原文件：已確認補回)
func KillGameProcess(gameType string) error {
	processMap := map[string]string{
		"GenshinCN":  "YuanShen.exe",
		"GenshinOS":  "GenshinImpact.exe",
		"StarRailCN": "StarRail.exe",
		"ZZZCN":      "ZenlessZoneZero.exe",
	}

	target, ok := processMap[gameType]
	if !ok {
		return fmt.Errorf("未知的遊戲類型: %s", gameType)
	}

	pids, _ := process.Pids()
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err == nil {
			name, _ := p.Name()
			if strings.EqualFold(name, target) {
				_ = p.Kill()
			}
		}
	}
	return nil
}

// StartProcess 啟動外部可執行文件 (僅在此處植入隱藏黑框邏輯)
func StartProcess(path string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// 使用 cmd /C start 啟動，使遊戲進程脫離啟動器獨立運行
		cmd = exec.Command("cmd", "/C", "start", "", path)

		// --- 這是解決黑框的核心改動，除此之外不影響任何原邏輯 ---
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	} else {
		cmd = exec.Command(path)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("無法啟動進程: %v", err)
	}

	// 釋放資源，不阻塞 Wails 主程序
	go func() {
		_ = cmd.Wait()
	}()

	return nil
}
