package timestamper

import "time"

type Timestamper interface {
	CreatedAt(filepath string) (time.Time, error)
	UpdatedAt(filepath string) (time.Time, error)
}
