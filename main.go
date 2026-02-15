package main

import (
	"context" // [修正] 补上遗漏的导入
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
	// 1. 预加载配置以获取上次保存的窗口状态
	cfg, err := logic.LoadConfig()
	if err != nil {
		fmt.Println("加载配置失败，将使用默认窗口大小:", err)
	}

	// 设置默认值
	if cfg.WindowWidth <= 0 {
		cfg.WindowWidth = 1024
	}
	if cfg.WindowHeight <= 0 {
		cfg.WindowHeight = 768
	}

	// 创建你的 App 实例
	app := NewApp()

	err = wails.Run(&options.App{
		Title:  "米哈游启动器增强版",
		Width:  cfg.WindowWidth,
		Height: cfg.WindowHeight,
		// [修正] X 和 Y 在 options.App 根层级是无效的，直接移除或放在 Windows 配置中
		// 我们这里移除它们，改为在启动後通过 runtime 设置位置（或使用 Bind 方式）
		// 如果你的 Wails 版本支持直接在 Windows 下设置，请参考官方文档
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 43, G: 43, B: 43, A: 1},
		OnStartup:        app.startup,
		// [修正] 补全 context.Context 类型定义
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			x, y := runtime.WindowGetPosition(ctx)
			w, h := runtime.WindowGetSize(ctx)

			// 再次读取配置以防运行期间有变动
			currentCfg, err := logic.LoadConfig()
			if err == nil {
				currentCfg.WindowX = x
				currentCfg.WindowY = y
				currentCfg.WindowWidth = w
				currentCfg.WindowHeight = h
				_ = logic.SaveConfig(currentCfg)
				fmt.Printf(">> [系统] 已记录窗口状态: %dx%d 坐标(%d,%d)\n", w, h, x, y)
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
