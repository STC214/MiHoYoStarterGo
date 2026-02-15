package main

import (
	"MiHoYoStarterGo/app_logic"
	"MiHoYoStarterGo/logic"
	"context"
)

type App struct {
	ctx          context.Context
	IsPaused     bool
	ShouldCancel bool
}

func NewApp() *App {
	return &App{
		IsPaused:     false,
		ShouldCancel: false,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- 環境與帳號邏輯 ---

func (a *App) PrepareAccountEnvironment(acc logic.Account) string {
	return app_logic.HandleEnvPatch(acc)
}

func (a *App) AddAccount(alias, user, pwd, gameID string) string {
	return app_logic.AddAccount(alias, user, pwd, gameID)
}

func (a *App) DeleteAccount(id string) string {
	return app_logic.DeleteAccount(id)
}

func (a *App) GetPlaintext(enc string) string {
	return app_logic.GetPlaintext(enc)
}

// --- 設置 ---

func (a *App) GetSettings() *logic.ConfigData {
	return app_logic.GetSettings()
}

func (a *App) SaveTheme(theme string) {
	app_logic.SaveTheme(theme)
}

func (a *App) SaveGamePaths(p map[string]string) {
	app_logic.SaveGamePaths(p)
}

func (a *App) SelectGameFile() string {
	return app_logic.SelectGameFile(a.ctx)
}

func (a *App) ExportBackup() string {
	return app_logic.ExportBackup()
}

// --- 監控與執行 ---

func (a *App) IsGameRunning(gameID string) bool {
	return app_logic.CheckGameRunning(gameID)
}

func (a *App) StartGame(gameID string) string {
	return app_logic.StartGameProcess(gameID)
}

func (a *App) StartMonitor(acc logic.Account) {
	// 调用 app_logic 中修复后的函数
	app_logic.StartAutomationMonitor(a.ctx, acc.GameID, acc.Username, acc.Password, acc.IsFirstLogin, &a.IsPaused, &a.ShouldCancel)
}

func (a *App) StopMonitor() {
	a.ShouldCancel = true
}

func (a *App) TogglePauseMonitor() {
	a.IsPaused = !a.IsPaused
}

func (a *App) GetMonitorStatus() string {
	if a.IsPaused {
		return "已暫停"
	}
	return "運行中"
}
