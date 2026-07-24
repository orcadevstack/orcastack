package gatewayapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var communitySlugPattern = regexp.MustCompile(`[^a-z0-9-]+`)

type publicDeployment struct {
	ID                 string `json:"id"`
	ProjectSlug        string `json:"project_slug"`
	ProjectName        string `json:"project_name"`
	RepositoryName     string `json:"repository_name"`
	Environment        string `json:"environment"`
	TargetKind         string `json:"target_kind"`
	Status             string `json:"status"`
	BuildStatus        string `json:"build_status"`
	CommitSHA          string `json:"commit_sha"`
	UpdatedAt          string `json:"updated_at"`
	DashboardPath      string `json:"dashboard_path"`
	CanAccessDashboard bool   `json:"can_access_dashboard"`
}

type communityPost struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Kind           string `json:"kind"`
	Title          string `json:"title"`
	Excerpt        string `json:"excerpt"`
	Markdown       string `json:"markdown"`
	MediaURL       string `json:"media_url"`
	AuthorUsername string `json:"author_username"`
	Featured       bool   `json:"featured"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type communityContributor struct {
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	Contributions int  `json:"contributions"`
}

type communityEvent struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	StartsAt  string `json:"starts_at"`
	EndsAt    string `json:"ends_at"`
	Location  string `json:"location"`
	EventURL  string `json:"event_url"`
}

type publicHomeResponse struct {
	Deployments []publicDeployment `json:"deployments"`
	RecentBuilds []publicDeployment `json:"recent_builds"`
	Community   struct {
		Posts       int `json:"posts"`
		Contributors int `json:"contributors"`
		Events      int `json:"events"`
	} `json:"community"`
	ViewerRole string `json:"viewer_role"`
	UpdatedAt  string `json:"updated_at"`
}

type publicCommunityResponse struct {
	Posts        []communityPost        `json:"posts"`
	Contributors []communityContributor `json:"contributors"`
	Events       []communityEvent       `json:"events"`
	CanPublish   bool                   `json:"can_publish"`
	ViewerRole   string                 `json:"viewer_role"`
}

type createCommunityPostRequest struct {
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt"`
	Markdown string `json:"markdown"`
	MediaURL string `json:"media_url"`
}

func handlePublicHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	session, authenticated := optionalSession(r)
	response, err := loadPublicHome(r.Context(), session, authenticated)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, response)
}

func handlePublicCommunity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	session, authenticated := optionalSession(r)
	response, err := loadPublicCommunity(r.Context(), session, authenticated)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, response)
}

func handleCommunityPosts(w http.ResponseWriter, r *http.Request, session sessionRecord) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !canPublishCommunity(session) {
		writeError(w, http.StatusForbidden, errors.New("community publishing permission is required"))
		return
	}
	var request createCommunityPostRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid community post payload"))
		return
	}
	post, err := createCommunityPost(r.Context(), session, request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSONStatus(w, http.StatusCreated, post)
}

func handlePublicHubAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Action       string `json:"action"`
		ResourceType string `json:"resource_type"`
		ResourceID   string `json:"resource_id"`
	}
	if err := decodeJSON(r, &request); err != nil || request.Action != "open" {
		writeError(w, http.StatusBadRequest, errors.New("invalid public hub audit payload"))
		return
	}
	if request.ResourceType != "deployment" && request.ResourceType != "community-post" && request.ResourceType != "resource" {
		writeError(w, http.StatusBadRequest, errors.New("invalid public hub audit resource"))
		return
	}
	session, authenticated := optionalSession(r)
	actor, role := "anonymous", "viewer"
	if authenticated {
		actor, role = session.User.Username, publicViewerRole(session, true)
	}
	db, err := signupDB()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	if _, err := db.ExecContext(r.Context(), `insert into public_hub_audit_log (id, actor, role, action, resource_type, resource_id) values ($1, $2, $3, $4, $5, $6)`, uuid.NewString(), actor, role, request.Action, request.ResourceType, strings.TrimSpace(request.ResourceID)); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func optionalSession(r *http.Request) (sessionRecord, bool) {
	token := bearerToken(r)
	if token == "" {
		return sessionRecord{}, false
	}
	return lookupSession(token)
}

func publicViewerRole(session sessionRecord, authenticated bool) string {
	if !authenticated {
		return "viewer"
	}
	if isAdminSession(session) {
		return "admin"
	}
	if sessionHasPermission(session, "deployments:write") || sessionHasPermission(session, "pipelines:write") {
		return "developer"
	}
	return "viewer"
}

func canAccessDeploymentDashboard(session sessionRecord, authenticated bool) bool {
	return authenticated && (isAdminSession(session) || sessionHasPermission(session, "deployments:write"))
}

func canPublishCommunity(session sessionRecord) bool {
	return isAdminSession(session) || sessionHasPermission(session, "community:write")
}

func loadPublicHome(ctx context.Context, session sessionRecord, authenticated bool) (publicHomeResponse, error) {
	db, err := signupDB()
	if err != nil {
		return publicHomeResponse{}, err
	}
	canAccess := canAccessDeploymentDashboard(session, authenticated)
	rows, err := db.QueryContext(ctx, `
		select d.id::text, p.slug, p.name, r.name, d.environment, d.target_kind, d.status,
			pr.status, pr.commit_sha, coalesce(pr.finished_at, pr.started_at, d.created_at)
		from deployments d
		join pipeline_runs pr on pr.id = d.pipeline_run_id
		join repositories r on r.id = pr.repository_id
		join projects p on p.id = r.project_id
		where p.visibility = 'public' or $1
		order by coalesce(pr.finished_at, pr.started_at, d.created_at) desc
		limit 12
	`, authenticated)
	if err != nil {
		return publicHomeResponse{}, fmt.Errorf("load public deployments: %w", err)
	}
	defer rows.Close()
	deployments := make([]publicDeployment, 0)
	for rows.Next() {
		var item publicDeployment
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.ProjectSlug, &item.ProjectName, &item.RepositoryName, &item.Environment, &item.TargetKind, &item.Status, &item.BuildStatus, &item.CommitSHA, &updatedAt); err != nil {
			return publicHomeResponse{}, err
		}
		item.CanAccessDashboard = canAccess
		if canAccess {
			item.DashboardPath = "/app/deployments"
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		deployments = append(deployments, item)
	}
	if err := rows.Err(); err != nil {
		return publicHomeResponse{}, err
	}

	var response publicHomeResponse
	response.Deployments = deployments
	response.RecentBuilds = deployments
	response.ViewerRole = publicViewerRole(session, authenticated)
	response.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := db.QueryRowContext(ctx, `select count(*), count(distinct author_username) from community_posts where published = true and (visibility = 'public' or $1)`, authenticated).Scan(&response.Community.Posts, &response.Community.Contributors); err != nil {
		return publicHomeResponse{}, err
	}
	if err := db.QueryRowContext(ctx, `select count(*) from community_events where published = true and starts_at >= now()`).Scan(&response.Community.Events); err != nil {
		return publicHomeResponse{}, err
	}
	return response, nil
}

func loadPublicCommunity(ctx context.Context, session sessionRecord, authenticated bool) (publicCommunityResponse, error) {
	db, err := signupDB()
	if err != nil {
		return publicCommunityResponse{}, err
	}
	response := publicCommunityResponse{Posts: []communityPost{}, Contributors: []communityContributor{}, Events: []communityEvent{}, CanPublish: authenticated && canPublishCommunity(session), ViewerRole: publicViewerRole(session, authenticated)}
	rows, err := db.QueryContext(ctx, `select id::text, slug, kind, title, excerpt, markdown, media_url, author_username, featured, created_at, updated_at from community_posts where published = true and (visibility = 'public' or $1) order by featured desc, updated_at desc`, authenticated)
	if err != nil {
		return response, err
	}
	for rows.Next() {
		var post communityPost
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&post.ID, &post.Slug, &post.Kind, &post.Title, &post.Excerpt, &post.Markdown, &post.MediaURL, &post.AuthorUsername, &post.Featured, &createdAt, &updatedAt); err != nil {
			rows.Close()
			return response, err
		}
		post.CreatedAt, post.UpdatedAt = createdAt.UTC().Format(time.RFC3339), updatedAt.UTC().Format(time.RFC3339)
		response.Posts = append(response.Posts, post)
	}
	rows.Close()
	contributors, err := db.QueryContext(ctx, `select p.author_username, coalesce(u.display_name, p.author_username), count(*) from community_posts p left join users u on u.username = p.author_username where p.published = true group by p.author_username, u.display_name order by count(*) desc, p.author_username limit 12`)
	if err != nil {
		return response, err
	}
	for contributors.Next() {
		var contributor communityContributor
		if err := contributors.Scan(&contributor.Username, &contributor.DisplayName, &contributor.Contributions); err != nil {
			contributors.Close()
			return response, err
		}
		response.Contributors = append(response.Contributors, contributor)
	}
	contributors.Close()
	events, err := db.QueryContext(ctx, `select id::text, title, summary, starts_at, ends_at, location, event_url from community_events where published = true and starts_at >= now() order by starts_at limit 12`)
	if err != nil {
		return response, err
	}
	for events.Next() {
		var event communityEvent
		var startsAt time.Time
		var endsAt *time.Time
		if err := events.Scan(&event.ID, &event.Title, &event.Summary, &startsAt, &endsAt, &event.Location, &event.EventURL); err != nil {
			events.Close()
			return response, err
		}
		event.StartsAt = startsAt.UTC().Format(time.RFC3339)
		if endsAt != nil {
			event.EndsAt = endsAt.UTC().Format(time.RFC3339)
		}
		response.Events = append(response.Events, event)
	}
	events.Close()
	return response, nil
}

func createCommunityPost(ctx context.Context, session sessionRecord, request createCommunityPostRequest) (communityPost, error) {
	title, markdown := strings.TrimSpace(request.Title), strings.TrimSpace(request.Markdown)
	if title == "" || markdown == "" {
		return communityPost{}, errors.New("community post title and markdown are required")
	}
	kind := strings.ToLower(strings.TrimSpace(request.Kind))
	if kind != "discussion" && kind != "tutorial" && kind != "announcement" && kind != "project" {
		return communityPost{}, errors.New("invalid community post kind")
	}
	slug := strings.Trim(communitySlugPattern.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if slug == "" {
		return communityPost{}, errors.New("community post slug is required")
	}
	id, now := uuid.NewString(), time.Now().UTC()
	db, err := signupDB()
	if err != nil {
		return communityPost{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return communityPost{}, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `insert into community_posts (id, slug, kind, title, excerpt, markdown, media_url, author_username) values ($1, $2, $3, $4, $5, $6, $7, $8)`, id, slug, kind, title, strings.TrimSpace(request.Excerpt), markdown, strings.TrimSpace(request.MediaURL), session.User.Username)
	if err != nil {
		return communityPost{}, fmt.Errorf("create community post: %w", err)
	}
	_, err = tx.ExecContext(ctx, `insert into public_hub_audit_log (id, actor, role, action, resource_type, resource_id) values ($1, $2, $3, 'create', 'community-post', $4)`, uuid.NewString(), session.User.Username, publicViewerRole(session, true), id)
	if err != nil {
		return communityPost{}, err
	}
	if err := tx.Commit(); err != nil {
		return communityPost{}, err
	}
	return communityPost{ID: id, Slug: slug, Kind: kind, Title: title, Excerpt: strings.TrimSpace(request.Excerpt), Markdown: markdown, MediaURL: strings.TrimSpace(request.MediaURL), AuthorUsername: session.User.Username, CreatedAt: now.Format(time.RFC3339), UpdatedAt: now.Format(time.RFC3339)}, nil
}
