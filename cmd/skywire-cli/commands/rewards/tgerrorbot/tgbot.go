// Package clirewardstgerrorbot cmd/skywire-cli/commands/rewards/tgerrorbot/tgbot.go
package clirewardstgerrorbot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bitfield/script"
	"github.com/spf13/cobra"
	tele "gopkg.in/telebot.v3"
)

var filePath string

func init() {
	RootCmd.Flags().StringVarP(&filePath, "watch", "w", "/tmp/skywire.log", "log file to watch")
}

// RootCmd contains the telegram error bot command
var RootCmd = &cobra.Command{
	Use:   "errbot",
	Short: "error notification telegram bot",
	Long:  "error notification telegram bot",
	Run: func(_ *cobra.Command, _ []string) {
		chatIDStr := os.Getenv("TG_CHAT_ID")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			log.Fatalf("failed to parse chat ID from env TG_CHAT_ID: %v", err)
		}
		pref := tele.Settings{
			Token:  os.Getenv("TG_BOT_TOKEN"),
			Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		}

		b, err := tele.NewBot(pref)
		if err != nil {
			log.Fatal(err)
			return
		}
		msg := "Error notification bot started."
		fmt.Println(msg)
		_, err = b.Send(&tele.Chat{ID: chatID}, msg)
		if err != nil {
			log.Printf("Error sending message to Telegram chat: %s", err)
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lastModTime, err := os.Stat(filePath)
		if err != nil {
			log.Fatal(err)
			return
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					fmt.Println("Stopping file watcher.")
					return
				default:
					time.Sleep(2 * time.Second)
					fileInfo, err := os.Stat(filePath)
					if err != nil {
						log.Printf("Error checking file info: %s", err)
						continue
					}

					if fileInfo.ModTime().After(lastModTime.ModTime()) && time.Since(fileInfo.ModTime()) > 1*time.Second {
						lastLine, err := script.File(filePath).Last(1).String()
						if err != nil {
							log.Printf("Error getting last line of file: %v", err)
							continue
						}
						if lastLine != "" {
							msg := fmt.Sprintf("`%s`", lastLine)
							fmt.Println(msg)

							_, err = b.Send(&tele.Chat{ID: chatID}, msg)
							if err != nil {
								log.Printf("Error sending message to Telegram chat: %s", err)
								continue
							}
						}
						lastModTime = fileInfo
					}
				}
			}
		}()

		<-stop
		fmt.Println("Received termination signal, shutting down bot.")
		cancel()
		b.Stop()
	},
}
