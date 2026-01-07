package run

import (
	"context"
	"testing"

	"charm.land/fantasy"

	"ayo/internal/agent"
)

func TestBuildMessagesOmitsEmpty(t *testing.T) {
	r := &Runner{}
	ag := agent.Agent{CombinedSystem: "", SkillsPrompt: "", Model: "m"}
	msgs := r.buildMessages(ag, "hi")
	if len(msgs) != 1 {
		t.Fatalf("expected single user message, got %d", len(msgs))
	}
	if msgs[0].Role != fantasy.MessageRoleUser {
		t.Fatalf("expected user role, got %s", msgs[0].Role)
	}
	// Check content contains "hi"
	content := getTextContent(msgs[0])
	if content != "hi" {
		t.Fatalf("expected user message to contain 'hi': got %q", content)
	}
}

func TestBuildMessagesOrdersSystemSkillsUser(t *testing.T) {
	r := &Runner{}
	ag := agent.Agent{CombinedSystem: "SYS", SkillsPrompt: "SKILLS", Model: "m"}
	msgs := r.buildMessages(ag, "hi")
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	wantRoles := []fantasy.MessageRole{fantasy.MessageRoleSystem, fantasy.MessageRoleSystem, fantasy.MessageRoleUser}
	wantContents := []string{"SYS", "SKILLS", "hi"}
	for i := range wantRoles {
		if msgs[i].Role != wantRoles[i] {
			t.Fatalf("msg %d role mismatch: got %s, want %s", i, msgs[i].Role, wantRoles[i])
		}
		content := getTextContent(msgs[i])
		if content != wantContents[i] {
			t.Fatalf("msg %d content should be %q: got %q", i, wantContents[i], content)
		}
	}
}

func TestRunChatStopsAfterEmptyModel(t *testing.T) {
	r := &Runner{sessions: make(map[string]*Session)}
	_, err := r.runChat(context.Background(), agent.Agent{Model: ""}, nil)
	if err == nil {
		t.Fatalf("expected error from empty model")
	}
}

// getTextContent extracts text content from a fantasy message.
func getTextContent(msg fantasy.Message) string {
	for _, part := range msg.Content {
		if tp, ok := part.(fantasy.TextPart); ok {
			return tp.Text
		}
	}
	return ""
}
