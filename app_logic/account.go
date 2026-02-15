package app_logic

import (
	"MiHoYoStarterGo/logic"
	"fmt"
	"time"
)

func AddAccount(alias, user, pwd, gameID string) string {
	cfg, _ := logic.LoadConfig()
	encPwd, _ := logic.EncryptString(pwd)
	newAcc := logic.Account{
		ID:    fmt.Sprintf("%d", time.Now().UnixNano()),
		Alias: alias, Username: user, Password: encPwd,
		GameID: gameID, IsFirstLogin: true, CreateTime: time.Now().Unix(),
	}
	cfg.Accounts = append(cfg.Accounts, newAcc)
	logic.SaveConfig(cfg)
	return "SUCCESS"
}

func DeleteAccount(id string) string {
	cfg, _ := logic.LoadConfig()
	var newList []logic.Account
	for _, acc := range cfg.Accounts {
		if acc.ID != id {
			newList = append(newList, acc)
		}
	}
	cfg.Accounts = newList
	logic.SaveConfig(cfg)
	return "SUCCESS"
}

func GetPlaintext(enc string) string {
	res, _ := logic.DecryptString(enc)
	return res
}
