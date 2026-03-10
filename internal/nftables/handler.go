package nftables

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
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/enable", h.enable)
	r.Post("/{id}/disable", h.disable)
	r.Post("/sync", h.sync)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	groups, err := h.svc.GetAll()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, groups)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	group, err := h.svc.GetByID(id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "group not found")
		return
	}
	httputil.JSON(w, http.StatusOK, group)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	group, err := h.svc.Create(req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, group)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateGroupRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	group, err := h.svc.Update(id, req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, group)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, nil)
}

func (h *Handler) enable(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	group, err := h.svc.SetEnabled(id, true)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, group)
}

func (h *Handler) disable(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	group, err := h.svc.SetEnabled(id, false)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, group)
}

func (h *Handler) sync(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.SyncRules(); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, nil)
}

func (h *Handler) parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
