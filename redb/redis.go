package redb

import (
	"github.com/rueian/rueidis"
)

func NewClient(endpoints []string) (rueidis.Client, error) {
	c, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"127.0.0.1:6379"},
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
