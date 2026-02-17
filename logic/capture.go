package logic

import (
	"fmt"
	"image"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	modUser32               = syscall.NewLazyDLL("user32.dll")
	procPrintWindow         = modUser32.NewProc("PrintWindow")
	procEnumWindows         = modUser32.NewProc("EnumWindows")
	procGetWindowThreadPID  = modUser32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible     = modUser32.NewProc("IsWindowVisible")
	procGetWindowTextLength = modUser32.NewProc("GetWindowTextLengthW")
)

func PrintWindow(hwnd win.HWND, hdc win.HDC, nFlags uint32) bool {
	ret, _, _ := procPrintWindow.Call(uintptr(hwnd), uintptr(hdc), uintptr(nFlags))
	return ret != 0
}

func GetWindowHandleByProcessName(exeName string) win.HWND {
	var targetHwnd win.HWND
	processes, _ := process.Processes()
	var targetPid uint32

	for _, p := range processes {
		name, _ := p.Name()
		if strings.EqualFold(name, exeName) {
			targetPid = uint32(p.Pid)
			break
		}
	}
	if targetPid == 0 {
		return 0
	}

	cb := syscall.NewCallback(func(hwnd win.HWND, lParam uintptr) uintptr {
		var windowPid uint32
		procGetWindowThreadPID.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&windowPid)))
		visible, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
		if windowPid == targetPid && visible != 0 {
			textLen, _, _ := procGetWindowTextLength.Call(uintptr(hwnd))
			if textLen > 0 {
				targetHwnd = hwnd
				return 0
			}
		}
		return 1
	})

	procEnumWindows.Call(cb, 0)
	return targetHwnd
}

func GetWindowHandleByGameID(gameID string) win.HWND {
	if gameID == "StarRailCN" {
		return GetWindowHandleByProcessName("StarRail.exe")
	}
	if gameID == "ZZZCN" {
		return GetWindowHandleByProcessName("ZenlessZoneZero.exe")
	}

	title := "原神"
	if gameID == "GenshinOS" {
		title = "Genshin Impact"
	}
	return win.FindWindow(nil, syscall.StringToUTF16Ptr(title))
}

func CaptureWindow(windowName string) (*image.RGBA, error) {
	hwnd := win.FindWindow(nil, syscall.StringToUTF16Ptr(windowName))
	if hwnd == 0 {
		return nil, fmt.Errorf("window not found: %s", windowName)
	}
	return CaptureWindowByHandle(hwnd)
}

func CaptureWindowByHandle(hwnd win.HWND) (*image.RGBA, error) {
	if hwnd == 0 {
		return nil, fmt.Errorf("invalid window handle")
	}

	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid window size")
	}

	hdcScreen := win.GetDC(hwnd)
	defer win.ReleaseDC(hwnd, hdcScreen)

	hdcMem := win.CreateCompatibleDC(hdcScreen)
	defer win.DeleteDC(hdcMem)

	hBitmap := win.CreateCompatibleBitmap(hdcScreen, width, height)
	defer win.DeleteObject(win.HGDIOBJ(hBitmap))

	win.SelectObject(hdcMem, win.HGDIOBJ(hBitmap))
	PrintWindow(hwnd, hdcMem, 2)

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	var bi win.BITMAPINFO
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
	bi.BmiHeader.BiWidth = width
	bi.BmiHeader.BiHeight = -height
	bi.BmiHeader.BiPlanes = 1
	bi.BmiHeader.BiBitCount = 32
	bi.BmiHeader.BiCompression = win.BI_RGB

	win.GetDIBits(hdcScreen, hBitmap, 0, uint32(height), (*byte)(unsafe.Pointer(&img.Pix[0])), &bi, win.DIB_RGB_COLORS)
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}
	return img, nil
}
