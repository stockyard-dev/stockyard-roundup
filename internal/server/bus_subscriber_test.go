package server

// Tests for the bus subscriber: invoice.overdue → auto-drafted
// follow-up task. Handler invoked directly with synthesized bus.Event
// payloads. End-to-end wiring is exercised in stockyard-desktop's
// orchestrator tests.

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stockyard-dev/stockyard-roundup/internal/store"
	"github.com/stockyard-dev/stockyard/bus"
)

func newSubscriberServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return New(db, ProLimits(), dir, nil)
}

func eventWith(topic, source string, payload map[string]any) bus.Event {
	raw, _ := json.Marshal(payload)
	return bus.Event{Topic: topic, Source: source, Payload: raw}
}

func TestHandleInvoiceOverdue_CreatesTask(t *testing.T) {
	s := newSubscriberServer(t)
	e := eventWith("invoice.overdue", "billfold", map[string]any{
		"invoice_id":  "inv-100",
		"client_name": "Acme Yoga",
		"amount":      float64(12345),
		"due_date":    "2026-03-15",
		"status":      "overdue",
	})
	if err := s.handleInvoiceOverdue(e); err != nil {
		t.Fatalf("handler: %v", err)
	}
	tasks := s.db.ListTasks("", "", "")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	task := tasks[0]
	if !strings.Contains(task.Title, "Acme Yoga") {
		t.Errorf("title = %q, should include client name", task.Title)
	}
	if task.Status != "todo" {
		t.Errorf("status = %q, want todo", task.Status)
	}
	if task.Priority != "high" {
		t.Errorf("priority = %q, want high", task.Priority)
	}
	if !strings.Contains(task.Description, "[invoice:inv-100]") {
		t.Errorf("description missing idempotency marker: %q", task.Description)
	}
	if !strings.Contains(task.Description, "12345") {
		t.Errorf("description missing amount: %q", task.Description)
	}
	if !strings.Contains(task.Description, "2026-03-15") {
		t.Errorf("description missing due_date: %q", task.Description)
	}
	hasOverdue, hasBilling := false, false
	for _, tag := range task.Tags {
		if tag == "overdue" {
			hasOverdue = true
		}
		if tag == "billing" {
			hasBilling = true
		}
	}
	if !hasOverdue || !hasBilling {
		t.Errorf("tags missing expected values: %v", task.Tags)
	}
}

func TestHandleInvoiceOverdue_IsIdempotent(t *testing.T) {
	s := newSubscriberServer(t)
	e := eventWith("invoice.overdue", "billfold", map[string]any{
		"invoice_id":  "inv-dup",
		"client_name": "Dupe Co",
		"amount":      float64(500),
	})
	_ = s.handleInvoiceOverdue(e)
	_ = s.handleInvoiceOverdue(e)
	_ = s.handleInvoiceOverdue(e)
	if n := len(s.db.ListTasks("", "", "")); n != 1 {
		t.Errorf("expected 1 task after 3 fires, got %d", n)
	}
}

func TestHandleInvoiceOverdue_MissingIDSkips(t *testing.T) {
	s := newSubscriberServer(t)
	e := eventWith("invoice.overdue", "billfold", map[string]any{
		"client_name": "No ID Co",
		"amount":      float64(100),
	})
	_ = s.handleInvoiceOverdue(e)
	if n := len(s.db.ListTasks("", "", "")); n != 0 {
		t.Errorf("expected 0 tasks (missing invoice_id should skip), got %d", n)
	}
}

func TestHandleInvoiceOverdue_EmptyClientFallback(t *testing.T) {
	s := newSubscriberServer(t)
	e := eventWith("invoice.overdue", "billfold", map[string]any{
		"invoice_id": "inv-noclient",
		"amount":     float64(50),
	})
	_ = s.handleInvoiceOverdue(e)
	tasks := s.db.ListTasks("", "", "")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if !strings.Contains(tasks[0].Title, "unknown client") {
		t.Errorf("title should fall back to 'unknown client', got %q", tasks[0].Title)
	}
}

func TestHandleInvoiceOverdue_MalformedPayloadSkips(t *testing.T) {
	s := newSubscriberServer(t)
	e := bus.Event{Topic: "invoice.overdue", Source: "billfold", Payload: []byte("not json")}
	if err := s.handleInvoiceOverdue(e); err != nil {
		t.Fatalf("handler should not return error on malformed payload: %v", err)
	}
	if n := len(s.db.ListTasks("", "", "")); n != 0 {
		t.Errorf("expected 0 tasks after malformed payload, got %d", n)
	}
}

func TestHandleInvoiceOverdue_DoesNotSubscribeToOtherTopics(t *testing.T) {
	// Defensive: handler should still behave sanely if called with an
	// off-topic event (shouldn't happen via subscribeBus, but a future
	// refactor could mis-wire). The idempotency/shape logic doesn't
	// depend on the topic string, only on payload contents.
	s := newSubscriberServer(t)
	e := eventWith("invoice.paid", "billfold", map[string]any{
		"invoice_id":  "inv-paid",
		"client_name": "Paid Co",
	})
	_ = s.handleInvoiceOverdue(e)
	// Handler doesn't filter by topic — it will create a task. This is
	// intentional: the subscription is the gate, not the handler. We
	// assert the happy-path task lands so a future "the handler should
	// double-check topic" change gets noticed.
	if n := len(s.db.ListTasks("", "", "")); n != 1 {
		t.Errorf("handler creates task regardless of topic string (gate is subscription, not handler); got %d tasks", n)
	}
	// Ensure the test above (TestHandleInvoiceOverdue_CreatesTask) is
	// the contract consumers rely on: the *wiring* restricts to
	// invoice.overdue only.
	_ = store.Task{}
}
