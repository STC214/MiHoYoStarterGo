package logic

import (
	"context"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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

type loginRuntimeState struct {
	userFilled   bool
	pwdFilled    bool
	userFieldY   int
	enterClicked bool

	agreementClicked   bool
	agreementClickedAt time.Time
}

type preOCRGateState struct {
	ocrEnabled  bool
	lastHintAt  time.Time
	lastHintMsg string
}

func isDebugMode() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MHY_DEBUG")))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

func getOCRIntervalByGame(gameID string) time.Duration {
	switch gameID {
	case "GenshinCN", "GenshinOS":
		return 300 * time.Millisecond
	default:
		return 300 * time.Millisecond
	}
}

func StartAutomationMonitor(ctx context.Context, gameID, user, pwd string, isFirst bool, directEnterFastPath bool, pause *atomic.Bool, cancel *atomic.Bool) {
	_ = isFirst
	rand.Seed(time.Now().UnixNano())
	debugMode := isDebugMode()

	go func() {
		fmt.Printf("[system] monitor started: game=%s user=%s\n", gameID, user)
		runtime.EventsEmit(ctx, "monitor_status", "流程已启动，正在等待游戏窗口...")
		if directEnterFastPath {
			runtime.EventsEmit(ctx, "monitor_status", "快速完成模式已启用：命中成功组后直接回写并结束")
		}

		tmpRawPath := fmt.Sprintf("temp_raw_%d.png", time.Now().UnixNano())
		defer func() {
			_ = os.Remove(tmpRawPath)
		}()

		accountPwdSwitched := false
		lastOCRSnapshot := ""
		lastStatusTip := ""
		lastMatchLog := ""
		zzzState := &zzzRuntimeState{}
		loginState := &loginRuntimeState{userFieldY: -1}
		preOCRGate := &preOCRGateState{}

		ticker := time.NewTicker(getOCRIntervalByGame(gameID))
		defer ticker.Stop()

		for range ticker.C {
			if cancel.Load() {
				fmt.Println("[system] cancel signal received, monitor stopped")
				runtime.EventsEmit(ctx, "monitor_finished", "CANCELLED")
				return
			}
			if pause.Load() || !IsGameRunning(gameID) {
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
			if shouldDelayOCRByScene(gameID) && !preOCRGate.ocrEnabled {
				ready, hint := isOCRWarmupReady(gameID, img)
				if !ready {
					if time.Since(preOCRGate.lastHintAt) >= 2*time.Second {
						preOCRGate.lastHintAt = time.Now()
						if hint == "" {
							hint = "游戏启动中，检测到黑白启动页，暂不进行OCR..."
						}
						if hint != preOCRGate.lastHintMsg {
							preOCRGate.lastHintMsg = hint
							runtime.EventsEmit(ctx, "monitor_status", hint)
						}
					}
					continue
				}
				preOCRGate.ocrEnabled = true
				preOCRGate.lastHintMsg = ""
				runtime.EventsEmit(ctx, "monitor_status", "启动门禁已解除，开始OCR识别")
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

			var rect win.RECT
			win.GetWindowRect(hwnd, &rect)
			windowWidth := int(rect.Right - rect.Left)
			windowHeight := int(rect.Bottom - rect.Top)
			groups := detectScreenGroups(gameID, textPoints, img, windowWidth, windowHeight)

			if groups.success {
				fmt.Println("[success] success-group matched, writing account data back")
				if directEnterFastPath {
					runtime.EventsEmit(ctx, "monitor_status", "命中快速完成条件：检测到成功组，正在回写并结束流程...")
				} else {
					runtime.EventsEmit(ctx, "monitor_status", "检测到成功组，正在回写账号数据...")
				}
				if err := finalizeAccountStorage(gameID, user); err != nil {
					fmt.Printf("[error] finalize account storage failed: %v\n", err)
					runtime.EventsEmit(ctx, "monitor_finished", "FAILED")
				} else {
					runtime.EventsEmit(ctx, "monitor_finished", "SUCCESS")
				}
				return
			}

			actionTaken := false

			if gameID == "StarRailCN" && !accountPwdSwitched && groups.accountPwd {
				if x, y, ok := findKeywordCenter(textPoints, []string{"账号密码"}); ok {
					runtime.EventsEmit(ctx, "monitor_status", "识别到 +86/账号密码 组，正在切换账号密码登录...")
					randClick(int(rect.Left)+x, int(rect.Top)+y, 8, 4)
					accountPwdSwitched = true
					time.Sleep(500 * time.Millisecond)
					actionTaken = true
				}
			}

			if gameID == "ZZZCN" {
				if !actionTaken && tryZZZFlow(ctx, hwnd, textPoints, user, pwd, zzzState) {
					actionTaken = true
				}
				if actionTaken {
					continue
				}
			}

			if !actionTaken && gameID != "ZZZCN" && groups.agreement && !loginState.agreementClicked {
				if clickSecondBlackAgreement(hwnd, img, textPoints) {
					loginState.agreementClicked = true
					loginState.agreementClickedAt = time.Now()
					actionTaken = true
				}
			}

			if !actionTaken && gameID != "ZZZCN" && groups.login {
				runtime.EventsEmit(ctx, "monitor_status", "识别到登录组，开始执行登录流程...")
				executeFullSequenceByHandle(hwnd, img, textPoints, user, pwd, loginState)
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
				runtime.EventsEmit(ctx, "monitor_status", "绝区零：识别到 +86/账号密码 组，点击账号密码")
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

type screenGroups struct {
	success    bool
	login      bool
	agreement  bool
	accountPwd bool
}

func detectScreenGroups(gameID string, points []TextPoint, frame image.Image, windowWidth, windowHeight int) screenGroups {
	g := screenGroups{}
	g.success = isDirectEnterScreen(gameID, points, frame, windowWidth, windowHeight)
	g.agreement = hasAnyKeywordStrict(points, []string{"不同意"}) && hasAnyKeywordStrict(points, []string{"同意"})

	switch gameID {
	case "StarRailCN":
		g.login = isLoginPage(gameID, points)
		g.accountPwd = isAccountPwdGroup(points)
	case "ZZZCN":
		g.login = isZZZScreenB(points)
		g.accountPwd = isAccountPwdGroup(points)
		if isZZZScreenC(points, windowHeight) {
			g.success = true
		}
	default:
		g.login = isLoginPage(gameID, points)
	}
	return g
}

func isAccountPwdGroup(points []TextPoint) bool {
	return hasAnyKeyword(points, []string{"+86"}) && hasAnyKeyword(points, []string{"账号密码"})
}

func shouldDelayOCRByScene(gameID string) bool {
	switch gameID {
	case "GenshinCN", "GenshinOS", "StarRailCN", "ZZZCN":
		return true
	default:
		return false
	}
}

func isOCRWarmupReady(gameID string, img image.Image) (bool, string) {
	if !shouldDelayOCRByScene(gameID) {
		return true, ""
	}
	if isPureBlackWhiteSplash(img) {
		return false, "检测到纯黑白/白黑启动页，暂不进行OCR"
	}
	return true, ""
}

func isPureBlackWhiteSplash(img image.Image) bool {
	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return false
	}

	stepX := maxInt(6, w/90)
	stepY := maxInt(6, h/90)
	total := 0
	mono := 0
	dark := 0
	bright := 0
	colorBuckets := make(map[int]int, 16)

	for y := b.Min.Y; y < b.Max.Y; y += stepY {
		for x := b.Min.X; x < b.Max.X; x += stepX {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r := int(r16 >> 8)
			g := int(g16 >> 8)
			bl := int(b16 >> 8)
			lum := (299*r + 587*g + 114*bl) / 1000
			total++
			if lum <= 55 {
				dark++
			}
			if lum >= 220 {
				bright++
			}
			deltaMax := maxInt(absInt(r-g), maxInt(absInt(r-bl), absInt(g-bl)))
			if deltaMax <= 16 {
				mono++
			}
			qr := r >> 5
			qg := g >> 5
			qb := bl >> 5
			key := (qr << 4) | (qg << 2) | qb
			colorBuckets[key]++
		}
	}
	if total == 0 {
		return false
	}

	mainThreshold := maxInt(2, total/40)
	mainColors := 0
	for _, c := range colorBuckets {
		if c >= mainThreshold {
			mainColors++
		}
	}
	monoRatio := float64(mono) / float64(total)
	bwRatio := float64(dark+bright) / float64(total)
	return monoRatio >= 0.93 && bwRatio >= 0.88 && mainColors <= 3
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
	keywords := []string{"账号密码", "验证码", "发送", "输入手机号", "手机号", "邮箱", "输入密码", "密码", "同意", "登录", "进入游戏", "点击进入", "开始游戏", "登出"}
	if gameID == "ZZZCN" {
		keywords = []string{"+86", "账号密码", "输入手机号", "手机号", "邮箱", "用户", "点击进入"}
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
		hasPhoneInput := hasAnyKeyword(points, []string{"输入手机号邮箱", "输入手机号", "手机号邮箱", "输入账号", "手机号", "邮箱"})
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

func hasAnyKeywordStrict(points []TextPoint, keywords []string) bool {
	for _, p := range points {
		for _, kw := range keywords {
			if containsKeywordStrict(p.Text, kw) {
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

func executeFullSequenceByHandle(hwnd win.HWND, frame *image.RGBA, points []TextPoint, user, pwd string, st *loginRuntimeState) {
	if hwnd == 0 {
		return
	}
	if st == nil {
		st = &loginRuntimeState{userFieldY: -1}
	}

	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	left, top := int(rect.Left), int(rect.Top)

	if !st.userFilled {
		for _, p := range points {
			if !hasAnyKeyword([]TextPoint{p}, []string{"手机号", "邮箱", "输入手机号", "输入账号", "账号密码"}) {
				continue
			}
			randClick(left+p.X, top+p.Y, 10, 2)
			time.Sleep(260 * time.Millisecond)
			typeAction(user)
			st.userFieldY = p.Y
			st.userFilled = true
			time.Sleep(280 * time.Millisecond)
			break
		}
	}

	if !st.pwdFilled {
		for _, p := range points {
			if !hasAnyKeyword([]TextPoint{p}, []string{"密码", "输入密码"}) || hasAnyKeyword([]TextPoint{p}, []string{"忘记"}) {
				continue
			}
			if st.userFieldY >= 0 && absInt(p.Y-st.userFieldY) < 18 {
				continue
			}
			randClick(left+p.X, top+p.Y, 10, 2)
			time.Sleep(260 * time.Millisecond)
			typeAction(pwd)
			st.pwdFilled = true
			time.Sleep(260 * time.Millisecond)
			break
		}
	}

	if st.userFilled && st.pwdFilled && !st.enterClicked {
		for _, p := range points {
			if containsKeywordSmart(p.Text, "进入游戏") {
				randClick(left+p.X, top+p.Y+5, 15, 5)
				st.enterClicked = true
				return
			}
		}
	}

	if st.userFilled && st.pwdFilled && !st.agreementClicked {
		if clickSecondBlackAgreement(hwnd, frame, points) {
			st.agreementClicked = true
			st.agreementClickedAt = time.Now()
		}
	}
}

func clickSecondBlackAgreement(hwnd win.HWND, frame *image.RGBA, points []TextPoint) bool {
	_ = frame

	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	left, top := int(rect.Left), int(rect.Top)

	hasNotAgree := false
	hasAgree := false
	for _, p := range points {
		if containsKeywordStrict(p.Text, "不同意") {
			hasNotAgree = true
		}
		if containsKeywordStrict(p.Text, "同意") && !containsKeywordStrict(p.Text, "不同意") {
			hasAgree = true
		}
	}
	if !(hasNotAgree && hasAgree) {
		fmt.Println("[agreement] strict gate not met: need both '不同意' and '同意' in same OCR frame")
		return false
	}

	hits := make([]TextPoint, 0, 6)
	for _, p := range points {
		if containsKeywordSmart(p.Text, "同意") {
			hits = append(hits, p)
		}
	}
	if len(hits) < 1 {
		fmt.Println("[agreement] no '同意' candidate in current OCR frame")
		return false
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Y == hits[j].Y {
			return hits[i].X < hits[j].X
		}
		return hits[i].Y < hits[j].Y
	})

	pureAgree := make([]TextPoint, 0, len(hits))
	notAgree := make([]TextPoint, 0, 2)
	for _, h := range hits {
		if containsKeywordStrict(h.Text, "不同意") {
			notAgree = append(notAgree, h)
			continue
		}
		if containsKeywordStrict(h.Text, "同意") {
			pureAgree = append(pureAgree, h)
		}
	}

	for i := 0; i < len(hits)-1; i++ {
		if !containsKeywordStrict(hits[i].Text, "不同意") {
			continue
		}
		next := hits[i+1]
		if containsKeywordStrict(next.Text, "同意") && !containsKeywordStrict(next.Text, "不同意") {
			fmt.Printf("[agreement] click next after '不同意': %q @(%d,%d)\n", next.Text, next.X, next.Y)
			randClick(left+next.X, top+next.Y+4, 10, 4)
			time.Sleep(1 * time.Second)
			return true
		}
	}

	for _, n := range notAgree {
		bestIdx := -1
		bestDx := 1 << 30
		for i, a := range pureAgree {
			dy := absInt(a.Y - n.Y)
			if dy > 36 {
				continue
			}
			dx := a.X - n.X
			if dx <= 0 {
				continue
			}
			if dx < bestDx {
				bestDx = dx
				bestIdx = i
			}
		}
		if bestIdx >= 0 {
			target := pureAgree[bestIdx]
			fmt.Printf("[agreement] click same-row right '同意': %q @(%d,%d)\n", target.Text, target.X, target.Y)
			randClick(left+target.X, top+target.Y+4, 10, 4)
			time.Sleep(1 * time.Second)
			return true
		}
	}

	if len(pureAgree) > 0 {
		target := pureAgree[len(pureAgree)-1]
		fmt.Printf("[agreement] fallback click last pure '同意': %q @(%d,%d)\n", target.Text, target.X, target.Y)
		randClick(left+target.X, top+target.Y+4, 10, 4)
		time.Sleep(1 * time.Second)
		return true
	}

	fmt.Println("[agreement] only '不同意' detected, skip click")
	return false
}

func isDirectEnterScreen(gameID string, points []TextPoint, frame image.Image, windowWidth, windowHeight int) bool {
	if gameID == "StarRailCN" {
		return isStarRailSuccessFeature(points, frame, windowWidth, windowHeight)
	}

	yMin := (windowHeight * 4) / 5
	for _, p := range points {
		if p.Y < yMin {
			continue
		}
		if containsKeywordSmart(p.Text, "点击进入") || containsKeywordSmart(p.Text, "开始游戏") {
			return true
		}
	}
	return false
}

func isStarRailSuccessFeature(points []TextPoint, frame image.Image, windowWidth, windowHeight int) bool {
	if windowWidth <= 0 || windowHeight <= 0 {
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := make(chan bool, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		matched := isStarRailSuccessByText(points, windowWidth, windowHeight)
		select {
		case resultCh <- matched:
		case <-ctx.Done():
		}
	}()

	go func() {
		defer wg.Done()
		matched := isStarRailSuccessByImage(ctx, frame, windowWidth, windowHeight)
		select {
		case resultCh <- matched:
		case <-ctx.Done():
		}
	}()

	matched := false
	for i := 0; i < 2; i++ {
		r := <-resultCh
		if r {
			matched = true
			cancel()
			break
		}
	}
	wg.Wait()
	return matched
}

func isStarRailSuccessByText(points []TextPoint, windowWidth, windowHeight int) bool {
	yBottom := (windowHeight * 4) / 5
	xRight := (windowWidth * 9) / 10

	for _, p := range points {
		if p.Y >= yBottom && containsKeywordSmart(p.Text, "点击进入") {
			fmt.Println("[starrail-success] matched by OCR keyword: 点击进入")
			return true
		}
		if p.Y >= yBottom && containsKeywordSmart(p.Text, "开始游戏") {
			fmt.Println("[starrail-success] matched by OCR keyword: 开始游戏")
			return true
		}
		if p.X >= xRight && containsKeywordSmart(p.Text, "登出") {
			fmt.Println("[starrail-success] matched by OCR keyword: 登出")
			return true
		}
	}
	return false
}

func isStarRailSuccessByImage(ctx context.Context, frame image.Image, windowWidth, windowHeight int) bool {
	if frame == nil || windowWidth <= 0 || windowHeight <= 0 {
		return false
	}
	if matchStarRailClickEnterByImage(ctx, frame, windowWidth, windowHeight) {
		fmt.Println("[starrail-success] matched by image template: 点击进入")
		return true
	}
	if matchStarRailLogoutByImage(ctx, frame, windowWidth, windowHeight) {
		fmt.Println("[starrail-success] matched by image template: 登出")
		return true
	}
	return false
}

func matchStarRailClickEnterByImage(ctx context.Context, img image.Image, w, h int) bool {
	yStart := (h * 4) / 5
	if yStart >= h {
		return false
	}

	patchW := maxInt(96, w/8)
	patchH := maxInt(28, h/24)
	stepX := maxInt(10, patchW/3)
	stepY := maxInt(8, patchH/2)

	for y := yStart; y+patchH <= h; y += stepY {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		for x := 0; x+patchW <= w; x += stepX {
			total, blueN, whiteN := 0, 0, 0
			for yy := y; yy < y+patchH; yy += 3 {
				for xx := x; xx < x+patchW; xx += 3 {
					r16, g16, b16, _ := img.At(xx, yy).RGBA()
					r := int(r16 >> 8)
					g := int(g16 >> 8)
					bl := int(b16 >> 8)
					lum := (299*r + 587*g + 114*bl) / 1000
					total++
					if bl >= g+10 && g >= r-12 && bl >= 70 {
						blueN++
					}
					if lum >= 185 && absInt(r-g) <= 35 && absInt(r-bl) <= 35 && absInt(g-bl) <= 35 {
						whiteN++
					}
				}
			}
			if total == 0 {
				continue
			}
			blueRatio := float64(blueN) / float64(total)
			whiteRatio := float64(whiteN) / float64(total)
			if blueRatio >= 0.20 && whiteRatio >= 0.04 && whiteRatio <= 0.35 {
				return true
			}
		}
	}
	return false
}

func matchStarRailLogoutByImage(ctx context.Context, img image.Image, w, h int) bool {
	xStart := (w * 9) / 10
	if xStart >= w {
		return false
	}

	patchW := maxInt(42, w/28)
	patchH := maxInt(72, h/13)
	stepX := maxInt(4, patchW/4)
	stepY := maxInt(10, patchH/4)

	for x := xStart; x+patchW <= w; x += stepX {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		for y := 0; y+patchH <= h; y += stepY {
			total, purpleN, whiteN := 0, 0, 0
			topWhite, bottomWhite := 0, 0
			for yy := y; yy < y+patchH; yy += 2 {
				for xx := x; xx < x+patchW; xx += 2 {
					r16, g16, b16, _ := img.At(xx, yy).RGBA()
					r := int(r16 >> 8)
					g := int(g16 >> 8)
					bl := int(b16 >> 8)
					lum := (299*r + 587*g + 114*bl) / 1000
					total++
					if bl >= r+8 && bl >= g+8 && bl >= 70 {
						purpleN++
					}
					if lum >= 175 && absInt(r-g) <= 40 && absInt(r-bl) <= 40 && absInt(g-bl) <= 40 {
						whiteN++
						if yy < y+patchH/2 {
							topWhite++
						} else {
							bottomWhite++
						}
					}
				}
			}
			if total == 0 {
				continue
			}
			purpleRatio := float64(purpleN) / float64(total)
			whiteRatio := float64(whiteN) / float64(total)
			topWhiteRatio := float64(topWhite) / float64(total)
			bottomWhiteRatio := float64(bottomWhite) / float64(total)
			if purpleRatio >= 0.20 && whiteRatio >= 0.08 && topWhiteRatio >= 0.02 && bottomWhiteRatio >= 0.01 {
				return true
			}
		}
	}
	return false
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

func containsKeywordStrict(text, keyword string) bool {
	nText := normalizeOCRText(text)
	nKeyword := normalizeOCRText(keyword)
	if nText == "" || nKeyword == "" {
		return false
	}
	return strings.Contains(nText, nKeyword)
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
		"。", "", "，", "", ",", "", ".", "", "：", "", ":", "",
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
