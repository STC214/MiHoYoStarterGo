package app_logic

import (
	"MiHoYoStarterGo/logic"
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func GetSettings() *logic.ConfigData {
	cfg, _ := logic.LoadConfig()
	return cfg
}

func SaveTheme(theme string) {
	cfg, _ := logic.LoadConfig()
	cfg.Theme = theme
	logic.SaveConfig(cfg)
}

func SaveGamePaths(p map[string]string) {
	cfg, _ := logic.LoadConfig()
	cfg.GamePaths = p
	logic.SaveConfig(cfg)
}

func SelectGameFile(ctx context.Context) string {
	res, _ := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title:   "选择游戏执行文件",
		Filters: []runtime.FileFilter{{DisplayName: "EXE", Pattern: "*.exe"}},
	})
	return res
}

func ExportBackup() string {
	cfg, _ := logic.LoadConfig()
	fileName, _ := logic.ExportPlaintextBackup(cfg)
	return fileName
}
