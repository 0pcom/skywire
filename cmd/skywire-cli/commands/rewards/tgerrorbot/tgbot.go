// Package clirewardstgerrorbot cmd/skywire-cli/commands/rewards/tgerrorbot/tgbot.go
package clirewardstgerrorbot

import (
	"bufio"
	"bytes"
	"fmt"
	"log"

	"time"

	"github.com/bitfield/script"
	"github.com/spf13/cobra"
)

var filePath string

// RootCmd contains the telegram error bot command
var RootCmd = &cobra.Command{
	Use:   "errbot",
	Short: "error notification test bot",
	Long:  "error notification test bot",
	Run: func(_ *cobra.Command, _ []string) {

		buf := new(bytes.Buffer)

		go func() {
			for {
				// Execute the command and capture the output
				_, _ = script.Exec(`sudo bash -c 'skywire visor -p --loglvl error'`).Tee(buf).Stdout() //nolint

				// Read output line by line
				scanner := bufio.NewScanner(buf)
				for scanner.Scan() {
					line := scanner.Text()
					msg := fmt.Sprintf("%s", line) // Prepare message

					// Print each line as a message to stdout
					fmt.Println("Message to send:", msg)
				}

				if err := scanner.Err(); err != nil {
					log.Printf("Error reading command output: %s", err)
				}

				// Clear the buffer after processing
				buf.Reset()

				// Optional: delay before restarting
				log.Println("Command exited. Restarting in 5 seconds...")
				time.Sleep(5 * time.Second)
			}
		}()

		// Keep the main routine alive
		select {}
	},
}
