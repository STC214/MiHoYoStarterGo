package app_logic

import (
	"MiHoYoStarterGo/logic"
	"fmt"
	"strings"
	"time"
)

func AddAccount(alias, user, pwd, gameID string) string {
	// 1. 严格校验：禁止空信息存入数据库
	if strings.TrimSpace(alias) == "" || strings.TrimSpace(user) == "" || strings.TrimSpace(pwd) == "" {
		return "MISSING_FIELDS"
	}

	// 2. 加载配置
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "LOAD_CONFIG_FAILED"
	}

	// 3. 密码加密
	encPwd, err := logic.EncryptString(pwd)
	if err != nil {
		return "ENCRYPT_FAILED"
	}

	// 4. 构建完整账号对象（不省略任何已有逻辑）
	newAcc := logic.Account{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Alias:        alias,
		Username:     user,
		Password:     encPwd,
		GameID:       gameID, // 确保此处 gameID 正确映射
		Token:        "",     // 新账号 Token 初始为空
		IsFirstLogin: true,   // 默认为首次登录以触发自动化
		CreateTime:   time.Now().Unix(),
	}

	// 5. 追加到列表并保存
	if cfg.Accounts == nil {
		cfg.Accounts = []logic.Account{}
	}
	cfg.Accounts = append(cfg.Accounts, newAcc)

	err = logic.SaveConfig(cfg)
	if err != nil {
		return "SAVE_FAILED"
	}

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

func UpdateAccount(id, alias, user, pwd string) string {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(alias) == "" || strings.TrimSpace(user) == "" {
		return "MISSING_FIELDS"
	}

	cfg, err := logic.LoadConfig()
	if err != nil {
		return "LOAD_CONFIG_FAILED"
	}

	target := -1
	for i, acc := range cfg.Accounts {
		if acc.ID == id {
			target = i
			break
		}
	}
	if target < 0 {
		return "ACCOUNT_NOT_FOUND"
	}

	cfg.Accounts[target].Alias = alias
	cfg.Accounts[target].Username = user

	// Empty password means keep existing encrypted password.
	if strings.TrimSpace(pwd) != "" {
		encPwd, err := logic.EncryptString(pwd)
		if err != nil {
			return "ENCRYPT_FAILED"
		}
		cfg.Accounts[target].Password = encPwd
	}

	if err := logic.SaveConfig(cfg); err != nil {
		return "SAVE_FAILED"
	}
	return "SUCCESS"
}

func SaveAccountsOrder(ids []string) string {
	cfg, err := logic.LoadConfig()
	if err != nil {
		return "LOAD_CONFIG_FAILED"
	}

	if len(ids) != len(cfg.Accounts) {
		return "ORDER_LENGTH_MISMATCH"
	}

	accMap := make(map[string]logic.Account, len(cfg.Accounts))
	for _, acc := range cfg.Accounts {
		accMap[acc.ID] = acc
	}

	seen := make(map[string]bool, len(ids))
	ordered := make([]logic.Account, 0, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return "ORDER_INVALID_ID"
		}
		if seen[id] {
			return "ORDER_DUPLICATE_ID"
		}
		acc, ok := accMap[id]
		if !ok {
			return "ORDER_ID_NOT_FOUND"
		}
		seen[id] = true
		ordered = append(ordered, acc)
	}

	cfg.Accounts = ordered
	if err := logic.SaveConfig(cfg); err != nil {
		return "SAVE_FAILED"
	}
	return "SUCCESS"
}
