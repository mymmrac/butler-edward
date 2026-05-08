package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool/memory"
)

//nolint:lll
const systemPrompt = `
You are a helpful butler named Edward.
You have access to different tools to help you with your tasks. Try solving the task without using any tools first.

CRITICAL INSTRUCTIONS:
	1. If you do not know the answer, use a tool to find it.
	2. Once you receive the observation from the tool, DO NOT call the tool again unless needed.
	3. Present the tool's output to the user in readable format.
	4. Do not pretend to "call" the tool and actually do nothing, always use tools to find the answer if not know from general knowledge.
	5. Analyze tool call results and in case of errors, try to resolve them.
	6. Use memory to remember and recall things that are relevant to the user and can be used in the future.
	7. When remembering things, use keywords that are short and concise and don't contain the thing itself.
	8. Don't call tools without reason. Make sure that you are using the tools only when needed.

CONTEXT:
	1. The current directory is "./" and all paths are relative to it, and it's the root path.
	2. All keywords for memory are listed below. Use only them to recall and forget things.
`

const maxSessionNameLength = 64

//nolint:lll
const sessionNameSystemPrompt = `
Your only task is to create very short, concise and readable chat names based on user's messages that represent the chat intent.
Output only the generated name, no quotes needed, spaces allowed; it must be less than 64 characters.
`

func (a *Agent) buildSystemPrompt(ctx context.Context, lc *loopContext) (string, error) {
	memories, err := a.storage.ListPrefix(ctx, lc.userID, memory.KeyPrefix)
	if err != nil {
		return "", fmt.Errorf("list memories: %w", err)
	}

	sb := &strings.Builder{}
	_, _ = sb.WriteString(a.normalizeContent(systemPrompt))

	_, _ = sb.WriteString("\n\nMEMORY KEYWORDS:\n")
	hasKeywords := false
	for keyword := range memories {
		hasKeywords = true
		_, _ = sb.WriteString("- " + strings.TrimPrefix(keyword, memory.KeyPrefix) + "\n")
	}
	if !hasKeywords {
		_, _ = sb.WriteString("No keywords remembered yet.\n")
	}

	return sb.String(), nil
}
