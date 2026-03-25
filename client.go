package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const SERVER_ADDR = "127.0.0.1:8080"

var (
	successCount int64
	failCount    int64
	timeoutCount int64
)

func runClient(id int, wg *sync.WaitGroup, message string, dialTimeout time.Duration) {
	defer wg.Done()

	conn, err := net.DialTimeout("tcp", SERVER_ADDR, dialTimeout)
	if err != nil {
		atomic.AddInt64(&failCount, 1)
		log.Printf("[CLIENT %d] Connect failed: %v", id, err)
		return
	}
	defer conn.Close()

	// Set read deadline
	conn.SetDeadline(time.Now().Add(dialTimeout))

	_, err = conn.Write([]byte(message))
	if err != nil {
		atomic.AddInt64(&failCount, 1)
		log.Printf("[CLIENT %d] Write failed: %v", id, err)
		return
	}

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			atomic.AddInt64(&timeoutCount, 1)
			log.Printf("[CLIENT %d] Timeout reading response", id)
		} else {
			atomic.AddInt64(&failCount, 1)
			log.Printf("[CLIENT %d] Read failed: %v", id, err)
		}
		return
	}

	atomic.AddInt64(&successCount, 1)
	log.Printf("[CLIENT %d] Response: %q", id, string(buf[:n]))
}

func main() {
	numClients := flag.Int("clients", 100, "Number of concurrent clients")
	message := flag.String("msg", "LOAD_TEST_PING", "Message to send")
	dialTimeout := flag.Duration("timeout", 5*time.Second, "Dial/read timeout per client")
	waves := flag.Int("waves", 1, "Number of load waves to send")
	flag.Parse()

	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("[LOAD TEST] Starting: clients=%d waves=%d timeout=%s msg=%q",
		*numClients, *waves, *dialTimeout, *message)

	totalStart := time.Now()

	for w := 1; w <= *waves; w++ {
		log.Printf("[WAVE %d/%d] Launching %d concurrent clients...", w, *waves, *numClients)
		waveStart := time.Now()

		var wg sync.WaitGroup
		for i := 1; i <= *numClients; i++ {
			wg.Add(1)
			go runClient(i, &wg, *message, *dialTimeout)
		}
		wg.Wait()

		log.Printf("[WAVE %d/%d] Done in %s", w, *waves, time.Since(waveStart).Round(time.Millisecond))

		if w < *waves {
			time.Sleep(500 * time.Millisecond) // brief pause between waves
		}
	}

	elapsed := time.Since(totalStart)
	total := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&failCount) + atomic.LoadInt64(&timeoutCount)

	fmt.Println()
	fmt.Println("========== LOAD TEST RESULTS ==========")
	fmt.Printf("Total clients  : %d\n", total)
	fmt.Printf("Successes      : %d\n", atomic.LoadInt64(&successCount))
	fmt.Printf("Failures       : %d\n", atomic.LoadInt64(&failCount))
	fmt.Printf("Timeouts       : %d\n", atomic.LoadInt64(&timeoutCount))
	fmt.Printf("Total time     : %s\n", elapsed.Round(time.Millisecond))
	if elapsed.Seconds() > 0 {
		fmt.Printf("Throughput     : %.1f req/s\n", float64(total)/elapsed.Seconds())
	}
	fmt.Println("=======================================")
}
