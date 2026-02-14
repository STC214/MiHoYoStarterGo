package logic

import (
	"encoding/hex"
	"fmt"
)

// FinalizeAccountData 在登录成功后抓取注册表 Token 并保存
func FinalizeAccountData(gameID, username string) error {
	// 1. 读取当前注册表里的 Token
	tokenBytes, err := ReadToken(gameID)
	if err != nil {
		return fmt.Errorf("读取注册表失败: %v", err)
	}
	tokenHex := hex.EncodeToString(tokenBytes)

	// 2. 加载配置文件
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// 3. 找到对应的账号并更新
	found := false
	for i, acc := range cfg.Accounts {
		if acc.Username == username && acc.GameID == gameID {
			cfg.Accounts[i].Token = tokenHex
			cfg.Accounts[i].IsFirstLogin = false // 标记为非首次登录
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未在配置中找到账号: %s", username)
	}

	// 4. 写回文件
	return SaveConfig(cfg)
}
