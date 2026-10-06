package blockchain

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"toy-blockchain/block"
)

func MineBlock(b *block.Block, difficulty int) (int, time.Duration) {

	start := time.Now()

	attempts := 0

	target := strings.Repeat("0", difficulty)

	for {

		hash := block.CalculateHash(*b)
		attempts++

		if strings.HasPrefix(hash, target) {

			b.Hash = hash

			return attempts, time.Since(start)
		}

		b.Nonce++
	}
}
func MineBlockConcurrent(b *block.Block, difficulty int, workers int) (int, time.Duration) {

	start := time.Now()
	fmt.Println("Starting concurrent mining with", workers, "workers")

	target := strings.Repeat("0", difficulty)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex

	var attempts int64
	found := false
	initialBlock := *b
	for i := 0; i < workers; i++ {

		wg.Add(1)

		go func(workerID int) {

			defer wg.Done()

			tempBlock := initialBlock

			nonce := workerID

			for {

				select {
				case <-ctx.Done():
					return
				default:
				}

				tempBlock.Nonce = nonce

				hash := block.CalculateHash(tempBlock)

				atomic.AddInt64(&attempts, 1)

				if strings.HasPrefix(hash, target) {

					mu.Lock()

					if !found {
						found = true

						b.Nonce = nonce
						b.Hash = hash

						fmt.Println("Worker", workerID, "found nonce:", nonce)

						cancel()
					}

					mu.Unlock()

					return
				}

				// Each worker searches different nonce values
				nonce += workers
			}

		}(i)
	}

	wg.Wait()

	return int(attempts), time.Since(start)
}
