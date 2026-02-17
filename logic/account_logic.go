package logic

import (
	"encoding/hex"
	"fmt"
)

// FinalizeAccountData reads token from registry and writes it back into config.
func FinalizeAccountData(gameID, username string) error {
	tokenBytes, err := ReadToken(gameID)
	if err != nil {
		return fmt.Errorf("read token failed: %w", err)
	}
	tokenHex := hex.EncodeToString(tokenBytes)

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	found := false
	for i, acc := range cfg.Accounts {
		if acc.Username == username && acc.GameID == gameID {
			cfg.Accounts[i].Token = tokenHex
			cfg.Accounts[i].DeviceFingerprint = GetDeviceFingerprint()
			cfg.Accounts[i].IsFirstLogin = false
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("account not found in config: %s", username)
	}

	return SaveConfig(cfg)
}
