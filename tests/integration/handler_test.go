package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"
)

func parseIDFromResponse(t *testing.T, data []byte) string {
	t.Helper()
	var res struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("failed to parse upload response: %v (raw: %s)", err, string(data))
	}
	return res.ID
}

func uploadTestFileWithName(t *testing.T, filename string, content []byte) (string, int) {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write to part failed: %v", err)
	}
	w.Close()

	req, err := http.NewRequest("POST", "http://localhost:3000/API/upload", &b)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// return raw body for debugging along with status code
		return string(body), resp.StatusCode
	}

	id := parseIDFromResponse(t, body)
	return id, resp.StatusCode
}

func uploadTestFile(t *testing.T) string {
	id, _ := uploadTestFileWithName(t, "test.csv", []byte("name,email\nJohn Doe,john@example.com\nJane Doe,none\n"))
	// id contains the UUID
	return id
}

func downloadWithRetry(t *testing.T, id string) []byte {
	t.Helper()
	// small retry loop, matching previous style (wait for worker to complete)
	for i := 0; i < 30; i++ {
		resp, err := http.Get("http://localhost:3000/API/download/" + id)
		if err != nil {
			t.Fatalf("download request failed: %v", err)
		}
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusLocked {
			// still processing; wait a bit and retry
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return data
		}

		// any other status considered a failure for this test
		t.Fatalf("unexpected status from download: %d body: %s", resp.StatusCode, string(data))
	}
	t.Fatal("download failed after retries")
	return nil
}

func TestUploadAndDownloadFlow(t *testing.T) {
	id := uploadTestFile(t)
	output := downloadWithRetry(t, id)

	expected := "name,email,has_email\nJohn Doe,john@example.com,true\nJane Doe,none,false\n"
	if string(output) != expected {
		t.Errorf("downloaded data mismatch:\nGot:\n%s\nExpected:\n%s", output, expected)
	}
}

func TestUpload_MissingFile(t *testing.T) {
	// Send multipart request without "file" field
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// no file added
	w.Close()
	req, err := http.NewRequest("POST", "http://localhost:3000/API/upload", &b)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected bad request when missing file, got 200")
	}
}

func TestUpload_InvalidExtension(t *testing.T) {
	// Try uploading a .exe (invalid extension) and expect 400
	_, status := uploadTestFileWithName(t, "bad.exe", []byte("binary-content"))
	if status == http.StatusOK {
		t.Fatalf("expected upload to be rejected for invalid extension, got 200")
	}
}

func TestDownload_NotFound(t *testing.T) {
	// Attempt to download with a valid-looking UUID but that doesn't exist.
	// Using a known invalid ID string (but valid UUID format).
	invalidId := "11111111-1111-1111-1111-111111111111"
	resp, err := http.Get("http://localhost:3000/API/download/" + invalidId)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected not found or error for missing id, got 200")
	}
}
