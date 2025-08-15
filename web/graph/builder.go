package graph

import (
	"fmt"
	"strings"

	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/agents/interfaces"
)

// Build generates a DOT language string to represent the agent hierarchy.
func Build(agent interfaces.LlmAgent) string {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  bgcolor=\"#f8f9fa\";\n")
	sb.WriteString("  node [shape=box, style=\"rounded,filled\", fillcolor=\"#ffffff\", fontname=\"Inter\"];\n")
	sb.WriteString("  edge [fontname=\"Inter\"];\n\n")

	buildNode(&sb, agent)

	sb.WriteString("}\n")
	return sb.String()
}

func buildNode(sb *strings.Builder, agent interfaces.LlmAgent) {
	agentID := sanitizeID(agent.GetName())

	// Define the node for the current agent
	switch a := agent.(type) {
	case *agents.BaseLlmAgent:
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\\n(LLM Agent)\", fillcolor=\"#e0eafc\"];\n", agentID, agent.GetName()))
		for _, tool := range a.GetTools() {
			toolID := sanitizeID(tool.Name())
			sb.WriteString(fmt.Sprintf("  %s [label=\"%s\\n(Tool)\", shape=cylinder, fillcolor=\"#fff3cd\"];\n", toolID, tool.Name()))
			sb.WriteString(fmt.Sprintf("  %s -> %s;\n", agentID, toolID))
		}
	case *agents.SequentialAgent:
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\\n(Sequential Workflow)\", fillcolor=\"#d1e7dd\"];\n", agentID, agent.GetName()))
		var prevSubAgentID string
		for i, subAgent := range a.SubAgents {
			subAgentID := sanitizeID(subAgent.GetName())
			buildNode(sb, subAgent)
			if i == 0 {
				sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"start\"];\n", agentID, subAgentID))
			} else {
				sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"next\"];\n", prevSubAgentID, subAgentID))
			}
			prevSubAgentID = subAgentID
		}
	case *agents.ParallelAgent:
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\\n(Parallel Workflow)\", fillcolor=\"#d1e7dd\"];\n", agentID, agent.GetName()))
		for _, subAgent := range a.SubAgents {
			subAgentID := sanitizeID(subAgent.GetName())
			buildNode(sb, subAgent)
			sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"concurrent\"];\n", agentID, subAgentID))
		}
	case *agents.LoopAgent:
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\\n(Loop Workflow)\", fillcolor=\"#d1e7dd\"];\n", agentID, agent.GetName()))
		subAgentID := sanitizeID(a.SubAgent.GetName())
		buildNode(sb, a.SubAgent)
		sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"start loop\"];\n", agentID, subAgentID))
		sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"repeat\", style=dashed, constraint=false];\n", subAgentID, subAgentID))
	}
}

func sanitizeID(name string) string {
	// DOT IDs cannot contain spaces or special characters.
	r := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return r.Replace(name)
}
