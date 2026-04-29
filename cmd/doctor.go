package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/ui"
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
	settings := &core.Settings{}
	if err := settings.Read(); err != nil {
		return false, err.Error()
	}
	env := settings.GetCurrentEnvironment()
	if len(env) == 0 {
		return true, "No active provider"
	}
	url := env[constants.EnvAnthropicBaseURL]
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

	for _, path := range paths {
		dir := path
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			dir = path[:len(path)-len("/")]
			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == '/' || path[i] == '\\' {
					dir = path[:i]
					break
				}
			}
		}
		if _, err := os.Stat(dir); err == nil {
			// Directory exists, check writability
			testFile := dir + "/.aix-write-test"
			if f, err := os.Create(testFile); err == nil {
				f.Close()
				os.Remove(testFile)
			} else {
				return false, fmt.Sprintf("cannot write to %s", dir)
			}
		}
	}
	return true, "All directories writable"
}
