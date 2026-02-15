package app_logic

import (
	"MiHoYoStarterGo/logic"
	"encoding/hex"
	"fmt"
)

func HandleEnvPatch(acc logic.Account) string {
	// 1. 检查游戏是否正在运行
	if logic.IsGameRunning(acc.GameID) {
		return "GAME_RUNNING"
	}

	// 2. 检查是否满足注入条件：有 Token 且不是第一次登录
	if acc.Token != "" && !acc.IsFirstLogin {
		fmt.Printf(">> [Patch] 正在为账号 %s 预注入环境...\n", acc.Alias)

		// 写入注册表
		tokenBytes, _ := hex.DecodeString(acc.Token)
		if err := logic.WriteToken(acc.GameID, tokenBytes); err != nil {
			return "REGISTRY_FAILED"
		}

		// 写入硬件指纹 (device.go)
		// logic.ApplyDeviceProfile(acc)

		return "ENV_READY"
	}

	return "SKIP"
}
