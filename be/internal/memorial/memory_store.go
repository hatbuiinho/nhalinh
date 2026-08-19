package memorial

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type MemoryStore struct {
	mu        sync.RWMutex
	houses    map[string]House
	members   map[string]map[string]string
	areas     map[string]Area
	positions map[string]Position
	tablets   map[string]Tablet
	spirits   map[string]Spirit
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{houses: map[string]House{}, members: map[string]map[string]string{}, areas: map[string]Area{}, positions: map[string]Position{}, tablets: map[string]Tablet{}, spirits: map[string]Spirit{}}
}
func (s *MemoryStore) ListHouses(_ context.Context, a Actor) ([]House, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []House{}
	for _, v := range s.houses {
		r := a.Role
		if a.Role != "admin" && !a.AllHouses {
			if s.members[v.ID][a.ID] == "" {
				continue
			}
		}
		v.AccessRole = r
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func (s *MemoryStore) CreateHouse(_ context.Context, v House) (House, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.houses[v.ID] = v
	return v, nil
}
func (s *MemoryStore) UpdateHouse(_ context.Context, v House) (House, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.houses[v.ID]
	if !ok {
		return House{}, ErrNotFound
	}
	v.CreatedAt = old.CreatedAt
	s.houses[v.ID] = v
	return v, nil
}
func (s *MemoryStore) DeleteHouse(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.houses[id]; !ok {
		return ErrNotFound
	}
	delete(s.houses, id)
	return nil
}
func (s *MemoryStore) AccessRole(_ context.Context, a Actor, h string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.houses[h]; !ok {
		return "", ErrNotFound
	}
	if a.Role == "admin" || a.AllHouses {
		return a.Role, nil
	}
	r := s.members[h][a.ID]
	if r == "" {
		return "", ErrForbidden
	}
	return a.Role, nil
}
func (s *MemoryStore) ListAreas(_ context.Context, _ Actor, h string) ([]Area, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Area{}
	for _, v := range s.areas {
		if v.HouseID != h {
			continue
		}
		for _, p := range s.positions {
			if p.AreaID != v.ID {
				continue
			}
			v.PositionCount++
			for _, t := range s.tablets {
				if t.PositionID != p.ID {
					continue
				}
				v.TabletCount++
				for _, sp := range s.spirits {
					if sp.TabletID == t.ID {
						v.SpiritCount++
					}
				}
			}
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}
func (s *MemoryStore) CreateArea(_ context.Context, v Area) (Area, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.areas {
		if x.HouseID == v.HouseID && x.Code == v.Code {
			return Area{}, ErrConflict
		}
	}
	s.areas[v.ID] = v
	return v, nil
}
func (s *MemoryStore) AreaCode(_ context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.areas[id]
	if !ok {
		return "", ErrNotFound
	}
	return v.Code, nil
}
func (s *MemoryStore) ListPositions(_ context.Context, _ Actor, areaID string) ([]Position, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Position{}
	for _, v := range s.positions {
		if v.AreaID != areaID {
			continue
		}
		a := s.areas[areaID]
		h := s.houses[a.HouseID]
		v.AreaCode = a.Code
		v.HouseID = h.ID
		v.HouseName = h.Name
		for _, t := range s.tablets {
			if t.PositionID == v.ID {
				v.TabletCount++
				for _, spirit := range s.spirits {
					if spirit.TabletID == t.ID {
						v.SpiritCount++
					}
				}
			}
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RowNumber == out[j].RowNumber {
			return out[i].ColumnNumber < out[j].ColumnNumber
		}
		return out[i].RowNumber < out[j].RowNumber
	})
	return out, nil
}
func (s *MemoryStore) SearchPositions(_ context.Context, _ Actor, o PositionSearchOptions) ([]Position, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Position{}
	query := fold(o.Query)
	for _, v := range s.positions {
		a := s.areas[v.AreaID]
		if a.HouseID != o.HouseID {
			continue
		}
		h := s.houses[a.HouseID]
		v.AreaCode = a.Code
		v.HouseID = h.ID
		v.HouseName = h.Name
		for _, tablet := range s.tablets {
			if tablet.PositionID == v.ID {
				v.TabletCount++
				for _, spirit := range s.spirits {
					if spirit.TabletID == tablet.ID {
						v.SpiritCount++
					}
				}
			}
		}
		haystack := fold(fmt.Sprintf("%s %s %s %s %d %d", v.Name, v.Notes, v.AreaCode, v.HouseName, v.RowNumber, v.ColumnNumber))
		if query == "" || strings.Contains(haystack, query) {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		leftExact, rightExact := fold(out[i].Name) == query, fold(out[j].Name) == query
		if leftExact != rightExact {
			return leftExact
		}
		if out[i].RowNumber == out[j].RowNumber {
			return out[i].ColumnNumber < out[j].ColumnNumber
		}
		return out[i].RowNumber < out[j].RowNumber
	})
	if len(out) > o.Limit {
		out = out[:o.Limit]
	}
	return out, nil
}
func (s *MemoryStore) ListOccupancyPositions(_ context.Context, _ Actor, houseID string) ([]Position, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Position{}
	unplaced := 0
	for _, spirit := range s.spirits {
		if spirit.HouseID == houseID && spirit.TabletID == "" {
			unplaced++
		}
	}
	for _, v := range s.positions {
		a := s.areas[v.AreaID]
		if a.HouseID != houseID {
			continue
		}
		h := s.houses[houseID]
		v.AreaCode = a.Code
		v.HouseID = h.ID
		v.HouseName = h.Name
		for _, tablet := range s.tablets {
			if tablet.PositionID != v.ID {
				continue
			}
			v.TabletCount++
			for _, spirit := range s.spirits {
				if spirit.TabletID == tablet.ID {
					v.SpiritCount++
				}
			}
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].AreaCode != out[j].AreaCode {
			return out[i].AreaCode < out[j].AreaCode
		}
		if out[i].RowNumber != out[j].RowNumber {
			return out[i].RowNumber < out[j].RowNumber
		}
		return out[i].ColumnNumber < out[j].ColumnNumber
	})
	return out, unplaced, nil
}
func (s *MemoryStore) CreatePosition(_ context.Context, v Position) (Position, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.positions {
		if x.AreaID == v.AreaID && (x.Name == v.Name || (x.RowNumber == v.RowNumber && x.ColumnNumber == v.ColumnNumber)) {
			return Position{}, ErrConflict
		}
	}
	s.positions[v.ID] = v
	return v, nil
}
func (s *MemoryStore) CreatePositions(_ context.Context, positions []Position) ([]Position, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	created := make([]Position, 0, len(positions))
	for _, v := range positions {
		duplicate := false
		for _, existing := range s.positions {
			if existing.AreaID == v.AreaID && (existing.Name == v.Name || (existing.RowNumber == v.RowNumber && existing.ColumnNumber == v.ColumnNumber)) {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		s.positions[v.ID] = v
		created = append(created, v)
	}
	return created, nil
}
func (s *MemoryStore) UpdatePosition(_ context.Context, v Position) (Position, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.positions[v.ID]
	if !ok {
		return Position{}, ErrNotFound
	}
	for id, x := range s.positions {
		if id != v.ID && x.AreaID == v.AreaID && (x.Name == v.Name || (x.RowNumber == v.RowNumber && x.ColumnNumber == v.ColumnNumber)) {
			return Position{}, ErrConflict
		}
	}
	v.CreatedAt = old.CreatedAt
	s.positions[v.ID] = v
	return v, nil
}
func (s *MemoryStore) ListTablets(_ context.Context, _ Actor, positionID string) ([]Tablet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Tablet{}
	for _, v := range s.tablets {
		if v.PositionID != positionID {
			continue
		}
		p := s.positions[positionID]
		a := s.areas[p.AreaID]
		h := s.houses[a.HouseID]
		v.HouseID = h.ID
		v.HouseName = h.Name
		v.AreaID = a.ID
		v.AreaCode = a.Code
		v.PositionName = p.Name
		v.RowNumber = p.RowNumber
		v.ColumnNumber = p.ColumnNumber
		for _, sp := range s.spirits {
			if sp.TabletID == v.ID {
				v.SpiritCount++
			}
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func (s *MemoryStore) CreateTablet(_ context.Context, v Tablet) (Tablet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.tablets {
		if x.PositionID == v.PositionID && x.Name == v.Name {
			return Tablet{}, ErrConflict
		}
	}
	s.tablets[v.ID] = v
	return v, nil
}
func (s *MemoryStore) CreateTabletWithSpirits(_ context.Context, v Tablet, spirits []Spirit) (Tablet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, x := range s.tablets {
		if x.PositionID == v.PositionID && x.Name == v.Name {
			return Tablet{}, ErrConflict
		}
	}
	s.tablets[v.ID] = v
	for _, spirit := range spirits {
		s.spirits[spirit.ID] = spirit
	}
	v.SpiritCount = len(spirits)
	return v, nil
}
func (s *MemoryStore) UpdateTabletWithSpirits(_ context.Context, v Tablet, spirits []Spirit) (Tablet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.tablets[v.ID]
	if !ok {
		return Tablet{}, ErrNotFound
	}
	for id, tablet := range s.tablets {
		if id != v.ID && tablet.PositionID == v.PositionID && tablet.Name == v.Name {
			return Tablet{}, ErrConflict
		}
	}
	for _, spirit := range spirits {
		if existing, found := s.spirits[spirit.ID]; found {
			if existing.TabletID != v.ID {
				return Tablet{}, fmt.Errorf("%w: spirit does not belong to tablet", ErrInvalidInput)
			}
			spirit.CreatedAt = existing.CreatedAt
		} else if spirit.CreatedAt.IsZero() {
			return Tablet{}, fmt.Errorf("%w: spirit does not belong to tablet", ErrInvalidInput)
		}
	}
	v.CreatedAt = old.CreatedAt
	s.tablets[v.ID] = v
	kept := make(map[string]bool, len(spirits))
	for _, spirit := range spirits {
		s.spirits[spirit.ID] = spirit
		kept[spirit.ID] = true
	}
	for id, spirit := range s.spirits {
		if spirit.TabletID == v.ID && !kept[id] {
			delete(s.spirits, id)
		}
	}
	v.SpiritCount = len(spirits)
	return v, nil
}
func (s *MemoryStore) ListSpirits(_ context.Context, a Actor, o SearchOptions) ([]Spirit, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Spirit{}
	q := fold(o.Query)
	for _, v := range s.spirits {
		v = s.join(v)
		if a.Role != "admin" && !a.AllHouses && s.members[v.HouseID][a.ID] == "" {
			continue
		}
		if o.HouseID != "" && v.HouseID != o.HouseID || o.AreaID != "" && v.AreaID != o.AreaID || o.PositionID != "" && v.PositionID != o.PositionID || o.TabletID != "" && v.TabletID != o.TabletID {
			continue
		}
		hay := fold(strings.Join([]string{v.FullName, v.DharmaName, v.BirthYear, v.DeathYear, v.Age, v.BurialPlace, v.Sender, v.SentMonth, v.Notes, v.PositionName, v.TabletName, v.AreaCode, v.HouseName}, " "))
		if q != "" && !strings.Contains(hay, q) {
			continue
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FullName < out[j].FullName })
	total := len(out)
	start := min(o.Offset, total)
	end := min(start+o.Limit, total)
	return out[start:end], total, nil
}
func (s *MemoryStore) GetSpirit(_ context.Context, a Actor, id string) (Spirit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.spirits[id]
	if !ok {
		return Spirit{}, ErrNotFound
	}
	v = s.join(v)
	if a.Role != "admin" && !a.AllHouses && s.members[v.HouseID][a.ID] == "" {
		return Spirit{}, ErrForbidden
	}
	return v, nil
}
func (s *MemoryStore) CreateSpirit(_ context.Context, v Spirit) (Spirit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spirits[v.ID] = v
	return v, nil
}
func (s *MemoryStore) CreateSpirits(_ context.Context, items []Spirit) ([]Spirit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		s.spirits[item.ID] = item
	}
	return items, nil
}
func (s *MemoryStore) UpdateSpirit(_ context.Context, v Spirit) (Spirit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.spirits[v.ID]
	if !ok {
		return Spirit{}, ErrNotFound
	}
	v.CreatedAt = old.CreatedAt
	s.spirits[v.ID] = v
	return v, nil
}
func (s *MemoryStore) PatchSpirit(_ context.Context, id, field, value string, updatedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.spirits[id]
	if !ok {
		return ErrNotFound
	}
	switch field {
	case "full_name":
		v.FullName = value
	case "dharma_name":
		v.DharmaName = value
	case "birth_year":
		v.BirthYear = value
	case "death_year":
		v.DeathYear = value
	case "age":
		v.Age = value
	case "burial_place":
		v.BurialPlace = value
	case "sender":
		v.Sender = value
	case "sent_month":
		v.SentMonth = value
	case "notes":
		v.Notes = value
	default:
		return ErrInvalidInput
	}
	v.UpdatedAt = updatedAt
	s.spirits[id] = v
	return nil
}
func (s *MemoryStore) DeleteSpirit(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.spirits[id]; !ok {
		return ErrNotFound
	}
	delete(s.spirits, id)
	return nil
}
func (s *MemoryStore) HouseIDForArea(_ context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.areas[id]
	if !ok {
		return "", ErrNotFound
	}
	return v.HouseID, nil
}
func (s *MemoryStore) HouseIDForPosition(_ context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.positions[id]
	if !ok {
		return "", ErrNotFound
	}
	a, ok := s.areas[p.AreaID]
	if !ok {
		return "", ErrNotFound
	}
	return a.HouseID, nil
}
func (s *MemoryStore) HouseIDForTablet(_ context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tablets[id]
	if !ok {
		return "", ErrNotFound
	}
	p := s.positions[t.PositionID]
	return s.areas[p.AreaID].HouseID, nil
}
func (s *MemoryStore) HouseIDForSpirit(_ context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.spirits[id]
	if !ok {
		return "", ErrNotFound
	}
	if v.HouseID != "" {
		return v.HouseID, nil
	}
	t := s.tablets[v.TabletID]
	p := s.positions[t.PositionID]
	return s.areas[p.AreaID].HouseID, nil
}
func (s *MemoryStore) join(v Spirit) Spirit {
	if v.TabletID == "" {
		h := s.houses[v.HouseID]
		v.HouseName = h.Name
		return v
	}
	t := s.tablets[v.TabletID]
	p := s.positions[t.PositionID]
	a := s.areas[p.AreaID]
	h := s.houses[a.HouseID]
	v.TabletName = t.Name
	v.PositionID = p.ID
	v.PositionName = p.Name
	v.AreaID = a.ID
	v.AreaCode = a.Code
	v.HouseID = h.ID
	v.HouseName = h.Name
	return v
}
func fold(v string) string {
	value := strings.NewReplacer("đ", "d", "Đ", "D", "-", "").Replace(strings.TrimSpace(v))
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return unicode.ToLower(r)
	}, norm.NFD.String(value))
}
