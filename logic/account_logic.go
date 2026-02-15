package logic

import (
	"encoding/hex"
	"fmt"
)

// FinalizeAccountData 鍦ㄧ櫥褰曟垚鍔熷悗鎶撳彇娉ㄥ唽琛?Token 骞朵繚瀛?
func FinalizeAccountData(gameID, username string) error {
	// 1. 璇诲彇褰撳墠娉ㄥ唽琛ㄩ噷鐨?Token
	tokenBytes, err := ReadToken(gameID)
	if err != nil {
		return fmt.Errorf("璇诲彇娉ㄥ唽琛ㄥけ璐? %v", err)
	}
	tokenHex := hex.EncodeToString(tokenBytes)

	// 2. 鍔犺浇閰嶇疆鏂囦欢
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// 3. 鎵惧埌瀵瑰簲鐨勮处鍙峰苟鏇存柊
	found := false
	for i, acc := range cfg.Accounts {
		if acc.Username == username && acc.GameID == gameID {
			cfg.Accounts[i].Token = tokenHex
			cfg.Accounts[i].DeviceFingerprint = GetDeviceFingerprint()
			cfg.Accounts[i].IsFirstLogin = false // 鏍囪涓洪潪棣栨鐧诲綍
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("鏈湪閰嶇疆涓壘鍒拌处鍙? %s", username)
	}

	// 4. 鍐欏洖鏂囦欢
	return SaveConfig(cfg)
}
