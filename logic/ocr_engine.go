package logic

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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
	RightX int // 新增：右边界坐标
	Width  int
	Height int
}

func RecognizeWithPos(imagePath string) ([]TextPoint, error) {
	exeRelPath := "./bin/PaddleOCR-json_v1.4.1/PaddleOCR-json.exe"
	absExe, _ := filepath.Abs(exeRelPath)
	absImg, _ := filepath.Abs(imagePath)

	cmd := exec.Command(absExe, "--image_path="+absImg)
	cmd.Dir = filepath.Dir(absExe)

	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	jsonStart := strings.Index(outputStr, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("OCR 无效输出")
	}

	var raw OCRRawResponse
	if err := json.Unmarshal([]byte(outputStr[jsonStart:]), &raw); err != nil {
		return nil, err
	}

	var results []TextPoint
	for _, item := range raw.Data {
		if len(item.Box) == 4 {
			results = append(results, TextPoint{
				Text:   item.Text,
				X:      (item.Box[0][0] + item.Box[2][0]) / 2,
				Y:      (item.Box[0][1] + item.Box[2][1]) / 2,
				LeftX:  item.Box[0][0],
				RightX: item.Box[1][0],
				Width:  item.Box[2][0] - item.Box[0][0],
				Height: item.Box[2][1] - item.Box[0][1],
			})
		}
	}
	return results, nil
}
