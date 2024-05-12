package timestamper

import (
	"os/exec"
	"strings"
	"time"
)

type GitTimestamper struct {
}

func NewGit() GitTimestamper {
	return GitTimestamper{}
}

func (gt GitTimestamper) CreatedAt(filepath string) (time.Time, error) {
	log, err := getGitLogDates(filepath)
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, log[len(log)-1])
}

func (gt GitTimestamper) UpdatedAt(filepath string) (time.Time, error) {
	log, err := getGitLogDates(filepath)
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, log[0])
}

func getGitLogDates(filepath string) ([]string, error) {
	cmd := exec.Command("git", "log", "--follow", "--format=%ad", "--date=iso8601-strict", filepath)

	var out strings.Builder
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return []string{}, err
	}

	return strings.Split(strings.TrimSpace(out.String()), "\n"), nil
}
