package tracking_service

import (
	"context"
	
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

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
