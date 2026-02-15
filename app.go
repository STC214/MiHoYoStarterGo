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
	return &App{IsPaused: false, ShouldCancel: false}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// --- 环境与账号逻辑转发 ---

// PrepareAccountEnvironment 环境准备补丁 (满足条件：游戏未运行 && 非首次登录)
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

// --- 配置与设置转发 ---

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

// --- 监控与执行转发 ---

func (a *App) IsGameRunning(gameID string) bool {
	return logic.IsGameRunning(gameID)
}

func (a *App) StartGameExecution(gameID string) string {
	return app_logic.StartGame(gameID)
}

func (a *App) ForceStartMonitor(acc logic.Account) string {
	a.IsPaused, a.ShouldCancel = false, false
	// 启动异步监控流
	go app_logic.RunMonitor(a.ctx, acc, &a.IsPaused, &a.ShouldCancel)
	return "SUCCESS"
}

func (a *App) TogglePause() string {
	a.IsPaused = !a.IsPaused
	return "OK"
}

func (a *App) StopMonitor() string {
	a.ShouldCancel = true
	return "SUCCESS"
}
