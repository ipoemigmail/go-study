package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
)

var (
	topic   string
	brokers []string

	rootCmd = &cobra.Command{
		Use:   "kafka-reader",
		Short: "kafka-reader",
		Run: func(cmd *cobra.Command, args []string) {
			runStreaming(topic, brokers)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&topic, "topic", "t", "", "topic")
	rootCmd.PersistentFlags().StringSliceVarP(&brokers, "brokers", "b", []string{}, "brokers")
	rootCmd.MarkPersistentFlagRequired("topic")
	rootCmd.MarkPersistentFlagRequired("brokers")
}

func runStreaming(topic string, brokers []string) {
	log.Printf("Streaming Start...\n")
	groupId := uuid.New().String()
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupId,
		StartOffset: kafka.LastOffset,
		Topic:       topic,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     100 * time.Millisecond,
	})
	ctx := context.Background()
	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			break
		}
		//fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		fmt.Printf("%s\n", string(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
