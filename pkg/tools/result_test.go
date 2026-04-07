package tools

import "testing"

func TestErrorResult(t *testing.T) {
	msg := "test error"
	res := ErrorResult(msg)

	expectedLLM := "Error: " + msg
	if res.ForLLM != expectedLLM {
		t.Errorf("expected ForLLM %q, got %q", expectedLLM, res.ForLLM)
	}

	expectedUser := "Sorry, I encountered an error: " + msg
	if res.ForUser != expectedUser {
		t.Errorf("expected ForUser %q, got %q", expectedUser, res.ForUser)
	}

	if !res.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestSuccessResult(t *testing.T) {
	llmMsg := "success for llm"
	userMsg := "success for user"
	res := SuccessResult(llmMsg, userMsg)

	if res.ForLLM != llmMsg {
		t.Errorf("expected ForLLM %q, got %q", llmMsg, res.ForLLM)
	}

	if res.ForUser != userMsg {
		t.Errorf("expected ForUser %q, got %q", userMsg, res.ForUser)
	}

	if res.IsError {
		t.Error("expected IsError to be false")
	}

	if res.Silent {
		t.Error("expected Silent to be false")
	}
}

func TestSilentResult(t *testing.T) {
	llmMsg := "silent for llm"
	res := SilentResult(llmMsg)

	if res.ForLLM != llmMsg {
		t.Errorf("expected ForLLM %q, got %q", llmMsg, res.ForLLM)
	}

	if res.ForUser != "" {
		t.Errorf("expected empty ForUser, got %q", res.ForUser)
	}

	if !res.Silent {
		t.Error("expected Silent to be true")
	}

	if res.IsError {
		t.Error("expected IsError to be false")
	}
}
