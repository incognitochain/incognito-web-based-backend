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

func connectDB(endpoint []string, user string, pass string) error {
	var err error
	fmt.Println("endpoint: ", endpoint)
	db, err = redb.NewClient(endpoint, user, pass)
	return err
}

func connectMQ(serviceID uuid.UUID, endpoint []string, user string, pass string) error {
	var redisClient redis.Cmdable
	if len(endpoint) > 1 {
		redisClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    endpoint,
			Username: user,
			Password: pass,
		})
	} else {
		redisClient = redis.NewClient(&redis.Options{Addr: endpoint[0],
			Username: user,
			Password: pass})
	}

	var err error
	rdmq, err = rmq.OpenConnectionWithRedisClient("worker-"+serviceID.String(), redisClient, nil)
	return err
}
