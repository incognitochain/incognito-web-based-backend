package submitproof

import (
	"fmt"

	"github.com/adjust/rmq/v4"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/redb"
	"github.com/rueian/rueidis"
)

var db rueidis.Client
var rdmq rmq.Connection

func connectDB(endpoint []string) error {
	var err error
	fmt.Println("endpoint: ", endpoint)
	db, err = redb.NewClient(endpoint)
	return err
}

func connectMQ(serviceID uuid.UUID, endpoint []string) error {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: endpoint,
	})
	var err error
	rdmq, err = rmq.OpenConnectionWithRedisClient("worker-"+serviceID.String(), client, nil)
	return err
}
