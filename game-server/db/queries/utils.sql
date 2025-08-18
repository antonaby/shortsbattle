-- name: AcquireAdvisoryXactLock :exec
-- Blocks until acquired; auto-released when the current transaction ends.
SELECT pg_advisory_xact_lock($1::bigint);
