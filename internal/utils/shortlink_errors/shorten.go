package shortlink_errors

import "errors"

var (
    ErrBlacklistedID = errors.New("custom ID is blacklisted")
    ErrIDExists      = errors.New("custom ID already exists")
    ErrGenerateID    = errors.New("failed to generate ID")
    ErrSaveShortlink = errors.New("failed to save short link")
)