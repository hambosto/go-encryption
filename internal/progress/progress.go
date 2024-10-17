package progress

import (
	"fmt"
	"time"
)

// ProgressReporter tracks the progress of an operation and provides visual feedback.
// It can operate in stealth mode, suppressing output if needed.
type ProgressReporter struct {
	stealthMode bool      // Indicates if stealth mode is enabled
	startTime   time.Time // Time when the operation started
	lastUpdate  time.Time // Time of the last status update
}

// NewProgressReporter initializes a new ProgressReporter instance.
// It takes a boolean parameter to enable or disable stealth mode.
func NewProgressReporter(stealthMode bool) *ProgressReporter {
	return &ProgressReporter{
		stealthMode: stealthMode,
	}
}

// Start marks the beginning of an operation and records the start time.
// It prints a message indicating that the operation has started unless stealth mode is enabled.
func (p *ProgressReporter) Start(operation string) {
	p.startTime = time.Now() // Record the current time as start time
	p.lastUpdate = p.startTime
	if !p.stealthMode {
		fmt.Printf("\n%s... Started\n", operation) // Notify the start of the operation
	}
}

// Update reports the current status and progress of the operation.
// It takes a status message and a progress value (between 0.0 and 1.0).
// In stealth mode, no updates are printed. The method throttles updates to avoid flooding.
func (p ProgressReporter) Update(status string, progress float64) {
	if p.stealthMode {
		return // Suppress output in stealth mode
	}
	now := time.Now()
	if now.Sub(p.lastUpdate) < time.Millisecond {
		return // Throttle updates to avoid excessive output
	}
	p.lastUpdate = now

	// Select an emoji based on the progress value
	var emoji string
	switch {
	case progress < 0.25:
		emoji = "🚶" // Walking
	case progress < 0.5:
		emoji = "🏃" // Running
	case progress < 0.75:
		emoji = "🚀" // Rocketing
	default:
		emoji = "✅" // Completed
	}

	// Print the current status and emoji, overwriting the previous line
	fmt.Printf("\r%-50s", fmt.Sprintf("%s %s", status, emoji))
}

// Complete indicates that the operation has finished.
// It prints a completion message along with the duration of the operation unless stealth mode is enabled.
func (p *ProgressReporter) Complete(message string) {
	if !p.stealthMode {
		duration := time.Since(p.startTime).Round(time.Millisecond) // Calculate the duration
		fmt.Printf("\n%s (took %v)\n", message, duration)           // Notify completion and duration
	}
}

