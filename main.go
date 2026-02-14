package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建你的 App 实例
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "米哈遊啟動器增強版",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 43, G: 43, B: 43, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app, // 这样前端就能调用 app.go 里的方法了
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
