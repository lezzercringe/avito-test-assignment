package api

import (
	"encoding/json/v2"
	"net/http"

	"go.uber.org/zap"
)

type ErrorCode string

const (
	CodeNotFound    ErrorCode = "NOT_FOUND"
	CodeTeamExists  ErrorCode = "TEAM_EXISTS"
	CodePRExists    ErrorCode = "PR_EXISTS"
	CodePRMerged    ErrorCode = "PR_MERGED"
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"
	CodeBadRequest  ErrorCode = "BAD_REQUEST"
)

func RespondJSON(w http.ResponseWriter, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		zap.L().Error("failed to encode json response",
			zap.Error(err), zap.Any("response", v))
	}
	w.Write(body)
}

type errorDTO struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type errorResponse struct {
	Error errorDTO `json:"error"`
}

func Error(
	w http.ResponseWriter,
	code int,
	errcode ErrorCode,
	msg string,
) {
	w.WriteHeader(code)

	resp := errorResponse{
		Error: errorDTO{
			Message: msg,
			Code:    string(errcode),
		},
	}

	RespondJSON(w, resp)
}

func BadRequest(w http.ResponseWriter) {
	Error(w, http.StatusBadRequest, CodeBadRequest, "bad request")
}

func InternalServerError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error", "INTERAL_SERVER_ERROR")
}
