package logic

import (
	"context"
	"encoding/hex"
	"fmt"
	"image/png"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-vgo/robotgo"
	"github.com/lxn/win"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func StartAutomationMonitor(ctx context.Context, gameID, user, pwd string, isFirst bool, pause *bool, cancel *bool) {
	_ = isFirst
	rand.Seed(time.Now().UnixNano())

	go func() {
		fmt.Printf("[系统] 监控启动: game=%s user=%s\n", gameID, user)
		runtime.EventsEmit(ctx, "monitor_status", "流程已启动，正在等待游戏窗口...")
		tmpImgPath := fmt.Sprintf("temp_%d.png", time.Now().UnixNano())
		starRailSwitched := false
		lastOCRSnapshot := ""
		lastStatusTip := ""

		ticker := time.NewTicker(300 * time.Millisecond)
		defer func() {
			ticker.Stop()
			_ = os.Remove(tmpImgPath)
		}()

		for range ticker.C {
			if *cancel {
				fmt.Println("[系统] 收到取消信号，结束监控")
				runtime.EventsEmit(ctx, "monitor_finished", "CANCELLED")
				return
			}
			if *pause {
				continue
			}
			if !IsGameRunning(gameID) {
				continue
			}

			hwnd := GetWindowHandleByGameID(gameID)
			if hwnd == 0 {
				continue
			}

			win.SetForegroundWindow(hwnd)
			var rect win.RECT
			win.GetWindowRect(hwnd, &rect)
			time.Sleep(100 * time.Millisecond)

			img, err := CaptureWindowByHandle(hwnd)
			if err != nil {
				continue
			}
			f, err := os.Create(tmpImgPath)
			if err != nil {
				continue
			}
			_ = png.Encode(f, img)
			_ = f.Close()

			if gameID == "StarRailCN" {
				if df, derr := os.Create("debug_capture_starrail.png"); derr == nil {
					_ = png.Encode(df, img)
					_ = df.Close()
				}
			}

			textPoints, err := RecognizeWithPos(tmpImgPath)
			if err != nil {
				continue
			}
			if len(textPoints) > 0 {
				tip := fmt.Sprintf("OCR识别成功：%d 条文本", len(textPoints))
				if tip != lastStatusTip {
					lastStatusTip = tip
					runtime.EventsEmit(ctx, "monitor_status", tip)
				}
			}

			snapshot := buildOCRSnapshot(textPoints, 24)
			if snapshot != lastOCRSnapshot {
				lastOCRSnapshot = snapshot
				fmt.Printf("[OCR][%s]\n%s\n", gameID, snapshot)
			}

			if gameID == "StarRailCN" && !starRailSwitched && isStarRailVerifyPage(textPoints) {
				if x, y, ok := findKeywordCenter(textPoints, []string{"账号密码", "賬號密碼"}); ok {
					left, top := int(rect.Left), int(rect.Top)
					fmt.Println("[StarRail] 检测到验证码页，点击“账号密码”切换登录方式")
					runtime.EventsEmit(ctx, "monitor_status", "已识别验证码页，正在切换到账号密码登录...")
					randClick(left+x, top+y, 8, 4)
					starRailSwitched = true
					time.Sleep(500 * time.Millisecond)
					continue
				}
			}

			if isLoginPage(gameID, textPoints) {
				fmt.Println("[流水线] 检测到登录界面，开始填充账号密码")
				runtime.EventsEmit(ctx, "monitor_status", "已识别登录界面，正在自动填充账号密码...")
				executeFullSequenceByHandle(hwnd, textPoints, user, pwd)
			}

			windowHeight := int(rect.Bottom - rect.Top)
			if isConfirmedImageA(textPoints, windowHeight) {
				fmt.Println("========================================")
				fmt.Println("[成功] 检测到登录成功，开始写回账号数据")
				runtime.EventsEmit(ctx, "monitor_status", "已识别登录成功，正在写回账号数据...")

				err := finalizeAccountStorage(gameID, user)
				if err != nil {
					fmt.Printf("[错误] 数据写回失败: %v\n", err)
					runtime.EventsEmit(ctx, "monitor_finished", "FAILED")
				} else {
					fmt.Println("[成功] config.json 已更新")
					runtime.EventsEmit(ctx, "monitor_finished", "SUCCESS")
				}
				fmt.Println("========================================")
				return
			}
		}
	}()
}

func isStarRailVerifyPage(points []TextPoint) bool {
	hasCode := hasKeyword(points, "验证码")
	hasSend := hasKeyword(points, "发送")
	hasAccountPwd := hasAnyKeyword(points, []string{"账号密码", "賬號密碼"})
	return hasCode && hasSend && hasAccountPwd
}

func isLoginPage(gameID string, points []TextPoint) bool {
	if gameID == "StarRailCN" {
		hasPhoneInput := hasAnyKeyword(points, []string{"输入手机号/邮箱", "输入手机号", "手机号/邮箱", "輸入手機號/郵箱", "輸入手機號"})
		hasPwdInput := hasAnyKeyword(points, []string{"输入密码", "輸入密碼", "密码", "密碼"})
		hasEnter := hasAnyKeyword(points, []string{"进入游戏", "進入遊戲"})
		hasForgot := hasAnyKeyword(points, []string{"忘记密码", "忘記密碼"})
		return hasPhoneInput && hasPwdInput && hasEnter && hasForgot
	}

	hasAccount := hasAnyKeyword(points, []string{"手机号", "手機號", "邮箱", "郵箱", "输入手机号", "輸入手機號"})
	hasAgreement := hasAnyKeyword(points, []string{"同意", "已阅读", "已閱讀"})
	return hasAccount && hasAgreement
}

func hasAnyKeyword(points []TextPoint, keywords []string) bool {
	for _, p := range points {
		for _, kw := range keywords {
			if strings.Contains(p.Text, kw) {
				return true
			}
		}
	}
	return false
}

func hasKeyword(points []TextPoint, keyword string) bool {
	for _, p := range points {
		if strings.Contains(p.Text, keyword) {
			return true
		}
	}
	return false
}

func findKeywordCenter(points []TextPoint, keywords []string) (int, int, bool) {
	for _, p := range points {
		for _, kw := range keywords {
			if strings.Contains(p.Text, kw) {
				return p.X, p.Y, true
			}
		}
	}
	return 0, 0, false
}

func buildOCRSnapshot(points []TextPoint, maxItems int) string {
	if len(points) == 0 {
		return "(empty)"
	}
	if maxItems <= 0 {
		maxItems = len(points)
	}

	items := make([]string, 0, maxItems+1)
	for i, p := range points {
		if i >= maxItems {
			items = append(items, fmt.Sprintf("... (%d more)", len(points)-maxItems))
			break
		}
		items = append(items, fmt.Sprintf("[%d] \"%s\" @(%d,%d)", i+1, p.Text, p.X, p.Y))
	}
	return strings.Join(items, "\n")
}

func isConfirmedImageA(points []TextPoint, windowHeight int) bool {
	bottomThreshold := (windowHeight / 8) * 7
	hasTargetWord := false
	hasInterference := false

	for _, p := range points {
		if p.Y > bottomThreshold {
			txt := p.Text
			if strings.Contains(txt, "进入游戏") || strings.Contains(txt, "点击进入") || strings.Contains(txt, "進入遊戲") {
				hasTargetWord = true
			} else if containsChinese(txt) {
				hasInterference = true
			}
		}
	}
	return hasTargetWord && !hasInterference
}

func finalizeAccountStorage(gameID, username string) error {
	time.Sleep(2 * time.Second)

	tokenBytes, err := ReadToken(gameID)
	if err != nil {
		return err
	}
	tokenHex := hex.EncodeToString(tokenBytes)

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	found := false
	for i, acc := range cfg.Accounts {
		if acc.Username == username && acc.GameID == gameID {
			cfg.Accounts[i].Token = tokenHex
			cfg.Accounts[i].DeviceFingerprint = GetDeviceFingerprint()
			cfg.Accounts[i].IsFirstLogin = false
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("account not found in config")
	}
	return SaveConfig(cfg)
}

func executeFullSequenceByHandle(hwnd win.HWND, points []TextPoint, user, pwd string) {
	if hwnd == 0 {
		return
	}

	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	left, top := int(rect.Left), int(rect.Top)

	filledUser := false
	filledPwd := false

	for _, p := range points {
		if !filledUser && hasAnyKeyword([]TextPoint{p}, []string{"手机号", "手機號", "邮箱", "郵箱", "输入手机号", "輸入手機號", "账号密码", "賬號密碼"}) {
			randClick(left+p.X, top+p.Y, 10, 2)
			typeAction(user)
			filledUser = true
		}
		if !filledPwd && hasAnyKeyword([]TextPoint{p}, []string{"密码", "密碼", "输入密码", "輸入密碼"}) && !hasAnyKeyword([]TextPoint{p}, []string{"忘记", "忘記"}) {
			randClick(left+p.X, top+p.Y, 10, 2)
			typeAction(pwd)
			filledPwd = true
		}
	}

	for _, p := range points {
		if strings.Contains(p.Text, "同意") || strings.Contains(p.Text, "已阅读") || strings.Contains(p.Text, "已閱讀") {
			if strings.HasPrefix(p.Text, "①") || strings.HasPrefix(p.Text, "☐") || strings.HasPrefix(p.Text, "□") || strings.HasPrefix(p.Text, "◯") || strings.HasPrefix(p.Text, "○") {
				charCount := utf8.RuneCountInString(p.Text)
				if charCount > 0 {
					singleCharWidth := p.Width / charCount
					randClickInCircle(left+p.LeftX+(singleCharWidth/2), top+p.Y, 4)
					time.Sleep(400 * time.Millisecond)
				}
			}
		}
	}

	for _, p := range points {
		if (strings.Contains(p.Text, "进入") || strings.Contains(p.Text, "進入") || strings.Contains(p.Text, "登录") || strings.Contains(p.Text, "登錄") || strings.Contains(p.Text, "开始") || strings.Contains(p.Text, "開始")) && len(p.Text) <= 16 {
			randClick(left+p.X, top+p.Y+5, 15, 5)
		}
	}
}

func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FA5 {
			return true
		}
	}
	return false
}

func randClickInCircle(x, y, radius int) {
	r := float64(radius) * math.Sqrt(rand.Float64())
	theta := rand.Float64() * 2 * math.Pi
	robotgo.Move(x+int(r*math.Cos(theta)), y+int(r*math.Sin(theta)))
	robotgo.Click("left", false)
}

func randClick(x, y, rx, ry int) {
	robotgo.Move(x+rand.Intn(rx*2)-rx, y+rand.Intn(ry*2)-ry)
	robotgo.Click("left", false)
}

func typeAction(s string) {
	time.Sleep(200 * time.Millisecond)
	robotgo.KeyTap("a", "control")
	robotgo.KeyTap("backspace")
	time.Sleep(100 * time.Millisecond)
	robotgo.TypeStr(s)
}
