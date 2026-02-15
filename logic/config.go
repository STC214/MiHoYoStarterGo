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
	"time"
)

const configFileName = "config.json"

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

// ConfigData 整體配置文件結構
type ConfigData struct {
	Theme        string            `json:"theme"`         // theme-darcula 或 theme-monokai
	EnabledTags  []string          `json:"enabled_tags"`  // 啟用的標籤
	Accounts     []Account         `json:"accounts"`      // 賬號列表
	WindowWidth  int               `json:"window_width"`  // 窗口寬度
	WindowHeight int               `json:"window_height"` // 窗口高度
	WindowX      int               `json:"window_x"`      // 修复：补全 WindowX
	WindowY      int               `json:"window_y"`      // 修复：补全 WindowY
	GamePaths    map[string]string `json:"game_paths"`    // 遊戲路徑映射
}

// --- 配置 IO 邏輯 ---

func LoadConfig() (*ConfigData, error) {
	var config ConfigData
	file, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigData{
				Theme:        "theme-darcula",
				EnabledTags:  []string{"GenshinCN"},
				WindowWidth:  1024,
				WindowHeight: 768,
				WindowX:      0,
				WindowY:      0,
				Accounts:     []Account{},
				GamePaths:    make(map[string]string),
			}, nil
		}
		return nil, err
	}
	err = json.Unmarshal(file, &config)
	if config.GamePaths == nil {
		config.GamePaths = make(map[string]string)
	}
	if config.Accounts == nil {
		config.Accounts = []Account{}
	}
	return &config, err
}

func SaveConfig(config *ConfigData) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFileName, data, 0644)
}

func ExportPlaintextBackup(config *ConfigData) (string, error) {
	backupName := fmt.Sprintf("backup_%d.json", time.Now().Unix())
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	err = os.WriteFile(backupName, data, 0644)
	return backupName, err
}

// ---------------- 密碼加密/解密 (AES) ----------------

var commonKey = []byte("MHY-STARTER-GOGO-12345678-SAFE-!") // 32 bytes for AES-256

func EncryptString(plaintext string) (string, error) {
	block, err := aes.NewCipher(commonKey)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
	return hex.EncodeToString(ciphertext), nil
}

func DecryptString(cryptoText string) (string, error) {
	ciphertext, _ := hex.DecodeString(cryptoText)
	block, err := aes.NewCipher(commonKey)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
