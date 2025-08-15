package common

import (
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
)

func IsMessageEmpty(msg modelstypes.Message) bool {
	return len(msg.Parts) == 0
}
