package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
)

func CapturePointWithCountdown(ctx context.Context, winLeft, winTop, width, height int, countdown time.Duration) (Point, error) {
	select {
	case <-ctx.Done():
		return Point{}, ctx.Err()
	case <-time.After(countdown):
	}

	mx, my := robotgo.GetMousePos()
	x := mx - winLeft
	y := my - winTop
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > width-1 {
		x = width - 1
	}
	if y > height-1 {
		y = height - 1
	}
	return Point{X: x, Y: y}, nil
}

func SaveZZZPointProfile(profile ZZZPointProfile) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	out := make([]ZZZPointProfile, 0, len(cfg.ZZZPoints)+1)
	replaced := false
	for _, p := range cfg.ZZZPoints {
		if p.Width == profile.Width && p.Height == profile.Height {
			out = append(out, profile)
			replaced = true
			continue
		}
		out = append(out, p)
	}
	if !replaced {
		out = append(out, profile)
	}
	cfg.ZZZPoints = out
	return SaveConfig(cfg)
}

func LoadZZZPointProfile(width, height int) (ZZZPointProfile, bool, bool, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return ZZZPointProfile{}, false, false, err
	}
	if len(cfg.ZZZPoints) == 0 {
		return ZZZPointProfile{}, false, false, nil
	}
	for _, p := range cfg.ZZZPoints {
		if p.Width == width && p.Height == height {
			return p, true, true, nil
		}
	}

	best := cfg.ZZZPoints[0]
	bestScore := absInt(best.Width-width) + absInt(best.Height-height)
	for _, p := range cfg.ZZZPoints[1:] {
		score := absInt(p.Width-width) + absInt(p.Height-height)
		if score < bestScore {
			best = p
			bestScore = score
		}
	}
	return best, false, true, nil
}

func ValidateZZZProfile(p ZZZPointProfile) error {
	if p.Width <= 0 || p.Height <= 0 {
		return fmt.Errorf("invalid resolution")
	}
	if p.Account == (Point{}) || p.Password == (Point{}) || p.Agreement == (Point{}) || p.Enter == (Point{}) {
		return fmt.Errorf("incomplete zzz points")
	}
	return nil
}
