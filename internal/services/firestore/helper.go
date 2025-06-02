package firestore_service

import (
	"time"
	"cloud.google.com/go/firestore"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
)


func (s *FirestoreServiceImpl) buildQuery(clickLogsQuery dto.ClickLogsQuery) (query firestore.Query) {
	query = s.client.Collection("click_logs").Where("short_id", "==", clickLogsQuery.ShortID)

	// Filter range waktu
	if !clickLogsQuery.After.IsZero() {
		query = query.Where("timestamp", ">", clickLogsQuery.After)
	}
	if !clickLogsQuery.Before.IsZero() {
		query = query.Where("timestamp", "<", clickLogsQuery.Before)
	}

	// Urutan
	if clickLogsQuery.OrderDesc {
		query = query.OrderBy("timestamp", firestore.Desc)
	} else {
		query = query.OrderBy("timestamp", firestore.Asc)
	}

	// Cursor
	if clickLogsQuery.Cursor != "" {
		parsedCursor, err := time.Parse(time.RFC3339Nano, clickLogsQuery.Cursor)
		if err == nil {
			query = query.StartAfter(parsedCursor)
		}
	}

	// Limit
	if clickLogsQuery.Limit <= 0 || clickLogsQuery.Limit > 100 {
		clickLogsQuery.Limit = 50
	}
	query = query.Limit(clickLogsQuery.Limit)

	return query
}