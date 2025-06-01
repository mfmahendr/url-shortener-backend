package firestore_service

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"google.golang.org/api/iterator"
)

type ClickLog interface {
	AddClickLog(ctx context.Context, doc interface{}) error
	GetClickLogs(ctx context.Context, shortID string) ([]models.ClickLog, error)
	GetAnalytics(ctx context.Context, shortID string) (int64, []models.ClickLog, error)
}

func (s *FirestoreServiceImpl) AddClickLog(ctx context.Context, doc interface{}) error {
	_, _, err := s.client.Collection("click_logs").Add(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to add click_logs: %w", err)
	}
	return nil
}

func (s *FirestoreServiceImpl) GetClickLogs(ctx context.Context, shortID string) ([]models.ClickLog, error) {
	iter := s.client.Collection("click_logs").
		Where("short_id", "==", shortID).
		OrderBy("timestamp", firestore.Desc).
		Limit(100).
		Documents(ctx)
	defer iter.Stop()

	var logs []models.ClickLog
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error retrieving document: %v\n", err)
			return nil, shortlink_errors.ErrFailedRetrieveData
		}

		var clickLog models.ClickLog
		if err := doc.DataTo(&clickLog); err != nil {
			fmt.Printf("Error converting document data to ClickLog: %v\n", err)
			return nil, shortlink_errors.ErrFailedRetrieveData
		}

		logs = append(logs, clickLog)
	}

	return logs, nil
}


func (s *FirestoreServiceImpl) GetAnalytics(ctx context.Context, shortID string) (int64, []models.ClickLog, error) {
	iter := s.client.Collection("click_logs").
		Where("short_id", "==", shortID).
		OrderBy("timestamp", firestore.Desc).
		Limit(100).
		Documents(ctx)
	defer iter.Stop()

	var logs []models.ClickLog
	var count int64 = 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error retrieving document: %v\n", err)
			return 0, nil, shortlink_errors.ErrFailedRetrieveData
		}

		var clickLog models.ClickLog
		if err := doc.DataTo(&clickLog); err != nil {
			fmt.Printf("Error converting document data to ClickLog: %v\n", err)
			return 0, nil, shortlink_errors.ErrFailedRetrieveData
		}

		logs = append(logs, clickLog)
		count++
	}

	return count, logs, nil
}
