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

// Account 賬號信息結構
type Account struct {
	ID           string `json:"id"`
	Alias        string `json:"alias"`          // 賬號別名
	Username     string `json:"username"`       // 賬號名
	Password     string `json:"password"`       // 加密後的密碼
	GameID       string `json:"game_id"`        // 遊戲類型 (GenshinCN, etc.)
	Token        string `json:"token"`          // 加密後的註冊表 Token (Hex格式)
	IsFirstLogin bool   `json:"is_first_login"` // 是否為首次登錄
	CreateTime   int64  `json:"create_time"`    // 創建時間
}

// ConfigData 整體配置文件結構 [新增 GamePaths 字段]
type ConfigData struct {
	Theme        string            `json:"theme"`         // theme-darcula 或 theme-monokai
	EnabledTags  []string          `json:"enabled_tags"`  // 啟用的標籤
	Accounts     []Account         `json:"accounts"`      // 賬號列表
	WindowWidth  int               `json:"window_width"`  // 窗口寬度
	WindowHeight int               `json:"window_height"` // 窗口高度
	WindowX      int               `json:"window_x"`      // 窗口 X 坐標
	WindowY      int               `json:"window_y"`      // 窗口 Y 坐標
	GamePaths    map[string]string `json:"game_paths"`    // [新增] 存儲各遊戲 .exe 絕對路徑
}

const configFileName = "config.json" //

// --- 加密核心邏輯 ---

// getSecretKey 根據設備指紋生成 32 字節的 AES 密鑰
func getSecretKey() []byte {
	fingerprint := GetDeviceFingerprint()                // 獲取設備指紋
	hash := md5.Sum([]byte(fingerprint + "_mhy_secret")) // 加鹽處理
	return []byte(hex.EncodeToString(hash[:]))           // 返回 32 字節 Key
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
	return hex.EncodeToString(ciphertext), nil //
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
	return string(plaintext), nil //
}

// --- 文件操作邏輯 ---

// LoadConfig 從本地加載配置
func LoadConfig() (*ConfigData, error) {
	var config ConfigData
	file, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			// 默認初始化配置，給定默認窗口大小
			return &ConfigData{
				Theme:        "theme-darcula",
				EnabledTags:  []string{"GenshinCN"},
				WindowWidth:  1024,
				WindowHeight: 768,
				GamePaths:    make(map[string]string), // 初始化路徑地圖
			}, nil
		}
		return nil, err
	}
	err = json.Unmarshal(file, &config)

	// 確保 GamePaths 不為 nil，防止前端調用報錯
	if config.GamePaths == nil {
		config.GamePaths = make(map[string]string)
	}

	return &config, err
}

// SaveConfig 保存配置到本地
func SaveConfig(config *ConfigData) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFileName, data, 0644) //
}

// ExportPlaintextBackup 明文導出備份 (需求：方便遷移)
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
	return backupFile, err //
}
