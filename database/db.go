package database

import (
	"context"
	"log"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func ConnectDB(dbName string, mongoAddr string, network string) error {
	wc := writeconcern.New(writeconcern.W(1), writeconcern.J(true))
	err := mgm.SetDefaultConfig(nil, dbName, options.Client().ApplyURI(mongoAddr).SetWriteConcern(wc))
	if err != nil {
		return err
	}
	_, cd, _, _ := mgm.DefaultConfigs()
	err = cd.Ping(context.Background(), nil)
	if err != nil {
		return err
	}
	log.Println("Database Connected!")

	err = DBCreateUnshieldTxIndex()
	if err != nil {
		return err
	}

	err = DBCreateShieldTxIndex()
	if err != nil {
		return err
	}
	err = DBCreateFeeIndex()
	if err != nil {
		return err
	}
	err = DBCreateNetworkIndex()
	if err != nil {
		return err
	}

	err = DBCreateIndex()
	if err != nil {
		return err
	}

	err = DBCreatePappsIndex()
	if err != nil {
		return err
	}

	err = DBCreatePappSupportTokenIndex()
	if err != nil {
		return err
	}

	err = DBCreateInterSwapDataIndex()
	if err != nil {
		return err
	}

	err = DBCreateOpenSeaIndex()
	if err != nil {
		return err
	}

	DBCreateDefaultNetworkInfo(network)

	// pdao:
	err = DBCreateProposalIndex()
	if err != nil {
		return err
	}
	// pblur:
	err = DBCreatePBlurIndex()
	if err != nil {
		return err
	}

	err = DBCreateVoteIndex()
	if err != nil {
		return err
	}

	return nil
}
