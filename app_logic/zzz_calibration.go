package app_logic

import (
	"MiHoYoStarterGo/logic"
	"context"
	"fmt"
	"time"

	"github.com/lxn/win"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func CalibrateZZZPoints(ctx context.Context) string {
	hwnd := logic.GetWindowHandleByGameID("ZZZCN")
	if hwnd == 0 {
		return "ZZZ_WINDOW_NOT_FOUND"
	}

	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)
	left := int(rect.Left)
	top := int(rect.Top)

	type step struct {
		key   string
		label string
	}
	steps := []step{
		{key: "account", label: "账号输入框"},
		{key: "password", label: "密码输入框"},
		{key: "agreement", label: "同意协议勾选点"},
		{key: "enter", label: "进入游戏点击点"},
	}

	points := map[string]logic.Point{}
	runtime.EventsEmit(ctx, "zzz_calibration_progress", map[string]any{
		"phase": "start",
		"step":  0,
		"total": len(steps),
		"text":  "开始绝区零坐标标定（倒计时自动记录鼠标位置）",
	})

	for i, s := range steps {
		runtime.EventsEmit(ctx, "zzz_calibration_progress", map[string]any{
			"phase": "prompt",
			"step":  i + 1,
			"total": len(steps),
			"label": s.label,
			"text":  fmt.Sprintf("步骤 %d/%d：请在 3 秒内将鼠标移到“%s”", i+1, len(steps), s.label),
		})
		p, err := logic.CapturePointWithCountdown(ctx, left, top, width, height, 3*time.Second)
		if err != nil {
			runtime.EventsEmit(ctx, "zzz_calibration_progress", map[string]any{
				"phase": "error",
				"step":  i + 1,
				"total": len(steps),
				"text":  "标定失败：倒计时采集坐标超时",
			})
			return "CAPTURE_TIMEOUT"
		}
		points[s.key] = p
		runtime.EventsEmit(ctx, "zzz_calibration_progress", map[string]any{
			"phase": "captured",
			"step":  i + 1,
			"total": len(steps),
			"label": s.label,
			"x":     p.X,
			"y":     p.Y,
			"text":  fmt.Sprintf("步骤 %d/%d 已记录：%s", i+1, len(steps), s.label),
		})
	}

	profile := logic.ZZZPointProfile{
		Width:     width,
		Height:    height,
		Account:   points["account"],
		Password:  points["password"],
		Agreement: points["agreement"],
		Enter:     points["enter"],
	}
	if err := logic.ValidateZZZProfile(profile); err != nil {
		return "INVALID_PROFILE"
	}
	if err := logic.SaveZZZPointProfile(profile); err != nil {
		return "SAVE_FAILED"
	}

	runtime.EventsEmit(ctx, "zzz_calibration_progress", map[string]any{
		"phase": "done",
		"step":  len(steps),
		"total": len(steps),
		"text":  "绝区零坐标标定完成并已保存",
	})
	return "SUCCESS"
}
