package main

import (
	"fmt"
	"os"
	"path/filepath"
	"golang.org/x/sys/windows/registry"
)

// RegisterPXGFProtocol hooks the custom URI scheme into the Windows Registry.
// This must be run with Administrator privileges.
func RegisterPXGFProtocol() error {
	// Get the absolute path of your compiled game executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable path: %w", err)
	}
	exePath = filepath.Clean(exePath)

	// 1. Create the base key under HKEY_CLASSES_ROOT
	key, _, err := registry.CreateKey(registry.CLASSES_ROOT, `pxgf`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create pxgf registry key (try running as Admin): %w", err)
	}
	defer key.Close()

	// Set protocol identifiers
	if err := key.SetStringValue("", "URL:PXGF Noisogic Protocol"); err != nil {
		return err
	}
	if err := key.SetStringValue("URL Protocol", ""); err != nil {
		return err
	}

	// 2. Create the shell\open\command subkey tree
	cmdKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, `pxgf\shell\open\command`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create shell command tree: %w", err)
	}
	defer cmdKey.Close()

	// The "%1" argument tells Windows to forward the full URL string (e.g., pxgf://launch/1) to your Go app
	commandStr := fmt.Sprintf(`"%s" "%%1"`, exePath)
	if err := cmdKey.SetStringValue("", commandStr); err != nil {
		return err
	}

	fmt.Println("[PXGF] Protocol successfully registered to Windows shell!")
	return nil
}

// CheckArgsAndLaunch intercepts the launch arguments if fired by the browser
func CheckArgsAndLaunch() {
	if len(os.Args) > 1 {
		launchURL := os.Args[1]
		if filepath.Ext(launchURL) == "" && (filepath.HasPrefix(launchURL, "pxgf://") || filepath.HasPrefix(launchURL, "pxgf:")) {
			fmt.Printf("[Noisogic Engine] Booting up via protocol hook!\n")
			fmt.Printf("Target Request Payload: %s\n", launchURL)
			
			// Your G3D/OpenGL window boot logic goes here!
			// Extract the Place ID out of the string and fetch the .pxgf file.
			os.Exit(0)
		}
	}
}

// Inside pxgf_protocol.go
func CheckArgsAndLaunch() {
	if len(os.Args) > 1 {
		launchURL := os.Args[1]
		
		// If the link came from the website browser (e.g., pxgf://launch?src=http://...)
		if strings.HasPrefix(launchURL, "pxgf://") {
			// Basic text parsing to slice out the raw HTTP download URL from the string
			parts := strings.Split(launchURL, "src=")
			if len(parts) > 1 {
				downloadTarget := parts[1]
				
				// Un-escape any URL encoding characters
				downloadTarget = strings.ReplaceAll(downloadTarget, "%3A", ":")
				downloadTarget = strings.ReplaceAll(downloadTarget, "%2F", "/")

				fmt.Println("[Launcher] Protocol valid. Waking up Raylib render architecture...")
				
				// TRIGGER RENDERING EXECUTABLE STEP!
				StartEngineRender(downloadTarget)
				os.Exit(0)
			}
		}
	}
}