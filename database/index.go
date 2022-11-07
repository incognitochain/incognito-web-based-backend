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
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}, {Key: "created_at", Value: bsonx.Int32(1)}},
		},
	}
	_, err := mgm.Coll(&common.ShieldTxData{}).Indexes().CreateMany(context.Background(), shieldTxModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	return nil
}

func DBCreateFeeIndex() error {
	startTime := time.Now()
	feeModel := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "created_at", Value: bsonx.Int32(1)}},
		},
	}
	_, err := mgm.Coll(&common.ExternalNetworksFeeData{}).Indexes().CreateMany(context.Background(), feeModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	return nil
}

func DBCreateNetworkIndex() error {
	startTime := time.Now()
	networkInfoModel := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "network", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := mgm.Coll(&common.BridgeNetworkData{}).Indexes().CreateMany(context.Background(), networkInfoModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	return nil
}

func DBCreatePappsIndex() error {
	startTime := time.Now()
	pappsModel := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "network", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := mgm.Coll(&common.PAppsEndpointData{}).Indexes().CreateMany(context.Background(), pappsModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	return nil
}

func DBCreateIndex() error {
	startTime := time.Now()
	externalTxModel := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "txhash", Value: bsonx.Int32(1)}, {Key: "network", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bsonx.Doc{{Key: "increquesttx", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "created_at", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}, {Key: "network", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}, {Key: "will_redeposit", Value: bsonx.Int32(1)}, {Key: "redeposit_submitted", Value: bsonx.Int32(1)}},
		},
	}
	_, err := mgm.Coll(&common.ExternalTxStatus{}).Indexes().CreateMany(context.Background(), externalTxModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	pappsModel := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "inctx", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},

		{
			Keys: bsonx.Doc{{Key: "externaltx", Value: bsonx.Int32(1)}},
			// Options: options.Index().SetUnique(true),
		},
		{
			Keys: bsonx.Doc{{Key: "created_at", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}, {Key: "type", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}, {Key: "refundsubmitted", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}, {Key: "refundsubmitted", Value: bsonx.Int32(1)}, {Key: "refundpfee", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "outchain_status", Value: bsonx.Int32(1)}},
		},
	}
	_, err = mgm.Coll(&common.PappTxData{}).Indexes().CreateMany(context.Background(), pappsModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	pappsAddressModel := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "network", Value: bsonx.Int32(1)}, {Key: "type", Value: bsonx.Int32(1)}},
		},
	}
	_, err = mgm.Coll(&common.PappVaultData{}).Indexes().CreateMany(context.Background(), pappsAddressModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	feeRefundModel := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "increquesttx", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}},
		},
	}
	_, err = mgm.Coll(&common.RefundFeeData{}).Indexes().CreateMany(context.Background(), feeRefundModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	dexSwapModel := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "inctx", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "status", Value: bsonx.Int32(1)}},
		},
	}
	_, err = mgm.Coll(&common.DexSwapTrackData{}).Indexes().CreateMany(context.Background(), dexSwapModel)
	if err != nil {
		log.Printf("failed to index coins in %v", time.Since(startTime))
		return err
	}

	return nil
}

func DBCreatePappSupportTokenIndex() error {
	pappTokenModel := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "tokenid", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "contractid", Value: bsonx.Int32(1)}},
		},
		{
			Keys: bsonx.Doc{{Key: "verify", Value: bsonx.Int32(1)}},
		},
	}
	_, err := mgm.Coll(&common.PappSupportedTokenData{}).Indexes().CreateMany(context.Background(), pappTokenModel)
	if err != nil {
		log.Println("failed to index tokens")
		return err
	}
	return nil
}
