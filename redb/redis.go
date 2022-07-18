package redb

import (
	"github.com/rueian/rueidis"
)

func NewClient(endpoints []string) (rueidis.Client, error) {
	c, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: endpoints,
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
