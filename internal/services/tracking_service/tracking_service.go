package tracking_service

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	firestoreService "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/iterator"
)

type TrackingService interface {
	GetClickCount(ctx context.Context, shortID string) (int64, error)
	TrackClick(ctx context.Context, shortID, ip, userAgent string) error
	StreamClickLogs(ctx context.Context, w http.ResponseWriter, query dto.ClickLogsQuery) error
	GetAnalytics(ctx context.Context, query dto.ClickLogsQuery) (*dto.AnalyticsDTO, error)
}

type TrackingServiceImpl struct {
	firestore firestoreService.ClickLog
	redis     *redis.Client
}

func New(fs firestoreService.ClickLog, redis *redis.Client) TrackingService {
	return &TrackingServiceImpl{firestore: fs, redis: redis}
}

func (t *TrackingServiceImpl) TrackClick(ctx context.Context, shortID, ip, userAgent string) error {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}
	// redis
	if err := t.redis.Incr(ctx, "clicks:"+shortID).Err(); err != nil {
		return err
	}

	clickLog := &models.ClickLog{
		ShortID:   shortID,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	// save to firestore
	return t.firestore.AddClickLog(ctx, clickLog)
}

func (t *TrackingServiceImpl) GetClickCount(ctx context.Context, shortID string) (int64, error) {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return 0, shortlink_errors.ErrValidateRequest
	}

	count, err := t.redis.Get(ctx, "clicks:"+shortID).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, shortlink_errors.ErrFailedRetrieveData
	}

	return count, nil
}

func (s *TrackingServiceImpl) StreamClickLogs(ctx context.Context, w http.ResponseWriter, query dto.ClickLogsQuery) error {
	if err := validators.Validate.Var(query.ShortID, "short_id"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}

	iter, err := s.firestore.StreamClickLogs(ctx, query.ShortID)
	if err != nil {
		return err
	}
	defer iter.Stop()

	format := ctx.Value(utils.ExportFormatKey)
	bufWriter := bufio.NewWriter(w)
	defer bufWriter.Flush()
	switch format {
	case "csv":
		err := s.streamForCSV(bufWriter, iter)
		if err != nil {
			return err
		}
	case "json":
		err := s.streamForJSON(bufWriter, iter)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TrackingServiceImpl) streamForCSV(w io.Writer, iter *firestore.DocumentIterator) (err error) {
	csvWriter := csv.NewWriter(w)
	if err = csvWriter.Write([]string{"timestamp", "ip", "user_agent"}); err != nil {
		return
	}

	defer func() {
		csvWriter.Flush()
		if flushErr := csvWriter.Error(); flushErr != nil {
			log.Printf("CSV flush error: %v", flushErr)
			err = flushErr
		}
	}()

	for {
		var click models.ClickLog
		click, err = s.getClickLog(iter)
		if err != nil {
			if err == iterator.Done {
				err = nil
				break
			}
			return
		}

		if err = csvWriter.Write([]string{
			click.Timestamp.Format(time.RFC3339),
			click.IP,
			click.UserAgent,
		}); err != nil {
			log.Printf("Error writing to CSV: %v", err)
			// return fmt.Errorf("failed to write to CSV: %w", err)
		}
	}

	return
}

func (s *TrackingServiceImpl) streamForJSON(w io.Writer, iter *firestore.DocumentIterator) (err error) {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		return errors.New("writer must be a *bufio.Writer")
	}

	encoder := json.NewEncoder(bw)
	bw.Write([]byte("["))
	defer func() {
		if err == nil {
			bw.Write([]byte("]"))
		} else {
			bw.Write([]byte("]\n")) // Clean up partial JSON
		}
		bw.Flush()
	}()

	first := true
	for {
		var click models.ClickLog
		click, err = s.getClickLog(iter)
		if err != nil {
			if err == iterator.Done {
				err = nil
				break
			}
			log.Printf("Error iterating: %v", err)
			// return fmt.Errorf("failed to iterate over documents: %w", err)
			return err
		}

		if !first {
			bw.Write([]byte(","))
		}
		first = false

		// encode object
		if err = encoder.Encode(&struct {
			Timestamp string `json:"timestamp"`
			IP        string `json:"ip"`
			UserAgent string `json:"user_agent"`
		}{
			Timestamp: click.Timestamp.Format(time.RFC3339),
			IP:        click.IP,
			UserAgent: click.UserAgent,
		}); err != nil {
			log.Printf("Error encoding JSON: %v", err)
			break
		}
	}

	return nil
}

func (*TrackingServiceImpl) getClickLog(iter *firestore.DocumentIterator) (models.ClickLog, error) {
	doc, err := iter.Next()
	if err != nil {
		return models.ClickLog{}, err
	}

	var click models.ClickLog
	return click, doc.DataTo(&click)
}

func (t *TrackingServiceImpl) GetAnalytics(ctx context.Context, query dto.ClickLogsQuery) (*dto.AnalyticsDTO, error) {
	if err := validators.Validate.Struct(query); err != nil {
		return nil, shortlink_errors.ErrValidateRequest
	}

	logs, nextCursor, err := t.firestore.GetClickLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	dtoLogs := make([]dto.ClickLogDTO, 0, len(logs))
	var count int64 = 0
	for _, l := range logs {
		dtoLogs = append(dtoLogs, dto.ClickLogDTO{
			Timestamp: l.Timestamp,
			IP:        l.IP,
			UserAgent: l.UserAgent,
		})
		count++
	}

	responseData := &dto.AnalyticsDTO{
		ShortID:     query.ShortID,
		TotalClicks: count,
		Clicks:      dtoLogs,
		NextCursor:  nextCursor,
	}

	return responseData, nil
}
