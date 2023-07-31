package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("starting!")

	client := redis.NewClient(&redis.Options{Network: "tcp", Addr: "127.0.0.1:6379"})
	defer client.Close()

	locker := redislock.New(client)

	ctx := context.Background()

	//ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	//defer cancel()

	// Retry every 100ms, for up-to 3x
	backoff := redislock.LimitRetry(redislock.LinearBackoff(1000*time.Millisecond), 10)

	// Obtain lock with retry
	// weird short ttl also makes the context timeout????????????????????
	lock, err := locker.Obtain(ctx, "my-key", 8*time.Second, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}
	defer lock.Release(ctx)

	for i := 0; i < 15; i++ {
		time.Sleep(1000 * time.Millisecond)
		if ttl, err := lock.TTL(ctx); err != nil {
			log.Fatalln(err)
		} else if ttl > 0 {
			fmt.Println("Yay, I still have my lock!")
		} else {
			fmt.Println("else!")
		}
	}

	fmt.Println("I have a lock!")
}
