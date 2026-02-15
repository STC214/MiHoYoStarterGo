package app_logic

import (
	"MiHoYoStarterGo/logic"
	"fmt"
	"image/png"
	"os"
	"strings"
	"time"
)

func CaptureDebugWindow(gameID string) string {
	hwnd := logic.GetWindowHandleByGameID(gameID)
	if hwnd == 0 {
		return "FAILED_CAPTURE"
	}

	img, err := logic.CaptureWindowByHandle(hwnd)
	if err != nil {
		return "FAILED_CAPTURE"
	}

	safeID := strings.ToLower(strings.ReplaceAll(gameID, " ", "_"))
	fileName := fmt.Sprintf("debug_capture_%s_%d.png", safeID, time.Now().Unix())
	f, err := os.Create(fileName)
	if err != nil {
		return "FAILED_WRITE"
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "FAILED_WRITE"
	}
	return fileName
}
