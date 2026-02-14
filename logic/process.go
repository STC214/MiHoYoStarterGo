package logic

import (
	"strings"

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
