// Seed binary: registers all 4 activity workflow definitions and the
// signin_daily_unique rule used by the signin flow. Idempotent — re-run
// after schema/version bumps and it will upsert.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mall-rule-rpc/rule"
	"mall-workflow-rpc/workflow"
)

var (
	workflowAddr = flag.String("workflow", "127.0.0.1:9012", "workflow rpc address")
	ruleAddr     = flag.String("rule", "127.0.0.1:9011", "rule rpc address")
)

type wfDef struct {
	code        string
	description string
	statesJson  string
	transJson   string
}

var workflowDefinitions = []wfDef{
	{
		code:        "signin_v1",
		description: "daily sign-in flow",
		statesJson: `{
  "initial": "REGISTERED",
  "states": ["REGISTERED", "CHECKED_IN", "AWARDED", "DONE", "REJECTED"],
  "terminal": ["DONE", "REJECTED"]
}`,
		transJson: `[
  {"from": "REGISTERED", "trigger": "check_in", "to": "CHECKED_IN"},
  {"from": "REGISTERED", "trigger": "reject",   "to": "REJECTED"},
  {"from": "CHECKED_IN", "trigger": "award",    "to": "AWARDED"},
  {"from": "AWARDED",    "trigger": "done",     "to": "DONE"}
]`,
	},
	{
		code:        "lottery_v1",
		description: "weighted-random lottery flow",
		statesJson: `{
  "initial": "REGISTERED",
  "states": ["REGISTERED", "PARTICIPATING", "WON", "LOST", "REWARDED", "NOTIFIED", "COMPENSATED"],
  "terminal": ["NOTIFIED", "LOST", "COMPENSATED"]
}`,
		transJson: `[
  {"from": "REGISTERED",    "trigger": "win",       "to": "WON"},
  {"from": "REGISTERED",    "trigger": "lose",      "to": "LOST"},
  {"from": "WON",           "trigger": "reward",    "to": "REWARDED"},
  {"from": "REWARDED",      "trigger": "notify",    "to": "NOTIFIED"},
  {"from": "WON",           "trigger": "compensate","to": "COMPENSATED"}
]`,
	},
	{
		code:        "seckill_v1",
		description: "seckill flow with Redis-Lua hot path and DTM SAGA",
		statesJson: `{
  "initial": "REGISTERED",
  "states": ["REGISTERED", "RESERVED", "CONFIRMED", "REJECTED", "COMPENSATED"],
  "terminal": ["CONFIRMED", "REJECTED", "COMPENSATED"]
}`,
		transJson: `[
  {"from": "REGISTERED", "trigger": "buy",        "to": "RESERVED"},
  {"from": "REGISTERED", "trigger": "reject",     "to": "REJECTED"},
  {"from": "RESERVED",   "trigger": "confirm",    "to": "CONFIRMED"},
  {"from": "RESERVED",   "trigger": "compensate", "to": "COMPENSATED"}
]`,
	},
	{
		code:        "coupon_v1",
		description: "coupon claim flow",
		statesJson: `{
  "initial": "REGISTERED",
  "states": ["REGISTERED", "ISSUED", "NOTIFIED", "REJECTED"],
  "terminal": ["NOTIFIED", "REJECTED"]
}`,
		transJson: `[
  {"from": "REGISTERED", "trigger": "claim",  "to": "ISSUED"},
  {"from": "REGISTERED", "trigger": "reject", "to": "REJECTED"},
  {"from": "ISSUED",     "trigger": "notify", "to": "NOTIFIED"}
]`,
	},
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wfConn, err := grpc.NewClient(*workflowAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("workflow dial: %v", err)
	}
	defer wfConn.Close()
	rConn, err := grpc.NewClient(*ruleAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("rule dial: %v", err)
	}
	defer rConn.Close()

	wfCli := workflow.NewWorkflowClient(wfConn)
	rCli := rule.NewRuleClient(rConn)

	// rule
	rRes, err := rCli.CreateRule(ctx, &rule.CreateRuleReq{
		Code:        "signin_daily_unique",
		Description: "user has not yet signed in today",
		Expression:  "participation_count_today == 0",
	})
	if err != nil {
		fmt.Printf("[rule] CreateRule: %v (probably exists)\n", err)
	} else {
		fmt.Printf("[rule] created id=%d code=signin_daily_unique\n", rRes.Id)
	}

	// 4 workflow definitions
	for _, d := range workflowDefinitions {
		res, err := wfCli.RegisterDefinition(ctx, &workflow.RegisterDefinitionReq{
			Code:            d.code,
			Description:     d.description,
			StatesJson:      d.statesJson,
			TransitionsJson: d.transJson,
		})
		if err != nil {
			log.Fatalf("[workflow] RegisterDefinition %s: %v", d.code, err)
		}
		fmt.Printf("[workflow] registered %-12s id=%d\n", d.code, res.Id)
	}
}
