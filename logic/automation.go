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
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/go-vgo/robotgo"
	"github.com/lxn/win"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartAutomationMonitor 核心监控逻辑
// 新增了 cancel *bool 参数用于接收用户的手动停止指令
func StartAutomationMonitor(ctx context.Context, gameID, user, pwd string, isFirst bool, pause *bool, cancel *bool) {
	rand.Seed(time.Now().UnixNano())

	go func() {
		fmt.Printf("[系统] 监控启动：正在为账号 %s 监视登录状态...\n", user)
		tmpImgPath := fmt.Sprintf("temp_%d.png", time.Now().Unix()) // 使用动态文件名避免冲突

		// 根据游戏 ID 确定窗口标题
		windowName := "原神"
		if gameID == "StarRailCN" {
			windowName = "崩坏：星穹铁道"
		} else if gameID == "ZZZCN" {
			windowName = "绝区零"
		}

		ticker := time.NewTicker(300 * time.Millisecond)
		defer func() {
			ticker.Stop()
			os.Remove(tmpImgPath) // 协程结束时清理临时图片资源
		}()

		for range ticker.C {
			// 1. [新增] 检查手动取消信号
			if *cancel {
				fmt.Printf("[系统] 接收到用户终止信号，已回收 OCR 资源并关闭协程。\n")
				return // 彻底退出协程
			}

			// 2. 检查暂停状态
			if *pause {
				continue
			}

			// 3. 检查进程是否运行
			if !IsGameRunning(gameID) {
				continue
			}

			// 4. 寻找窗口句柄并置顶
			hwnd := win.FindWindow(nil, syscall.StringToUTF16Ptr(windowName))
			var rect win.RECT
			if hwnd != 0 {
				win.SetForegroundWindow(hwnd)
				win.GetWindowRect(hwnd, &rect)
				time.Sleep(100 * time.Millisecond)
			} else {
				continue
			}

			// 5. 截图
			img, err := CaptureWindow(windowName)
			if err != nil {
				continue
			}
			f, err := os.Create(tmpImgPath)
			if err != nil {
				continue
			}
			png.Encode(f, img)
			f.Close()

			// 6. OCR 识别
			textPoints, err := RecognizeWithPos(tmpImgPath)
			if err != nil {
				continue
			}

			// 7. 判定画面 B (登录框) 并执行填充
			if isImageBStrict(textPoints) {
				fmt.Println("[流水线] 检测到登录界面，开始执行账密填充...")
				executeFullSequence(windowName, textPoints, user, pwd)
			}

			// 8. 判定画面 A (进入游戏成功) 并提取 Token
			windowHeight := int(rect.Bottom - rect.Top)
			if isConfirmedImageA(textPoints, windowHeight) {
				fmt.Println("========================================")
				fmt.Println(" [验证成功] 已登录成功，正在提取数据...")

				err := finalizeAccountStorage(gameID, user)

				if err != nil {
					fmt.Printf(" [错误] 数据存档失败: %v\n", err)
					runtime.EventsEmit(ctx, "monitor_finished", "FAILED")
				} else {
					fmt.Println(" [成功] 数据已更新至 config.json")
					runtime.EventsEmit(ctx, "monitor_finished", "SUCCESS")
				}
				fmt.Println("========================================")
				return // 任务完成，退出协程
			}
		}
	}()
}

// ---------------- 判定与数据存储逻辑 ----------------

func isImageBStrict(points []TextPoint) bool {
	hasAccountField := false
	hasAgreement := false
	for _, p := range points {
		t := p.Text
		if strings.Contains(t, "手机号") || strings.Contains(t, "邮箱") {
			hasAccountField = true
		}
		if strings.Contains(t, "同意") || strings.Contains(t, "已阅读") {
			hasAgreement = true
		}
	}
	return hasAccountField && hasAgreement
}

func isConfirmedImageA(points []TextPoint, windowHeight int) bool {
	bottomThreshold := (windowHeight / 8) * 7
	hasTargetWord := false
	hasInterference := false

	for _, p := range points {
		if p.Y > bottomThreshold {
			txt := p.Text
			if strings.Contains(txt, "进入游戏") || strings.Contains(txt, "点击进入") {
				hasTargetWord = true
			} else if containsChinese(txt) {
				hasInterference = true
			}
		}
	}
	return hasTargetWord && !hasInterference
}

func finalizeAccountStorage(gameID, username string) error {
	// 延迟读取注册表，确保游戏已经把最新数据写进去
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

// ---------------- 键鼠模拟执行 ----------------

func executeFullSequence(windowName string, points []TextPoint, user, pwd string) {
	hwnd := win.FindWindow(nil, syscall.StringToUTF16Ptr(windowName))
	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	left, top := int(rect.Left), int(rect.Top)

	// 1. 填写账号密码
	for _, p := range points {
		if strings.Contains(p.Text, "手机号") || strings.Contains(p.Text, "邮箱") {
			randClick(left+p.X, top+p.Y, 10, 2)
			typeAction(user)
		}
		if strings.Contains(p.Text, "密码") && !strings.Contains(p.Text, "忘记") {
			randClick(left+p.X, top+p.Y, 10, 2)
			typeAction(pwd)
		}
	}

	// 2. 勾选协议 (找特定的圆形符号)
	for _, p := range points {
		if strings.Contains(p.Text, "同意") && (strings.HasPrefix(p.Text, "①") || strings.HasPrefix(p.Text, "○") || strings.HasPrefix(p.Text, "〇")) {
			charCount := utf8.RuneCountInString(p.Text)
			if charCount > 0 {
				singleCharWidth := p.Width / charCount
				randClickInCircle(left+p.LeftX+(singleCharWidth/2), top+p.Y, 4)
				time.Sleep(400 * time.Millisecond)
			}
		}
	}

	// 3. 点击进入/登录
	for _, p := range points {
		if (strings.Contains(p.Text, "进入") || strings.Contains(p.Text, "开始")) && len(p.Text) <= 12 {
			randClick(left+p.X, top+p.Y+5, 15, 5)
		}
	}
}

// ---------------- 基础工具函数 ----------------

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
