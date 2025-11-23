package users

import (
	"encoding/json/v2"
	"errors"
	"net/http"

	"github.com/lezzercringe/avito-test-assignment/internal/api"
	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
)

type Handler struct {
	svc usecases.UserService
}

func (h *Handler) InjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/users/setIsActive", h.setIsActive)
	mux.HandleFunc("/users/getReview", h.getReview)
	mux.HandleFunc("/users/deactivate", h.deactivate)
}

func (h *Handler) setIsActive(w http.ResponseWriter, r *http.Request) {
	type DTO struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	type responseDTO struct {
		User struct {
			ID       string `json:"user_id"`
			Name     string `json:"username"`
			TeamName string `json:"team_name"`
			IsActive bool   `json:"is_active"`
		} `json:"user"`
	}

	var dto DTO
	if err := json.UnmarshalRead(r.Body, &dto); err != nil {
		api.BadRequest(w)
		return
	}

	res, err := h.svc.SetIsActive(r.Context(), dto.UserID, dto.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrNotFound):
			api.Error(w, http.StatusNotFound, api.CodeNotFound, "resource not found")
		default:
			api.InternalServerError(w)
		}
		return
	}

	api.RespondJSON(w, responseDTO{
		User: struct {
			ID       string `json:"user_id"`
			Name     string `json:"username"`
			TeamName string `json:"team_name"`
			IsActive bool   `json:"is_active"`
		}{
			ID:       res.ID,
			Name:     res.Name,
			TeamName: res.TeamName,
			IsActive: res.IsActive,
		},
	})

}

func (h *Handler) getReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		api.BadRequest(w)
		return
	}

	type responseDTO struct {
		UserID       string `json:"user_id"`
		PullRequests []struct {
			ID       string `json:"pull_request_id"`
			Name     string `json:"pull_request_name"`
			AuthorID string `json:"author_id"`
			Status   string `json:"status"`
		} `json:"pull_requests"`
	}

	prs, err := h.svc.GetReview(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrNotFound):
			api.Error(w, http.StatusNotFound, api.CodeNotFound, "resource not found")
		default:
			api.InternalServerError(w)
		}
		return
	}

	pullRequests := make([]struct {
		ID       string `json:"pull_request_id"`
		Name     string `json:"pull_request_name"`
		AuthorID string `json:"author_id"`
		Status   string `json:"status"`
	}, len(prs))

	for i, pr := range prs {
		pullRequests[i] = struct {
			ID       string `json:"pull_request_id"`
			Name     string `json:"pull_request_name"`
			AuthorID string `json:"author_id"`
			Status   string `json:"status"`
		}{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   string(pr.Status),
		}
	}

	api.RespondJSON(w, responseDTO{
		UserID:       userID,
		PullRequests: pullRequests,
	})
}

func (h *Handler) deactivate(w http.ResponseWriter, r *http.Request) {
	type DTO struct {
		UserIDs []string `json:"user_ids"`
	}
	type responseDTO struct {
		Users []struct {
			ID       string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"users"`
	}

	var dto DTO
	if err := json.UnmarshalRead(r.Body, &dto); err != nil {
		api.BadRequest(w)
		return
	}

	if len(dto.UserIDs) == 0 {
		api.BadRequest(w)
		return
	}

	users, err := h.svc.Deactivate(r.Context(), dto.UserIDs...)
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrNotFound):
			api.Error(w, http.StatusNotFound, api.CodeNotFound, "resource not found")
		default:
			api.InternalServerError(w)
		}
		return
	}

	userDTOs := make([]struct {
		ID       string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	}, len(users))

	for i, user := range users {
		userDTOs[i] = struct {
			ID       string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}{
			ID:       user.ID,
			Username: user.Name,
			IsActive: user.IsActive,
		}
	}

	api.RespondJSON(w, responseDTO{
		Users: userDTOs,
	})
}

func NewHandler(svc usecases.UserService) *Handler {
	return &Handler{svc: svc}
}
