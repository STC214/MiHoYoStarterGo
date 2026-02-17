package logic

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

func IsGameRunning(gameType string) bool {
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
		if err != nil {
			continue
		}
		name, _ := p.Name()
		if strings.EqualFold(name, target) {
			return true
		}
	}
	return false
}

func KillGameProcess(gameType string) error {
	processMap := map[string]string{
		"GenshinCN":  "YuanShen.exe",
		"GenshinOS":  "GenshinImpact.exe",
		"StarRailCN": "StarRail.exe",
		"ZZZCN":      "ZenlessZoneZero.exe",
	}

	target, ok := processMap[gameType]
	if !ok {
		return fmt.Errorf("unknown game type: %s", gameType)
	}

	pids, _ := process.Pids()
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		name, _ := p.Name()
		if strings.EqualFold(name, target) {
			_ = p.Kill()
		}
	}
	return nil
}

func StartProcess(path string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "start", "", path)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000,
		}
	} else {
		cmd = exec.Command(path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot start process: %w", err)
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}
