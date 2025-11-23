package prs

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"time"

	"github.com/lezzercringe/avito-test-assignment/internal/api"
	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
)

type Handler struct {
	svc usecases.PullRequestService
}

func (h *Handler) InjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/pullRequest/create", h.create)
	mux.HandleFunc("/pullRequest/merge", h.merge)
	mux.HandleFunc("/pullRequest/reassign", h.reassign)
}

type pullRequestDTO struct {
	ID        string    `json:"pull_request_id"`
	Name      string    `json:"pull_request_name"`
	AuthorID  string    `json:"author_id"`
	Status    string    `json:"status"`
	Reviewers []string  `json:"assigned_reviewers"`
	MergedAt  time.Time `json:"mergedAt,omitzero"`
}

func dtoFromView(dto *prs.PullRequestView) pullRequestDTO {
	return pullRequestDTO{
		ID:        dto.ID,
		Name:      dto.Name,
		AuthorID:  dto.AuthorID,
		Status:    dto.Status,
		Reviewers: dto.ReviewerIDs,
		MergedAt:  dto.MergedAt,
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	type DTO struct {
		ID       string `json:"pull_request_id"`
		Name     string `json:"pull_request_name"`
		AuthorID string `json:"author_id"`
	}
	type responseDTO struct {
		PR pullRequestDTO `json:"pr"`
	}

	var dto DTO
	if err := json.UnmarshalRead(r.Body, &dto); err != nil {
		api.BadRequest(w)
		return
	}

	res, err := h.svc.Create(r.Context(), usecases.CreateRequest{
		AuthorID: dto.AuthorID,
		Name:     dto.Name,
		ID:       dto.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrNotFound):
			api.Error(w, http.StatusNotFound, api.CodeNotFound, "resource not found")
		case errors.Is(err, errorsx.ErrAlreadyExists):
			api.Error(w, http.StatusConflict, api.CodePRExists, "PR id already exists")
		default:
			api.InternalServerError(w)
		}
		return
	}

	api.RespondJSON(w, responseDTO{
		PR: dtoFromView(res),
	})
}

func (h *Handler) merge(w http.ResponseWriter, r *http.Request) {
	type DTO struct {
		ID string `json:"pull_request_id"`
	}
	type responseDTO struct {
		PR pullRequestDTO `json:"pr"`
	}

	var dto DTO
	if err := json.UnmarshalRead(r.Body, &dto); err != nil {
		api.BadRequest(w)
		return
	}

	res, err := h.svc.Merge(r.Context(), dto.ID)
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrNotFound):
			api.Error(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		default:
			api.InternalServerError(w)
		}
		return
	}

	api.RespondJSON(w, responseDTO{
		PR: dtoFromView(res),
	})
}

func (h *Handler) reassign(w http.ResponseWriter, r *http.Request) {
	type DTO struct {
		ID            string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}
	type responseDTO struct {
		PR         pullRequestDTO `json:"pr"`
		ReplacedBy string         `json:"replaced_by"`
	}

	var dto DTO
	if err := json.UnmarshalRead(r.Body, &dto); err != nil {
		api.BadRequest(w)
		return
	}

	res, err := h.svc.ReassignReviewer(r.Context(), usecases.ReassignReviewerRequest{
		PullRequestID:    dto.ID,
		UserIDToReassign: dto.OldReviewerID,
	})
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrNotFound):
			api.Error(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		case errors.Is(err, errorsx.ErrModifyMergedPR):
			api.Error(w, http.StatusNotFound, "PR_MERGED", "cannot reassign on merged PR")
		default:
			api.InternalServerError(w)
		}
		return
	}

	api.RespondJSON(w, responseDTO{
		PR:         dtoFromView(res.PR),
		ReplacedBy: res.ReplacedByID,
	})
}

func NewHandler(svc usecases.PullRequestService) *Handler {
	return &Handler{svc: svc}
}
