package attachment

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/konfa-chat/hub/pkg/uuid"
)

// HTTPHandler is an HTTP handler for serving attachments
type HTTPHandler struct {
	storage Storage
}

// NewHTTPHandler creates a new HTTP handler for serving attachments
func NewHTTPHandler(storage Storage) *HTTPHandler {
	return &HTTPHandler{
		storage: storage,
	}
}

// ServeHTTP handles HTTP requests for attachments
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract attachment ID from the URL path
	// Expected format: /attachments/{uuid}[/filename]
	path := r.URL.Path
	// Skip the "/attachments/" prefix
	if len(path) <= 13 {
		http.Error(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	// Get the attachment ID
	remainingPath := path[13:]
	parts := strings.SplitN(remainingPath, "/", 2)
	idStr := parts[0]

	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	// Get metadata for the attachment to retrieve the original filename
	info, err := h.storage.GetInfo(r.Context(), id)
	if err != nil {
		// If we can't get metadata, continue with the download anyway
		// but we won't have the original filename
		info = AttachmentInfo{ID: id}
	}

	// Use URL filename if provided, otherwise use the original filename from storage
	filename := info.Filename
	if len(parts) > 1 && parts[1] != "" {
		filename = parts[1]
	}

	// Get the attachment from storage
	data, err := h.storage.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Attachment not found", http.StatusNotFound)
		return
	}
	defer data.Close()

	// Set appropriate Content-Type header based on filename extension
	contentType := "application/octet-stream"
	if filename != "" {
		if ext := filepath.Ext(filename); ext != "" {
			if mimeType := mime.TypeByExtension(ext); mimeType != "" {
				contentType = mimeType
			}
		}
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", contentType)
	if filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	} else {
		w.Header().Set("Content-Disposition", "attachment")
	}

	// Stream the attachment data to the response
	if _, err := io.Copy(w, data); err != nil {
		// Just log the error, since we've already started sending the response
		fmt.Printf("Error sending attachment: %v\n", err)
	}
}
