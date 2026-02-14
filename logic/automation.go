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

// StartAutomationMonitor 核心監控邏輯
// 新增了 cancel *bool 參數用於接收用戶的手動停止指令
func StartAutomationMonitor(ctx context.Context, gameID, user, pwd string, isFirst bool, pause *bool, cancel *bool) {
	rand.Seed(time.Now().UnixNano())

	go func() {
		fmt.Printf("[系統] 監控啟動：正在為賬號 %s 監視登錄狀態...\n", user)
		tmpImgPath := fmt.Sprintf("temp_%d.png", time.Now().Unix()) // 使用動態文件名避免衝突

		// 根據遊戲 ID 確定窗口標題
		windowName := "原神"
		if gameID == "StarRailCN" {
			windowName = "崩坏：星穹铁道"
		} else if gameID == "ZZZCN" {
			windowName = "绝区零"
		}

		ticker := time.NewTicker(1200 * time.Millisecond)
		defer func() {
			ticker.Stop()
			os.Remove(tmpImgPath) // 協程結束時清理臨時圖片資源
		}()

		for range ticker.C {
			// 1. [新增] 檢查手動取消信號
			if *cancel {
				fmt.Printf("[系統] 接收到用戶終止信號，已回收 OCR 資源並關閉協程。\n")
				return // 徹底退出協程
			}

			// 2. 檢查暫停狀態
			if *pause {
				continue
			}

			// 3. 檢查進程是否運行
			if !IsGameRunning(gameID) {
				continue
			}

			// 4. 尋找窗口句柄並置頂
			hwnd := win.FindWindow(nil, syscall.StringToUTF16Ptr(windowName))
			var rect win.RECT
			if hwnd != 0 {
				win.SetForegroundWindow(hwnd)
				win.GetWindowRect(hwnd, &rect)
				time.Sleep(100 * time.Millisecond)
			} else {
				continue
			}

			// 5. 截圖
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

			// 6. OCR 識別
			textPoints, err := RecognizeWithPos(tmpImgPath)
			if err != nil {
				continue
			}

			// 7. 判定畫面 B (登錄框) 並執行填充
			if isImageBStrict(textPoints) {
				fmt.Println("[流水線] 檢測到登錄界面，開始執行賬密填充...")
				executeFullSequence(windowName, textPoints, user, pwd)
			}

			// 8. 判定畫面 A (進入遊戲成功) 並提取 Token
			windowHeight := int(rect.Bottom - rect.Top)
			if isConfirmedImageA(textPoints, windowHeight) {
				fmt.Println("========================================")
				fmt.Println(" [驗證成功] 已登錄成功，正在提取數據...")

				err := finalizeAccountStorage(gameID, user)

				if err != nil {
					fmt.Printf(" [錯誤] 數據存檔失敗: %v\n", err)
					runtime.EventsEmit(ctx, "monitor_finished", "FAILED")
				} else {
					fmt.Println(" [成功] 數據已更新至 config.json")
					runtime.EventsEmit(ctx, "monitor_finished", "SUCCESS")
				}
				fmt.Println("========================================")
				return // 任務完成，退出協程
			}
		}
	}()
}

// ---------------- 判定與數據存儲邏輯 ----------------

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
	// 延遲讀取註冊表，確保遊戲已經把最新數據寫進去
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

// ---------------- 鍵鼠模擬執行 ----------------

func executeFullSequence(windowName string, points []TextPoint, user, pwd string) {
	hwnd := win.FindWindow(nil, syscall.StringToUTF16Ptr(windowName))
	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	left, top := int(rect.Left), int(rect.Top)

	// 1. 填寫賬號密碼
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

	// 2. 勾選協議 (找特定的圓形符號)
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

	// 3. 點擊進入/登錄
	for _, p := range points {
		if (strings.Contains(p.Text, "进入") || strings.Contains(p.Text, "开始")) && len(p.Text) <= 12 {
			randClick(left+p.X, top+p.Y+5, 15, 5)
		}
	}
}

// ---------------- 基礎工具函數 ----------------

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
