package logic

import (
	"context"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/lxn/win"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type loginStrategy string

const (
	loginStrategyGenshin  loginStrategy = "genshin_windowed"
	loginStrategyStarRail loginStrategy = "starrail_windowed"
)

type zzzRuntimeState struct {
	profileLoaded bool
	profileExact  bool
	profile       ZZZPointProfile

	lastClickAAt  time.Time
	lastClickBAt  time.Time
	lastHintNoCfg time.Time
}

func isDebugMode() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MHY_DEBUG")))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

func StartAutomationMonitor(ctx context.Context, gameID, user, pwd string, isFirst bool, pause *bool, cancel *bool) {
	_ = isFirst
	rand.Seed(time.Now().UnixNano())
	debugMode := isDebugMode()

	go func() {
		fmt.Printf("[system] monitor started: game=%s user=%s\n", gameID, user)
		runtime.EventsEmit(ctx, "monitor_status", "流程已启动，正在等待游戏窗口...")

		tmpRawPath := fmt.Sprintf("temp_raw_%d.png", time.Now().UnixNano())
		defer func() {
			_ = os.Remove(tmpRawPath)
		}()

		accountPwdSwitched := false
		lastOCRSnapshot := ""
		lastStatusTip := ""
		lastMatchLog := ""
		zzzState := &zzzRuntimeState{}

		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if *cancel {
				fmt.Println("[system] cancel signal received, monitor stopped")
				runtime.EventsEmit(ctx, "monitor_finished", "CANCELLED")
				return
			}
			if *pause || !IsGameRunning(gameID) {
				continue
			}

			hwnd := GetWindowHandleByGameID(gameID)
			if hwnd == 0 {
				continue
			}
			win.SetForegroundWindow(hwnd)
			time.Sleep(100 * time.Millisecond)

			img, err := CaptureWindowByHandle(hwnd)
			if err != nil {
				continue
			}
			if err := writePNG(tmpRawPath, img); err != nil {
				continue
			}
			textPoints, err := RecognizeWithPosContext(context.Background(), tmpRawPath)
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
			if debugMode {
				matches := collectTargetMatches(gameID, textPoints)
				if len(matches) > 0 {
					line := strings.Join(matches, " | ")
					if line != lastMatchLog {
						lastMatchLog = line
						fmt.Printf("[DEBUG][OCR_MATCH][%s] %s\n", gameID, line)
					}
				}
			}

			if gameID == "StarRailCN" && !accountPwdSwitched && isStarRailVerifyPage(textPoints) {
				if x, y, ok := findKeywordCenter(textPoints, []string{"账号密码"}); ok {
					var rect win.RECT
					win.GetWindowRect(hwnd, &rect)
					runtime.EventsEmit(ctx, "monitor_status", "已识别验证码页，正在切换到账号密码登录")
					randClick(int(rect.Left)+x, int(rect.Top)+y, 8, 4)
					accountPwdSwitched = true
					time.Sleep(500 * time.Millisecond)
					continue
				}
			}

			if gameID == "ZZZCN" {
				if tryZZZFlow(ctx, hwnd, textPoints, user, pwd, zzzState) {
					continue
				}
			}

			if isLoginPage(gameID, textPoints) {
				runtime.EventsEmit(ctx, "monitor_status", "已识别登录界面，开始执行登录流程")
				executeFullSequenceByHandle(hwnd, textPoints, user, pwd)
			}

			var rect win.RECT
			win.GetWindowRect(hwnd, &rect)
			if (gameID == "ZZZCN" && isZZZScreenC(textPoints, int(rect.Bottom-rect.Top))) || (gameID != "ZZZCN" && isConfirmedImageA(textPoints, int(rect.Bottom-rect.Top))) {
				fmt.Println("[success] login confirmed, writing account data back")
				runtime.EventsEmit(ctx, "monitor_status", "已识别登录成功，正在写回账号数据...")
				if err := finalizeAccountStorage(gameID, user); err != nil {
					fmt.Printf("[error] finalize account storage failed: %v\n", err)
					runtime.EventsEmit(ctx, "monitor_finished", "FAILED")
				} else {
					runtime.EventsEmit(ctx, "monitor_finished", "SUCCESS")
				}
				return
			}
		}
	}()
}

func tryZZZFlow(ctx context.Context, hwnd win.HWND, points []TextPoint, user, pwd string, st *zzzRuntimeState) bool {
	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	left, top := int(rect.Left), int(rect.Top)
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)

	if isZZZScreenA(points) {
		if time.Since(st.lastClickAAt) >= 1200*time.Millisecond {
			if x, y, ok := findKeywordCenter(points, []string{"账号密码"}); ok {
				runtime.EventsEmit(ctx, "monitor_status", "绝区零：识别到画面A，点击账号密码")
				randClick(left+x, top+y, 8, 4)
				st.lastClickAAt = time.Now()
			}
		}
		return true
	}

	if isZZZScreenB(points) {
		if !st.profileLoaded {
			p, exact, found, err := LoadZZZPointProfile(width, height)
			if err != nil {
				runtime.EventsEmit(ctx, "monitor_status", "绝区零：读取坐标失败")
				return true
			}
			if !found {
				if st.lastHintNoCfg.IsZero() || time.Since(st.lastHintNoCfg) > 3*time.Second {
					st.lastHintNoCfg = time.Now()
					runtime.EventsEmit(ctx, "monitor_status", "绝区零：未找到当前分辨率坐标，请先点击“记录坐标”")
				}
				return true
			}
			st.profile = p
			st.profileLoaded = true
			st.profileExact = exact
			runtime.EventsEmit(ctx, "monitor_status", "绝区零：已加载坐标，开始坐标登录")
		}
		if time.Since(st.lastClickBAt) < 1800*time.Millisecond {
			return true
		}
		executeZZZByPoints(hwnd, user, pwd, st.profile)
		st.lastClickBAt = time.Now()
		return true
	}

	return false
}

func executeZZZByPoints(hwnd win.HWND, user, pwd string, profile ZZZPointProfile) {
	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	left, top := int(rect.Left), int(rect.Top)
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)

	account := clampPoint(profile.Account, width, height)
	password := clampPoint(profile.Password, width, height)
	agreement := clampPoint(profile.Agreement, width, height)
	enter := clampPoint(profile.Enter, width, height)

	randClick(left+account.X, top+account.Y, 6, 3)
	time.Sleep(220 * time.Millisecond)
	typeAction(user)
	time.Sleep(280 * time.Millisecond)

	randClick(left+password.X, top+password.Y, 6, 3)
	time.Sleep(220 * time.Millisecond)
	typeAction(pwd)
	time.Sleep(280 * time.Millisecond)

	randClick(left+agreement.X, top+agreement.Y, 5, 2)
	time.Sleep(220 * time.Millisecond)
	randClick(left+enter.X, top+enter.Y, 8, 4)
}

func clampPoint(p Point, width, height int) Point {
	if p.X < 0 {
		p.X = 0
	}
	if p.Y < 0 {
		p.Y = 0
	}
	if p.X > width-1 {
		p.X = width - 1
	}
	if p.Y > height-1 {
		p.Y = height - 1
	}
	return p
}

func isZZZScreenA(points []TextPoint) bool {
	return hasAnyKeyword(points, []string{"+86"}) && hasAnyKeyword(points, []string{"账号密码"})
}

func isZZZScreenB(points []TextPoint) bool {
	hits := 0
	if hasAnyKeyword(points, []string{"输入手机号", "手机号"}) {
		hits++
	}
	if hasAnyKeyword(points, []string{"邮箱"}) {
		hits++
	}
	if hasAnyKeyword(points, []string{"用户"}) {
		hits++
	}
	return hits >= 2
}

func isZZZScreenC(points []TextPoint, windowHeight int) bool {
	yMin := (windowHeight * 4) / 5
	for _, p := range points {
		if p.Y < yMin {
			continue
		}
		if containsKeywordSmart(p.Text, "点击进入") {
			return true
		}
	}
	return false
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func collectTargetMatches(gameID string, points []TextPoint) []string {
	keywords := []string{"账号密码", "验证码", "发送", "输入手机号", "手机号", "邮箱", "输入密码", "密码", "同意", "登录", "进入游戏"}
	if gameID == "ZZZCN" {
		keywords = []string{"+86", "账号密码", "输入手机号", "邮箱", "用户", "点击进入"}
	}
	seen := make(map[string]struct{}, len(keywords))
	out := make([]string, 0, 8)
	for _, p := range points {
		for _, kw := range keywords {
			if containsKeywordSmart(p.Text, kw) {
				if _, ok := seen[p.Text]; ok {
					continue
				}
				seen[p.Text] = struct{}{}
				out = append(out, p.Text)
				break
			}
		}
	}
	return out
}

func isStarRailVerifyPage(points []TextPoint) bool {
	hasCode := hasKeyword(points, "验证码")
	hasSend := hasKeyword(points, "发送")
	hasAccountPwd := hasAnyKeyword(points, []string{"账号密码"})
	return hasCode && hasSend && hasAccountPwd
}

func isLoginPage(gameID string, points []TextPoint) bool {
	switch gameID {
	case "StarRailCN", "ZZZCN":
		return isLoginPageByStrategy(loginStrategyStarRail, points)
	default:
		return isLoginPageByStrategy(loginStrategyGenshin, points)
	}
}

func isLoginPageByStrategy(strategy loginStrategy, points []TextPoint) bool {
	switch strategy {
	case loginStrategyStarRail:
		hasPhoneInput := hasAnyKeyword(points, []string{"输入手机号/邮箱", "输入手机号", "手机号/邮箱", "输入账号", "手机号", "邮箱"})
		hasPwdInput := hasAnyKeyword(points, []string{"输入密码", "密码"})
		hasEnter := hasAnyKeyword(points, []string{"进入游戏", "登录", "开始游戏"})
		return hasPhoneInput && hasPwdInput && hasEnter
	default:
		hasAccount := hasAnyKeyword(points, []string{"手机号", "邮箱", "输入手机号"})
		hasAgreement := hasAnyKeyword(points, []string{"同意", "已阅读"})
		return hasAccount && hasAgreement
	}
}

func hasAnyKeyword(points []TextPoint, keywords []string) bool {
	for _, p := range points {
		for _, kw := range keywords {
			if containsKeywordSmart(p.Text, kw) {
				return true
			}
		}
	}
	return false
}

func hasKeyword(points []TextPoint, keyword string) bool {
	for _, p := range points {
		if containsKeywordSmart(p.Text, keyword) {
			return true
		}
	}
	return false
}

func findKeywordCenter(points []TextPoint, keywords []string) (int, int, bool) {
	for _, p := range points {
		for _, kw := range keywords {
			if containsKeywordSmart(p.Text, kw) {
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
		items = append(items, fmt.Sprintf("[%d] %q @(%d,%d)", i+1, p.Text, p.X, p.Y))
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
			if containsKeywordSmart(txt, "进入游戏") || containsKeywordSmart(txt, "点击进入") || containsKeywordSmart(txt, "登录") {
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
	userFieldY := -1

	for _, p := range points {
		if !filledUser && hasAnyKeyword([]TextPoint{p}, []string{"手机号", "邮箱", "输入手机号", "输入账号", "账号密码"}) {
			randClick(left+p.X, top+p.Y, 10, 2)
			time.Sleep(260 * time.Millisecond)
			typeAction(user)
			userFieldY = p.Y
			filledUser = true
			time.Sleep(280 * time.Millisecond)
		}
		if !filledPwd && hasAnyKeyword([]TextPoint{p}, []string{"密码", "输入密码"}) && !hasAnyKeyword([]TextPoint{p}, []string{"忘记"}) {
			if userFieldY >= 0 && absInt(p.Y-userFieldY) < 18 {
				continue
			}
			randClick(left+p.X, top+p.Y, 10, 2)
			time.Sleep(260 * time.Millisecond)
			typeAction(pwd)
			filledPwd = true
			time.Sleep(260 * time.Millisecond)
		}
	}

	for _, p := range points {
		if (containsKeywordSmart(p.Text, "进入") || containsKeywordSmart(p.Text, "登录") || containsKeywordSmart(p.Text, "开始")) && len(p.Text) <= 16 {
			randClick(left+p.X, top+p.Y+5, 15, 5)
		}
	}
}

func containsKeywordSmart(text, keyword string) bool {
	nText := normalizeOCRText(text)
	nKeyword := normalizeOCRText(keyword)
	if nText == "" || nKeyword == "" {
		return false
	}
	if strings.Contains(nText, nKeyword) {
		return true
	}

	textRunes := []rune(nText)
	kwRunes := []rune(nKeyword)
	maxDist := fuzzyDistanceThreshold(len(kwRunes))

	if len(textRunes) <= len(kwRunes) {
		return levenshteinDistanceRunes(textRunes, kwRunes) <= maxDist
	}

	minWin := len(kwRunes) - maxDist
	if minWin < 1 {
		minWin = 1
	}
	maxWin := len(kwRunes) + maxDist
	if maxWin > len(textRunes) {
		maxWin = len(textRunes)
	}

	for start := 0; start < len(textRunes); start++ {
		for win := minWin; win <= maxWin; win++ {
			end := start + win
			if end > len(textRunes) {
				break
			}
			if levenshteinDistanceRunes(textRunes[start:end], kwRunes) <= maxDist {
				return true
			}
		}
	}
	return false
}

func fuzzyDistanceThreshold(keywordLen int) int {
	switch {
	case keywordLen <= 3:
		return 1
	case keywordLen <= 6:
		return 2
	case keywordLen <= 10:
		return 3
	default:
		return 4
	}
}

func normalizeOCRText(s string) string {
	if s == "" {
		return ""
	}
	n := strings.ToLower(strings.TrimSpace(s))
	replacer := strings.NewReplacer(
		" ", "", "\t", "", "\n", "", "\r", "",
		"。", "", "，", "", "：", "", "；", "", "！", "", "？", "",
		"(", "", ")", "", "（", "", "）", "", "[", "", "]", "",
	)
	n = replacer.Replace(n)
	n = strings.ReplaceAll(n, "人游", "进入游戏")
	n = strings.ReplaceAll(n, "入游", "进入游戏")
	n = strings.ReplaceAll(n, "登入", "登录")
	return n
}

func levenshteinDistanceRunes(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = minInt(del, minInt(ins, sub))
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FA5 {
			return true
		}
	}
	return false
}

func randClick(x, y, rx, ry int) {
	robotgo.Move(x+rand.Intn(rx*2)-rx, y+rand.Intn(ry*2)-ry)
	robotgo.Click("left", false)
}

func typeAction(s string) {
	time.Sleep(260 * time.Millisecond)
	robotgo.KeyTap("a", "control")
	robotgo.KeyTap("backspace")
	time.Sleep(180 * time.Millisecond)
	for _, r := range s {
		robotgo.TypeStr(string(r))
		time.Sleep(28 * time.Millisecond)
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
