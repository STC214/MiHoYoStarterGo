//go:build windows

package main

import (
	_ "embed"
	"time"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

func (a *App) initSystemTray() {
	go systray.Run(func() {
		systray.SetIcon(trayIcon)
		systray.SetTitle("MiHoYoStarterGo")
		systray.SetTooltip("MiHoYoStarterGo")

		showItem := systray.AddMenuItem("显示主界面", "恢复主界面")
		systray.AddSeparator()
		quitItem := systray.AddMenuItem("退出", "退出程序")

		go func() {
			for range showItem.ClickedCh {
				if a.ctx == nil {
					continue
				}
				runtime.Show(a.ctx)
				runtime.WindowUnminimise(a.ctx)
				runtime.WindowShow(a.ctx)
			}
		}()

		go func() {
			for range quitItem.ClickedCh {
				if a.ctx != nil {
					runtime.Quit(a.ctx)
				}
				systray.Quit()
				return
			}
		}()
	}, func() {})

	go a.minimiseToTrayLoop()
}

func (a *App) minimiseToTrayLoop() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if a.ctx == nil {
			continue
		}
		if runtime.WindowIsMinimised(a.ctx) {
			runtime.Hide(a.ctx)
		}
	}
}
