package cache

import (
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.ClusterClient

func InitCluster(){
	addrs := os.Getenv("REDIS_CLUSTER_ADDRS")
	if addrs == "" {
		addrs = "redis1:6379,redis2:6379,redis3:6379"
	}
	nodes:=splitAddrs(addrs)
	Client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: nodes,
		PoolSize: 200,
		MinIdleConns:  50,
		DialTimeout:   2 * time.Second,
		ReadTimeout:   1 * time.Second,
		WriteTimeout:  1 * time.Second,
		RouteRandomly: false,
		RouteByLatency: true,
	})
}
func splitAddrs(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}