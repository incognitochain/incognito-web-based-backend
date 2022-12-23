package interswap

import (
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

var config wcommon.Config

// endpoint Be to call request to estimate swap info
var APIEndpoint string

func StartService(cfg wcommon.Config, serviceID uuid.UUID) error {
	network := cfg.NetworkID

	// TODO: 0xkraken review pubsub

	// start client
	// err := startPubsubClient(cfg.GGCProject, cfg.GGCAuth)
	// if err != nil {
	// 	return err
	// }

	// init topic instance
	// pappTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + PAPP_TX_TOPIC)
	// if err != nil {
	// 	panic(err)
	// }

	// init incognito client
	err := InitIncClient(network)
	if err != nil {
		return err
	}

	err = InitSupportedMidTokens(network)
	if err != nil {
		return err
	}

	APIEndpoint = "0.0.0.0:" + strconv.Itoa(cfg.Port)

	// submit OTA key to fullnode
	if cfg.ISIncPrivKey != "" {
		wl, err := wallet.Base58CheckDeserialize(cfg.ISIncPrivKey)
		if err != nil {
			panic(err)
		}
		if cfg.FullnodeAuthKey != "" {
			err = incClient.AuthorizedSubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType), cfg.FullnodeAuthKey, 0, false)
			if err != nil {
				if !strings.Contains(err.Error(), "has been submitted") {
					return err
				}
			}
		} else {
			err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
			if err != nil {
				if !strings.Contains(err.Error(), "has been submitted") {
					return err
				}
			}
		}
	}

	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	// TODO: 0xkraken review pubsub

	// init subscription

	// var pappSub *pubsub.Subscription
	// pappSubID := cfg.NetworkID + "_" + serviceID.String() + "_papp"
	// pappSub, err = psclient.CreateSubscription(context.Background(), pappSubID,
	// 	pubsub.SubscriptionConfig{Topic: pappTxTopic})
	// if err != nil {
	// 	pappSub = psclient.Subscription(pappSubID)
	// }
	// log.Println("pappSub.ID()", pappSub.ID())

	// start subscription to receive msg and req workers execute something

	// go func() {
	// 	ctx := context.Background()
	// 	err := pappSub.Receive(ctx, ProcessPappTxRequest)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }()

	return nil
}
