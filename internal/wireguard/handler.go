package wireguard

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
	r.Get("/status", h.status)
	r.Post("/enable", h.enable)
	r.Post("/disable", h.disable)
	r.Get("/peers", h.listPeers)
	r.Post("/peers", h.createPeer)
	r.Get("/peers/{id}", h.getPeer)
	r.Put("/peers/{id}", h.updatePeer)
	r.Delete("/peers/{id}", h.deletePeer)
	r.Get("/peers/{id}/config", h.getPeerConfig)
	r.Get("/peers/{id}/qrcode", h.getPeerQRCode)
	r.Post("/sync", h.syncPeers)
	return r
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.GetStatus()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, st)
}

func (h *Handler) enable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint   string `json:"endpoint"`
		ListenPort int    `json:"listenPort"`
		Address    string `json:"address"`
	}
	if err := httputil.Decode(r, &req); err != nil {
		// Allow empty body for already-configured setups
		req.Endpoint = ""
	}

	if err := h.svc.Enable(req.Endpoint, req.ListenPort, req.Address); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	st, _ := h.svc.GetStatus()
	httputil.JSON(w, http.StatusOK, st)
}

func (h *Handler) disable(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Disable(); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	st, _ := h.svc.GetStatus()
	httputil.JSON(w, http.StatusOK, st)
}

func (h *Handler) listPeers(w http.ResponseWriter, r *http.Request) {
	peers, err := h.svc.GetPeers()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, peers)
}

func (h *Handler) createPeer(w http.ResponseWriter, r *http.Request) {
	var req CreatePeerRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	peer, err := h.svc.CreatePeer(req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, peer)
}

func (h *Handler) getPeer(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	peer, err := h.svc.GetPeer(id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "peer not found")
		return
	}
	httputil.JSON(w, http.StatusOK, peer)
}

func (h *Handler) updatePeer(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdatePeerRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	peer, err := h.svc.UpdatePeer(id, req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, peer)
}

func (h *Handler) deletePeer(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.DeletePeer(id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, nil)
}

func (h *Handler) getPeerConfig(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	pc, err := h.svc.GetPeerConfig(id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, pc)
}

func (h *Handler) getPeerQRCode(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	qr, err := h.svc.GetPeerQRCode(id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(qr)))
	w.WriteHeader(http.StatusOK)
	w.Write(qr)
}

func (h *Handler) syncPeers(w http.ResponseWriter, r *http.Request) {
	count, err := h.svc.SyncPeers()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]int{"imported": count})
}

func (h *Handler) parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
