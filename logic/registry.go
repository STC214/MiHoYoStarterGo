package logic

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type GameRegConfig struct {
	Path   string
	Prefix string // 改用前綴匹配，增加兼容性
}

var GamePaths = map[string]GameRegConfig{
	"GenshinCN": {
		Path:   `Software\miHoYo\原神`,
		Prefix: "MIHOYOSDK_ADL_PROD_CN",
	},
	"GenshinOS": {
		Path:   `Software\Cognosphere\Genshin Impact`,
		Prefix: "MIHOYOSDK_ADL_PROD_OVERSEA",
	},
	"StarRailCN": {
		Path:   `Software\miHoYo\崩坏：星穹铁道`,
		Prefix: "MIHOYOSDK_ADL_PROD_CN",
	},
	"ZZZCN": {
		Path:   `Software\miHoYo\绝区零`,
		Prefix: "MIHOYOSDK_ADL_PROD_CN",
	},
}

// findActualKeyName 自動在註冊表路徑下尋找匹配前綴的真實鍵名
func findActualKeyName(k registry.Key, prefix string) (string, error) {
	names, err := k.ReadValueNames(0)
	if err != nil {
		return "", err
	}
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			return name, nil
		}
	}
	return "", fmt.Errorf("找不到以 %s 開頭的註冊表鍵", prefix)
}

func ReadToken(gameID string) ([]byte, error) {
	config, ok := GamePaths[gameID]
	if !ok {
		return nil, fmt.Errorf("未定義的遊戲 ID: %s", gameID)
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, config.Path, registry.READ)
	if err != nil {
		return nil, fmt.Errorf("無法打開註冊表路徑: %v", err)
	}
	defer k.Close()

	// 動態獲取鍵名，解決 h3123548890 可能變動的問題
	actualKey, err := findActualKeyName(k, config.Prefix)
	if err != nil {
		return nil, err
	}

	val, _, err := k.GetBinaryValue(actualKey)
	if err != nil {
		return nil, fmt.Errorf("讀取鍵值失敗 (%s): %v", actualKey, err)
	}
	return val, nil
}

func WriteToken(gameID string, tokenBytes []byte) error {
	config, ok := GamePaths[gameID]
	if !ok {
		return fmt.Errorf("未定義的遊戲 ID: %s", gameID)
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, config.Path, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("無法打開註冊表路徑進行寫入: %v", err)
	}
	defer k.Close()

	actualKey, err := findActualKeyName(k, config.Prefix)
	if err != nil {
		return err
	}

	return k.SetBinaryValue(actualKey, tokenBytes)
}
