package email

import (
	"strings"
	"testing"
)

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	msg := string(buildMessage("from@example.com", "to@example.com", "Hello\nWorld", "body line"))
	if !strings.Contains(msg, "Subject: HelloWorld\r\n") {
		t.Fatalf("expected sanitized subject, got %q", msg)
	}
	if !strings.Contains(msg, "From: from@example.com\r\n") {
		t.Fatalf("missing From header: %q", msg)
	}
	if !strings.HasSuffix(msg, "body line") {
		t.Fatalf("missing body: %q", msg)
	}
}

func TestNewSMTPSenderValidation(t *testing.T) {
	t.Parallel()

	_, err := NewSMTPSender(SMTPConfig{Host: "smtp.gmail.com", Port: "587"})
	if err == nil {
		t.Fatal("expected error when username/password missing")
	}

	sender, err := NewSMTPSender(SMTPConfig{
		Host:     "smtp.gmail.com",
		Port:     "587",
		Username: "a@gmail.com",
		Password: "abcd efgh ijkl mnop",
		From:     "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.cfg.Password != "abcdefghijklmnop" {
		t.Fatalf("expected spaces stripped from app password, got %q", sender.cfg.Password)
	}
	if sender.cfg.From != "a@gmail.com" {
		t.Fatalf("expected From to default to username, got %q", sender.cfg.From)
	}
}
