# Shared quota multiplier

The master panel's Settings / General field applies to new local traffic and
new node snapshot deltas. `quotaMultiplierPercent` is an integer percentage:
100 = 1x, 200 = 2x, 50 = 0.5x. Valid values are 1 through 10000.

Keep this setting at 100 on sub-nodes: billing is performed centrally on the
master. Do not independently enable weighted accounting on both levels.
Existing `client_traffics` counters are left intact. Node snapshot baselines
remain raw; global totals pushed to nodes are not multiplied on ingestion.
Inbound network totals remain physical bytes. Client upload/download and quota
usage are billed bytes, so the panel, subscriptions and Telegram bot agree.
Fractional bytes are rounded down per delta; additions saturate at int64 max.

Change this field through the existing settings Save action. No client reset,
quota change or expiry change is performed. Traffic already reported before a
change is never reprocessed. In-flight, not-yet-recorded deltas use the setting
read in their accounting transaction.

Deploy the rebuilt master backend and frontend together. Restarting only the
Telegram bot does not install this feature. Back up x-ui.db before deployment.

Verification commands (from the repository root):

    gofmt -w internal/web/service/quota_multiplier*.go internal/web/service/inbound_traffic.go internal/web/service/inbound_node.go internal/web/service/setting.go internal/web/entity/entity.go
    go test ./internal/web/service -run 'TestWeightedQuotaBytes|TestMultiplierChangeOnlyAffectsNewDeltas|Test.*Node.*Traffic|Test.*Traffic' -count=1
    cd frontend
    npm ci
    npm run typecheck
    npm run build
