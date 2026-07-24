package gatewayapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var headerSections = []string{"projects", "deployments", "profile"}

type headerNavigationItem struct {
	ID          string `json:"id"`
	Section     string `json:"section"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Icon        string `json:"icon"`
	Action      string `json:"action"`
	Count       *int   `json:"count,omitempty"`
}

type headerNavigationResponse struct {
	Organization string                 `json:"organization"`
	Tagline      string                 `json:"tagline"`
	Items        []headerNavigationItem `json:"items"`
	UpdatedAt    string                 `json:"updated_at"`
}

type headerAuditRequest struct {
	Action     string `json:"action"`
	Section    string `json:"section"`
	TargetPath string `json:"target_path"`
}

func handleHeaderNavigation(w http.ResponseWriter, r *http.Request, session sessionRecord) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	items, err := loadHeaderNavigation(r, session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, headerNavigationResponse{
		Organization: "OrcaStack",
		Tagline:      "Build. Secure. Ship. Operate.",
		Items:        items,
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	})
}

func loadHeaderNavigation(r *http.Request, session sessionRecord) ([]headerNavigationItem, error) {
	db, err := signupDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(r.Context(), `
		select id, section, label, description, path, icon, action, required_permission
		from dashboard_header_items
		where enabled = true
			and section in ('projects', 'deployments', 'profile')
		order by case section
			when 'projects' then 1
			when 'deployments' then 2
			when 'profile' then 3
			else 4
		end, sort_order
	`)
	if err != nil {
		return nil, fmt.Errorf("load dashboard header navigation: %w", err)
	}
	defer rows.Close()

	overview, err := overviewData()
	if err != nil {
		return nil, err
	}

	counts := map[string]int{
		"projects-overview":        len(overview.Repositories),
		"projects-reviews":         len(overview.Reviews),
		"deployments-pipelines":    len(overview.Pipelines),
		"deployments-environments": len(overview.Deployments),
	}
	items := make([]headerNavigationItem, 0)
	for rows.Next() {
		var item headerNavigationItem
		var requiredPermission string
		if err := rows.Scan(&item.ID, &item.Section, &item.Label, &item.Description, &item.Path, &item.Icon, &item.Action, &requiredPermission); err != nil {
			return nil, fmt.Errorf("scan dashboard header navigation: %w", err)
		}
		if !sessionHasPermission(session, requiredPermission) {
			continue
		}
		if count, ok := counts[item.ID]; ok {
			item.Count = &count
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read dashboard header navigation: %w", err)
	}
	return items, nil
}

func handleHeaderAudit(w http.ResponseWriter, r *http.Request, session sessionRecord) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request headerAuditRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid header audit payload"))
		return
	}
	if err := validateHeaderAudit(request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	db, err := signupDB()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	_, err = db.ExecContext(r.Context(), `
		insert into dashboard_header_audit_log (id, username, role, action, section, target_path)
		values ($1, $2, $3, $4, $5, $6)
	`, uuid.NewString(), session.User.Username, session.User.Role, request.Action, request.Section, request.TargetPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("record dashboard header audit: %w", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateHeaderAudit(request headerAuditRequest) error {
	if request.Action != "open" && request.Action != "navigate" {
		return errors.New("header audit action must be open or navigate")
	}
	sectionValid := false
	for _, section := range headerSections {
		if request.Section == section {
			sectionValid = true
			break
		}
	}
	if !sectionValid {
		return errors.New("invalid header audit section")
	}
	if request.TargetPath != "" && !strings.HasPrefix(request.TargetPath, "/app/") {
		return errors.New("header audit target must be an application path")
	}
	return nil
}

func sessionHasPermission(session sessionRecord, required string) bool {
	if required == "" {
		return true
	}
	for _, permission := range session.User.Permissions {
		if permission == required || (permission == "control-panel:admin" && required == "control-panel:read") {
			return true
		}
	}
	return false
}
