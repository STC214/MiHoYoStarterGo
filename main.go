package main

import (
	"context" // [修正] 補上遺漏的導入
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"MiHoYoStarterGo/logic"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. 預加載配置以獲取上次保存的窗口狀態
	cfg, err := logic.LoadConfig()
	if err != nil {
		fmt.Println("加載配置失敗，將使用默認窗口大小:", err)
	}

	// 設置默認值
	if cfg.WindowWidth <= 0 {
		cfg.WindowWidth = 1024
	}
	if cfg.WindowHeight <= 0 {
		cfg.WindowHeight = 768
	}

	// 創建你的 App 實例
	app := NewApp()

	err = wails.Run(&options.App{
		Title:  "米哈遊啟動器增強版",
		Width:  cfg.WindowWidth,
		Height: cfg.WindowHeight,
		// [修正] X 和 Y 在 options.App 根層級是無效的，直接移除或放在 Windows 配置中
		// 我們這裡移除它們，改為在啟動後通過 runtime 設置位置（或使用 Bind 方式）
		// 如果你的 Wails 版本支持直接在 Windows 下設置，請參考官方文檔
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 43, G: 43, B: 43, A: 1},
		OnStartup:        app.startup,
		// [修正] 補全 context.Context 類型定義
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			x, y := runtime.WindowGetPosition(ctx)
			w, h := runtime.WindowGetSize(ctx)

			// 再次讀取配置以防運行期間有變動
			currentCfg, err := logic.LoadConfig()
			if err == nil {
				currentCfg.WindowX = x
				currentCfg.WindowY = y
				currentCfg.WindowWidth = w
				currentCfg.WindowHeight = h
				_ = logic.SaveConfig(currentCfg)
				fmt.Printf(">> [系統] 已記錄窗口狀態: %dx%d 坐標(%d,%d)\n", w, h, x, y)
			}
			return false
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
