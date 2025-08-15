package sessions

import "github.com/KennethanCeyer/adk-go/models/types"

// Session stores the state of an interaction with an agent.
type Session struct {
	History []types.Message
	// Additional state can be stored here.
}
