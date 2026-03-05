package application

import (
	"time"

	"github.com/google/uuid"
)

func generateID() string {
	return uuid.NewString()
}

func now() int64 {
	return time.Now().Unix()
}
