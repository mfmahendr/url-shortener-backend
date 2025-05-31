package firestore_service

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"google.golang.org/api/iterator"
)

type ClickLog interface {
	AddClickLog(ctx context.Context, doc interface{}) error
	GetClickLog(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error)
	GetAnalytics(ctx context.Context, shortID string) (int, []models.ClickLog, error)
}

func (s *FirestoreServiceImpl) AddClickLog(ctx context.Context, doc interface{}) error {
	_, _, err := s.client.Collection("click_logs").Add(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to add click_logs: %w", err)
	}
	return nil
}

func (s *FirestoreServiceImpl) GetClickLog(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error) {
	docSnap, err := s.client.Collection("click_logs").Doc(shortID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get click log: %w", err)
	}
	return docSnap, nil
}

func (s *FirestoreServiceImpl) GetAnalytics(ctx context.Context, shortID string) (int, []models.ClickLog, error) {
	iter := s.client.Collection("click_logs").
		Where("short_id", "==", shortID).
		OrderBy("timestamp", firestore.Desc).
		Limit(100).
		Documents(ctx)

	var logs []models.ClickLog
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, nil, shortlink_errors.ErrFailedRetrieveData
		}

		data := doc.Data()

		timestamp, _ := data["timestamp"].(time.Time)
		ip, _ := data["ip"].(string)
		ua, _ := data["user_agent"].(string)

		logs = append(logs, models.ClickLog{
			Timestamp: timestamp,
			IP:        ip,
			UserAgent: ua,
		})
		count++
	}

	if count == 0 {
		return 0, nil, shortlink_errors.ErrNotFound
	}

	return count, logs, nil
}
