package gatewayapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var organizationSlugPattern = regexp.MustCompile(`[^a-z0-9-]+`)

type organizationSummary struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Website      string `json:"website"`
	Role         string `json:"role"`
	MemberCount  int    `json:"member_count"`
	TeamCount    int    `json:"team_count"`
	ProjectCount int    `json:"project_count"`
	CreatedAt    string `json:"created_at"`
}

type organizationMember struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type organizationTeam struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

type organizationProject struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	RepositoryName string `json:"repository_name"`
	TeamID         string `json:"team_id,omitempty"`
	TeamName       string `json:"team_name,omitempty"`
	BranchStrategy string `json:"branch_strategy"`
	DefaultBranch  string `json:"default_branch"`
	Status         string `json:"status"`
	CloneURL       string `json:"clone_url"`
	CreatedAt      string `json:"created_at"`
}

type organizationActivity struct {
	ID           string `json:"id"`
	Actor        string `json:"actor"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Summary      string `json:"summary"`
	OccurredAt   string `json:"occurred_at"`
}

type organizationDetail struct {
	Organization organizationSummary    `json:"organization"`
	Members      []organizationMember   `json:"members"`
	Teams        []organizationTeam     `json:"teams"`
	Projects     []organizationProject  `json:"projects"`
	Activity     []organizationActivity `json:"activity"`
}

type createOrganizationRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Website     string `json:"website"`
}

type createTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type addOrganizationMemberRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TeamID   string `json:"team_id"`
}

type createOrganizationProjectRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	RepositoryName string `json:"repository_name"`
	TeamID         string `json:"team_id"`
	BranchStrategy string `json:"branch_strategy"`
	DefaultBranch  string `json:"default_branch"`
}

func handleOrganizations(w http.ResponseWriter, r *http.Request, session sessionRecord) {
	switch r.Method {
	case http.MethodGet:
		organizations, err := listOrganizations(r.Context(), session)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, map[string]any{"organizations": organizations})
	case http.MethodPost:
		if !isAdminSession(session) {
			writeError(w, http.StatusForbidden, errors.New("platform administrator access is required to create an organization"))
			return
		}
		var request createOrganizationRequest
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid organization payload"))
			return
		}
		organization, err := createOrganization(r.Context(), session, request)
		if err != nil {
			writeOrganizationError(w, err)
			return
		}
		writeJSONStatus(w, http.StatusCreated, organization)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleOrganization(w http.ResponseWriter, r *http.Request, session sessionRecord) {
	pathParts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/organizations/"), "/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		writeError(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	slug := pathParts[0]
	if len(pathParts) == 1 && r.Method == http.MethodGet {
		detail, err := loadOrganizationDetail(r.Context(), session, slug)
		if err != nil {
			writeOrganizationError(w, err)
			return
		}
		writeJSON(w, detail)
		return
	}
	if len(pathParts) != 2 || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch pathParts[1] {
	case "teams":
		var request createTeamRequest
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid team payload"))
			return
		}
		team, err := createOrganizationTeam(r.Context(), session, slug, request)
		if err != nil {
			writeOrganizationError(w, err)
			return
		}
		writeJSONStatus(w, http.StatusCreated, team)
	case "members":
		var request addOrganizationMemberRequest
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid member payload"))
			return
		}
		member, err := addOrganizationMember(r.Context(), session, slug, request)
		if err != nil {
			writeOrganizationError(w, err)
			return
		}
		writeJSONStatus(w, http.StatusCreated, member)
	case "projects":
		var request createOrganizationProjectRequest
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid project payload"))
			return
		}
		project, err := createOrganizationProject(r.Context(), session, slug, request)
		if err != nil {
			writeOrganizationError(w, err)
			return
		}
		writeJSONStatus(w, http.StatusCreated, project)
	default:
		writeError(w, http.StatusNotFound, errors.New("organization operation not found"))
	}
}

func normalizeOrganizationSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = organizationSlugPattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func validateOrganizationRole(role string) error {
	switch role {
	case "owner", "maintainer", "developer":
		return nil
	default:
		return errors.New("organization role must be owner, maintainer, or developer")
	}
}

func validateBranchStrategy(strategy string) error {
	switch strategy {
	case "feature", "release", "hotfix":
		return nil
	default:
		return errors.New("branch strategy must be feature, release, or hotfix")
	}
}

func organizationRoleAllows(role, required string) bool {
	rank := map[string]int{"developer": 1, "maintainer": 2, "owner": 3}
	return rank[role] >= rank[required]
}

func organizationAccess(ctx context.Context, session sessionRecord, slug string) (string, string, error) {
	db, err := signupDB()
	if err != nil {
		return "", "", err
	}
	var organizationID, role string
	err = db.QueryRowContext(ctx, `
		select o.id, coalesce(m.role, '')
		from organizations o
		left join organization_members m on m.organization_id = o.id and m.username = $2
		where o.slug = $1
	`, slug, session.User.Username).Scan(&organizationID, &role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", errors.New("organization not found")
	}
	if err != nil {
		return "", "", fmt.Errorf("load organization access: %w", err)
	}
	if role == "" && isAdminSession(session) {
		role = "owner"
	}
	if role == "" {
		return "", "", errors.New("organization access is required")
	}
	return organizationID, role, nil
}

func listOrganizations(ctx context.Context, session sessionRecord) ([]organizationSummary, error) {
	db, err := signupDB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
		select o.id, o.slug, o.name, o.description, o.website, coalesce(m.role, ''), o.created_at,
			(select count(*) from organization_members om where om.organization_id = o.id),
			(select count(*) from organization_teams ot where ot.organization_id = o.id),
			(select count(*) from organization_projects op where op.organization_id = o.id)
		from organizations o
		left join organization_members m on m.organization_id = o.id and m.username = $1
		where $2 or m.id is not null
		order by o.name
	`, session.User.Username, isAdminSession(session))
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()
	organizations := make([]organizationSummary, 0)
	for rows.Next() {
		var organization organizationSummary
		var createdAt time.Time
		if err := rows.Scan(&organization.ID, &organization.Slug, &organization.Name, &organization.Description, &organization.Website, &organization.Role, &createdAt, &organization.MemberCount, &organization.TeamCount, &organization.ProjectCount); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		if organization.Role == "" && isAdminSession(session) {
			organization.Role = "owner"
		}
		organization.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		organizations = append(organizations, organization)
	}
	return organizations, rows.Err()
}

func createOrganization(ctx context.Context, session sessionRecord, request createOrganizationRequest) (organizationSummary, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return organizationSummary{}, errors.New("organization name is required")
	}
	slug := normalizeOrganizationSlug(request.Slug)
	if slug == "" {
		slug = normalizeOrganizationSlug(name)
	}
	if slug == "" {
		return organizationSummary{}, errors.New("organization slug is required")
	}
	db, err := signupDB()
	if err != nil {
		return organizationSummary{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return organizationSummary{}, err
	}
	defer tx.Rollback()
	organizationID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `insert into organizations (id, slug, name, description, website, created_by) values ($1, $2, $3, $4, $5, $6)`, organizationID, slug, name, strings.TrimSpace(request.Description), strings.TrimSpace(request.Website), session.User.Username)
	if err != nil {
		return organizationSummary{}, fmt.Errorf("create organization: %w", err)
	}
	_, err = tx.ExecContext(ctx, `insert into organization_members (id, organization_id, username, role) values ($1, $2, $3, 'owner')`, uuid.NewString(), organizationID, session.User.Username)
	if err != nil {
		return organizationSummary{}, fmt.Errorf("create organization owner: %w", err)
	}
	if err := insertOrganizationActivity(ctx, tx, organizationID, session.User.Username, "created", "organization", slug, "Created organization "+name); err != nil {
		return organizationSummary{}, err
	}
	if err := tx.Commit(); err != nil {
		return organizationSummary{}, err
	}
	return organizationSummary{ID: organizationID, Slug: slug, Name: name, Description: strings.TrimSpace(request.Description), Website: strings.TrimSpace(request.Website), Role: "owner", MemberCount: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

func createOrganizationTeam(ctx context.Context, session sessionRecord, slug string, request createTeamRequest) (organizationTeam, error) {
	organizationID, role, err := organizationAccess(ctx, session, slug)
	if err != nil {
		return organizationTeam{}, err
	}
	if !organizationRoleAllows(role, "maintainer") {
		return organizationTeam{}, errors.New("organization maintainer access is required")
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return organizationTeam{}, errors.New("team name is required")
	}
	teamSlug := normalizeOrganizationSlug(name)
	db, err := signupDB()
	if err != nil {
		return organizationTeam{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return organizationTeam{}, err
	}
	defer tx.Rollback()
	teamID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `insert into organization_teams (id, organization_id, slug, name, description, created_by) values ($1, $2, $3, $4, $5, $6)`, teamID, organizationID, teamSlug, name, strings.TrimSpace(request.Description), session.User.Username)
	if err != nil {
		return organizationTeam{}, fmt.Errorf("create organization team: %w", err)
	}
	if err := insertOrganizationActivity(ctx, tx, organizationID, session.User.Username, "created", "team", teamID, "Created team "+name); err != nil {
		return organizationTeam{}, err
	}
	if err := tx.Commit(); err != nil {
		return organizationTeam{}, err
	}
	return organizationTeam{ID: teamID, Slug: teamSlug, Name: name, Description: strings.TrimSpace(request.Description), CreatedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

func addOrganizationMember(ctx context.Context, session sessionRecord, slug string, request addOrganizationMemberRequest) (organizationMember, error) {
	organizationID, role, err := organizationAccess(ctx, session, slug)
	if err != nil {
		return organizationMember{}, err
	}
	if !organizationRoleAllows(role, "owner") {
		return organizationMember{}, errors.New("organization owner access is required")
	}
	username := strings.TrimSpace(request.Username)
	if username == "" {
		return organizationMember{}, errors.New("member username is required")
	}
	if err := validateOrganizationRole(request.Role); err != nil {
		return organizationMember{}, err
	}
	db, err := signupDB()
	if err != nil {
		return organizationMember{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return organizationMember{}, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `insert into organization_members (id, organization_id, username, role) values ($1, $2, $3, $4) on conflict (organization_id, username) do update set role = excluded.role`, uuid.NewString(), organizationID, username, request.Role)
	if err != nil {
		return organizationMember{}, fmt.Errorf("add organization member: %w", err)
	}
	if request.TeamID != "" {
		result, err := tx.ExecContext(ctx, `insert into organization_team_members (team_id, username) select id, $2 from organization_teams where id = $1 and organization_id = $3 on conflict do nothing`, request.TeamID, username, organizationID)
		if err != nil {
			return organizationMember{}, fmt.Errorf("assign team member: %w", err)
		}
		if rows, _ := result.RowsAffected(); rows == 0 {
			return organizationMember{}, errors.New("team not found in organization")
		}
	}
	if err := insertOrganizationActivity(ctx, tx, organizationID, session.User.Username, "assigned", "member", username, fmt.Sprintf("Assigned %s as %s", username, request.Role)); err != nil {
		return organizationMember{}, err
	}
	if err := tx.Commit(); err != nil {
		return organizationMember{}, err
	}
	return organizationMember{Username: username, Role: request.Role, CreatedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

func createOrganizationProject(ctx context.Context, session sessionRecord, slug string, request createOrganizationProjectRequest) (organizationProject, error) {
	organizationID, role, err := organizationAccess(ctx, session, slug)
	if err != nil {
		return organizationProject{}, err
	}
	if !organizationRoleAllows(role, "maintainer") {
		return organizationProject{}, errors.New("organization maintainer access is required")
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return organizationProject{}, errors.New("project name is required")
	}
	if err := validateBranchStrategy(request.BranchStrategy); err != nil {
		return organizationProject{}, err
	}
	repositoryName := sanitizeRepositoryName(request.RepositoryName)
	if repositoryName == "" {
		repositoryName = sanitizeRepositoryName(slug + "-" + name)
	}
	defaultBranch := strings.TrimSpace(request.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	db, err := signupDB()
	if err != nil {
		return organizationProject{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return organizationProject{}, err
	}
	defer tx.Rollback()
	projectID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `insert into organization_projects (id, organization_id, team_id, repository_name, name, description, branch_strategy, default_branch, created_by) values ($1, $2, nullif($3, '')::uuid, $4, $5, $6, $7, $8, $9)`, projectID, organizationID, request.TeamID, repositoryName, name, strings.TrimSpace(request.Description), request.BranchStrategy, defaultBranch, session.User.Username)
	if err != nil {
		return organizationProject{}, fmt.Errorf("create organization project: %w", err)
	}
	repository, err := createRepository(createRepositoryRequest{Name: repositoryName, Summary: strings.TrimSpace(request.Description), DefaultBranch: defaultBranch})
	if err != nil {
		return organizationProject{}, err
	}
	repositoryCreated := true
	defer func() {
		if repositoryCreated {
			_ = os.RemoveAll(repositoryDir(repositoryName))
		}
	}()
	_, err = tx.ExecContext(ctx, `update organization_projects set status = 'active' where id = $1`, projectID)
	if err != nil {
		return organizationProject{}, err
	}
	if err := insertOrganizationActivity(ctx, tx, organizationID, session.User.Username, "created", "project", projectID, "Created project "+name+" with "+request.BranchStrategy+" branching"); err != nil {
		return organizationProject{}, err
	}
	if err := tx.Commit(); err != nil {
		return organizationProject{}, err
	}
	repositoryCreated = false
	return organizationProject{ID: projectID, Name: name, Description: strings.TrimSpace(request.Description), RepositoryName: repositoryName, TeamID: request.TeamID, BranchStrategy: request.BranchStrategy, DefaultBranch: defaultBranch, Status: "active", CloneURL: repository.Repository.CloneURL, CreatedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

func loadOrganizationDetail(ctx context.Context, session sessionRecord, slug string) (organizationDetail, error) {
	organizationID, role, err := organizationAccess(ctx, session, slug)
	if err != nil {
		return organizationDetail{}, err
	}
	db, err := signupDB()
	if err != nil {
		return organizationDetail{}, err
	}
	var detail organizationDetail
	var createdAt time.Time
	err = db.QueryRowContext(ctx, `select id, slug, name, description, website, created_at from organizations where id = $1`, organizationID).Scan(&detail.Organization.ID, &detail.Organization.Slug, &detail.Organization.Name, &detail.Organization.Description, &detail.Organization.Website, &createdAt)
	if err != nil {
		return organizationDetail{}, err
	}
	detail.Organization.Role = role
	detail.Organization.CreatedAt = createdAt.UTC().Format(time.RFC3339)

	detail.Members, err = loadOrganizationMembers(ctx, db, organizationID)
	if err != nil {
		return organizationDetail{}, err
	}
	detail.Teams, err = loadOrganizationTeams(ctx, db, organizationID)
	if err != nil {
		return organizationDetail{}, err
	}
	detail.Projects, err = loadOrganizationProjects(ctx, db, organizationID)
	if err != nil {
		return organizationDetail{}, err
	}
	detail.Activity, err = loadOrganizationActivity(ctx, db, organizationID)
	if err != nil {
		return organizationDetail{}, err
	}
	detail.Organization.MemberCount = len(detail.Members)
	detail.Organization.TeamCount = len(detail.Teams)
	detail.Organization.ProjectCount = len(detail.Projects)
	return detail, nil
}

func loadOrganizationMembers(ctx context.Context, db *sql.DB, organizationID string) ([]organizationMember, error) {
	rows, err := db.QueryContext(ctx, `select username, role, created_at from organization_members where organization_id = $1 order by case role when 'owner' then 1 when 'maintainer' then 2 else 3 end, username`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := make([]organizationMember, 0)
	for rows.Next() {
		var member organizationMember
		var createdAt time.Time
		if err := rows.Scan(&member.Username, &member.Role, &createdAt); err != nil {
			return nil, err
		}
		member.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		members = append(members, member)
	}
	return members, rows.Err()
}

func loadOrganizationTeams(ctx context.Context, db *sql.DB, organizationID string) ([]organizationTeam, error) {
	rows, err := db.QueryContext(ctx, `select t.id, t.slug, t.name, t.description, t.created_at, (select count(*) from organization_team_members tm where tm.team_id = t.id) from organization_teams t where t.organization_id = $1 order by t.name`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teams := make([]organizationTeam, 0)
	for rows.Next() {
		var team organizationTeam
		var createdAt time.Time
		if err := rows.Scan(&team.ID, &team.Slug, &team.Name, &team.Description, &createdAt, &team.MemberCount); err != nil {
			return nil, err
		}
		team.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func loadOrganizationProjects(ctx context.Context, db *sql.DB, organizationID string) ([]organizationProject, error) {
	rows, err := db.QueryContext(ctx, `select p.id, p.name, p.description, p.repository_name, coalesce(p.team_id::text, ''), coalesce(t.name, ''), p.branch_strategy, p.default_branch, p.status, p.created_at from organization_projects p left join organization_teams t on t.id = p.team_id where p.organization_id = $1 order by p.name`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]organizationProject, 0)
	for rows.Next() {
		var project organizationProject
		var createdAt time.Time
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.RepositoryName, &project.TeamID, &project.TeamName, &project.BranchStrategy, &project.DefaultBranch, &project.Status, &createdAt); err != nil {
			return nil, err
		}
		project.CloneURL = repositoryCloneURL(project.RepositoryName)
		project.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func loadOrganizationActivity(ctx context.Context, db *sql.DB, organizationID string) ([]organizationActivity, error) {
	rows, err := db.QueryContext(ctx, `select id, actor, action, resource_type, resource_id, summary, occurred_at from organization_activity where organization_id = $1 order by occurred_at desc limit 50`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activity := make([]organizationActivity, 0)
	for rows.Next() {
		var event organizationActivity
		var occurredAt time.Time
		if err := rows.Scan(&event.ID, &event.Actor, &event.Action, &event.ResourceType, &event.ResourceID, &event.Summary, &occurredAt); err != nil {
			return nil, err
		}
		event.OccurredAt = occurredAt.UTC().Format(time.RFC3339)
		activity = append(activity, event)
	}
	return activity, rows.Err()
}

type organizationActivityExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func insertOrganizationActivity(ctx context.Context, executor organizationActivityExecutor, organizationID, actor, action, resourceType, resourceID, summary string) error {
	_, err := executor.ExecContext(ctx, `insert into organization_activity (id, organization_id, actor, action, resource_type, resource_id, summary) values ($1, $2, $3, $4, $5, $6, $7)`, uuid.NewString(), organizationID, actor, action, resourceType, resourceID, summary)
	if err != nil {
		return fmt.Errorf("record organization activity: %w", err)
	}
	return nil
}

func writeOrganizationError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := err.Error()
	switch {
	case strings.Contains(message, "required"), strings.Contains(message, "must be"):
		status = http.StatusBadRequest
	case strings.Contains(message, "access"):
		status = http.StatusForbidden
	case strings.Contains(message, "not found"):
		status = http.StatusNotFound
	case strings.Contains(message, "duplicate"), strings.Contains(message, "already exists"):
		status = http.StatusConflict
	}
	writeError(w, status, err)
}
