package teams

import (
	"encoding/json/v2"
	"errors"
	"net/http"

	"github.com/lezzercringe/avito-test-assignment/internal/api"
	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
)

type Handler struct {
	svc usecases.TeamService
}

func (h *Handler) InjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/team/add", h.add)
	mux.HandleFunc("/team/get", h.get)
}

func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	type DTO struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	}
	type responseDTO struct {
		Team struct {
			TeamName string `json:"team_name"`
			Members  []struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				IsActive bool   `json:"is_active"`
			} `json:"members"`
		} `json:"team"`
	}

	var dto DTO
	if err := json.UnmarshalRead(r.Body, &dto); err != nil {
		api.BadRequest(w)
		return
	}

	req := usecases.TeamView{
		Name:    dto.TeamName,
		Members: make([]usecases.TeamMemberView, len(dto.Members)),
	}

	for i, member := range dto.Members {
		req.Members[i] = usecases.TeamMemberView{
			ID:     member.UserID,
			Name:   member.Username,
			Active: member.IsActive,
		}
	}

	res, err := h.svc.AddTeam(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrAlreadyExists):
			api.Error(w, http.StatusConflict, api.CodeTeamExists, "team_name already exists")
		default:
			api.InternalServerError(w)
		}
		return
	}

	members := make([]struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	}, len(res.Members))

	for i, member := range res.Members {
		members[i] = struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}{
			UserID:   member.ID,
			Username: member.Name,
			IsActive: member.Active,
		}
	}

	api.RespondJSON(w, responseDTO{
		Team: struct {
			TeamName string `json:"team_name"`
			Members  []struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				IsActive bool   `json:"is_active"`
			} `json:"members"`
		}{
			TeamName: res.Name,
			Members:  members,
		},
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		api.BadRequest(w)
		return
	}

	type responseDTO struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	}

	res, err := h.svc.GetTeam(r.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, errorsx.ErrNotFound):
			api.Error(w, http.StatusNotFound, api.CodeNotFound, "resource not found")
		default:
			api.InternalServerError(w)
		}
		return
	}

	members := make([]struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	}, len(res.Members))

	for i, member := range res.Members {
		members[i] = struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}{
			UserID:   member.ID,
			Username: member.Name,
			IsActive: member.Active,
		}
	}

	api.RespondJSON(w, responseDTO{
		TeamName: res.Name,
		Members:  members,
	})
}

func NewHandler(svc usecases.TeamService) *Handler {
	return &Handler{svc: svc}
}
