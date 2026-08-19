package httpapi

import (
	"errors"
	"net/http"
	"nhalinh/be/internal/memorial"
	"strings"
)

type MemorialHandler struct{ service *memorial.Service }

func NewMemorialHandler(s *memorial.Service) *MemorialHandler { return &MemorialHandler{service: s} }

type housePayload struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Notes   string `json:"notes"`
	Active  *bool  `json:"active"`
}
type areaPayload struct {
	HouseID string `json:"house_id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Notes   string `json:"notes"`
}
type positionPayload struct {
	AreaID       string `json:"area_id"`
	RowNumber    int    `json:"row_number"`
	ColumnNumber int    `json:"column_number"`
	Notes        string `json:"notes"`
}
type positionsBatchPayload struct {
	AreaID    string            `json:"area_id"`
	Positions []positionPayload `json:"positions"`
}
type tabletPayload struct {
	PositionID string          `json:"position_id"`
	Name       string          `json:"name"`
	Notes      string          `json:"notes"`
	Spirits    []spiritPayload `json:"spirits"`
}
type spiritPayload struct {
	ID          string `json:"id"`
	HouseID     string `json:"house_id"`
	TabletID    string `json:"tablet_id"`
	FullName    string `json:"full_name"`
	DharmaName  string `json:"dharma_name"`
	BirthYear   string `json:"birth_year"`
	DeathYear   string `json:"death_year"`
	Age         string `json:"age"`
	ImageURL    string `json:"image_url"`
	BurialPlace string `json:"burial_place"`
	Sender      string `json:"sender"`
	SentMonth   string `json:"sent_month"`
	Notes       string `json:"notes"`
}
type spiritsBatchPayload struct {
	Spirits []spiritPayload `json:"spirits"`
}
type spiritPatchPayload struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

func actor(r *http.Request) memorial.Actor {
	u := currentUser(r.Context())
	return memorial.Actor{ID: u.ID, Role: u.Role, AllHouses: u.AllHouses}
}
func (h *MemorialHandler) Houses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, e := h.service.ListHouses(r.Context(), actor(r))
		h.write(w, v, e, http.StatusOK)
	case http.MethodPost:
		var p housePayload
		if !decode(w, r, &p) {
			return
		}
		v, e := h.service.CreateHouse(r.Context(), actor(r), memorial.HouseInput{Name: p.Name, Address: p.Address, Notes: p.Notes})
		h.write(w, v, e, http.StatusCreated)
	default:
		methodNotAllowed(w)
	}
}
func (h *MemorialHandler) House(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/spirit-houses/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 1 || parts[0] == "" {
		writeError(w, 404, "not_found", "Không tìm thấy Nhà Linh")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var p housePayload
		if !decode(w, r, &p) {
			return
		}
		active := true
		if p.Active != nil {
			active = *p.Active
		}
		v, e := h.service.UpdateHouse(r.Context(), actor(r), parts[0], memorial.HouseInput{Name: p.Name, Address: p.Address, Notes: p.Notes, Active: active})
		h.write(w, v, e, 200)
	case http.MethodDelete:
		e := h.service.DeleteHouse(r.Context(), actor(r), parts[0])
		h.write(w, nil, e, 204)
	default:
		methodNotAllowed(w)
	}
}
func (h *MemorialHandler) Areas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, e := h.service.ListAreas(r.Context(), actor(r), r.URL.Query().Get("house_id"))
		h.write(w, map[string]any{"areas": v}, e, 200)
	case http.MethodPost:
		var p areaPayload
		if !decode(w, r, &p) {
			return
		}
		v, e := h.service.CreateArea(r.Context(), actor(r), memorial.AreaInput{HouseID: p.HouseID, Code: p.Code, Name: p.Name, Notes: p.Notes})
		h.write(w, v, e, 201)
	default:
		methodNotAllowed(w)
	}
}
func (h *MemorialHandler) Positions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if houseID := r.URL.Query().Get("house_id"); houseID != "" {
			limit, _, ok := readPaginationWithMax(w, r, 100)
			if !ok {
				return
			}
			v, e := h.service.SearchPositions(r.Context(), actor(r), memorial.PositionSearchOptions{HouseID: houseID, Query: r.URL.Query().Get("q"), Limit: limit})
			h.write(w, map[string]any{"positions": v}, e, 200)
			return
		}
		v, e := h.service.ListPositions(r.Context(), actor(r), r.URL.Query().Get("area_id"))
		h.write(w, map[string]any{"positions": v}, e, 200)
	case http.MethodPost:
		var p positionPayload
		if !decode(w, r, &p) {
			return
		}
		v, e := h.service.CreatePosition(r.Context(), actor(r), memorial.PositionInput{AreaID: p.AreaID, RowNumber: p.RowNumber, ColumnNumber: p.ColumnNumber, Notes: p.Notes})
		h.write(w, v, e, 201)
	default:
		methodNotAllowed(w)
	}
}
func (h *MemorialHandler) Position(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/memorial-positions/"), "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, 404, "not_found", "Không tìm thấy vị trí")
		return
	}
	if r.Method != http.MethodPut {
		methodNotAllowed(w)
		return
	}
	var p positionPayload
	if !decode(w, r, &p) {
		return
	}
	v, e := h.service.UpdatePosition(r.Context(), actor(r), id, memorial.PositionInput{AreaID: p.AreaID, RowNumber: p.RowNumber, ColumnNumber: p.ColumnNumber, Notes: p.Notes})
	h.write(w, v, e, 200)
}
func (h *MemorialHandler) PositionsBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var p positionsBatchPayload
	if !decode(w, r, &p) {
		return
	}
	inputs := make([]memorial.PositionInput, 0, len(p.Positions))
	for _, position := range p.Positions {
		inputs = append(inputs, memorial.PositionInput{AreaID: p.AreaID, RowNumber: position.RowNumber, ColumnNumber: position.ColumnNumber, Notes: position.Notes})
	}
	v, e := h.service.CreatePositions(r.Context(), actor(r), inputs)
	h.write(w, map[string]any{"positions": v, "skipped_count": len(inputs) - len(v)}, e, http.StatusCreated)
}
func (h *MemorialHandler) Occupancy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	v, e := h.service.Occupancy(r.Context(), actor(r), r.URL.Query().Get("house_id"))
	h.write(w, v, e, http.StatusOK)
}
func (h *MemorialHandler) Tablets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, e := h.service.ListTablets(r.Context(), actor(r), r.URL.Query().Get("position_id"))
		h.write(w, map[string]any{"tablets": v}, e, 200)
	case http.MethodPost:
		var p tabletPayload
		if !decode(w, r, &p) {
			return
		}
		spirits := make([]memorial.SpiritInput, 0, len(p.Spirits))
		for _, spirit := range p.Spirits {
			spirits = append(spirits, spiritInput(spirit))
		}
		v, e := h.service.CreateTablet(r.Context(), actor(r), memorial.TabletInput{PositionID: p.PositionID, Name: p.Name, Notes: p.Notes, Spirits: spirits})
		h.write(w, v, e, 201)
	default:
		methodNotAllowed(w)
	}
}
func (h *MemorialHandler) Tablet(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/memorial-tablets/"), "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, 404, "not_found", "Không tìm thấy bài vị")
		return
	}
	if r.Method != http.MethodPut {
		methodNotAllowed(w)
		return
	}
	var p tabletPayload
	if !decode(w, r, &p) {
		return
	}
	spirits := make([]memorial.SpiritInput, 0, len(p.Spirits))
	for _, spirit := range p.Spirits {
		spirits = append(spirits, spiritInput(spirit))
	}
	v, e := h.service.UpdateTablet(r.Context(), actor(r), id, memorial.TabletInput{PositionID: p.PositionID, Name: p.Name, Notes: p.Notes, Spirits: spirits})
	h.write(w, v, e, 200)
}
func (h *MemorialHandler) Spirits(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit, offset, ok := readPaginationWithMax(w, r, 500)
		if !ok {
			return
		}
		v, total, e := h.service.ListSpirits(r.Context(), actor(r), memorial.SearchOptions{Query: r.URL.Query().Get("q"), HouseID: r.URL.Query().Get("house_id"), AreaID: r.URL.Query().Get("area_id"), PositionID: r.URL.Query().Get("position_id"), TabletID: r.URL.Query().Get("tablet_id"), Limit: limit, Offset: offset})
		h.write(w, map[string]any{"spirits": v, "total": total, "has_more": offset+len(v) < total}, e, 200)
	case http.MethodPost:
		var p spiritPayload
		if !decode(w, r, &p) {
			return
		}
		v, e := h.service.CreateSpirit(r.Context(), actor(r), spiritInput(p))
		h.write(w, v, e, 201)
	default:
		methodNotAllowed(w)
	}
}
func (h *MemorialHandler) SpiritsBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var p spiritsBatchPayload
	if !decode(w, r, &p) {
		return
	}
	inputs := make([]memorial.SpiritInput, 0, len(p.Spirits))
	for _, spirit := range p.Spirits {
		inputs = append(inputs, spiritInput(spirit))
	}
	v, e := h.service.CreateSpirits(r.Context(), actor(r), inputs)
	h.write(w, map[string]any{"spirits": v}, e, http.StatusCreated)
}
func (h *MemorialHandler) Spirit(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/spirits/"), "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, 404, "not_found", "Không tìm thấy Hương linh")
		return
	}
	switch r.Method {
	case http.MethodGet:
		v, e := h.service.GetSpirit(r.Context(), actor(r), id)
		h.write(w, v, e, 200)
	case http.MethodPut:
		var p spiritPayload
		if !decode(w, r, &p) {
			return
		}
		v, e := h.service.UpdateSpirit(r.Context(), actor(r), id, spiritInput(p))
		h.write(w, v, e, 200)
	case http.MethodPatch:
		var p spiritPatchPayload
		if !decode(w, r, &p) {
			return
		}
		v, e := h.service.PatchSpirit(r.Context(), actor(r), id, p.Field, p.Value)
		h.write(w, v, e, 200)
	case http.MethodDelete:
		e := h.service.DeleteSpirit(r.Context(), actor(r), id)
		h.write(w, nil, e, 204)
	default:
		methodNotAllowed(w)
	}
}
func spiritInput(p spiritPayload) memorial.SpiritInput {
	return memorial.SpiritInput{ID: p.ID, HouseID: p.HouseID, TabletID: p.TabletID, FullName: p.FullName, DharmaName: p.DharmaName, BirthYear: p.BirthYear, DeathYear: p.DeathYear, Age: p.Age, ImageURL: p.ImageURL, BurialPlace: p.BurialPlace, Sender: p.Sender, SentMonth: p.SentMonth, Notes: p.Notes}
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if e := readJSON(r, v); e != nil {
		writeError(w, 400, "invalid_json", "request body must be valid json")
		return false
	}
	return true
}
func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, 405, "method_not_allowed", "method is not allowed")
}
func (h *MemorialHandler) write(w http.ResponseWriter, v any, e error, status int) {
	if e == nil {
		if status == 204 {
			w.WriteHeader(status)
		} else {
			writeJSON(w, status, v)
		}
		return
	}
	switch {
	case errors.Is(e, memorial.ErrInvalidInput):
		writeError(w, 400, "invalid_input", e.Error())
	case errors.Is(e, memorial.ErrForbidden):
		writeError(w, 403, "forbidden", "Bạn không có quyền với Nhà Linh này")
	case errors.Is(e, memorial.ErrNotFound):
		writeError(w, 404, "not_found", "Không tìm thấy dữ liệu")
	case errors.Is(e, memorial.ErrConflict):
		writeError(w, 409, "conflict", "Mã hoặc vị trí đã tồn tại")
	default:
		writeError(w, 500, "internal_error", "internal server error")
	}
}
