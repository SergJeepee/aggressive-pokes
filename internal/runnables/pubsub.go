package runnables

import (
	"aggressive-pokes/internal/ltlogger"
	"aggressive-pokes/internal/stats"
	"aggressive-pokes/internal/utils"
	"cloud.google.com/go/pubsub"
	"context"
	"time"
)

func PubsubRunnable(logger ltlogger.Logger, projectId, topicName string, body []byte) func(reporter stats.Reporter) {
	pbClient, err := pubsub.NewClient(context.Background(), projectId)
	if err != nil {
		logger.Fatal("Cannot create pubsub client", "err", err)
	}
	topic := utils.GetPubsubTopic(logger, pbClient, topicName)
	return func(reporter stats.Reporter) {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := topic.Publish(ctx, &pubsub.Message{
			Data:       body,
			Attributes: map[string]string{"serverTime": time.Now().Format(time.RFC3339)},
		})

		<-result.Ready()
		_, err := result.Get(context.Background())
		if err != nil {
			reporter.ReportFailure("pubsub_publish_error", err.Error(), time.Since(start))
			return
		}
		reporter.Report("pubsub_publish_success", time.Since(start))
	}

}
