package engine

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// DownloadFailureKind classifies failures reported by the managed downloader.
type DownloadFailureKind string

const (
	// DownloadFailureGeneric is an unclassified downloader failure.
	DownloadFailureGeneric DownloadFailureKind = "generic"
	// DownloadFailureCompatibility indicates that a supported site changed in a
	// way the pinned downloader could not handle.
	DownloadFailureCompatibility DownloadFailureKind = "compatibility"
)

// DownloaderError contains structured diagnostics from the pinned yt-dlp
// executable. Callers can use errors.As without parsing the error message.
type DownloaderError struct {
	Version  string
	ExitCode int
	Details  string
	Kind     DownloadFailureKind
	Err      error
}

// Error implements the error interface.
func (e *DownloaderError) Error() string {
	version := e.Version
	if version == "" {
		version = "unknown"
	}
	if e.ExitCode >= 0 {
		return fmt.Sprintf("yt-dlp %s failed with exit code %d", version, e.ExitCode)
	}
	return fmt.Sprintf("yt-dlp %s failed", version)
}

// Unwrap exposes the execution error for errors.Is and errors.As.
func (e *DownloaderError) Unwrap() error {
	return e.Err
}

// IsCompatibilityFailure reports whether the pinned downloader appears unable
// to handle the current YouTube response or format behavior.
func (e *DownloaderError) IsCompatibilityFailure() bool {
	return e.Kind == DownloadFailureCompatibility
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func newDownloaderError(rawURL, version string, exitCode int, details string, err error) *DownloaderError {
	details = normalizeDownloaderDetails(details)
	return &DownloaderError{
		Version:  version,
		ExitCode: exitCode,
		Details:  details,
		Kind:     classifyDownloadFailure(rawURL, details),
		Err:      err,
	}
}

func classifyDownloadFailure(rawURL, details string) DownloadFailureKind {
	if !isYouTubeURL(rawURL) {
		return DownloadFailureGeneric
	}

	lower := strings.ToLower(details)
	compatibilityMarkers := []string{
		"sabr",
		"http error 403",
		"403: forbidden",
		"requested format is not available",
		"older than 90 days",
	}
	for _, marker := range compatibilityMarkers {
		if strings.Contains(lower, marker) {
			return DownloadFailureCompatibility
		}
	}
	return DownloadFailureGeneric
}

func isYouTubeURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "youtu.be" || host == "youtube.com" || strings.HasSuffix(host, ".youtube.com")
}

func normalizeDownloaderDetails(details string) string {
	details = ansiPattern.ReplaceAllString(details, "")
	lines := strings.Split(strings.ReplaceAll(details, "\r\n", "\n"), "\n")
	normalized := make([]string, 0, len(lines))
	skippingUpdateWarning := false
	previous := ""

	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if strings.HasPrefix(line, "WARNING: Your yt-dlp version") {
			skippingUpdateWarning = true
			continue
		}
		if skippingUpdateWarning {
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				continue
			}
			skippingUpdateWarning = false
		}
		if strings.TrimSpace(line) == "" {
			if len(normalized) > 0 && previous != "" {
				normalized = append(normalized, "")
				previous = ""
			}
			continue
		}
		if line == previous {
			continue
		}
		normalized = append(normalized, line)
		previous = line
	}

	return strings.TrimSpace(strings.Join(normalized, "\n"))
}
