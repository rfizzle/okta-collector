package main

import (
	"github.com/rfizzle/collector-helpers/outputs"
	"github.com/rfizzle/collector-helpers/state"
	"github.com/rfizzle/okta-collector/client"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

func main() {
	// Setup variables
	var maxMessages = int64(5000)

	// Setup Parameters via CLI or ENV
	if err := setupCliFlags(); err != nil {
		log.Fatalf("initialization failed: %v", err.Error())
	}

	// Setup log writer
	tmpWriter, err := outputs.NewTmpWriter()
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	// Setup the channels for handling async messages
	chnMessages := make(chan string, maxMessages)

	// Setup the Go Routine
	pollTime := viper.GetInt("schedule")

	// Start Poll
	go pollEvery(pollTime, chnMessages, tmpWriter)

	// Handle messages in the channel (this will keep the process running indefinitely)
	for message := range chnMessages {
		handleMessage(message, tmpWriter)
	}
}

func pollEvery(seconds int, resultsChannel chan<- string, tmpWriter *outputs.TmpWriter) {
	var currentState *state.State
	var err error

	// Setup State
	if state.Exists(viper.GetString("state-path")) {
		currentState, err = state.Restore(viper.GetString("state-path"))
		if err != nil {
			log.Fatalf("Error getting state: %v\n", err.Error())
		}
	} else {
		currentState = state.New()
	}

	for {
		log.Println("Getting data...")

		// Get events
		eventCount, lastPollTime := getEvents(currentState.LastPollTimestamp, resultsChannel)

		// Copy tmp file to correct outputs
		if eventCount > 0 {
			// Wait until the results channel has no more messages 0
			for len(resultsChannel) != 0 {
				<-time.After(time.Duration(1) * time.Second)
			}

			// Close and rotate file
			_ = tmpWriter.Rotate()

			if err := outputs.WriteToOutputs(tmpWriter.LastFilePath, lastPollTime.Format(time.RFC3339)); err != nil {
				log.Fatalf("Unable to write to output: %v", err)
			}

			// Remove temp file now
			err := os.Remove(tmpWriter.LastFilePath)
			if err != nil {
				log.Fatalf("Unable to remove tmp file: %v", err)
			}
		}

		// Let know that event has been processes
		log.Printf("%v events processed...\n", eventCount)

		// Update state
		currentState.LastPollTimestamp = lastPollTime.Format(time.RFC3339)
		state.Save(currentState, viper.GetString("state-path"))

		// Wait for x seconds until next poll
		<-time.After(time.Duration(seconds) * time.Second)
	}
}

func getEvents(timestamp string, resultChannel chan<- string) (int, time.Time) {
	// Get current time
	now := time.Now()

	// Build an Okta client
	oktaClient := client.NewClient(viper.GetString("okta-domain"), viper.GetString("okta-api-key"))

	// Get logs
	count, err := oktaClient.GetLogs(timestamp, now.Format(time.RFC3339), resultChannel)

	if err != nil {
		log.Fatalf("Unable to retrieve okta logs: %v", err)
	}

	return count, now
}

// Handle message in a channel
func handleMessage(message string, tmpWriter *outputs.TmpWriter) {
	if err := tmpWriter.WriteLog(message); err != nil {
		log.Fatalf("Unable to write to temp file: %v", err)
	}
}
