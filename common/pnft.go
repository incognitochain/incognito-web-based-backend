package common

import "github.com/kamva/mgm/v3"

type ListNftCache struct {
	mgm.DefaultModel `bson:",inline"`
	Address          string `bson:"status"`
	Data             string
}
