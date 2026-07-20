package main

import (
	"archive/zip"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows/registry"
)

const appVersion = "1.0.0"

//go:embed payload.zip
var payloadData embed.FS

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// InstallApp extracts the embedded payload.zip to the given target directory.
func (a *App) InstallApp(targetPath string, keepData bool) string {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		return fmt.Sprintf("Lỗi tạo thư mục cài đặt: %v", err)
	}

	// Explicitly remove old data if keepData is false (clean install)
	if !keepData {
		for _, dir := range []string{"data", "www", "temp"} {
			if err := os.RemoveAll(filepath.Join(targetPath, dir)); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", dir, err)
			}
		}
		os.Remove(filepath.Join(targetPath, "ctdodb.json"))
	}

	// Open the embedded file as a stream
	embedFile, err := payloadData.Open("payload.zip")
	if err != nil {
		return fmt.Sprintf("Lỗi đọc payload.zip từ bộ nhớ: %v", err)
	}
	defer embedFile.Close()

	// Write embed stream to a temp file to avoid loading 300MB+ to RAM
	tempFile, err := os.CreateTemp("", "ctdodb-payload-*.zip")
	if err != nil {
		return fmt.Sprintf("Lỗi tạo file tạm: %v", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		tempFile.Close()
		os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, embedFile); err != nil {
		return fmt.Sprintf("Lỗi ghi file tạm: %v", err)
	}
	tempFile.Close()

	// Open zip from temp file
	zipReader, err := zip.OpenReader(tempPath)
	if err != nil {
		return fmt.Sprintf("Lỗi đọc file zip tạm: %v", err)
	}
	defer zipReader.Close()

	// Count total files for progress calculation
	totalFiles := len(zipReader.File)

	// Extract files with progress reporting
	for i, f := range zipReader.File {
		fpath := filepath.Join(targetPath, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(targetPath)+string(os.PathSeparator)) {
			return fmt.Sprintf("Lỗi đường dẫn zip hợp lệ: %s", fpath)
		}

		cleanName := filepath.ToSlash(f.Name)

		// ALWAYS skip extracting ctdodb.json from payload to prevent leaking absolute dev paths.
		if cleanName == "ctdodb.json" {
			continue
		}

		if keepData {
			if strings.HasPrefix(cleanName, "data/") || strings.HasPrefix(cleanName, "www/") || strings.HasPrefix(cleanName, "temp/") {
				if _, err := os.Stat(fpath); err == nil {
					continue
				}
			}
		}

		isDir := f.FileInfo().IsDir() || strings.HasSuffix(f.Name, "/") || strings.HasSuffix(f.Name, "\\")
		if isDir {
			if stat, err := os.Stat(fpath); err == nil && !stat.IsDir() {
				os.Remove(fpath)
			}
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return fmt.Sprintf("Lỗi tạo thư mục: %v", err)
			}
			continue
		}

		parentDir := filepath.Dir(fpath)
		if stat, err := os.Stat(parentDir); err == nil && !stat.IsDir() {
			os.Remove(parentDir)
		}
		if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
			return fmt.Sprintf("Lỗi tạo thư mục: %v", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Sprintf("Lỗi tạo file: %v", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Sprintf("Lỗi mở file zip: %v", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Sprintf("Lỗi giải nén file: %v", err)
		}

		// Emit progress event (every 10 files or last file)
		if i%10 == 0 || i == totalFiles-1 {
			progress := float64(i+1) / float64(totalFiles) * 100
			runtime.EventsEmit(a.ctx, "install:progress", map[string]interface{}{
				"progress": progress,
				"current":  i + 1,
				"total":    totalFiles,
				"file":     f.Name,
			})
		}
	}

	// Register the application in Add/Remove Programs
	if err := a.registerUninstall(targetPath); err != nil {
		fmt.Printf("Warning: Failed to register uninstall entry: %v\n", err)
	}

	return "success"
}

// CreateDesktopShortcut creates a desktop shortcut and Start Menu shortcut
func (a *App) CreateDesktopShortcut(targetPath string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("Không tìm thấy Desktop: %v", err)
	}
	desktopPath := filepath.Join(homeDir, "Desktop", "ctdodb.lnk")
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "ctdodb.lnk")

	targetExe := filepath.Join(targetPath, "ctdodb.exe")

	// Use powershell to create shortcuts
	script := fmt.Sprintf(`
		$WshShell = New-Object -comObject WScript.Shell
		$Shortcut = $WshShell.CreateShortcut("%s")
		$Shortcut.TargetPath = "%s"
		$Shortcut.WorkingDirectory = "%s"
		$Shortcut.Save()
		
		$Shortcut2 = $WshShell.CreateShortcut("%s")
		$Shortcut2.TargetPath = "%s"
		$Shortcut2.WorkingDirectory = "%s"
		$Shortcut2.Save()
	`, desktopPath, targetExe, targetPath, startMenuPath, targetExe, targetPath)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("Lỗi tạo Shortcut: %v", err)
	}

	return "success"
}

// LaunchApp opens the installed ctdodb.exe
func (a *App) LaunchApp(targetPath string) string {
	targetExe := filepath.Join(targetPath, "ctdodb.exe")
	cmd := exec.Command(targetExe)
	cmd.Dir = targetPath
	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("Lỗi khởi chạy app: %v", err)
	}
	return "success"
}

// SelectDirectory opens a directory selection dialog
func (a *App) SelectDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Chọn thư mục cài đặt",
	})
	if err != nil {
		return ""
	}
	return dir
}

// registerUninstall adds the application to Windows Add/Remove Programs
func (a *App) registerUninstall(targetPath string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\CTDODB`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()

	exePath := filepath.Join(targetPath, "ctdodb.exe")

	homeDir, _ := os.UserHomeDir()
	desktopPath := filepath.Join(homeDir, "Desktop", "ctdodb.lnk")
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "ctdodb.lnk")

	uninstallCmd := fmt.Sprintf(`cmd.exe /c "timeout /t 2 >nul & rmdir /s /q \"%s\" & del /q \"%s\" & del /q \"%s\" & reg delete HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\CTDODB /f"`, targetPath, desktopPath, startMenuPath)

	k.SetStringValue("DisplayName", "CTDODB Local Server")
	k.SetStringValue("DisplayIcon", exePath)
	k.SetStringValue("UninstallString", uninstallCmd)
	k.SetStringValue("DisplayVersion", appVersion)
	k.SetStringValue("Publisher", "ctdoteam")
	k.SetStringValue("InstallLocation", targetPath)
	k.SetDWordValue("NoModify", 1)
	k.SetDWordValue("NoRepair", 1)

	return nil
}