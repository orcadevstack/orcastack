package runnerapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type blockingRunner struct {
	release chan struct{}
}

func (runner blockingRunner) Run(ctx context.Context, workflow string) (string, error) {
	select {
	case <-runner.release:
		return "ok", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type recordingRunner struct {
	done chan struct{}
}

func (runner recordingRunner) Run(_ context.Context, workflow string) (string, error) {
	close(runner.done)
	return "runner output for " + workflow, nil
}

func TestCreatePipelineCanBeRetrieved(t *testing.T) {
	mux := http.NewServeMux()
	runner := blockingRunner{release: make(chan struct{})}
	registerWithRunner(mux, runner)
	t.Cleanup(func() { close(runner.release) })

	payload := []byte(`{"project":"orcastack/platform","ref":"main","trigger":"manual"}`)
	request := httptest.NewRequest(http.MethodPost, "/pipelines", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("create pipeline status = %d, want %d", response.Code, http.StatusAccepted)
	}

	var created Pipeline
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatalf("decode created pipeline: %v", err)
	}
	if created.ID == "" {
		t.Fatal("created pipeline ID is empty")
	}
	if created.Status != "queued" {
		t.Fatalf("created pipeline status = %q, want queued", created.Status)
	}

	request = httptest.NewRequest(http.MethodGet, "/pipelines/"+created.ID, nil)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("get pipeline status = %d, want %d", response.Code, http.StatusOK)
	}

	var retrieved Pipeline
	if err := json.NewDecoder(response.Body).Decode(&retrieved); err != nil {
		t.Fatalf("decode retrieved pipeline: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Fatalf("retrieved pipeline ID = %q, want %q", retrieved.ID, created.ID)
	}
}

func TestPipelineExecutionCompletesAndPublishesLogs(t *testing.T) {
	mux := http.NewServeMux()
	runner := recordingRunner{done: make(chan struct{})}
	registerWithRunner(mux, runner)

	payload := []byte(`{"project":"orcastack/platform","ref":"main","trigger":"manual","workflow":"api-test"}`)
	request := httptest.NewRequest(http.MethodPost, "/pipelines", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("create pipeline status = %d, want %d", response.Code, http.StatusAccepted)
	}

	var created Pipeline
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatalf("decode created pipeline: %v", err)
	}
	<-runner.done

	request = httptest.NewRequest(http.MethodGet, "/pipelines/"+created.ID, nil)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	var completed Pipeline
	if err := json.NewDecoder(response.Body).Decode(&completed); err != nil {
		t.Fatalf("decode completed pipeline: %v", err)
	}
	if completed.Status != "success" {
		t.Fatalf("completed pipeline status = %q, want success", completed.Status)
	}
	if len(completed.Jobs) != 1 || completed.Jobs[0].FinishedAt == "" {
		t.Fatalf("completed pipeline job = %#v, want one finished job", completed.Jobs)
	}

	request = httptest.NewRequest(http.MethodGet, completed.Jobs[0].LogEndpoint, nil)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("get logs status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "runner output for api-test" {
		t.Fatalf("logs = %q, want runner output", response.Body.String())
	}
}

func TestCreatePipelineRejectsUnsupportedWorkflow(t *testing.T) {
	mux := http.NewServeMux()
	registerWithRunner(mux, recordingRunner{done: make(chan struct{})})

	payload := []byte(`{"project":"orcastack/platform","ref":"main","workflow":"arbitrary-shell"}`)
	request := httptest.NewRequest(http.MethodPost, "/pipelines", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("create unsupported workflow status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
