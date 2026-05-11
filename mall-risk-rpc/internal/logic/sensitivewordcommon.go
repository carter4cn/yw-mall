package logic

import (
	"context"
	"strings"
	"sync"
	"time"

	"mall-risk-rpc/risk"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// In-memory cache of active sensitive words with periodic refresh (60s TTL).
// Used by CheckText to avoid hitting the DB on every check.
type sensitiveWordCache struct {
	mu        sync.RWMutex
	words     []*sensitiveWordRow
	loadedAt  time.Time
	ttl       time.Duration
}

var sensitiveCache = &sensitiveWordCache{ttl: 60 * time.Second}

type sensitiveWordRow struct {
	Id         uint64 `db:"id"`
	Word       string `db:"word"`
	Category   string `db:"category"`
	Action     string `db:"action"`
	Status     int64  `db:"status"`
	CreateTime int64  `db:"create_time"`
}

const sensitiveWordCols = "id, word, category, action, status, create_time"

func toSensitiveWordProto(r *sensitiveWordRow) *risk.SensitiveWord {
	return &risk.SensitiveWord{
		Id:         int64(r.Id),
		Word:       r.Word,
		Category:   r.Category,
		Action:     r.Action,
		Status:     int32(r.Status),
		CreateTime: r.CreateTime,
	}
}

// loadAllActive returns active sensitive words, refreshing the cache as needed.
func (c *sensitiveWordCache) loadAllActive(ctx context.Context, db sqlx.SqlConn) ([]*sensitiveWordRow, error) {
	c.mu.RLock()
	if time.Since(c.loadedAt) < c.ttl && c.words != nil {
		out := c.words
		c.mu.RUnlock()
		return out, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.loadedAt) < c.ttl && c.words != nil {
		return c.words, nil
	}
	var rows []*sensitiveWordRow
	if err := db.QueryRowsCtx(ctx, &rows,
		"SELECT "+sensitiveWordCols+" FROM sensitive_word WHERE status=1"); err != nil {
		return nil, err
	}
	c.words = rows
	c.loadedAt = time.Now()
	return rows, nil
}

func (c *sensitiveWordCache) invalidate() {
	c.mu.Lock()
	c.loadedAt = time.Time{}
	c.words = nil
	c.mu.Unlock()
}

// scan finds all sensitive-word matches inside text.
func scanText(text string, words []*sensitiveWordRow) []*sensitiveWordRow {
	if text == "" || len(words) == 0 {
		return nil
	}
	matches := make([]*sensitiveWordRow, 0)
	for _, w := range words {
		if w.Word != "" && strings.Contains(text, w.Word) {
			matches = append(matches, w)
		}
	}
	return matches
}
