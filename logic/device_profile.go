package logic

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const deviceProfileRoot = `Software\MiHoYoStarterGo\Profiles`

// ApplySavedDeviceFingerprint writes the saved device fingerprint to a stable profile location.
// This keeps account-bound hardware metadata synchronized before launching login flow.
func ApplySavedDeviceFingerprint(gameID, fingerprint string) error {
	keyPath := fmt.Sprintf(`%s\%s`, deviceProfileRoot, gameID)
	k, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open profile registry failed: %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue("device_fingerprint", fingerprint); err != nil {
		return fmt.Errorf("write device fingerprint failed: %w", err)
	}
	return nil
}
