package pblur

import (
	"encoding/base64"
	"fmt"
)

func EncodeData(data, key string) (string, error) {
	x, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	y := ""

	for i := 0; i < len(x); i++ {
		y += string(x[i] ^ key[i%len(key)])
	}
	fmt.Println("y:", y)
	return y, nil
}
