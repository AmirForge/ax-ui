package service

import (
    "fmt"
    "math"
    "strconv"

    "github.com/mhsanaei/3x-ui/v3/internal/database/model"
    "gorm.io/gorm"
)

// Read inside the traffic transaction: a setting change affects only later
// deltas, never the already committed counters or the node's raw baseline.
func quotaMultiplierPercent(tx *gorm.DB) (int64, error) {
    var rows []model.Setting
    if err := tx.Where("key = ?", "quotaMultiplierPercent").Limit(1).Find(&rows).Error; err != nil {
        return 0, err
    }
    if len(rows) == 0 { return 100, nil }
    value, err := strconv.ParseInt(rows[0].Value, 10, 64)
    if err != nil || value < 1 || value > 10000 {
        return 0, fmt.Errorf("invalid quota multiplier percent: %q", rows[0].Value)
    }
    return value, nil
}

// Integer quotient/remainder avoids overflow and floating point precision loss.
// Fractions of a byte are rounded down (less than one byte per delta).
func weightedQuotaBytes(bytes, percent int64) int64 {
    if bytes <= 0 { return 0 }
    whole, fraction := bytes / 100, bytes % 100
    extra := fraction * percent / 100
    if whole > (math.MaxInt64-extra)/percent { return math.MaxInt64 }
    return whole * percent + extra
}
