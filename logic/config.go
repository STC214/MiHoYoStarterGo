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
	"path/filepath"
	"time"
)

const configFileName = "config.json"

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func appConfigPath() (exePath string, cwdPath string) {
	cwdPath = configFileName
	exePath = configFileName

	exe, err := os.Executable()
	if err != nil {
		return exePath, cwdPath
	}

	if realExe, err := filepath.EvalSymlinks(exe); err == nil {
		exe = realExe
	}

	exePath = filepath.Join(filepath.Dir(exe), configFileName)
	return exePath, cwdPath
}

func readConfigPath() string {
	exePath, cwdPath := appConfigPath()
	if fileExists(exePath) {
		return exePath
	}
	if fileExists(cwdPath) {
		return cwdPath
	}
	return exePath
}

func writeConfigPath() string {
	exePath, cwdPath := appConfigPath()
	if fileExists(exePath) || !fileExists(cwdPath) {
		return exePath
	}
	return cwdPath
}

type Account struct {
	ID                string `json:"id"`
	Alias             string `json:"alias"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	GameID            string `json:"game_id"`
	Token             string `json:"token"`
	DeviceFingerprint string `json:"device_fingerprint"`
	IsFirstLogin      bool   `json:"is_first_login"`
	CreateTime        int64  `json:"create_time"`
}

type ConfigData struct {
	Theme        string            `json:"theme"`
	EnabledTags  []string          `json:"enabled_tags"`
	Accounts     []Account         `json:"accounts"`
	WindowWidth  int               `json:"window_width"`
	WindowHeight int               `json:"window_height"`
	WindowX      int               `json:"window_x"`
	WindowY      int               `json:"window_y"`
	GamePaths    map[string]string `json:"game_paths"`
	ZZZPoints    []ZZZPointProfile `json:"zzz_points,omitempty"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ZZZPointProfile struct {
	Width     int   `json:"width"`
	Height    int   `json:"height"`
	Account   Point `json:"account"`
	Password  Point `json:"password"`
	Agreement Point `json:"agreement"`
	Enter     Point `json:"enter"`
}

func LoadConfig() (*ConfigData, error) {
	var config ConfigData
	file, err := os.ReadFile(readConfigPath())
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
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	if config.GamePaths == nil {
		config.GamePaths = make(map[string]string)
	}
	if config.Accounts == nil {
		config.Accounts = []Account{}
	}
	if config.ZZZPoints == nil {
		config.ZZZPoints = []ZZZPointProfile{}
	}
	return &config, nil
}

func SaveConfig(config *ConfigData) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	target := writeConfigPath()
	if err := os.WriteFile(target, data, 0644); err == nil {
		return nil
	}

	_, cwdPath := appConfigPath()
	if target != cwdPath {
		return os.WriteFile(cwdPath, data, 0644)
	}
	return os.WriteFile(target, data, 0644)
}

func ExportPlaintextBackup(config *ConfigData) (string, error) {
	backupName := fmt.Sprintf("backup_%d.json", time.Now().Unix())
	target := filepath.Join(filepath.Dir(writeConfigPath()), backupName)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(target, data, 0644); err == nil {
		return backupName, nil
	}

	if err := os.WriteFile(backupName, data, 0644); err != nil {
		return "", err
	}
	return backupName, nil
}

var commonKey = []byte("MHY-STARTER-GOGO-12345678-SAFE-!")

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
