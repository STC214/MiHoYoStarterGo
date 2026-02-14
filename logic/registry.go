package logic

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type GameRegConfig struct {
	Path   string
	Prefix string // 改用前缀匹配，增加兼容性
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

// findActualKeyName 自动在注册表路径下寻找匹配前缀的真实键名
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
	return "", fmt.Errorf("找不到以 %s 开头的注册表键", prefix)
}

func ReadToken(gameID string) ([]byte, error) {
	config, ok := GamePaths[gameID]
	if !ok {
		return nil, fmt.Errorf("未定义的游戏 ID: %s", gameID)
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, config.Path, registry.READ)
	if err != nil {
		return nil, fmt.Errorf("无法打开注册表路径: %v", err)
	}
	defer k.Close()

	// 动态获取键名，解决 h3123548890 可能变动的问题
	actualKey, err := findActualKeyName(k, config.Prefix)
	if err != nil {
		return nil, err
	}

	val, _, err := k.GetBinaryValue(actualKey)
	if err != nil {
		return nil, fmt.Errorf("读取键值失败 (%s): %v", actualKey, err)
	}
	return val, nil
}

func WriteToken(gameID string, tokenBytes []byte) error {
	config, ok := GamePaths[gameID]
	if !ok {
		return fmt.Errorf("未定义的游戏 ID: %s", gameID)
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, config.Path, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("无法打开注册表路径进行写入: %v", err)
	}
	defer k.Close()

	actualKey, err := findActualKeyName(k, config.Prefix)
	if err != nil {
		return err
	}

	return k.SetBinaryValue(actualKey, tokenBytes)
}
