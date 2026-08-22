package memorial

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service { return &Service{store: store, now: now} }

func (s *Service) ListHouses(ctx context.Context, actor Actor) ([]House, error) {
	return s.store.ListHouses(ctx, actor)
}
func (s *Service) CreateHouse(ctx context.Context, actor Actor, in HouseInput) (House, error) {
	if actor.Role != "admin" {
		return House{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return House{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	now := s.now().UTC()
	return s.store.CreateHouse(ctx, House{ID: newID("house"), Name: in.Name, Address: strings.TrimSpace(in.Address), Notes: strings.TrimSpace(in.Notes), Active: true, AccessRole: "admin", CreatedAt: now, UpdatedAt: now})
}
func (s *Service) UpdateHouse(ctx context.Context, actor Actor, id string, in HouseInput) (House, error) {
	if err := s.requireWrite(ctx, actor, id); err != nil {
		return House{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return House{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	return s.store.UpdateHouse(ctx, House{ID: id, Name: in.Name, Address: strings.TrimSpace(in.Address), Notes: strings.TrimSpace(in.Notes), Active: in.Active, UpdatedAt: s.now().UTC()})
}
func (s *Service) DeleteHouse(ctx context.Context, actor Actor, id string) error {
	if actor.Role != "admin" {
		return ErrForbidden
	}
	return s.store.DeleteHouse(ctx, id)
}
func (s *Service) ListAreas(ctx context.Context, actor Actor, houseID string) ([]Area, error) {
	if _, err := s.store.AccessRole(ctx, actor, houseID); err != nil {
		return nil, err
	}
	return s.store.ListAreas(ctx, actor, houseID)
}
func (s *Service) CreateArea(ctx context.Context, actor Actor, in AreaInput) (Area, error) {
	if err := s.requireWrite(ctx, actor, in.HouseID); err != nil {
		return Area{}, err
	}
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	if in.Code == "" {
		return Area{}, fmt.Errorf("%w: code is required", ErrInvalidInput)
	}
	now := s.now().UTC()
	return s.store.CreateArea(ctx, Area{ID: newID("area"), HouseID: in.HouseID, Code: in.Code, Name: strings.TrimSpace(in.Name), Notes: strings.TrimSpace(in.Notes), CreatedAt: now, UpdatedAt: now})
}
func (s *Service) ListPositions(ctx context.Context, actor Actor, areaID string) ([]Position, error) {
	house, err := s.store.HouseIDForArea(ctx, areaID)
	if err != nil {
		return nil, err
	}
	if _, err = s.store.AccessRole(ctx, actor, house); err != nil {
		return nil, err
	}
	return s.store.ListPositions(ctx, actor, areaID)
}
func (s *Service) SearchPositions(ctx context.Context, actor Actor, o PositionSearchOptions) ([]Position, error) {
	o.HouseID = strings.TrimSpace(o.HouseID)
	o.Query = strings.TrimSpace(o.Query)
	if o.HouseID == "" {
		return nil, fmt.Errorf("%w: house_id is required", ErrInvalidInput)
	}
	if o.Limit == 0 {
		o.Limit = 30
	}
	if o.Limit < 1 || o.Limit > 100 {
		return nil, fmt.Errorf("%w: limit must be between 1 and 100", ErrInvalidInput)
	}
	if _, err := s.store.AccessRole(ctx, actor, o.HouseID); err != nil {
		return nil, err
	}
	return s.store.SearchPositions(ctx, actor, o)
}
func (s *Service) Occupancy(ctx context.Context, actor Actor, houseID string) (Occupancy, error) {
	houseID = strings.TrimSpace(houseID)
	if houseID == "" {
		return Occupancy{}, fmt.Errorf("%w: house_id is required", ErrInvalidInput)
	}
	if _, err := s.store.AccessRole(ctx, actor, houseID); err != nil {
		return Occupancy{}, err
	}
	areas, err := s.store.ListAreas(ctx, actor, houseID)
	if err != nil {
		return Occupancy{}, err
	}
	positions, unplaced, err := s.store.ListOccupancyPositions(ctx, actor, houseID)
	if err != nil {
		return Occupancy{}, err
	}
	out := Occupancy{HouseID: houseID, Positions: positions}
	out.Summary.AreaCount = len(areas)
	out.Summary.UnplacedSpiritCount = unplaced
	areaIndexes := make(map[string]int, len(areas))
	for _, area := range areas {
		areaIndexes[area.ID] = len(out.Areas)
		out.Areas = append(out.Areas, OccupancyArea{ID: area.ID, Code: area.Code, Name: area.Name})
	}
	for _, position := range positions {
		out.Summary.PositionCount++
		out.Summary.TabletCount += position.TabletCount
		out.Summary.SpiritCount += position.SpiritCount
		if position.TabletCount == 0 {
			out.Summary.EmptyPositionCount++
		} else {
			out.Summary.UsedPositionCount++
		}
		areaIndex, ok := areaIndexes[position.AreaID]
		if !ok {
			continue
		}
		area := &out.Areas[areaIndex]
		area.PositionCount++
		area.TabletCount += position.TabletCount
		area.SpiritCount += position.SpiritCount
		if position.TabletCount == 0 {
			area.EmptyPositionCount++
		}
	}
	out.Summary.SpiritCount += unplaced
	return out, nil
}
func (s *Service) CreatePosition(ctx context.Context, actor Actor, in PositionInput) (Position, error) {
	positions, err := s.preparePositions(ctx, actor, []PositionInput{in})
	if err != nil {
		return Position{}, err
	}
	return s.store.CreatePosition(ctx, positions[0])
}
func (s *Service) CreatePositions(ctx context.Context, actor Actor, inputs []PositionInput) ([]Position, error) {
	positions, err := s.preparePositions(ctx, actor, inputs)
	if err != nil {
		return nil, err
	}
	return s.store.CreatePositions(ctx, positions)
}
func (s *Service) preparePositions(ctx context.Context, actor Actor, inputs []PositionInput) ([]Position, error) {
	if len(inputs) == 0 || len(inputs) > 500 {
		return nil, fmt.Errorf("%w: positions must contain between 1 and 500 rows", ErrInvalidInput)
	}
	areaID := strings.TrimSpace(inputs[0].AreaID)
	house, err := s.store.HouseIDForArea(ctx, areaID)
	if err != nil {
		return nil, err
	}
	if err = s.requireWrite(ctx, actor, house); err != nil {
		return nil, err
	}
	areaCode, err := s.store.AreaCode(ctx, areaID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	positions := make([]Position, 0, len(inputs))
	for index, in := range inputs {
		if strings.TrimSpace(in.AreaID) != areaID {
			return nil, fmt.Errorf("%w: row %d belongs to another area", ErrInvalidInput, index+1)
		}
		if in.RowNumber < 1 || in.ColumnNumber < 1 {
			return nil, fmt.Errorf("%w: row %d: row_number and column_number must be positive", ErrInvalidInput, index+1)
		}
		name := fmt.Sprintf("%d%s-%d", in.ColumnNumber, strings.ToUpper(areaCode), in.RowNumber)
		positions = append(positions, Position{ID: newID("position"), AreaID: areaID, Name: name, RowNumber: in.RowNumber, ColumnNumber: in.ColumnNumber, Notes: strings.TrimSpace(in.Notes), CreatedAt: now, UpdatedAt: now})
	}
	return positions, nil
}
func (s *Service) UpdatePosition(ctx context.Context, actor Actor, id string, in PositionInput) (Position, error) {
	house, err := s.store.HouseIDForPosition(ctx, id)
	if err != nil {
		return Position{}, err
	}
	if err = s.requireWrite(ctx, actor, house); err != nil {
		return Position{}, err
	}
	targetHouse, err := s.store.HouseIDForArea(ctx, in.AreaID)
	if err != nil {
		return Position{}, err
	}
	if targetHouse != house {
		return Position{}, fmt.Errorf("%w: cannot move position to another house", ErrInvalidInput)
	}
	if in.RowNumber < 1 || in.ColumnNumber < 1 {
		return Position{}, fmt.Errorf("%w: row_number and column_number must be positive", ErrInvalidInput)
	}
	areaCode, err := s.store.AreaCode(ctx, in.AreaID)
	if err != nil {
		return Position{}, err
	}
	name := fmt.Sprintf("%d%s-%d", in.ColumnNumber, strings.ToUpper(areaCode), in.RowNumber)
	return s.store.UpdatePosition(ctx, Position{ID: id, AreaID: in.AreaID, Name: name, RowNumber: in.RowNumber, ColumnNumber: in.ColumnNumber, Notes: strings.TrimSpace(in.Notes), UpdatedAt: s.now().UTC()})
}
func (s *Service) ListTablets(ctx context.Context, actor Actor, positionID string) ([]Tablet, error) {
	house, err := s.store.HouseIDForPosition(ctx, positionID)
	if err != nil {
		return nil, err
	}
	if _, err = s.store.AccessRole(ctx, actor, house); err != nil {
		return nil, err
	}
	return s.store.ListTablets(ctx, actor, positionID)
}
func (s *Service) CreateTablet(ctx context.Context, actor Actor, in TabletInput) (Tablet, error) {
	house, err := s.store.HouseIDForPosition(ctx, in.PositionID)
	if err != nil {
		return Tablet{}, err
	}
	if err = s.requireWrite(ctx, actor, house); err != nil {
		return Tablet{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return Tablet{}, fmt.Errorf("%w: tablet name is required", ErrInvalidInput)
	}
	existingSpiritIDs := make([]string, 0, len(in.ExistingSpiritIDs))
	seenExistingIDs := make(map[string]bool, len(in.ExistingSpiritIDs))
	for _, rawID := range in.ExistingSpiritIDs {
		id := strings.TrimSpace(rawID)
		if id == "" || seenExistingIDs[id] {
			return Tablet{}, fmt.Errorf("%w: existing spirit ids must be unique and non-empty", ErrInvalidInput)
		}
		seenExistingIDs[id] = true
		existingSpiritIDs = append(existingSpiritIDs, id)
	}
	if len(in.Spirits)+len(existingSpiritIDs) == 0 || len(in.Spirits)+len(existingSpiritIDs) > 500 {
		return Tablet{}, fmt.Errorf("%w: a tablet must contain between 1 and 500 spirits", ErrInvalidInput)
	}
	now := s.now().UTC()
	tablet := Tablet{ID: newID("tablet"), PositionID: in.PositionID, Name: name, Notes: strings.TrimSpace(in.Notes), CreatedAt: now, UpdatedAt: now}
	spirits := make([]Spirit, 0, len(in.Spirits))
	for index, spiritInput := range in.Spirits {
		spiritInput.TabletID = tablet.ID
		spiritInput.HouseID = house
		spirit, normalizeErr := s.normalizeSpirit(spiritInput)
		if normalizeErr != nil {
			return Tablet{}, fmt.Errorf("%w: spirit row %d: %v", ErrInvalidInput, index+1, normalizeErr)
		}
		spirit.ID = newID("spirit")
		spirit.CreatedAt = now
		spirit.UpdatedAt = now
		spirits = append(spirits, spirit)
	}
	return s.store.CreateTabletWithSpirits(ctx, tablet, spirits, existingSpiritIDs, house)
}
func (s *Service) UpdateTablet(ctx context.Context, actor Actor, id string, in TabletInput) (Tablet, error) {
	house, err := s.store.HouseIDForTablet(ctx, id)
	if err != nil {
		return Tablet{}, err
	}
	if err = s.requireWrite(ctx, actor, house); err != nil {
		return Tablet{}, err
	}
	targetHouse, err := s.store.HouseIDForPosition(ctx, in.PositionID)
	if err != nil {
		return Tablet{}, err
	}
	if targetHouse != house {
		return Tablet{}, fmt.Errorf("%w: cannot move tablet to another house", ErrInvalidInput)
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return Tablet{}, fmt.Errorf("%w: tablet name is required", ErrInvalidInput)
	}
	if len(in.Spirits) == 0 || len(in.Spirits) > 500 {
		return Tablet{}, fmt.Errorf("%w: a tablet must contain between 1 and 500 spirits", ErrInvalidInput)
	}
	now := s.now().UTC()
	tablet := Tablet{ID: id, PositionID: in.PositionID, Name: name, Notes: strings.TrimSpace(in.Notes), UpdatedAt: now}
	spirits := make([]Spirit, 0, len(in.Spirits))
	seenIDs := make(map[string]bool, len(in.Spirits))
	for index, spiritInput := range in.Spirits {
		spiritInput.TabletID = id
		spiritInput.HouseID = house
		spirit, normalizeErr := s.normalizeSpirit(spiritInput)
		if normalizeErr != nil {
			return Tablet{}, fmt.Errorf("%w: spirit row %d: %v", ErrInvalidInput, index+1, normalizeErr)
		}
		spirit.ID = strings.TrimSpace(spiritInput.ID)
		if spirit.ID == "" {
			spirit.ID = newID("spirit")
			spirit.CreatedAt = now
		} else if seenIDs[spirit.ID] {
			return Tablet{}, fmt.Errorf("%w: duplicate spirit id on row %d", ErrInvalidInput, index+1)
		}
		seenIDs[spirit.ID] = true
		spirit.UpdatedAt = now
		spirits = append(spirits, spirit)
	}
	return s.store.UpdateTabletWithSpirits(ctx, tablet, spirits)
}
func (s *Service) ListSpirits(ctx context.Context, actor Actor, o SearchOptions) ([]Spirit, int, error) {
	o.Query = strings.TrimSpace(o.Query)
	if o.Limit == 0 {
		o.Limit = 20
	}
	if o.Limit < 1 || o.Limit > 500 || o.Offset < 0 {
		return nil, 0, fmt.Errorf("%w: invalid pagination", ErrInvalidInput)
	}
	if o.HouseID != "" {
		if _, err := s.store.AccessRole(ctx, actor, o.HouseID); err != nil {
			return nil, 0, err
		}
	}
	return s.store.ListSpirits(ctx, actor, o)
}
func (s *Service) GetSpirit(ctx context.Context, actor Actor, id string) (Spirit, error) {
	return s.store.GetSpirit(ctx, actor, id)
}
func (s *Service) CreateSpirit(ctx context.Context, actor Actor, in SpiritInput) (Spirit, error) {
	items, err := s.prepareNewSpirits(ctx, actor, []SpiritInput{in})
	if err != nil {
		return Spirit{}, err
	}
	return s.store.CreateSpirit(ctx, items[0])
}
func (s *Service) CreateSpirits(ctx context.Context, actor Actor, inputs []SpiritInput) ([]Spirit, error) {
	items, err := s.prepareNewSpirits(ctx, actor, inputs)
	if err != nil {
		return nil, err
	}
	return s.store.CreateSpirits(ctx, items)
}
func (s *Service) prepareNewSpirits(ctx context.Context, actor Actor, inputs []SpiritInput) ([]Spirit, error) {
	if len(inputs) == 0 || len(inputs) > 500 {
		return nil, fmt.Errorf("%w: spirits must contain between 1 and 500 rows", ErrInvalidInput)
	}
	now := s.now().UTC()
	items := make([]Spirit, 0, len(inputs))
	tabletHouses := map[string]string{}
	writableHouses := map[string]bool{}
	for index, in := range inputs {
		house := strings.TrimSpace(in.HouseID)
		tabletID := strings.TrimSpace(in.TabletID)
		if tabletID != "" {
			var ok bool
			house, ok = tabletHouses[tabletID]
			if !ok {
				var err error
				house, err = s.store.HouseIDForTablet(ctx, tabletID)
				if err != nil {
					return nil, err
				}
				tabletHouses[tabletID] = house
			}
		}
		if house == "" {
			return nil, fmt.Errorf("%w: row %d: house_id is required when tablet_id is empty", ErrInvalidInput, index+1)
		}
		if !writableHouses[house] {
			if err := s.requireWrite(ctx, actor, house); err != nil {
				return nil, err
			}
			writableHouses[house] = true
		}
		in.HouseID = house
		item, err := s.normalizeSpirit(in)
		if err != nil {
			return nil, fmt.Errorf("%w: row %d: %v", ErrInvalidInput, index+1, err)
		}
		item.ID = newID("spirit")
		item.CreatedAt = now
		item.UpdatedAt = now
		items = append(items, item)
	}
	return items, nil
}
func (s *Service) UpdateSpirit(ctx context.Context, actor Actor, id string, in SpiritInput) (Spirit, error) {
	house, err := s.store.HouseIDForSpirit(ctx, id)
	if err != nil {
		return Spirit{}, err
	}
	if err = s.requireWrite(ctx, actor, house); err != nil {
		return Spirit{}, err
	}
	targetHouse := strings.TrimSpace(in.HouseID)
	if strings.TrimSpace(in.TabletID) != "" {
		targetHouse, err = s.store.HouseIDForTablet(ctx, in.TabletID)
		if err != nil {
			return Spirit{}, err
		}
	}
	if targetHouse == "" {
		targetHouse = house
	}
	if targetHouse != house {
		return Spirit{}, fmt.Errorf("%w: cannot move spirit to another house", ErrInvalidInput)
	}
	in.HouseID = targetHouse
	item, err := s.normalizeSpirit(in)
	if err != nil {
		return Spirit{}, err
	}
	item.ID = id
	item.UpdatedAt = s.now().UTC()
	return s.store.UpdateSpirit(ctx, item)
}
func (s *Service) PatchSpirit(ctx context.Context, actor Actor, id, field, value string) (Spirit, error) {
	allowed := map[string]bool{"full_name": true, "dharma_name": true, "birth_year": true, "death_year": true, "age": true, "burial_place": true, "sender": true, "sent_month": true, "notes": true}
	if !allowed[field] {
		return Spirit{}, fmt.Errorf("%w: field cannot be patched", ErrInvalidInput)
	}
	value = strings.TrimSpace(value)
	if field == "full_name" && value == "" {
		return Spirit{}, fmt.Errorf("%w: full_name is required", ErrInvalidInput)
	}
	house, err := s.store.HouseIDForSpirit(ctx, id)
	if err != nil {
		return Spirit{}, err
	}
	if err = s.requireWrite(ctx, actor, house); err != nil {
		return Spirit{}, err
	}
	if err = s.store.PatchSpirit(ctx, id, field, value, s.now().UTC()); err != nil {
		return Spirit{}, err
	}
	return s.store.GetSpirit(ctx, actor, id)
}
func (s *Service) DeleteSpirit(ctx context.Context, actor Actor, id string) error {
	house, err := s.store.HouseIDForSpirit(ctx, id)
	if err != nil {
		return err
	}
	if err = s.requireWrite(ctx, actor, house); err != nil {
		return err
	}
	return s.store.DeleteSpirit(ctx, id)
}
func (s *Service) normalizeSpirit(in SpiritInput) (Spirit, error) {
	in.FullName = strings.TrimSpace(in.FullName)
	if in.FullName == "" || strings.TrimSpace(in.HouseID) == "" {
		return Spirit{}, fmt.Errorf("%w: house_id and full_name are required", ErrInvalidInput)
	}
	return Spirit{HouseID: strings.TrimSpace(in.HouseID), TabletID: strings.TrimSpace(in.TabletID), FullName: in.FullName, DharmaName: strings.TrimSpace(in.DharmaName), BirthYear: strings.TrimSpace(in.BirthYear), DeathYear: strings.TrimSpace(in.DeathYear), Age: strings.TrimSpace(in.Age), ImageURL: strings.TrimSpace(in.ImageURL), BurialPlace: strings.TrimSpace(in.BurialPlace), Sender: strings.TrimSpace(in.Sender), SentMonth: strings.TrimSpace(in.SentMonth), Notes: strings.TrimSpace(in.Notes)}, nil
}
func (s *Service) requireWrite(ctx context.Context, a Actor, h string) error {
	r, e := s.store.AccessRole(ctx, a, h)
	if e != nil {
		return e
	}
	if a.Role == "admin" || r == "editor" {
		return nil
	}
	return ErrForbidden
}
func newID(prefix string) string {
	var b [12]byte
	if _, e := rand.Read(b[:]); e != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}
func Is(err, target error) bool { return errors.Is(err, target) }
