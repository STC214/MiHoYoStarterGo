//go:build windows

package main

import (
	_ "embed"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/lxn/win"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

var (
	trayHookMu      sync.Mutex
	trayHookReady   bool
	trayHookOrigWnd uintptr
	trayHookApp     *App
	trayWndProcPtr  = syscall.NewCallback(trayWindowProc)
)

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

	go a.installTrayDoubleClickRestoreHook()
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

func (a *App) installTrayDoubleClickRestoreHook() {
	const (
		tries             = 80
		wmSystrayMessage  = 0x0401 // WM_USER + 1, consistent with getlantern/systray windows impl
		wmLButtonDblClk   = 0x0203
		classNameSystray  = "SystrayClass"
		hookRetryInterval = 250 * time.Millisecond
	)

	classPtr, _ := syscall.UTF16PtrFromString(classNameSystray)
	for i := 0; i < tries; i++ {
		hwnd := win.FindWindow(classPtr, nil)
		if hwnd != 0 {
			oldProc := win.SetWindowLongPtr(hwnd, win.GWLP_WNDPROC, trayWndProcPtr)
			if oldProc != 0 {
				trayHookMu.Lock()
				trayHookOrigWnd = uintptr(oldProc)
				trayHookApp = a
				trayHookReady = true
				trayHookMu.Unlock()
				return
			}
		}
		time.Sleep(hookRetryInterval)
	}

	_ = wmSystrayMessage
	_ = wmLButtonDblClk
}

func trayWindowProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	const (
		wmSystrayMessage = 0x0401
		wmLButtonDblClk  = 0x0203
	)

	trayHookMu.Lock()
	app := trayHookApp
	orig := trayHookOrigWnd
	ready := trayHookReady
	trayHookMu.Unlock()

	if ready && msg == wmSystrayMessage && lParam == wmLButtonDblClk && app != nil && app.ctx != nil {
		go func() {
			runtime.Show(app.ctx)
			runtime.WindowUnminimise(app.ctx)
			runtime.WindowShow(app.ctx)
		}()
	}

	if orig != 0 {
		return win.CallWindowProc(orig, hwnd, msg, wParam, lParam)
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}
