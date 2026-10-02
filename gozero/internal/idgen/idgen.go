package idgen

import (
	"fmt"
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

var generator = sync.OnceValues(func() (*sonyflake.Sonyflake, error) {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if flake == nil {
		return nil, fmt.Errorf("initialize sonyflake")
	}
	return flake, nil
})

func NextID() (int64, error) {
	flake, err := generator()
	if err != nil {
		return 0, err
	}
	id, err := flake.NextID()
	if err != nil {
		return 0, fmt.Errorf("generate sonyflake ID: %w", err)
	}
	return int64(id), nil
}
