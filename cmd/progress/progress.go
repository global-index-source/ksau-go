package progress

import (
	"fmt"
	"strings"
	"time"
)

// ProgressStyle represents different progress bar styles
type ProgressStyle string

const (
	StyleBasic   ProgressStyle = "basic"   // [=====>     ]
	StyleBlocks  ProgressStyle = "blocks"  // █████░░░░░
	StyleModern  ProgressStyle = "modern"  // ○○●●●○○○
	StyleEmoji   ProgressStyle = "emoji"   // 🟦🟦🟦⬜⬜
	StyleMinimal ProgressStyle = "minimal" // 43% | 4.2MB/s
)

// ValidStyles returns a list of valid progress bar styles
func ValidStyles() []ProgressStyle {
	return []ProgressStyle{
		StyleBasic,
		StyleBlocks,
		StyleModern,
		StyleEmoji,
		StyleMinimal,
	}
}

// ProgressTracker keeps track of upload progress
type ProgressTracker struct {
	TotalSize     int64
	UploadedSize  int64
	StartTime     time.Time
	LastUpdate    time.Time
	Style         ProgressStyle
	CustomEmoji   string
	Width         int
	LastChunkSize int64
	LastSpeed     float64
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalSize int64, style ProgressStyle) *ProgressTracker {
	return &ProgressTracker{
		TotalSize:   totalSize,
		StartTime:   time.Now(),
		LastUpdate:  time.Now(),
		Style:       style,
		Width:       40, // default width
		CustomEmoji: "🟦",
	}
}

// UpdateProgress updates the progress and displays the progress bar
func (p *ProgressTracker) UpdateProgress(uploadedSize int64) {
	p.UploadedSize = uploadedSize
	now := time.Now()
	elapsed := now.Sub(p.LastUpdate).Seconds()

	// Calculate speed (bytes per second)
	chunkSize := uploadedSize - p.LastChunkSize
	speed := float64(chunkSize) / elapsed
	if elapsed >= 1.0 { // Update speed calculation if at least 1 second has passed
		p.LastSpeed = speed
		p.LastUpdate = now
		p.LastChunkSize = uploadedSize
	}

	p.displayProgress()
}

func (p *ProgressTracker) displayProgress() {
	percent := float64(p.UploadedSize) * 100 / float64(p.TotalSize)

	var progressBar string
	switch p.Style {
	case StyleBasic:
		progressBar = p.basicStyle(percent)
	case StyleBlocks:
		progressBar = p.blockStyle(percent)
	case StyleModern:
		progressBar = p.modernStyle(percent)
	case StyleEmoji:
		progressBar = p.emojiStyle(percent)
	case StyleMinimal:
		progressBar = p.minimalStyle(percent)
	default:
		progressBar = p.basicStyle(percent)
	}

	// Clear line and show progress
	fmt.Printf("\r\033[K%s", progressBar)
}

func (p *ProgressTracker) basicStyle(percent float64) string {
	width := p.Width - 2 // Account for brackets
	complete := int((percent / 100) * float64(width))
	return fmt.Sprintf("[%s>%s] %.1f%% | %s/s",
		strings.Repeat("=", complete),
		strings.Repeat(" ", width-complete-1),
		percent,
		formatBytes(p.LastSpeed))
}

func (p *ProgressTracker) blockStyle(percent float64) string {
	width := p.Width
	complete := int((percent / 100) * float64(width))
	return fmt.Sprintf("%s%s %.1f%% | %s/s",
		strings.Repeat("█", complete),
		strings.Repeat("░", width-complete),
		percent,
		formatBytes(p.LastSpeed))
}

func (p *ProgressTracker) modernStyle(percent float64) string {
	width := p.Width
	complete := int((percent / 100) * float64(width))
	return fmt.Sprintf("%s%s %.1f%% | %s/s",
		strings.Repeat("●", complete),
		strings.Repeat("○", width-complete),
		percent,
		formatBytes(p.LastSpeed))
}

func (p *ProgressTracker) emojiStyle(percent float64) string {
	width := p.Width / 2 // Emojis are typically double-width
	complete := int((percent / 100) * float64(width))
	emoji := p.CustomEmoji
	if emoji == "" {
		emoji = "🟦"
	}
	return fmt.Sprintf("%s%s %.1f%% | %s/s",
		strings.Repeat(emoji, complete),
		strings.Repeat("⬜", width-complete),
		percent,
		formatBytes(p.LastSpeed))
}

func (p *ProgressTracker) minimalStyle(percent float64) string {
	timeElapsed := time.Since(p.StartTime)
	eta := time.Duration(float64(timeElapsed)*(100/percent) - float64(timeElapsed))
	if percent == 0 {
		eta = 0
	}

	return fmt.Sprintf("%.1f%% | %s/s | %s/%s | ETA: %s",
		percent,
		formatBytes(p.LastSpeed),
		formatBytes(float64(p.UploadedSize)),
		formatBytes(float64(p.TotalSize)),
		formatDuration(eta))
}

// Finish prints final progress and moves to next line
func (p *ProgressTracker) Finish() {
	p.UpdateProgress(p.TotalSize)
	fmt.Println()
}

// Helper functions
func formatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.1f B", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", bytes/div, "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
