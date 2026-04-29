package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics",
	Long:  "Check system configuration, registry, settings, and token storage.",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println(ui.TitleStyle.Render("aix Diagnostics"))
	fmt.Println()

	checks := []struct {
		name string
		fn   func() (bool, string)
	}{
		{"Registry file", checkRegistry},
		{"Settings file", checkSettings},
		{"Token storage", checkTokenStorage},
		{"Network connectivity", checkNetwork},
		{"File permissions", checkPermissions},
	}

	var failed int
	for _, check := range checks {
		ok, msg := check.fn()
		status := ui.Checkmark()
		if !ok {
			status = ui.Cross()
			failed++
		}
		fmt.Printf("%s %s: %s\n", status, check.name, msg)
	}

	fmt.Println()
	if failed == 0 {
		fmt.Println(ui.Success("All checks passed!"))
	} else {
		fmt.Println(ui.Warning(fmt.Sprintf("%d check(s) failed", failed)))
	}

	return nil
}

func checkRegistry() (bool, string) {
	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		return false, err.Error()
	}
	providers, _ := registry.GetAll()
	return true, fmt.Sprintf("%d provider(s) configured", len(providers))
}

func checkSettings() (bool, string) {
	url := os.Getenv(constants.EnvAnthropicBaseURL)
	if url == "" {
		return true, "No active provider (ANTHROPIC_BASE_URL not set)"
	}
	return true, fmt.Sprintf("Active: %s", url)
}

func checkTokenStorage() (bool, string) {
	tokenMgr := core.NewTokenManager()
	backend := tokenMgr.GetStorageInfo()
	return true, fmt.Sprintf("Using %s", backend)
}

func checkNetwork() (bool, string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.anthropic.com")
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	return true, "OK"
}

func checkPermissions() (bool, string) {
	paths := []string{
		constants.RegistryPath(),
		constants.SettingsPath(),
		constants.TokenDir(),
	}

	var missing []string
	for _, path := range paths {
		dir := path
		if info, err := os.Stat(path); err == nil {
			if !info.IsDir() {
				dir = filepath.Dir(path)
			}
		} else {
			// Path doesn't exist; check parent directory
			dir = filepath.Dir(path)
			missing = append(missing, filepath.Base(path))
		}

		testFile := filepath.Join(dir, ".aix-write-test")
		if f, err := os.Create(testFile); err == nil {
			f.Close()
			os.Remove(testFile)
		} else {
			return false, fmt.Sprintf("cannot write to %s", dir)
		}
	}

	if len(missing) > 0 {
		return true, fmt.Sprintf("All directories writable (%d path(s) will be created: %v)", len(missing), missing)
	}
	return true, "All directories writable"
}
