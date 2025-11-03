package firestore_service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
)


func (s *FirestoreServiceImpl) buildClickLogsQuery(shortID string, clickLogsQuery dto.ClickLogsQuery) (query firestore.Query) {
	query = s.client.Collection("click_logs").Where("short_id", "==", shortID)

	// Filter the range of click logs
	if !clickLogsQuery.After.IsZero() {
		query = query.Where("timestamp", ">", clickLogsQuery.After)
	}
	if !clickLogsQuery.Before.IsZero() {
		query = query.Where("timestamp", "<", clickLogsQuery.Before)
	}

	// pagination query
	query = buildPaginationQuery(clickLogsQuery.PaginationQuery, query)

	return query
}

func buildPaginationQuery(pq dto.PaginationQuery, query firestore.Query) firestore.Query {
	if pq.OrderDesc {
		query = query.OrderBy("timestamp", firestore.Desc)
	} else {
		query = query.OrderBy("timestamp", firestore.Asc)
	}

	// Cursor
	if pq.Cursor != "" {
		parsedCursor, err := time.Parse(time.RFC3339Nano, pq.Cursor)
		if err == nil {
			query = query.StartAfter(parsedCursor)
		}
	}

	// Limit
	if pq.Limit <= 0 || pq.Limit > 100 {
		pq.Limit = 50
	}
	query = query.Limit(pq.Limit)
	return query
}

func checkDocumentExists(ctx context.Context, q firestore.Query) (bool, error) {
	_, err := q.Limit(1).Documents(ctx).Next()
	if err == iterator.Done {
		return true, shortlink_errors.ErrNotFound
	}
	if err != nil {
		return true, fmt.Errorf("failed to check documents: %w", err)
	}
	return false, nil
}

func normalizeURLForBlacklist(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	path := strings.TrimSuffix(u.EscapedPath(), "/")
	if path == "" {
		return host
	}
	return host + path
}