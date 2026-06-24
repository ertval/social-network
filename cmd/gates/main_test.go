package main

import (
	"strings"
	"testing"

	"social-network/internal/gates"
)

func TestExtractSuggestion(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{msg: "checked: X | why: Y | status: FAIL - Z | debug: run 'make format'", want: "run 'make format'"},
		{msg: "checked: X | why: Y | status: OK - all good", want: ""},
		{msg: "checked: X | why: Y | status: FAIL - Z | debug: run 'golangci-lint run' to review", want: "run 'golangci-lint run' to review"},
		{msg: "", want: ""},
		{msg: "no pipes", want: ""},
	}

	for _, tt := range tests {
		got := extractSuggestion(tt.msg)
		if got != tt.want {
			t.Errorf("extractSuggestion(%q) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestExtractReason(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{
			msg:  "checked: format | why: consistency | status: OK - all files correctly styled",
			want: "all files correctly styled",
		},
		{
			msg:  "checked: lint | why: quality | status: FAIL - violations found | debug: run 'linter'",
			want: "violations found",
		},
		{
			msg:  "checked: coverage | why: coverage | status: SKIP - not applicable",
			want: "not applicable",
		},
		{
			msg:  "no structured format",
			want: "no structured format",
		},
	}

	for _, tt := range tests {
		got := extractReason(tt.msg)
		if got != tt.want {
			t.Errorf("extractReason(%q) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestPlainMessage_PASS(t *testing.T) {
	r := gates.Result{
		Gate:    "format",
		Status:  "PASS",
		Message: "checked: format | why: consistency | status: OK - all files correctly styled",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "[PASS]") {
		t.Errorf("plainMessage(PASS) should contain [PASS], got: %q", got)
	}
	if !strings.Contains(got, "all files correctly styled") {
		t.Errorf("plainMessage(PASS) should contain reason, got: %q", got)
	}
	if !strings.Contains(got, "format") {
		t.Errorf("plainMessage(PASS) should contain gate name, got: %q", got)
	}
}

func TestPlainMessage_FAIL(t *testing.T) {
	r := gates.Result{
		Gate:    "lint",
		Status:  "FAIL",
		Message: "checked: lint | why: quality | status: FAIL - violations found | debug: run 'golangci-lint run'",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "[FAIL]") {
		t.Errorf("plainMessage(FAIL) should contain [FAIL], got: %q", got)
	}
	if !strings.Contains(got, "violations found → run 'golangci-lint run'") {
		t.Errorf("plainMessage(FAIL) should contain reason + suggestion, got: %q", got)
	}
}

func TestPlainMessage_SKIP(t *testing.T) {
	r := gates.Result{
		Gate:    "coverage",
		Status:  "SKIP",
		Message: "checked: coverage | why: coverage | status: SKIP - not applicable",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "[SKIP]") {
		t.Errorf("plainMessage(SKIP) should contain [SKIP], got: %q", got)
	}
	if !strings.Contains(got, "not applicable") {
		t.Errorf("plainMessage(SKIP) should contain reason, got: %q", got)
	}
	if !strings.Contains(got, "coverage") {
		t.Errorf("plainMessage(SKIP) should contain gate name, got: %q", got)
	}
}

func TestPlainMessage_FAIL_Multiline(t *testing.T) {
	r := gates.Result{
		Gate:    "boundaries",
		Status:  "FAIL",
		Message: "checked: X | why: Y | status: FAIL - depguard violations:\ninvalid import | debug: run 'golangci-lint run'",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "depguard violations: invalid import") {
		t.Errorf("plainMessage(multiline) should contain collapsed reason, got: %q", got)
	}
	if !strings.Contains(got, "→") {
		t.Errorf("plainMessage(FAIL) should contain suggestion arrow, got: %q", got)
	}
}

func TestPlainSummary(t *testing.T) {
	report := gates.Report{
		Overall: "FAIL",
		Gates: []gates.Result{
			{Gate: "format", Status: "PASS", Message: "ok"},
			{Gate: "lint", Status: "FAIL", Message: "checked: lint | why: quality | status: FAIL - err | debug: fix it"},
			{Gate: "coverage", Status: "SKIP", Message: "ok"},
		},
	}

	got := plainSummary(report)
	if !strings.Contains(got, "1 passed") {
		t.Errorf("plainSummary should show pass count, got: %s", got)
	}
	if !strings.Contains(got, "1 failed") {
		t.Errorf("plainSummary should show fail count, got: %s", got)
	}
	if !strings.Contains(got, "1 skipped") {
		t.Errorf("plainSummary should show skip count, got: %s", got)
	}
	if !strings.Contains(got, "Some gates failed") {
		t.Errorf("plainSummary should show failure message, got: %s", got)
	}
}

func TestPlainSummary_AllPass(t *testing.T) {
	report := gates.Report{
		Overall: "PASS",
		Gates: []gates.Result{
			{Gate: "format", Status: "PASS", Message: "ok"},
			{Gate: "lint", Status: "PASS", Message: "ok"},
		},
	}
	got := plainSummary(report)
	if !strings.Contains(got, "All gates passed") {
		t.Errorf("plainSummary(all pass) should show success, got: %s", got)
	}
}

func TestPlainHeader(t *testing.T) {
	got := plainHeader()
	if !strings.Contains(got, "Review Gates") {
		t.Errorf("plainHeader should contain title, got: %s", got)
	}
	if strings.Contains(got, "\033") {
		t.Errorf("plainHeader should not contain ANSI codes, got: %s", got)
	}
	if strings.Contains(got, "━") {
		t.Errorf("plainHeader should not contain unicode box-drawing, got: %s", got)
	}
}
