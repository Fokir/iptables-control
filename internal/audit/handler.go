package audit

import (
	"net/http"
	"strconv"

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
	r.Get("/", h.list)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	result, err := h.svc.List(ListParams{
		Page:       page,
		Limit:      limit,
		Action:     q.Get("action"),
		EntityType: q.Get("entityType"),
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, result)
}
