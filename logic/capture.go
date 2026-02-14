package logic

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	modUser32       = syscall.NewLazyDLL("user32.dll")
	procPrintWindow = modUser32.NewProc("PrintWindow")
)

// PrintWindow 手动调用 user32.dll 中的函数
func PrintWindow(hwnd win.HWND, hdc win.HDC, nFlags uint32) bool {
	ret, _, _ := procPrintWindow.Call(
		uintptr(hwnd),
		uintptr(hdc),
		uintptr(nFlags),
	)
	return ret != 0
}

func CaptureWindow(windowName string) (*image.RGBA, error) {
	// 1. 查找窗口句柄
	hwnd := win.FindWindow(nil, syscall.StringToUTF16Ptr(windowName))
	if hwnd == 0 {
		return nil, fmt.Errorf("未找到窗口: %s", windowName)
	}

	// 2. 获取实际窗口大小
	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("窗口大小异常")
	}

	// 3. 准备绘图上下文
	hdcScreen := win.GetDC(hwnd)
	defer win.ReleaseDC(hwnd, hdcScreen)

	hdcMem := win.CreateCompatibleDC(hdcScreen)
	defer win.DeleteDC(hdcMem)

	hBitmap := win.CreateCompatibleBitmap(hdcScreen, width, height)
	defer win.DeleteObject(win.HGDIOBJ(hBitmap))

	win.SelectObject(hdcMem, win.HGDIOBJ(hBitmap))

	// 4. 使用我们自定义的 PrintWindow
	// 参数 2 (PW_RENDERFULLCONTENT) 尝试抓取硬件加速内容
	PrintWindow(hwnd, hdcMem, 2)

	// 5. 转换为 image.RGBA
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	var bi win.BITMAPINFO
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
	bi.BmiHeader.BiWidth = width
	bi.BmiHeader.BiHeight = -height // 负值确保图像不是颠倒的
	bi.BmiHeader.BiPlanes = 1
	bi.BmiHeader.BiBitCount = 32
	bi.BmiHeader.BiCompression = win.BI_RGB

	win.GetDIBits(hdcScreen, hBitmap, 0, uint32(height), (*byte)(unsafe.Pointer(&img.Pix[0])), &bi, win.DIB_RGB_COLORS)

	// 6. 修正 Alpha 通道，确保图片不透明
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}

	return img, nil
}
