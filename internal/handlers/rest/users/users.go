package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"usersservice/internal/handlers"
	"usersservice/internal/handlers/rest/utils"
	"usersservice/internal/models/domain"
	"usersservice/internal/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	log     *zap.Logger
	service handlers.Service
}

func New(log *zap.Logger, service handlers.Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.rest.users.GetUsers"
	log := h.log.With(zap.String("op", op))

	paginationData := utils.GetPagination(r)

	notes, err := h.service.GetUsers(r.Context(), paginationData.Offset, paginationData.Limit)
	if err != nil {
		log.Error("failed to get users", zap.Error(err))
		http.Error(w, "failed to get users", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(notes); err != nil {
		log.Error("failed to encode users response", zap.Error(err))
		http.Error(w, "failed to encode users response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.rest.users.GetUserById"
	log := h.log.With(zap.String("op", op))

	userIds := r.URL.Query().Get("id")
	if userIds == "" {
		http.Error(w, "missing user ID", http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(userIds)
	if err != nil {
		log.Error("invalid user ID format", zap.String("userId", userIds), zap.Error(err))
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserById(r.Context(), userId)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Warn("user not found", zap.String("userId", userId.String()))
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		log.Error("failed to get user by ID", zap.Error(err))
		http.Error(w, "failed to get user by ID", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Error("failed to encode user response", zap.Error(err))
		http.Error(w, "failed to encode user response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Insert(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.rest.users.Insert"
	log := h.log.With(zap.String("op", op))

	request := struct {
		User domain.User `json:"user"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Insert(r.Context(), request.User); err != nil {
		if errors.Is(err, service.ErrInvalidArgument) {
			log.Warn("invalid user data", zap.Any("user", request.User), zap.Error(err))
			http.Error(w, "invalid user data", http.StatusConflict)
			return
		}
		log.Error("failed to insert user", zap.Error(err))
		http.Error(w, "failed to insert user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.rest.users.ChangePassword"
	log := h.log.With(zap.String("op", op))

	userIds := r.URL.Query().Get("id")
	if userIds == "" {
		http.Error(w, "missing user ID", http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(userIds)
	if err != nil {
		log.Error("invalid user ID format", zap.String("userId", userIds), zap.Error(err))
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
		return
	}

	request := struct {
		NewPassword string `json:"newPassword"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ChangePassword(r.Context(), userId, request.NewPassword); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Warn("user not found", zap.String("userId", userId.String()))
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidArgument) {
			log.Warn("invalid password", zap.String("userId", userId.String()), zap.Error(err))
			http.Error(w, "invalid password", http.StatusBadRequest)
			return
		}
		log.Error("failed to change password", zap.Error(err))
		http.Error(w, "failed to change password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.rest.users.Update"
	log := h.log.With(zap.String("op", op))

	userIds := r.URL.Query().Get("id")
	if userIds == "" {
		http.Error(w, "missing user ID", http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(userIds)
	if err != nil {
		log.Error("invalid user ID format", zap.String("userId", userIds), zap.Error(err))
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
		return
	}

	request := struct {
		User domain.User `json:"user"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Update(r.Context(), userId, request.User); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Warn("user not found for update", zap.String("userId", userId.String()))
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidArgument) {
			log.Warn("invalid user data for update", zap.Any("user", request.User), zap.Error(err))
			http.Error(w, "invalid user data", http.StatusBadRequest)
			return
		}
		log.Error("failed to update user", zap.Error(err))
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.rest.users.Delete"
	log := h.log.With(zap.String("op", op))

	userIds := r.URL.Query().Get("id")
	if userIds == "" {
		http.Error(w, "missing user ID", http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(userIds)
	if err != nil {
		log.Error("invalid user ID format", zap.String("userId", userIds), zap.Error(err))
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), userId); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Warn("user not found for deletion", zap.String("userId", userId.String()))
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Error("failed to delete user", zap.Error(err))
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
