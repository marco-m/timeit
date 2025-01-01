// This code is released under the MIT License
// Copyright (c) 2023 Marco Molteni and the timeit contributors.

package pytestsim

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/alecthomas/kong"
)

func Main() int {
	if err := run(); err != nil {
		fmt.Println("pytestsim:", err)
		return 1
	}
	return 0
}

type msg struct {
	workerId int
	name     string
	status   string
}

type config struct {
	NumWorkers int           `help:"Number of workers." default:"8"`
	Seed       int64         `help:"Seed for the PRNG (default: current time)." placeholder:"N"`
	MinDur     time.Duration `help:"min job duration in Go time units (eg: 1h2m3s4ms)." default:"500ms"`
	MaxDur     time.Duration `help:"max job duration in Go time units (eg: 1h2m3s4ms)." default:"5000ms"`
}

func run() error {
	var cfg config
	kong.Parse(&cfg,
		kong.Name("pytestsim"),
		kong.Description("Simple simulator of the output of pytest."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: false,
			Summary: true,
		}))

	if cfg.MinDur >= cfg.MaxDur {
		return fmt.Errorf("--mindur must be less than --maxdur")
	}
	if cfg.Seed == 0 {
		// default
		cfg.Seed = time.Now().UnixMilli()
	}
	fmt.Printf("cfg: %+v\n", cfg)
	fmt.Printf("some more output that is not a test name\n")

	names := names()
	numJobs := len(names)
	jobsCh := make(chan string, 2*cfg.NumWorkers)
	outputCh := make(chan msg, 2*cfg.NumWorkers)
	printerDone := make(chan bool)
	var wg sync.WaitGroup

	// Printer goroutine.
	go func() {
		terminated := 0
		for msg := range outputCh {
			if msg.status == "" {
				fmt.Println(msg.name)
			} else {
				terminated++
				pct := terminated * 100 / numJobs
				fmt.Printf("[gw%d] [%d%%] %s %s\n", msg.workerId, pct, msg.status, msg.name)
			}
		}
		printerDone <- true
	}()

	// Init the pool with the worker goroutines.
	for workerId := 1; workerId <= cfg.NumWorkers; workerId++ {
		wg.Add(1)
		workerId := workerId // avoid workerId loop capture
		go func() {
			defer wg.Done()
			worker(cfg, workerId, jobsCh, outputCh)
		}()
	}

	// Distribute the jobs
	for jobId := 0; jobId < numJobs; jobId++ {
		jobsCh <- names[jobId]
	}
	close(jobsCh) // This will make the workers to terminate.

	wg.Wait()
	// Now we can safely close outputCh, which will make the printer goroutine
	// terminate
	close(outputCh)

	<-printerDone
	fmt.Println("pytestsim finished")

	return nil
}

// worker simulator.
func worker(cfg config, workerId int, jobsCh <-chan string, outputCh chan<- msg) {
	// A rand.Source (and so a rand.Rand) is not safe for concurrent use, this
	// is why we have one per goroutine.
	seed := uint64(cfg.Seed) + uint64(workerId)
	rnd := rand.New(rand.NewPCG(seed, seed+100))
	for name := range jobsCh {
		msg := msg{workerId: workerId, name: name}
		outputCh <- msg

		jitter := rnd.Int64N(int64(cfg.MaxDur - cfg.MinDur))
		sleep := cfg.MinDur + time.Duration(jitter)
		time.Sleep(sleep)

		msg.status = "PASSED"
		outputCh <- msg
	}
}
