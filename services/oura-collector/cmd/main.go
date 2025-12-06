package main

import (
	"context"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/oura-collector/internal/client"
	"github.com/asian-code/myapp-kubernetes/services/oura-collector/internal/config"
	"github.com/asian-code/myapp-kubernetes/services/shared/logger"
)

func main() {
	log := logger.Init("oura-collector")
	cfg := config.Load()

	log.Info("Starting oura-collector")

	// Fetch data from Oura
	ouraClient := client.New(cfg.OuraAPIKey, log)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	date := time.Now().Format("2006-01-02")

	// Fetch sleep data
	sleepData, err := ouraClient.GetSleepData(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to fetch sleep data")
	} else {
		log.WithField("sleep_score", sleepData.Score).Info("Successfully fetched sleep data")
		if err := ouraClient.SendToProcessor(ctx, cfg.ProcessorURL, sleepData); err != nil {
			log.WithError(err).Error("Failed to send sleep data to processor")
		}
	}

	// Fetch activity data
	activityData, err := ouraClient.GetActivityData(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to fetch activity data")
	} else {
		log.WithField("activity_score", activityData.Score).Info("Successfully fetched activity data")
		if err := ouraClient.SendToProcessor(ctx, cfg.ProcessorURL, activityData); err != nil {
			log.WithError(err).Error("Failed to send activity data to processor")
		}
	}

	// Fetch readiness data
	readinessData, err := ouraClient.GetReadinessData(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to fetch readiness data")
	} else {
		log.WithField("readiness_score", readinessData.Score).Info("Successfully fetched readiness data")
		if err := ouraClient.SendToProcessor(ctx, cfg.ProcessorURL, readinessData); err != nil {
			log.WithError(err).Error("Failed to send readiness data to processor")
		}
	}

	log.Info("oura-collector completed successfully")
}
