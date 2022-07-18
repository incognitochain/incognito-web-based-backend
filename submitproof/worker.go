package submitproof

import (
	"log"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/pkg/errors"
)

func worker() error {
	consumerTag := "submitproof_worker"

	// cleanup, err := tracers.SetupTracer(consumerTag)
	// if err != nil {
	// 	log.FATAL.Fatalln("Unable to instantiate a tracer:", err)
	// }
	// defer cleanup()

	server, err := startServer()
	if err != nil {
		return err
	}

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker := server.NewWorker(consumerTag, 0)

	// Here we inject some custom code for error handling,
	// start and end of task hooks, useful for metrics for example.
	errorhandler := func(err error) {
		log.Println("I am an error handler:", err)
	}

	pretaskhandler := func(signature *tasks.Signature) {
		log.Println("I am a start of task handler for:", signature.Name)
	}

	posttaskhandler := func(signature *tasks.Signature) {
		log.Println("I am an end of task handler for:", signature.Name)
	}

	worker.SetPostTaskHandler(posttaskhandler)
	worker.SetErrorHandler(errorhandler)
	worker.SetPreTaskHandler(pretaskhandler)

	return worker.Launch()
}

func StartWorker(keylist []string, network string, cfg wcommon.Config) error {
	config = cfg
	keyList = keylist

	err := connectDB(cfg.DatabaseURLs)
	if err != nil {
		return err
	}

	switch network {
	case "mainnet":
		incClient, err = incclient.NewMainNetClient()
	case "testnet": // testnet2
		incClient, err = incclient.NewTestNetClient()
	case "testnet1":
		incClient, err = incclient.NewTestNet1Client()
	case "devnet":
		return errors.New("unsupported network")
	}
	if err != nil {
		return err
	}

	for _, v := range keyList {
		wl, err := wallet.Base58CheckDeserialize(v)
		if err != nil {
			panic(err)
		}
		err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
		if err != nil {
			return err
		}
	}
	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	return nil
}
