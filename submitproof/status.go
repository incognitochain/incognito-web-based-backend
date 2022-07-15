package submitproof

import (
	"context"
	"fmt"
)

func getShieldTxStatus(externalTxHash string, networkID int, tokenID string) (int, error) {
	ctx := context.Background()
	key := buildShieldStatusKey(externalTxHash, networkID, tokenID)
	value, err := db.Do(ctx, db.B().Get().Key(key).Build()).AsInt64()
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

func updateShieldTxStatus(externalTxHash string, networkID int, tokenID string, status int) error {
	ctx := context.Background()

	key := buildShieldStatusKey(externalTxHash, networkID, tokenID)
	value := fmt.Sprint(status)
	return db.Do(ctx, db.B().Set().Key(key).Value(value).Nx().Build()).Error()
}

func setShieldTxStatusError(externalTxHash string, networkID int, tokenID string, errStr string) error {
	ctx := context.Background()

	key := buildShieldStatusKey(externalTxHash, networkID, tokenID)
	value := fmt.Sprint(errStr)
	return db.Do(ctx, db.B().Set().Key(key).Value(value).Nx().Build()).Error()
}

func buildShieldStatusKey(externalTxHash string, networkID int, tokenID string) string {
	return externalTxHash + "_" + fmt.Sprint(networkID) + "_" + tokenID
}
