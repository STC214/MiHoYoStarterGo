package logic

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/yusufpapurcu/wmi"
)

type Win32_Processor struct {
	ProcessorId string
}

type Win32_BaseBoard struct {
	SerialNumber string
}

// GetDeviceFingerprint 获取硬件唯一指纹
func GetDeviceFingerprint() string {
	var cpu []Win32_Processor
	var board []Win32_BaseBoard

	// 1. 获取 CPU ID
	err := wmi.Query("SELECT ProcessorId FROM Win32_Processor", &cpu)
	cpuID := "UnknownCPU"
	if err == nil && len(cpu) > 0 {
		cpuID = strings.TrimSpace(cpu[0].ProcessorId)
	}

	// 2. 获取主板序列号
	err = wmi.Query("SELECT SerialNumber FROM Win32_BaseBoard", &board)
	boardID := "UnknownBoard"
	if err == nil && len(board) > 0 {
		boardID = strings.TrimSpace(board[0].SerialNumber)
	}

	// 3. 按照原项目的逻辑拼接并计算 MD5
	raw := fmt.Sprintf("%s-%s", cpuID, boardID)
	hash := md5.Sum([]byte(raw))

	return fmt.Sprintf("%x", hash)
}
