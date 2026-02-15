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

// --- 闀ㄦ澘顣ㄩ懜鍥ц祴閾忕喖鍊ф潛?---

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

// --- 鐟奉厾鐤?---

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

// --- 閻╋絾甯﹂懜鍥х吋鐞?---

func (a *App) IsGameRunning(gameID string) bool {
	return logic.IsGameRunning(gameID)
}

func (a *App) StartGame(gameID string) string {
	return app_logic.StartGame(gameID)
}

func (a *App) StartMonitor(acc logic.Account) {
	a.IsPaused = false
	a.ShouldCancel = false
	app_logic.RunMonitor(a.ctx, acc, &a.IsPaused, &a.ShouldCancel)
}

func (a *App) ExecuteLoginAction(acc logic.Account, action string) string {
	return app_logic.ExecuteLoginAction(a.ctx, acc, action, &a.IsPaused, &a.ShouldCancel)
}

func (a *App) StopMonitor() {
	a.ShouldCancel = true
}

func (a *App) TogglePauseMonitor() {
	a.IsPaused = !a.IsPaused
}

func (a *App) GetMonitorStatus() string {
	if a.IsPaused {
		return "PAUSED"
	}
	return "RUNNING"
}
