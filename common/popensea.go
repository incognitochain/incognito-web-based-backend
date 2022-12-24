package common

import (
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
	"github.com/kamva/mgm/v3"
)

type OpenseaCollectionData struct {
	mgm.DefaultModel `bson:",inline"`
	Address          string                    `bson:"address"`
	Name             string                    `bson:"name"`
	Detail           popensea.CollectionDetail `bson:"detail"`
}

type OpenseaAssetData struct {
	mgm.DefaultModel `bson:",inline"`
	UID              string             `bson:"uid"`
	Address          string             `bson:"address"`
	TokenID          string             `bson:"token_id"`
	Name             string             `bson:"name"`
	Detail           popensea.NFTDetail `bson:"detail"`
}
