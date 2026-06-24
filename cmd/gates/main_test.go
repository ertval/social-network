package main

import (
	"strings"
	"testing"

	"social-network/internal/gates"
)

func TestHighlightStatus(t *testing.T) {
	saved := noColor
	noColor = true
	defer func() { noColor = saved }()

	tests := []struct {
		msg  string
		want string
	}{
		{msg: "no pipes", want: "no pipes"},
		{msg: "checked: formatting | why: consistency | status: OK - all good", want: "status: OK - all good | reason: Checks formatting consistency"},
		{msg: "checked: lint | why: quality | status: FAIL - violations | debug: fix it", want: "status: FAIL - violations | reason: Checks lint quality | debug: fix it"},
	}
	noCol := func(s string) string { return s }
	for _, tt := range tests {
		got := highlightStatus(tt.msg, noCol)
		if got != tt.want {
			t.Errorf("highlightStatus(%q) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestPlainMessage(t *testing.T) {
	r := gates.Result{
		Gate:    "format",
		Status:  "PASS",
		Message: "checked: Go formatting | why: consistency | status: OK - all good",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "[PASS]") {
		t.Errorf("plainMessage should contain [PASS], got: %q", got)
	}
	if !strings.Contains(got, "status: OK") {
		t.Errorf("plainMessage should have status first, got: %q", got)
	}
	if !strings.Contains(got, "reason: Checks Go formatting consistency") {
		t.Errorf("plainMessage should contain merged reason, got: %q", got)
	}
	if !strings.Contains(got, "format") {
		t.Errorf("plainMessage should contain gate name, got: %q", got)
	}
}

func TestPlainMessage_FailWithDebug(t *testing.T) {
	r := gates.Result{
		Gate:    "lint",
		Status:  "FAIL",
		Message: "checked: lint checks | why: quality | status: FAIL - violations found | debug: run 'golangci-lint run'",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "status: FAIL - violations found") {
		t.Errorf("plainMessage(FAIL) should show status first, got: %q", got)
	}
	if !strings.Contains(got, "debug:") {
		t.Errorf("plainMessage(FAIL) should contain debug, got: %q", got)
	}
}

func TestPlainMessage_NoStructuredFormat(t *testing.T) {
	r := gates.Result{
		Gate:    "custom",
		Status:  "PASS",
		Message: "plain unstructured message",
	}
	got := plainMessage(r)
	if !strings.Contains(got, "plain unstructured message") {
		t.Errorf("plainMessage(no-structure) should pass through original message, got: %q", got)
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
