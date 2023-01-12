package common

import (
	"time"

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

type OpenseaDefaultCollectionData struct {
	mgm.DefaultModel `bson:",inline"`
	Address          string `bson:"address"`
	Slug             string `bson:"slug"`
	Verify           bool   `bson:"verify"`
}

// TODO: opensea
type OpenseaOfferData struct {
	mgm.DefaultModel `bson:",inline"`
	Receiver         string    `bson:"receiver"`
	NFTID            string    `bson:"nft_id"`
	NFTCollection    string    `bson:"collection_id"`
	OfferHash        string    `bson:"offer_hash"`
	TimeoutAt        time.Time `bson:"timeout_at"`
	Status           string    `bson:"status"`
	OfferTxInc       string    `bson:"offer_tx_inc"`
	CancelTxInc      string    `bson:"cancel_tx_inc"`
	// CancelAdapterTx  string    `bson:"cancel_apdapter_tx"`
	CancelOpenseaTx string `bson:"cancel_opensea_tx"` //not used yet
	ReshieldTx      string `bson:"reshield_tx_inc"`
	OfferSubmitted  bool   `bson:"offer_submitted"`
	ClaimNFTTx      string `bson:"claim_nft_tx"`
}

type OpenseaOfferDetail struct {
	OfferToken  string `bson:"offer_token" json:"offer_token"`
	OfferAmount string `bson:"offer_amount" json:"offer_amount"`
	Offer       string `bson:"offer" json:"offer"`
}
