package service

import (
    "math"
    "testing"
)

func TestWeightedQuotaBytes(t *testing.T) {
    for _, tc := range []struct{ bytes, percent, want int64 }{
        {0, 100, 0}, {-1, 200, 0}, {1000, 100, 1000},
        {1000, 200, 2000}, {1000, 50, 500}, {123, 125, 153},
        {math.MaxInt64, 100, math.MaxInt64}, {math.MaxInt64, 200, math.MaxInt64},
    } {
        if got := weightedQuotaBytes(tc.bytes, tc.percent); got != tc.want {
            t.Errorf("weightedQuotaBytes(%d, %d) = %d; want %d", tc.bytes, tc.percent, got, tc.want)
        }
    }
}

func TestMultiplierChangeOnlyAffectsNewDeltas(t *testing.T) {
    // Existing usage is never multiplied again; node baselines remain raw.
    used := int64(1000)
    baseline := int64(2000)
    snapshot := int64(2100)
    used += weightedQuotaBytes(snapshot-baseline, 200)
    baseline = snapshot
    used += weightedQuotaBytes(snapshot-baseline, 200)
    if used != 1200 { t.Fatalf("duplicate node snapshot was billed: %d", used) }
    snapshot = 2300
    used += weightedQuotaBytes(snapshot-baseline, 50)
    if used != 1300 { t.Fatalf("historical usage changed: %d", used) }
}
