package database

import (
	"context"
	"log"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

func DBCreateShieldTxIndex() error {
	startTime := time.Now()
	shieldTxModel := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "externaltx", Value: bsonx.Int32(1)}, {Key: "networkid", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}},
		},
	}
	_, err := mgm.Coll(&common.ShieldTxData{}).Indexes().CreateMany(context.Background(), shieldTxModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	return nil
}