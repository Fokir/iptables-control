package settings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sokol/system-control/internal/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.get)
	r.Put("/", h.update)
	return r
}

type settingsDTO struct {
	ForceTLS12 bool `json:"forceTls12"`
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, settingsDTO{ForceTLS12: h.svc.ForceTLS12()})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req settingsDTO
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SetForceTLS12(req.ForceTLS12); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, settingsDTO{ForceTLS12: h.svc.ForceTLS12()})
}
