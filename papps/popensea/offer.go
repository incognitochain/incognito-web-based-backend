package popensea

import (
	"errors"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// TODO: opensea
func GenOfferOrder(tokenContract string, receiverAddress string, offerAdapterAddress string, amount string, endTime int64, item NFTDetail) (*OrderComponents, error) {
	var order OrderComponents
	if len(item.SeaportSellOrders) == 0 {
		return nil, errors.New("can't create offer order without SeaportSellOrders")
	}
	sellOrderParam := item.SeaportSellOrders[0].ProtocolData.Parameters
	currentTime := time.Now().Unix()
	order.OrderType = 0
	order.Counter = big.NewInt(0) // ? always 0
	order.StartTime = new(big.Int).SetInt64(currentTime)
	if endTime > currentTime {
		order.EndTime = new(big.Int).SetInt64(endTime)
	} else {
		order.EndTime = new(big.Int).SetInt64(currentTime + 100)
	}
	order.Offerer = common.HexToAddress(offerAdapterAddress)
	order.ConduitKey = common.HexToHash(sellOrderParam.ConduitKey)
	order.Zone = common.HexToAddress(sellOrderParam.Zone)
	order.ZoneHash = common.HexToHash(sellOrderParam.ZoneHash)

	c := 10
	randSalt := make([]byte, c)
	_, err := rand.Read(randSalt)
	if err != nil {
		return nil, errors.New("can't create offer order salt " + err.Error())
	}

	order.Salt = common.BytesToHash(randSalt).Big()

	tokenId, _ := new(big.Int).SetString(sellOrderParam.Offer[0].IdentifierOrCriteria, 10)
	offerAmount, _ := new(big.Int).SetString(amount, 10)
	nftOffer := OfferItem{
		ItemType:             1,
		Token:                common.HexToAddress(tokenContract),
		IdentifierOrCriteria: big.NewInt(0),
		StartAmount:          offerAmount,
		EndAmount:            offerAmount,
	}
	order.Offer = append(order.Offer, nftOffer)

	feeD := new(big.Int).SetInt64(10000)
	openseaFee := new(big.Int).SetInt64(0)
	for _, v := range item.Collection.Fees.OpenseaFees {
		percent := new(big.Int).SetUint64(v)
		openseaFee = offerAmount.Div(feeD, offerAmount).Mul(percent, offerAmount)
	}

	collectionFee := new(big.Int).SetInt64(0)
	for _, v := range item.Collection.Fees.SellerFees {
		percent := new(big.Int).SetUint64(v)
		collectionFee = offerAmount.Div(feeD, offerAmount).Mul(percent, offerAmount)
	}

	considerationNFT := ConsiderationItem{
		ItemType:             2,
		Token:                common.HexToAddress(sellOrderParam.Offer[0].Token),
		IdentifierOrCriteria: tokenId,
		StartAmount:          big.NewInt(1),
		EndAmount:            big.NewInt(1),
	}
	considerationOpenseaFee := ConsiderationItem{
		ItemType:             1,
		Token:                common.HexToAddress(tokenContract),
		IdentifierOrCriteria: big.NewInt(0),
		StartAmount:          openseaFee,
		EndAmount:            openseaFee,
	}

	considerationCollectionFee := ConsiderationItem{
		ItemType:             1,
		Token:                common.HexToAddress(tokenContract),
		IdentifierOrCriteria: big.NewInt(0),
		StartAmount:          collectionFee,
		EndAmount:            collectionFee,
	}

	order.Consideration = append(order.Consideration, considerationNFT)
	order.Consideration = append(order.Consideration, considerationOpenseaFee)
	if collectionFee.Int64() != 0 {
		order.Consideration = append(order.Consideration, considerationCollectionFee)
	}

	return &order, nil

	// offer := opensea.OrderComponents{
	// 	Offerer: v2.OpenseaOfferAddr,
	// 	Zone:    common.HexToAddress("0x0000000000000000000000000000000000000000"),
	// 	Offer: []opensea.OfferItem{
	// 		{
	// 			ItemType:             1,
	// 			Token:                common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
	// 			IdentifierOrCriteria: big.NewInt(0),
	// 			StartAmount:          offerAmount,
	// 			EndAmount:            offerAmount,
	// 		},
	// 	},
	// 	Consideration: []opensea.ConsiderationItem{
	// 		{
	// 			ItemType:             2,
	// 			Token:                common.HexToAddress("0x8b0e17927a58392BBc42faeD1Cb41abE3A43A50C"),
	// 			IdentifierOrCriteria: big.NewInt(0),
	// 			StartAmount:          big.NewInt(1),
	// 			EndAmount:            big.NewInt(1),
	// 			Recipient:            common.HexToAddress("0x2f6F03F1b43Eab22f7952bd617A24AB46E970dF7"),
	// 		},
	// 		{
	// 			ItemType:             1,
	// 			Token:                common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
	// 			IdentifierOrCriteria: big.NewInt(0),
	// 			StartAmount:          big.NewInt(25000000000000),
	// 			EndAmount:            big.NewInt(25000000000000),
	// 			Recipient:            common.HexToAddress("0x0000a26b00c1F0DF003000390027140000fAa719"),
	// 		},
	// 	},
	// 	OrderType:  0,
	// 	StartTime:  big.NewInt(1673077447),
	// 	EndTime:    big.NewInt(1673336642),
	// 	ZoneHash:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
	// 	Salt:       salt,
	// 	ConduitKey: common.HexToHash("0x0000007b02230091a7ed01230072f7006a004d60a8d4e71d599b8104250f0000"),
	// 	Counter:    big.NewInt(0),
	// }
}
