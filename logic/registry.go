package logic

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type GameRegConfig struct {
	Path   string
	Prefix string
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
	return "", fmt.Errorf("cannot find registry value with prefix %s", prefix)
}

func ReadToken(gameID string) ([]byte, error) {
	config, ok := GamePaths[gameID]
	if !ok {
		return nil, fmt.Errorf("unknown game id: %s", gameID)
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, config.Path, registry.READ)
	if err != nil {
		return nil, fmt.Errorf("cannot open registry key: %w", err)
	}
	defer k.Close()

	actualKey, err := findActualKeyName(k, config.Prefix)
	if err != nil {
		return nil, err
	}

	val, _, err := k.GetBinaryValue(actualKey)
	if err != nil {
		return nil, fmt.Errorf("cannot read binary value %s: %w", actualKey, err)
	}
	return val, nil
}

func WriteToken(gameID string, tokenBytes []byte) error {
	config, ok := GamePaths[gameID]
	if !ok {
		return fmt.Errorf("unknown game id: %s", gameID)
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, config.Path, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("cannot open registry key for write: %w", err)
	}
	defer k.Close()

	actualKey, err := findActualKeyName(k, config.Prefix)
	if err != nil {
		return err
	}

	return k.SetBinaryValue(actualKey, tokenBytes)
}
