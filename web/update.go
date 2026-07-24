package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type versionResponse struct {
	Version string `json:"version"`
}

type checkUpdateResponse struct {
	HasUpdate     bool   `json:"has_update"`
	LatestVersion string `json:"latest_version"`
	DownloadURL   string `json:"download_url"`
	ReleaseNotes  string `json:"release_notes"`
}

type applyUpdateRequest struct {
	DownloadURL string `json:"download_url"`
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Body    string        `json:"body"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

const githubRepo = "cn-maul/Gentry"

func (s *WebServer) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, NewSuccessResponse(versionResponse{Version: s.version}))
}

func (s *WebServer) checkUpdate(c *gin.Context) {
	release, err := fetchLatestRelease()
	if err != nil {
		c.JSON(http.StatusOK, NewSuccessResponse(checkUpdateResponse{HasUpdate: false}))
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(s.version, "v")

	if currentVersion == "dev" || currentVersion == latestVersion {
		c.JSON(http.StatusOK, NewSuccessResponse(checkUpdateResponse{HasUpdate: false}))
		return
	}

	assetName := platformAssetName()
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	c.JSON(http.StatusOK, NewSuccessResponse(checkUpdateResponse{
		HasUpdate:     true,
		LatestVersion: release.TagName,
		DownloadURL:   downloadURL,
		ReleaseNotes:  release.Body,
	}))
}

func (s *WebServer) applyUpdate(c *gin.Context) {
	var req applyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.DownloadURL == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "缺少下载地址"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(gin.H{"message": "更新下载中，完成后将自动重启"}))

	go func() {
		if err := performUpdate(req.DownloadURL); err != nil {
			log.Printf("[Update] 升级失败: %v", err)
		}
	}()
}

func fetchLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func platformAssetName() string {
	if runtime.GOOS == "windows" {
		return "gentry-windows-amd64.exe"
	}
	return "gentry-linux-amd64"
}

func performUpdate(downloadURL string) error {
	time.Sleep(1 * time.Second)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前程序路径失败: %w", err)
	}
	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("解析绝对路径失败: %w", err)
	}
	execDir := filepath.Dir(execPath)

	log.Printf("[Update] 当前程序路径: %s", execPath)

	tmpFile := filepath.Join(execDir, ".gentry-update-tmp")
	backupFile := execPath + ".bak"

	if err := downloadFile(downloadURL, tmpFile); err != nil {
		return fmt.Errorf("下载更新文件失败: %w", err)
	}
	log.Printf("[Update] 下载完成: %s", tmpFile)

	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("设置执行权限失败: %w", err)
	}

	if err := os.Rename(execPath, backupFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("备份旧程序失败: %w", err)
	}
	log.Printf("[Update] 已备份旧程序: %s", backupFile)

	if err := os.Rename(tmpFile, execPath); err != nil {
		os.Rename(backupFile, execPath)
		return fmt.Errorf("替换程序失败，已恢复: %w", err)
	}
	log.Printf("[Update] 新程序已就位，准备重启")

	if runtime.GOOS == "windows" {
		return restartWindows(execPath)
	}
	return restartUnix(execPath)
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载返回 %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func restartUnix(execPath string) error {
	log.Printf("[Update] 正在重启...")
	time.Sleep(500 * time.Millisecond)
	return syscall.Exec(execPath, os.Args, os.Environ())
}

func restartWindows(execPath string) error {
	batContent := fmt.Sprintf(`@echo off
timeout /t 3 /nobreak >nul
move "%s" "%s" >nul
start "" "%s"
del "%%~f0"`, execPath+".bak", execPath+".bak.old", execPath)

	batPath := filepath.Join(filepath.Dir(execPath), "gentry-upgrade.bat")
	if err := os.WriteFile(batPath, []byte(batContent), 0755); err != nil {
		return fmt.Errorf("创建升级脚本失败: %w", err)
	}

	cmd := exec.Command("cmd", "/c", "start", "/b", batPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动升级脚本失败: %w", err)
	}

	log.Printf("[Update] 升级脚本已启动，程序即将退出: %s", batPath)
	os.Exit(0)
	return nil
}