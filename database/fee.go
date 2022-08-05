package database

import (
	"context"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
)

func DBSaveFeetTable(data common.ExternalNetworksFeeData) error {
	data.Creating()
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalNetworksFeeData{}).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}
