package config

type Config struct {
	Name        string
	Concurrency int `json:",default=50"`
	Redis       struct {
		Addr     string
		Password string `json:",optional"`
		DB       int    `json:",default=0"`
	}
	Queues map[string]int

	// H-2 settlement
	OrderDSN              string `json:",optional"`
	PaymentDSN            string `json:",optional"`
	SettlementDelaySec    int    `json:",default=259200"` // 3 days
	SettlementIntervalSec int    `json:",default=300"`    // 5 minutes
}
