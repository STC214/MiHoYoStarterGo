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

// Account 璐﹀彿淇℃伅缁撴瀯
type Account struct {
	ID                string `json:"id"`
	Alias             string `json:"alias"`              // 璐﹀彿鍒悕
	Username          string `json:"username"`           // 璐﹀彿鍚?
	Password          string `json:"password"`           // 鍔犲瘑寰岀殑瀵嗙爜
	GameID            string `json:"game_id"`            // 娓告垙绫诲瀷 (GenshinCN, etc.)
	Token             string `json:"token"`              // 鍔犲瘑寰岀殑娉ㄥ唽琛?Token (Hex鏍煎紡)
	DeviceFingerprint string `json:"device_fingerprint"` // saved hardware profile fingerprint
	IsFirstLogin      bool   `json:"is_first_login"`     // 鏄惁涓洪娆＄櫥褰?
	CreateTime        int64  `json:"create_time"`        // 鍒涘缓鏃堕棿
}

// ConfigData 鏁翠綋閰嶇疆鏂囦欢缁撴瀯
type ConfigData struct {
	Theme        string            `json:"theme"`         // theme-darcula 鎴?theme-monokai
	EnabledTags  []string          `json:"enabled_tags"`  // 鍚敤鐨勬爣绛?
	Accounts     []Account         `json:"accounts"`      // 璐﹀彿鍒楄〃
	WindowWidth  int               `json:"window_width"`  // 绐楀彛瀹藉害
	WindowHeight int               `json:"window_height"` // 绐楀彛楂樺害
	WindowX      int               `json:"window_x"`      // 淇锛氳ˉ鍏?WindowX
	WindowY      int               `json:"window_y"`      // 淇锛氳ˉ鍏?WindowY
	GamePaths    map[string]string `json:"game_paths"`    // 娓告垙璺緞鏄犲皠
}

// --- 閰嶇疆 IO 閫昏緫 ---

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

// ---------------- 瀵嗙爜鍔犲瘑/瑙ｅ瘑 (AES) ----------------

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
