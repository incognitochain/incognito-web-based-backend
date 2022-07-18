package submitproof

import (
	"context"
	"fmt"

	"github.com/rueian/rueidis"
)

func getShieldTxStatus(externalTxHash string, networkID int, tokenID string) (int, error) {
	ctx := context.Background()
	key := ShieldStatusPrefix + buildShieldStatusKey(externalTxHash, networkID, tokenID)
	value, err := db.Do(ctx, db.B().Get().Key(key).Build()).AsInt64()
	if err != nil {
		if err.Error() == "redis nil message" {
			return 0, nil
		}
		return 0, err
	}
	return int(value), nil
}

func updateShieldTxStatus(externalTxHash string, networkID int, tokenID string, status int) error {
	ctx := context.Background()

	key := ShieldStatusPrefix + buildShieldStatusKey(externalTxHash, networkID, tokenID)
	value := fmt.Sprint(status)
	action := func(rd rueidis.DedicatedClient) error {
		result := rd.DoMulti(ctx, db.B().Set().Key(key).Value(value).Build(), rd.B().Persist().Key(key).Build())
		for _, rs := range result {
			if rs.RedisError() != nil {
				return rs.RedisError()
			}
		}
		return nil
	}

	return db.Dedicated(action)
}

func setShieldTxStatusError(externalTxHash string, networkID int, tokenID string, errStr string) error {
	ctx := context.Background()

	key := ShieldErrorPrefix + buildShieldStatusKey(externalTxHash, networkID, tokenID)
	value := fmt.Sprint(errStr)
	return db.Do(ctx, db.B().Set().Key(key).Value(value).Nx().Build()).Error()
}

func buildShieldStatusKey(externalTxHash string, networkID int, tokenID string) string {
	return externalTxHash + "_" + fmt.Sprint(networkID) + "_" + tokenID
}
