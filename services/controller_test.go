package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adruzhkin/hacker-news-reader-golang/repo"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	return New(10, repo.NewStoryRepo(), repo.NewUserRepo(), 100, &http.Client{})
}

func overrideItemURL(t *testing.T, serverURL string) {
	t.Helper()
	orig := itemURL
	itemURL = serverURL + "/item/%v.json"
	t.Cleanup(func() { itemURL = orig })
}

func overrideBaseURL(t *testing.T, serverURL string) {
	t.Helper()
	orig := baseURL
	baseURL = serverURL + "/topstories.json?limitToFirst=%d&orderBy=%s"
	t.Cleanup(func() { baseURL = orig })
}

// --- doRequest tests ---

func TestDoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	svc := newTestService(t)
	body, err, shouldRetry := svc.doRequest(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shouldRetry {
		t.Error("shouldRetry = true, want false")
	}
	if string(body) != `{"id":1}` {
		t.Errorf("body = %q, want %q", string(body), `{"id":1}`)
	}
}

func TestDoRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	svc := newTestService(t)
	_, err, shouldRetry := svc.doRequest(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !shouldRetry {
		t.Error("shouldRetry = false, want true for 500")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("error = %q, want to contain 'server error'", err.Error())
	}
}

func TestDoRequest_ClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	svc := newTestService(t)
	_, err, shouldRetry := svc.doRequest(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if shouldRetry {
		t.Error("shouldRetry = true, want false for 404")
	}
	if !strings.Contains(err.Error(), "client error") {
		t.Errorf("error = %q, want to contain 'client error'", err.Error())
	}
}

func TestDoRequest_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := newTestService(t)
	_, err, _ := svc.doRequest(ctx, "http://localhost:0/never")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

// --- doWithRetry tests ---

func TestDoWithRetry_SuccessFirstAttempt(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(`"ok"`))
	}))
	defer server.Close()

	svc := newTestService(t)
	body, err := svc.doWithRetry(context.Background(), server.URL, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `"ok"` {
		t.Errorf("body = %q, want %q", string(body), `"ok"`)
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("server hit %d times, want 1", got)
	}
}

func TestDoWithRetry_RetriesOnServerError(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n <= 2 {
			w.WriteHeader(500)
			return
		}
		w.Write([]byte(`"ok"`))
	}))
	defer server.Close()

	svc := newTestService(t)
	body, err := svc.doWithRetry(context.Background(), server.URL, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `"ok"` {
		t.Errorf("body = %q, want %q", string(body), `"ok"`)
	}
	if got := hits.Load(); got != 3 {
		t.Errorf("server hit %d times, want 3", got)
	}
}

func TestDoWithRetry_StopsOnClientError(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(404)
	}))
	defer server.Close()

	svc := newTestService(t)
	_, err := svc.doWithRetry(context.Background(), server.URL, 3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("server hit %d times, want 1 (should not retry 404)", got)
	}
}

func TestDoWithRetry_ExhaustsRetries(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(500)
	}))
	defer server.Close()

	svc := newTestService(t)
	_, err := svc.doWithRetry(context.Background(), server.URL, 3)
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	if got := hits.Load(); got != 3 {
		t.Errorf("server hit %d times, want 3", got)
	}
}

func TestDoWithRetry_RespectsContextCancellation(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(500)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	svc := newTestService(t)
	_, err := svc.doWithRetry(ctx, server.URL, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should have been cancelled before all 10 attempts
	if got := hits.Load(); got >= 10 {
		t.Errorf("server hit %d times, expected fewer than 10 due to context cancellation", got)
	}
}

// --- Fetch function tests ---

func TestFetchStoryIDs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[10,20,30]`))
	}))
	defer server.Close()
	overrideBaseURL(t, server.URL)

	svc := newTestService(t)
	ids, err := svc.FetchStoryIDs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 || ids[0] != 10 || ids[1] != 20 || ids[2] != 30 {
		t.Errorf("ids = %v, want [10 20 30]", ids)
	}
}

func TestFetchStoryIDs_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()
	overrideBaseURL(t, server.URL)

	svc := newTestService(t)
	_, err := svc.FetchStoryIDs(context.Background())
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("error = %q, want to contain 'parse error'", err.Error())
	}
}

func TestFetchStory_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":42,"by":"alice","kids":[100,101],"title":"Test Story"}`))
	}))
	defer server.Close()
	overrideItemURL(t, server.URL)

	svc := newTestService(t)
	story, err := svc.FetchStory(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if story.ID != 42 {
		t.Errorf("story.ID = %d, want 42", story.ID)
	}
	if story.CreatedBy != "alice" {
		t.Errorf("story.CreatedBy = %q, want 'alice'", story.CreatedBy)
	}
	if story.Title != "Test Story" {
		t.Errorf("story.Title = %q, want 'Test Story'", story.Title)
	}
	if len(story.Kids) != 2 {
		t.Errorf("len(story.Kids) = %d, want 2", len(story.Kids))
	}
}

func TestFetchComment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":100,"by":"bob","kids":[200],"parent":42}`))
	}))
	defer server.Close()
	overrideItemURL(t, server.URL)

	svc := newTestService(t)
	comment, err := svc.FetchComment(context.Background(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.ID != 100 {
		t.Errorf("comment.ID = %d, want 100", comment.ID)
	}
	if comment.CreatedBy != "bob" {
		t.Errorf("comment.CreatedBy = %q, want 'bob'", comment.CreatedBy)
	}
	if comment.IsDeleted {
		t.Error("comment.IsDeleted = true, want false")
	}
}

func TestFetchComment_DeletedComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"deleted":true,"id":100,"parent":42}`))
	}))
	defer server.Close()
	overrideItemURL(t, server.URL)

	svc := newTestService(t)
	comment, err := svc.FetchComment(context.Background(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !comment.IsDeleted {
		t.Error("comment.IsDeleted = false, want true")
	}
}

func TestFetchStory_SemaphoreLimit(t *testing.T) {
	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := concurrent.Add(1)
		// Track max concurrency
		for {
			old := maxConcurrent.Load()
			if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		concurrent.Add(-1)
		w.Write([]byte(fmt.Sprintf(`{"id":1,"by":"alice","title":"Test"}`)))
	}))
	defer server.Close()
	overrideItemURL(t, server.URL)

	// Semaphore of 1 — requests should be serialized
	svc := New(10, repo.NewStoryRepo(), repo.NewUserRepo(), 1, &http.Client{})

	done := make(chan struct{})
	go func() {
		for i := range 3 {
			svc.FetchStory(context.Background(), i)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for fetches to complete")
	}

	if got := maxConcurrent.Load(); got > 1 {
		t.Errorf("max concurrent requests = %d, want 1 (semaphore should serialize)", got)
	}
}
