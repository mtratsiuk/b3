package timestamper

import (
	"errors"
	"time"
)

type GitTimestamper struct {
}

func NewGit() GitTimestamper {
	return GitTimestamper{}
}

func (gt GitTimestamper) CreatedAt(filepath string) (time.Time, error) {

	return time.Time{}, errors.New("not implemented")
}

func (gt GitTimestamper) UpdatedAt(filepath string) (time.Time, error) {

	return time.Time{}, errors.New("not implemented")
}
