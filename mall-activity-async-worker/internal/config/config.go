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
}
