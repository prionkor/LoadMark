package k6

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Manager struct {
	Version string
}

const releaseBaseURL = "https://github.com/prionkor/xk6-output-loadmark/releases/download"

func (m Manager) Get() (string, error) {
	// cache
	cacheDir, err := m.cacheDir()
	if err != nil {
		return "", err
	}

	if executable, err := findExecutable(cacheDir); err == nil {
		return executable, nil
	}

	return m.download(cacheDir)
}

func (m Manager) assetName() (string, error) {
	switch {
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		return fmt.Sprintf("k6-loadmark-%s-linux-amd64.tar.gz", m.Version), nil

	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		return fmt.Sprintf("k6-loadmark-%s-linux-arm64.tar.gz", m.Version), nil

	case runtime.GOOS == "darwin" && runtime.GOARCH == "amd64":
		return fmt.Sprintf("k6-loadmark-%s-darwin-amd64.tar.gz", m.Version), nil

	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		return fmt.Sprintf("k6-loadmark-%s-darwin-arm64.tar.gz", m.Version), nil

	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return fmt.Sprintf("k6-loadmark-%s-windows-amd64.zip", m.Version), nil

	case runtime.GOOS == "windows" && runtime.GOARCH == "arm64":
		return fmt.Sprintf("k6-loadmark-%s-windows-arm64.zip", m.Version), nil

	default:
		return "", fmt.Errorf(
			"unsupported platform: %s/%s",
			runtime.GOOS,
			runtime.GOARCH,
		)
	}
}

func (m Manager) cacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine cache directory: %w", err)
	}

	return filepath.Join(
		cacheDir,
		"loadmark",
		"k6",
		m.Version,
		fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
	), nil
}

func (m Manager) download(cacheDir string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(cacheDir), 0755); err != nil {
		return "", fmt.Errorf("failed to create k6 cache directory: %w", err)
	}

	asset, err := m.assetName()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"%s/%s/%s",
		releaseBaseURL,
		m.Version,
		asset,
	)

	fmt.Println("Downloading k6 from:", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download k6: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to download k6: HTTP %s",
			resp.Status,
		)
	}

	archivePath := filepath.Join(
		filepath.Dir(cacheDir),
		asset,
	)

	file, err := os.Create(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to create archive: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save k6 archive: %w", err)
	}

	expected, err := m.expectedChecksum(asset)
	if err != nil {
		return "", err
	}

	if err := verifyChecksum(archivePath, expected); err != nil {
		return "", err
	}

	fmt.Println("k6 archive checksum verified")

	if err := extractArchive(archivePath, cacheDir); err != nil {
		return "", fmt.Errorf("failed to extract k6 archive: %w", err)
	}

	executable, err := findExecutable(cacheDir)
	if err != nil {
		return "", err
	}

	// cleanup the downloaded archive after extraction
	if err := os.Remove(archivePath); err != nil {
		return "", fmt.Errorf("failed to remove k6 archive: %w", err)
	}

	return executable, nil
}

func (m Manager) downloadChecksums() ([]byte, error) {
	url := fmt.Sprintf(
		"%s/%s/SHA256SUMS",
		releaseBaseURL,
		m.Version,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to download checksums: HTTP %s",
			resp.Status,
		)
	}

	return io.ReadAll(resp.Body)
}

func (m Manager) expectedChecksum(asset string) (string, error) {
	url := fmt.Sprintf(
		"%s/%s/SHA256SUMS",
		releaseBaseURL,
		m.Version,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to download checksums: HTTP %s",
			resp.Status,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksums: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)

		if len(fields) == 2 && fields[1] == asset {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("checksum not found for asset: %s", asset)
}

func verifyChecksum(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open archive for checksum: %w", err)
	}
	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actual := hex.EncodeToString(hash.Sum(nil))

	if actual != expected {
		return fmt.Errorf(
			"checksum mismatch: expected %s, got %s",
			expected,
			actual,
		)
	}

	return nil
}
func findExecutable(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read k6 cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		if runtime.GOOS == "windows" {
			if strings.EqualFold(filepath.Ext(entry.Name()), ".exe") {
				return path, nil
			}
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.Mode()&0111 != 0 {
			return path, nil
		}
	}

	return "", fmt.Errorf("k6 executable not found in %s", dir)
}

func extractArchive(archivePath, destination string) error {
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractTarGz(archivePath, destination)

	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(archivePath, destination)

	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}
func extractTarGz(archivePath, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to open gzip archive: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %w", err)
		}

		target, err := safeArchivePath(destination, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			out, err := os.OpenFile(
				target,
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
				os.FileMode(header.Mode),
			)
			if err != nil {
				return fmt.Errorf("failed to create extracted file: %w", err)
			}

			if _, err := io.Copy(out, tarReader); err != nil {
				out.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}

			if err := out.Close(); err != nil {
				return fmt.Errorf("failed to close extracted file: %w", err)
			}
		}
	}

	return nil
}
func extractZip(archivePath, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		target, err := safeArchivePath(destination, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		in, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}

		out, err := os.OpenFile(
			target,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			file.Mode(),
		)
		if err != nil {
			in.Close()
			return fmt.Errorf("failed to create extracted file: %w", err)
		}

		_, copyErr := io.Copy(out, in)
		in.Close()
		out.Close()

		if copyErr != nil {
			return fmt.Errorf("failed to extract file: %w", copyErr)
		}
	}

	return nil
}

func safeArchivePath(destination, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("archive contains absolute path: %s", name)
	}

	target := filepath.Join(destination, name)

	relative, err := filepath.Rel(destination, target)
	if err != nil {
		return "", fmt.Errorf("failed to resolve archive path %q: %w", name, err)
	}

	if relative == ".." ||
		strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive contains unsafe path: %s", name)
	}

	return target, nil
}
