package submitproof

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"github.com/pkg/errors"
)

func StartWorker(keylist []string, cfg wcommon.Config, serviceID uuid.UUID) error {
	config = cfg
	keyList = keylist
	network := cfg.NetworkID

	err := startPubsubClient(cfg.GGCProject, cfg.GGCAuth)
	if err != nil {
		return err
	}

	shieldTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + SHIELD_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	unshieldTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + UNSHIELD_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	pappTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + PAPP_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	err = initIncClient(network)
	if err != nil {
		return err
	}

	if len(keyList) == 0 {
		return errors.New("no keys")
	}

	for _, v := range keyList {
		wl, err := wallet.Base58CheckDeserialize(v)
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

	if config.IncKey != "" {
		wl, err := wallet.Base58CheckDeserialize(config.IncKey)
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

		// err = genShardsAccount(config.IncKey)
		// if err != nil {
		// 	return err
		// }

		// for _, v := range incShardsAccount {
		// 	wl, err := wallet.Base58CheckDeserialize(v)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}

	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	var shieldSub *pubsub.Subscription
	shieldSubID := cfg.NetworkID + "_" + serviceID.String() + "_shield"
	shieldSub, err = psclient.CreateSubscription(context.Background(), shieldSubID,
		pubsub.SubscriptionConfig{Topic: shieldTxTopic})
	if err != nil {
		shieldSub = psclient.Subscription(shieldSubID)
	}
	log.Println("shieldSub.ID()", shieldSub.ID())

	var pappSub *pubsub.Subscription
	pappSubID := cfg.NetworkID + "_" + serviceID.String() + "_papp"
	pappSub, err = psclient.CreateSubscription(context.Background(), pappSubID,
		pubsub.SubscriptionConfig{Topic: pappTxTopic})
	if err != nil {
		pappSub = psclient.Subscription(pappSubID)
	}
	log.Println("pappSub.ID()", pappSub.ID())

	var unshieldSub *pubsub.Subscription
	unshieldSubID := cfg.NetworkID + "_" + serviceID.String() + "_unshield"
	unshieldSub, err = psclient.CreateSubscription(context.Background(), unshieldSubID,
		pubsub.SubscriptionConfig{Topic: shieldTxTopic})
	if err != nil {
		unshieldSub = psclient.Subscription(unshieldSubID)
	}
	log.Println("unshieldSub.ID()", unshieldSub.ID())

	go func() {
		ctx := context.Background()
		err := shieldSub.Receive(ctx, ProcessShieldRequest)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		ctx := context.Background()
		err := unshieldSub.Receive(ctx, ProcessUnshieldTxRequest)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		ctx := context.Background()
		err := pappSub.Receive(ctx, ProcessPappTxRequest)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func ProcessShieldRequest(ctx context.Context, m *pubsub.Message) {
	task := SubmitProofShieldTask{}
	defer m.Ack()
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		// rejectDelivery(delivery, payload)
		log.Println("ProcessShieldRequest error decoding message", err)
		return
	}
	if time.Since(m.PublishTime) > time.Hour {
		errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, "timeout")
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[shieldtx]` shield/redeposit timeout 😵 exttx `%v` network `%v`\n", task.TxHash, task.NetworkID))
		return
	}

	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	incTx, paymentAddr, tokenID, linkedTokenID, err := submitProof(task.TxHash, task.TokenID, task.NetworkID, key)
	if err != nil {
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[submitProof]` create tx failed `%v`, tokenID `%v`, networkID `%v`, error: `%v`\n", task.TxHash, task.TokenID, task.NetworkID, err))
		errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, err.Error())
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			return
		}
		return
	}

	errdb := database.DBUpdateShieldOnChainTxInfo(task.TxHash, task.NetworkID, paymentAddr, incTx, tokenID, linkedTokenID)
	if errdb != nil {
		log.Println("DBUpdateShieldOnChainTxInfo error:", err)
		return
	}
	err = database.DBUpdateExternalTxSubmitedRedeposit(task.TxHash, true)
	if err != nil {
		log.Println("DBUpdateExternalTxSubmitedRedeposit error:", err)
		return
	}

	err = database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusPending, "")
	if err != nil {
		log.Println("DBUpdateShieldTxStatus err:", err)
		return
	}
}
