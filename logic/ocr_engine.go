package logic

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall" // 新增导入
)

type OCRRawResponse struct {
	Code int `json:"code"`
	Data []struct {
		Text string  `json:"text"`
		Box  [][]int `json:"box"`
	} `json:"data"`
}

type TextPoint struct {
	Text   string
	X      int
	Y      int
	LeftX  int
	RightX int
	Width  int
	Height int
}

func RecognizeWithPos(imagePath string) ([]TextPoint, error) {
	exeRelPath := "./bin/PaddleOCR-json_v1.4.1/PaddleOCR-json.exe"
	absExe, _ := filepath.Abs(exeRelPath)
	absImg, _ := filepath.Abs(imagePath)

	cmd := exec.Command(absExe, "--image_path="+absImg)
	cmd.Dir = filepath.Dir(absExe)

	// --- 核心修复：隐藏窗口属性 ---
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("OCR 执行失败: %v", err)
	}

	outputStr := string(output)
	jsonStart := strings.Index(outputStr, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("OCR 无效输出")
	}

	var raw OCRRawResponse
	if err := json.Unmarshal([]byte(outputStr[jsonStart:]), &raw); err != nil {
		return nil, err
	}

	var res []TextPoint
	for _, d := range raw.Data {
		res = append(res, TextPoint{
			Text:   d.Text,
			X:      (d.Box[0][0] + d.Box[1][0]) / 2,
			Y:      (d.Box[0][1] + d.Box[2][1]) / 2,
			LeftX:  d.Box[0][0],
			RightX: d.Box[1][0],
			Width:  d.Box[1][0] - d.Box[0][0],
			Height: d.Box[2][1] - d.Box[0][1],
		})
	}
	return res, nil
}
