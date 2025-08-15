package common

import (
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
)

// IsMessageEmpty checks if a message has no content.
// A message is considered empty if it has no parts.
func IsMessageEmpty(msg modelstypes.Message) bool {
	return len(msg.Parts) == 0
}
