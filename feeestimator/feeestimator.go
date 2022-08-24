package feeestimator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
)

const (
	checkFeeInterval = 5 * time.Second
)

func StartService(cfg common.Config) error {
	return checkFee()
}

func checkFee() error {
	ticker := time.NewTicker(checkFeeInterval)
	for range ticker.C {
		networkList, err := retrieveNetwork()
		if err != nil {
			log.Println(err)
			continue
		}
		feeList := make(map[string]uint64)
		for network, endpoints := range networkList {
			fee, err := getFee(network, endpoints)
			if err != nil {
				log.Println(err)
				continue
			}
			feeList[network] = fee
		}

		err = saveFeeData(feeList)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func retrieveNetwork() (map[string][]string, error) {
	result := make(map[string][]string)
	networks, err := database.DBGetBridgeNetworkInfos()
	for _, network := range networks {
		result[network.Network] = network.Endpoints
	}
	return result, err
}

func getFee(network string, endpoints []string) (uint64, error) {
	epCount := len(endpoints)
	i := 0
retry:
	if i == epCount {
		return 0, fmt.Errorf("can't find endpoints for network %v", network)
	}
	evmClient, err := ethclient.Dial(endpoints[i])
	if err != nil {
		i++
		goto retry
	}
	var errFee error
	var fee *big.Int
	log.Printf("get fee for network %v using endpoint %v \n", network, endpoints[i])
	switch network {
	case common.NETWORK_ETH:
		fee, errFee = getEthGasPrice(evmClient)
	case common.NETWORK_BSC:
		fee, errFee = getBscGasPrice(evmClient)
	case common.NETWORK_PLG:
		fee, errFee = getPlgGasPrice(evmClient)
	case common.NETWORK_FTM:
		fee, errFee = getFtmGasPrice(evmClient)
	default:
		return 0, errors.New("unsupported network")
	}
	if errFee != nil {
		log.Println("getFee", errFee)
		i++
		goto retry
	}
	if fee == nil {
		return 0, nil
	}

	return fee.Uint64(), nil
}

func saveFeeData(data map[string]uint64) error {
	var feeData common.ExternalNetworksFeeData
	feeData.Creating()
	feeData.Fees = data
	return database.DBSaveFeetTable(feeData)
}

func getEthGasPrice(c *ethclient.Client) (*big.Int, error) {
	gasPrice, err := SuggestGasPrice(c)
	if err != nil {
		return nil, err
	}

	return gasPrice, nil
}

func getBscGasPrice(c *ethclient.Client) (*big.Int, error) {
	// todo: get from network:
	gasPrice, err := SuggestGasPrice(c)
	if err != nil {
		return nil, err
	}

	return gasPrice, nil
}

func getPlgGasPrice(c *ethclient.Client) (*big.Int, error) {
	// todo: get from network:
	gasPrice, err := SuggestGasPrice(c)
	if err != nil {
		return nil, err
	}
	// speed up
	// gasPrice = gasPrice.Mul(gasPrice, big.NewInt(2))

	// fee := new(big.Int).Mul(big.NewInt(int64(1000000)), gasPrice) //todo: update gas limit.

	// fmt.Println("fee est bsc: gasPrice, fee", gasPrice, fee)

	// fee = fee.Mul(fee, big.NewInt(3))

	// fmt.Println("fee est bsc x3: ", fee)

	return gasPrice, nil
}

func getFtmGasPrice(c *ethclient.Client) (*big.Int, error) {
	gasPrice, err := SuggestGasPrice(c)
	if err != nil {
		return nil, err
	}
	return gasPrice, nil
}

func SuggestGasPrice(client *ethclient.Client) (*big.Int, error) {
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	fmt.Println("gas from SuggestGasPrice from chain ID: ", chainID, gasPrice)

	// // hardcode for polygon:
	// if chainID.Uint64() == 137 || chainID.Uint64() == 80001 {
	// 	// increase x1.5:
	// 	gasPrice = new(big.Int).Mul(big.NewInt(int64(15)), gasPrice)
	// 	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(int64(10)))

	// 	fmt.Println("gas x1.5 from SuggestGasPrice: ", gasPrice)
	// }
	return gasPrice, nil
}
