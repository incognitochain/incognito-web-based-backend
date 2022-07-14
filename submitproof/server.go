package submitproof

import (
	"github.com/RichardKnop/machinery/v1"
	mconfig "github.com/RichardKnop/machinery/v1/config"
)

func startServer() (*machinery.Server, error) {
	cnf := &mconfig.Config{
		DefaultQueue:    "machinery_tasks",
		ResultsExpireIn: 3600,
		Broker:          "redis://localhost:6379",
		ResultBackend:   "redis://localhost:6379",
		Redis: &mconfig.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}

	server, err := machinery.NewServer(cnf)
	if err != nil {
		return nil, err
	}

	// Register tasks
	tasks := map[string]interface{}{
		"submitproof": SubmitProofToIncognitoChain,
	}

	return server, server.RegisterTasks(tasks)
}
