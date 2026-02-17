package app_logic

import (
	"MiHoYoStarterGo/logic"
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

const (
	ActionRunningRestart    = "running_restart"
	ActionRunningManual     = "running_manual"
	ActionStoppedAutoStart  = "stopped_auto_start"
	ActionStoppedManualWait = "stopped_manual_wait"
	ActionCancel            = "cancel"
)

// ExecuteLoginAction handles all account-switch/login branches requested by UI.
func ExecuteLoginAction(ctx context.Context, acc logic.Account, action string, pause, cancel *atomic.Bool) string {
	pause.Store(false)
	cancel.Store(false)

	switch action {
	case ActionCancel:
		return "CANCELLED"

	case ActionRunningRestart:
		if !logic.IsGameRunning(acc.GameID) {
			return "GAME_NOT_RUNNING"
		}
		if err := logic.KillGameProcess(acc.GameID); err != nil {
			return "KILL_FAILED"
		}
		if !waitProcessStopped(acc.GameID, 8*time.Second) {
			return "KILL_TIMEOUT"
		}
		if err := patchEnvIfNeeded(acc); err != nil {
			return fmt.Sprintf("PATCH_FAILED:%v", err)
		}
		if res := StartGame(acc.GameID); res != "SUCCESS" {
			return res
		}
		RunMonitor(ctx, acc, pause, cancel, false)
		return "STARTED"

	case ActionRunningManual:
		if !logic.IsGameRunning(acc.GameID) {
			return "GAME_NOT_RUNNING"
		}
		RunMonitor(ctx, acc, pause, cancel, false)
		return "STARTED"

	case ActionStoppedAutoStart:
		if logic.IsGameRunning(acc.GameID) {
			return "GAME_RUNNING"
		}
		if err := patchEnvIfNeeded(acc); err != nil {
			return fmt.Sprintf("PATCH_FAILED:%v", err)
		}
		if res := StartGame(acc.GameID); res != "SUCCESS" {
			return res
		}
		RunMonitor(ctx, acc, pause, cancel, false)
		return "STARTED"

	case ActionStoppedManualWait:
		if logic.IsGameRunning(acc.GameID) {
			return "GAME_RUNNING"
		}
		directEnterFastPath := hasSavedEnvForManualSwitch(acc)
		if err := patchEnvIfNeeded(acc); err != nil {
			return fmt.Sprintf("PATCH_FAILED:%v", err)
		}
		// Monitor waits for process existence (300ms loop) and starts OCR flow once game starts.
		RunMonitor(ctx, acc, pause, cancel, directEnterFastPath)
		return "STARTED"

	default:
		return "INVALID_ACTION"
	}
}

func hasSavedEnvForManualSwitch(acc logic.Account) bool {
	return strings.TrimSpace(acc.Token) != "" && strings.TrimSpace(acc.DeviceFingerprint) != ""
}

func waitProcessStopped(gameID string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !logic.IsGameRunning(gameID) {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return !logic.IsGameRunning(gameID)
}

func patchEnvIfNeeded(acc logic.Account) error {
	// Special rule from requirement:
	// If first-login marker is true and both saved registry token + hardware info are absent,
	// skip env patch and go straight to process waiting/login flow.
	if acc.IsFirstLogin && strings.TrimSpace(acc.Token) == "" && strings.TrimSpace(acc.DeviceFingerprint) == "" {
		return nil
	}

	if strings.TrimSpace(acc.Token) != "" {
		tokenBytes, err := hex.DecodeString(acc.Token)
		if err != nil {
			return fmt.Errorf("invalid token hex")
		}
		if err := logic.WriteToken(acc.GameID, tokenBytes); err != nil {
			return err
		}
	}

	if strings.TrimSpace(acc.DeviceFingerprint) != "" {
		if err := logic.ApplySavedDeviceFingerprint(acc.GameID, acc.DeviceFingerprint); err != nil {
			return err
		}
	}

	return nil
}
