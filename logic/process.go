package logic

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall" // 确保导入，用於 Windows 隐藏窗口

	"github.com/shirou/gopsutil/v3/process"
)

// IsGameRunning 检查游戏进程是否存在
func IsGameRunning(gameType string) bool {
	// 映射游戏名到进程名
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

// KillGameProcess 杀死指定的游戏进程 (核对原文件：已确认补回)
func KillGameProcess(gameType string) error {
	processMap := map[string]string{
		"GenshinCN":  "YuanShen.exe",
		"GenshinOS":  "GenshinImpact.exe",
		"StarRailCN": "StarRail.exe",
		"ZZZCN":      "ZenlessZoneZero.exe",
	}

	target, ok := processMap[gameType]
	if !ok {
		return fmt.Errorf("未知的游戏类型: %s", gameType)
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

// StartProcess 启动外部可执行文件 (仅在此处植入隐藏黑框逻辑)
func StartProcess(path string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// 使用 cmd /C start 启动，使游戏进程脱离启动器独立运行
		cmd = exec.Command("cmd", "/C", "start", "", path)

		// --- 这是解决黑框的核心改动，除此之外不影响任何原逻辑 ---
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	} else {
		cmd = exec.Command(path)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("无法启动进程: %v", err)
	}

	// 释放资源，不阻塞 Wails 主程序
	go func() {
		_ = cmd.Wait()
	}()

	return nil
}
