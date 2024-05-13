package main

import (
	"aggressive-pokes/internal/ltlogger"
	"aggressive-pokes/internal/utils"
	"cloud.google.com/go/pubsub"
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	_ "net/http/pprof"
)

func main() {
	logger := ltlogger.New(true, "Stub Server", slog.LevelDebug)
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			logger.Fatal("Failed to start pprof")
		}
	}()

	httpChan := stubHttpServer(logger, "/pixel")
	pubsubChan := stubPubsubListener(logger, "local-project", "stub-topic", "stub-sub")

	logger.Info("Stub server started",
		"http_url", "http://127.0.0.1:8081/pixel",
		"pubsub_project")

	<-httpChan
	<-pubsubChan
}

func stubHttpServer(logger ltlogger.Logger, url string) chan struct{} {
	finishChan := make(chan struct{})
	go func() {
		mux := http.NewServeMux()
		mux.Handle(url, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			random := rand.Int31n(10)
			if random < 5 {
				time.Sleep(time.Duration(rand.Int31n(500)) * time.Millisecond)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if random < 8 {
				time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			time.Sleep(time.Duration(rand.Int31n(300)) * time.Millisecond)

			w.WriteHeader(http.StatusInternalServerError)
			return

		}))
		logger.Info("Running stub server on port 8081")
		logger.Error("Server stopped", "err", http.ListenAndServe("127.0.0.1:8081", mux))
		finishChan <- struct{}{}
	}()
	return finishChan

}

func stubPubsubListener(logger ltlogger.Logger, projectId, topicName, subscriptionName string) chan struct{} {
	finishChan := make(chan struct{})
	go func() {
		utils.SetPubsubEmulatorAddr()
		pbClient, err := pubsub.NewClient(context.Background(), projectId)
		if err != nil {
			logger.Fatal("Cannot create pubsub client", "err", err)
		}
		topic := utils.GetPubsubTopic(logger, pbClient, topicName)
		subscription := utils.GetPubsubSubscription(logger, pbClient, topic, subscriptionName)

		logger.Info("Listening for pubsub messages", "topic", topicName, "subscription", subscriptionName)
		err = subscription.Receive(context.Background(), func(ctx context.Context, msg *pubsub.Message) {
			if err == nil {
				logger.Info("Got message", "msg", msg.ID)
				msg.Ack()
			} else {
				msg.Nack()
			}
		})
		if err != nil {
			logger.Fatal("Failed to receive pubsub message", "topic", topicName, "subscription", subscriptionName, "err", err)
		}
		finishChan <- struct{}{}
	}()
	return finishChan
}
