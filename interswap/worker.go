package interswap

import (
	"context"
	"log"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/utxomanager"
)

var config wcommon.Config
var InterswapIncKeySets map[string]*wallet.KeyWallet
var UtxoManager *utxomanager.UTXOManager

func StartWorker(cfg wcommon.Config, serviceID uuid.UUID) error {
	network := cfg.NetworkID
	config = cfg

	// init incognito client
	err := InitIncClient(network, cfg)
	if err != nil {
		return err
	}
	UtxoManager = utxomanager.NewUTXOManager(incClient)
	log.Printf("UtxoManager len: %v\n", UtxoManager.Caches)

	// submit OTA key to fullnode
	if len(cfg.ISIncPrivKeys) > 0 {
		for _, key := range cfg.ISIncPrivKeys {
			wl, err := wallet.Base58CheckDeserialize(key)
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

	}

	go watchInterswapPendingTx(config)
	go monitorBalanceISIncKeys(config)

	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	return nil
}

func ProcessInterswapTxRequest(ctx context.Context, m *pubsub.Message) {
	taskDesc := m.Attributes["task"]
	switch taskDesc {
	case InterswapPdexPappTxTask:
		processInterswapPdexPappPathTask(ctx, m)
	}
}

type InterswapSubmitTxTask struct {
	TxID                string `json:"txid" bson:"txid"`
	TxRawBytes          []byte `json:"txraw" bson:"txraw"`
	FromToken           string `json:"fromtoken" bson:"fromtoken"`
	ToToken             string `json:"totoken" bson:"totoken"`
	MidToken            string `json:"midtoken" bson:"midtoken"`
	PathType            int    `json:"pathtype" bson:"pathtype"`
	FinalMinExpectedAmt uint64 `json:"final_minacceptedamount" bson:"final_minacceptedamount"`

	OTARefundFee string `json:"ota_refundfee" bson:"ota_refundfee"`
	OTARefund    string `json:"ota_refund" bson:"ota_refund"`
	OTAFromToken string `json:"ota_fromtoken" bson:"ota_fromtoken"`
	OTAToToken   string `json:"ota_totoken" bson:"ota_totoken"`

	AddOnTxID    string `json:"addon_txid" bson:"addon_txid"`
	PAppName     string `json:"papp_name" bson:"papp_name"`
	PAppNetwork  string `json:"papp_network" bson:"papp_network"`
	PAppContract string `json:"papp_contract" bson:"papp_contract"`

	Status    int    `json:"status" bson:"status"`
	StatusStr string `json:"statusstr" bson:"statusstr"`
	UserAgent string `json:"useragent" bson:"useragent"`
	Error     string `json:"error" bson:"error"`
}

func processInterswapPdexPappPathTask(ctx context.Context, m *pubsub.Message) {
	m.Ack()
	return
}
