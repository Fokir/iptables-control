package traffic

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
	r.Get("/nodes", h.allNodes)
	r.Get("/nodes/{id}", h.nodeTraffic)
	r.Get("/interfaces", h.allInterfaces)
	r.Get("/interfaces/{name}", h.interfaceTraffic)
	return r
}

func (h *Handler) allNodes(w http.ResponseWriter, r *http.Request) {
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "5m"
	}
	series, err := h.svc.GetAllNodeTraffic(interval)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, series)
}

func (h *Handler) nodeTraffic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "5m"
	}

	series, err := h.svc.GetNodeTraffic(id, interval)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, series)
}

func (h *Handler) allInterfaces(w http.ResponseWriter, r *http.Request) {
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "5m"
	}
	series, err := h.svc.GetAllInterfaceTraffic(interval)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, series)
}

func (h *Handler) interfaceTraffic(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "5m"
	}

	series, err := h.svc.GetInterfaceTraffic(name, interval)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, series)
}
