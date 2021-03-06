package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var logLines = []string{
	"Log line",
	"Hello, world",
	"This is a bit longer",
	"And this one is a much longer log line that also includes some fixed numbers like 1,000,000",
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	lpsStr := os.Getenv("LPS")
	if lpsStr == "" {
		logger.Error("LPS env variable not set")
		return
	}

	lps, err := strconv.Atoi(lpsStr)
	if err != nil {
		logger.Error("LPS env variable is not a number.", zap.Error(err))
		return
	}

	if lps <= 0 {
		logger.Error("LPS env variable must be positive number.", zap.Error(err))
		return
	}

	// plug SIGTERM signal into a channel.
	signalsChannel := make(chan os.Signal, 1)
	signal.Notify(signalsChannel, os.Interrupt, syscall.SIGTERM)

	uuid := uuid.New()
	uuidStr := uuid.String()

	startTime := time.Now()
	ticker := time.Tick(10 * time.Millisecond)
	var generatedLineCount int64 = 0

	logger.Info(fmt.Sprintf("Generating logs at %v logs per second. Ctrl-C to stop.", lps))

loop:
	for {
		select {
		case <-ticker:
			dur := time.Now().Sub(startTime)
			expectedLines := int64(dur.Seconds() * float64(lps))

			// No more than this each cycle. We have this limit to ensure we don't generate
			// a ton of lines each cycle when we begin to fallback if the LPS is too high.
			// Otherwise the application will become unresponsive to Ctrl-C signal.
			max := 10000
			for generatedLineCount < expectedLines && max > 0 {
				logLine := logLines[rand.Intn(len(logLines))]
				generatedLineCount++
				logger.Info(logLine, zap.Int64("counter", generatedLineCount), zap.String("service.instance.id", uuidStr))
				max--
			}

		case <-signalsChannel:
			logger.Info("Stopped.")
			break loop
		}
	}

	sinceStart := time.Now().Sub(startTime)
	logger.Info(fmt.Sprintf("Generated for %.2f sec, Printed total %v lines, %.1f per second",
		sinceStart.Seconds(), generatedLineCount, float64(generatedLineCount)/sinceStart.Seconds()))
}
