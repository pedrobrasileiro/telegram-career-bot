package telegram

import "testing"

func TestEmojiForTrackerStatus(t *testing.T) {
	cases := map[string]string{
		"Applied":   "📤",
		"Interview": "🎯",
		"Offer":     "🎉",
		"Rejected":  "❌",
		"Discarded": "🗑",
		"Evaluated": "📋",
		"SKIP":      "⏭",
		"Algo raro": "📌",
	}
	for status, want := range cases {
		if got := emojiForTrackerStatus(status); got != want {
			t.Errorf("emojiForTrackerStatus(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestEmojiForTask(t *testing.T) {
	cases := map[string]string{
		"scan":         "🔎",
		"avaliação":    "📊",
		"desconhecido": "⏳",
	}
	for taskType, want := range cases {
		if got := emojiForTask(taskType); got != want {
			t.Errorf("emojiForTask(%q) = %q, want %q", taskType, got, want)
		}
	}
}
