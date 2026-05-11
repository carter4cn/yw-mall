package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"mall-activity-async-worker/internal/config"
	"mall-activity-async-worker/internal/handlers"
	"mall-activity-async-worker/internal/settlement"

	"mall-common/configcenter"

	"github.com/hibiken/asynq"
)

var configFile = flag.String("f", "etc/worker.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, configcenter.ServiceKey("yw-mall", "activity-worker"), *configFile, &c)

	queues := c.Queues
	if len(queues) == 0 {
		queues = map[string]int{"default": 1}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if c.OrderDSN != "" && c.PaymentDSN != "" {
		settler, err := settlement.New(c.OrderDSN, c.PaymentDSN, c.SettlementDelaySec, c.SettlementIntervalSec)
		if err != nil {
			log.Fatalf("settlement init failed: %v", err)
		}
		go settler.Run(ctx)
		fmt.Printf("Settlement loop started: delay=%ds tick=%ds\n", c.SettlementDelaySec, c.SettlementIntervalSec)
	} else {
		fmt.Println("Settlement disabled (OrderDSN/PaymentDSN not set)")
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
