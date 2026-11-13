package url

import (
	"time"
)

type URL struct {
	ID          int64
	Code        string
	OriginalURL string
	CreatedAt   time.Time
}
