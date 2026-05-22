package latch

// helpers.go — small functions copied from dock that the moved
// handlers depend on. Kept here so latch-svc has no compile-time
// dependency on the dock package.

import (
	"crypto/rand"
	"encoding/hex"
)

// generateResourceID — 16 random bytes hex-encoded. Same shape as
// dock's store.go::generateResourceID. Used as primary key for new
// rows in latch_proxies / latch_rules / latch_profiles /
// latch_service_nodes / latch_service_node_agent_tokens.
func generateResourceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
