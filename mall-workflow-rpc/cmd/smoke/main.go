// Smoke test: drives a signin_v1 instance through REGISTERED → CHECKED_IN
// → AWARDED → DONE, exercising rule evaluation along the way.
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
	defId        = flag.Int64("def", 1, "workflow_definition id")
	ruleId       = flag.Int64("rule_id", 2, "rule id for signin_daily_unique")
	userId       = flag.Int64("user", 1001, "user id")
	activityId   = flag.Int64("activity", 9001, "activity id")
)

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
	wf := workflow.NewWorkflowClient(wfConn)
	r := rule.NewRuleClient(rConn)

	// 1. Evaluate the eligibility rule
	ev, err := r.Evaluate(ctx, &rule.EvaluateReq{
		RuleId: *ruleId,
		Context: &rule.RuleContext{
			UserId:                  *userId,
			UserTier:                "gold",
			ActivityId:              *activityId,
			ActivityType:            "signin",
			ParticipationCountToday: 0,
		},
	})
	if err != nil {
		log.Fatalf("rule.Evaluate: %v", err)
	}
	fmt.Printf("[1/5] rule.Evaluate result=%v latency_us=%d detail=%q\n", ev.Result, ev.LatencyUs, ev.Detail)
	if !ev.Result {
		log.Fatalf("rule failed: user already signed in today; abort")
	}

	// 2. StartInstance
	si, err := wf.StartInstance(ctx, &workflow.StartInstanceReq{
		DefinitionId: *defId,
		ActivityId:   *activityId,
		UserId:       *userId,
		PayloadJson:  fmt.Sprintf(`{"user_id":%d,"activity_id":%d}`, *userId, *activityId),
	})
	if err != nil {
		log.Fatalf("StartInstance: %v", err)
	}
	fmt.Printf("[2/5] StartInstance instance_id=%d state=%s\n", si.InstanceId, si.State)

	// 3. Fire check_in
	f1, err := wf.Fire(ctx, &workflow.FireReq{InstanceId: si.InstanceId, Trigger: "check_in"})
	if err != nil {
		log.Fatalf("Fire(check_in): %v", err)
	}
	fmt.Printf("[3/5] Fire(check_in) state=%s advanced=%v\n", f1.State, f1.Advanced)

	// 4. Fire award
	f2, err := wf.Fire(ctx, &workflow.FireReq{InstanceId: si.InstanceId, Trigger: "award"})
	if err != nil {
		log.Fatalf("Fire(award): %v", err)
	}
	fmt.Printf("[4/5] Fire(award) state=%s advanced=%v\n", f2.State, f2.Advanced)

	// 5. Fire done
	f3, err := wf.Fire(ctx, &workflow.FireReq{InstanceId: si.InstanceId, Trigger: "done"})
	if err != nil {
		log.Fatalf("Fire(done): %v", err)
	}
	fmt.Printf("[5/5] Fire(done) state=%s advanced=%v\n", f3.State, f3.Advanced)

	// 6. GetInstance and dump steps
	gi, err := wf.GetInstance(ctx, &workflow.IdReq{Id: si.InstanceId})
	if err != nil {
		log.Fatalf("GetInstance: %v", err)
	}
	fmt.Printf("\n[final] state=%s version=%d\n", gi.Instance.State, gi.Instance.Version)
	fmt.Printf("[steps]\n")
	for _, s := range gi.Steps {
		fmt.Printf("  %s -> %s  (trigger=%s, %dms)\n", s.FromState, s.ToState, s.Trigger, s.LatencyMs)
	}

	if gi.Instance.State != "DONE" {
		log.Fatalf("FAIL: expected final state DONE, got %s", gi.Instance.State)
	}
	fmt.Println("\nSMOKE OK")
}
