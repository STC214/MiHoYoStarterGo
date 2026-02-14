package logic

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Account 账号信息结构
type Account struct {
	ID           string `json:"id"`
	Alias        string `json:"alias"`          // 账号别名
	Username     string `json:"username"`       // 账号名
	Password     string `json:"password"`       // 加密后的密码
	GameID       string `json:"game_id"`        // 游戏类型 (GenshinCN, etc.)
	Token        string `json:"token"`          // 加密后的注册表 Token (Hex格式)
	IsFirstLogin bool   `json:"is_first_login"` // 是否为首次登录
	CreateTime   int64  `json:"create_time"`
}

// ConfigData 整体配置文件结构
type ConfigData struct {
	Theme       string    `json:"theme"` // theme-darcula 或 theme-monokai
	EnabledTags []string  `json:"enabled_tags"`
	Accounts    []Account `json:"accounts"`
}

const configFileName = "config.json"

// --- 加密核心逻辑 ---

// getSecretKey 根据设备指纹生成 32 字节的 AES 密钥
func getSecretKey() []byte {
	fingerprint := GetDeviceFingerprint()
	hash := md5.Sum([]byte(fingerprint + "_mhy_secret")) // 加盐处理
	return []byte(hex.EncodeToString(hash[:]))           // 返回 32 字节 Key
}

// EncryptString 加密字符串
func EncryptString(plaintext string) (string, error) {
	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptString 解密字符串
func DecryptString(ciphertextHex string) (string, error) {
	data, _ := hex.DecodeString(ciphertextHex)
	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// --- 文件操作逻辑 ---

// LoadConfig 从本地加载配置
func LoadConfig() (*ConfigData, error) {
	var config ConfigData
	file, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigData{Theme: "theme-darcula", EnabledTags: []string{"GenshinCN"}}, nil
		}
		return nil, err
	}
	err = json.Unmarshal(file, &config)
	return &config, err
}

// SaveConfig 保存配置到本地
func SaveConfig(config *ConfigData) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFileName, data, 0644)
}

// ExportPlaintextBackup 明文导出备份 (需求：方便迁移)
func ExportPlaintextBackup(config *ConfigData) (string, error) {
	type PlainAccount struct {
		Alias    string `json:"alias"`
		Game     string `json:"game"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var backup []PlainAccount
	for _, acc := range config.Accounts {
		pwd, _ := DecryptString(acc.Password)
		backup = append(backup, PlainAccount{
			Alias:    acc.Alias,
			Game:     acc.GameID,
			Username: acc.Username,
			Password: pwd,
		})
	}
	data, _ := json.MarshalIndent(backup, "", "  ")
	backupFile := "backup_accounts.json"
	err := os.WriteFile(backupFile, data, 0644)
	return backupFile, err
}
