package main

import (
	"flag"
	"fmt"
	"log"

	"mall-activity-async-worker/internal/config"
	"mall-activity-async-worker/internal/handlers"

	"mall-common/configcenter"

	"github.com/hibiken/asynq"
)

var configFile = flag.String("f", "etc/worker.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, "/mall/config/activity-async-worker", *configFile, &c)

	queues := c.Queues
	if len(queues) == 0 {
		queues = map[string]int{"default": 1}
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     c.Redis.Addr,
			Password: c.Redis.Password,
			DB:       c.Redis.DB,
		},
		asynq.Config{
			Concurrency: c.Concurrency,
			Queues:      queues,
		},
	)

	mux := asynq.NewServeMux()
	handlers.Register(mux)

	fmt.Printf("Starting asynq worker (concurrency=%d, queues=%v)...\n", c.Concurrency, queues)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq server stopped: %v", err)
	}
}
