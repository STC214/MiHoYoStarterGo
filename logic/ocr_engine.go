package logic

import (
	"encoding/json"
	"fmt"
	"math"
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

func parseOCRBox(box [][]int) (left, right, top, bottom int, ok bool) {
	if len(box) == 0 {
		return 0, 0, 0, 0, false
	}

	left = math.MaxInt
	right = math.MinInt
	top = math.MaxInt
	bottom = math.MinInt

	for _, pt := range box {
		if len(pt) < 2 {
			return 0, 0, 0, 0, false
		}
		x, y := pt[0], pt[1]
		if x < left {
			left = x
		}
		if x > right {
			right = x
		}
		if y < top {
			top = y
		}
		if y > bottom {
			bottom = y
		}
	}

	if right <= left || bottom <= top {
		return 0, 0, 0, 0, false
	}

	return left, right, top, bottom, true
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
	invalidCount := 0
	for _, d := range raw.Data {
		left, right, top, bottom, ok := parseOCRBox(d.Box)
		if !ok {
			invalidCount++
			continue
		}

		res = append(res, TextPoint{
			Text:   strings.TrimSpace(d.Text),
			X:      (left + right) / 2,
			Y:      (top + bottom) / 2,
			LeftX:  left,
			RightX: right,
			Width:  right - left,
			Height: bottom - top,
		})
	}

	if len(res) == 0 {
		if invalidCount > 0 {
			return nil, fmt.Errorf("OCR 结果结构异常：%d 条数据坐标无效", invalidCount)
		}
		return nil, fmt.Errorf("OCR 未识别到有效文本")
	}

	if invalidCount > 0 {
		fmt.Printf("[OCR] 跳过 %d 条无效坐标数据\n", invalidCount)
	}

	return res, nil
}
