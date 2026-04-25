package tools

import "testing"

func TestErrorResult(t *testing.T) {
	tests := []struct {
		name         string
		msg          string
		expectedLLM  string
		expectedUser string
	}{
		{
			name:         "Standard message",
			msg:          "something went wrong",
			expectedLLM:  "Error: something went wrong",
			expectedUser: "Sorry, I encountered an error: something went wrong",
		},
		{
			name:         "Empty message",
			msg:          "",
			expectedLLM:  "Error: ",
			expectedUser: "Sorry, I encountered an error: ",
		},
		{
			name:         "Special characters",
			msg:          "!@#$%^&*()",
			expectedLLM:  "Error: !@#$%^&*()",
			expectedUser: "Sorry, I encountered an error: !@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ErrorResult(tt.msg)

			if res.ForLLM != tt.expectedLLM {
				t.Errorf("expected ForLLM %q, got %q", tt.expectedLLM, res.ForLLM)
			}

			if res.ForUser != tt.expectedUser {
				t.Errorf("expected ForUser %q, got %q", tt.expectedUser, res.ForUser)
			}

			if !res.IsError {
				t.Error("expected IsError to be true")
			}

			if res.Silent {
				t.Error("expected Silent to be false")
			}

			if res.Async {
				t.Error("expected Async to be false")
			}
		})
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
