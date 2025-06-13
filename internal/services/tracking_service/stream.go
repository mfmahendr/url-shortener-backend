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
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"google.golang.org/api/iterator"
)

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