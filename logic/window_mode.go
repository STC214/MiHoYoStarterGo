package logic

import (
	"math"
	"unsafe"

	"github.com/lxn/win"
)

// IsWindowFullscreen checks whether the window bounds match its monitor bounds.
func IsWindowFullscreen(hwnd win.HWND) bool {
	if hwnd == 0 {
		return false
	}

	var windowRect win.RECT
	if !win.GetWindowRect(hwnd, &windowRect) {
		return false
	}

	monitor := win.MonitorFromWindow(hwnd, win.MONITOR_DEFAULTTONEAREST)
	if monitor == 0 {
		return false
	}

	monitorInfo := win.MONITORINFO{
		CbSize: uint32(unsafe.Sizeof(win.MONITORINFO{})),
	}
	if !win.GetMonitorInfo(monitor, &monitorInfo) {
		return false
	}

	// Keep a small tolerance for borderless fullscreen or DPI rounding.
	const tolerance = 2.0
	return math.Abs(float64(windowRect.Left-monitorInfo.RcMonitor.Left)) <= tolerance &&
		math.Abs(float64(windowRect.Top-monitorInfo.RcMonitor.Top)) <= tolerance &&
		math.Abs(float64(windowRect.Right-monitorInfo.RcMonitor.Right)) <= tolerance &&
		math.Abs(float64(windowRect.Bottom-monitorInfo.RcMonitor.Bottom)) <= tolerance
}
