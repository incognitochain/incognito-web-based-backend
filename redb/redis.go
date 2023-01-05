package redb

import (
	"github.com/rueian/rueidis"
)

func NewClient(endpoints []string, user string, pass string) (rueidis.Client, error) {
	c, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: endpoints,
		Username:    user,
		Password:    pass,
		ClientName:  "test",
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
