package memorial

import "time"

var (
	ErrNotFound     = errorString("not found")
	ErrInvalidInput = errorString("invalid input")
	ErrForbidden    = errorString("forbidden")
	ErrConflict     = errorString("conflict")
)

type errorString string

func (e errorString) Error() string { return string(e) }

type Actor struct {
	ID, Role  string
	AllHouses bool
}

type House struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Notes      string    `json:"notes"`
	Active     bool      `json:"active"`
	AccessRole string    `json:"access_role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type Area struct {
	ID            string    `json:"id"`
	HouseID       string    `json:"house_id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Notes         string    `json:"notes"`
	PositionCount int       `json:"position_count"`
	TabletCount   int       `json:"tablet_count"`
	SpiritCount   int       `json:"spirit_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type Position struct {
	ID           string    `json:"id"`
	AreaID       string    `json:"area_id"`
	HouseID      string    `json:"house_id"`
	HouseName    string    `json:"house_name"`
	AreaCode     string    `json:"area_code"`
	RowNumber    int       `json:"row_number"`
	ColumnNumber int       `json:"column_number"`
	Name         string    `json:"name"`
	Notes        string    `json:"notes"`
	TabletCount  int       `json:"tablet_count"`
	SpiritCount  int       `json:"spirit_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type OccupancySummary struct {
	AreaCount           int `json:"area_count"`
	PositionCount       int `json:"position_count"`
	EmptyPositionCount  int `json:"empty_position_count"`
	UsedPositionCount   int `json:"used_position_count"`
	TabletCount         int `json:"tablet_count"`
	SpiritCount         int `json:"spirit_count"`
	UnplacedSpiritCount int `json:"unplaced_spirit_count"`
}
type OccupancyArea struct {
	ID                 string `json:"id"`
	Code               string `json:"code"`
	Name               string `json:"name"`
	PositionCount      int    `json:"position_count"`
	EmptyPositionCount int    `json:"empty_position_count"`
	TabletCount        int    `json:"tablet_count"`
	SpiritCount        int    `json:"spirit_count"`
}
type Occupancy struct {
	HouseID   string           `json:"house_id"`
	Summary   OccupancySummary `json:"summary"`
	Areas     []OccupancyArea  `json:"areas"`
	Positions []Position       `json:"positions"`
}
type Tablet struct {
	ID           string    `json:"id"`
	PositionID   string    `json:"position_id"`
	HouseID      string    `json:"house_id"`
	HouseName    string    `json:"house_name"`
	AreaID       string    `json:"area_id"`
	AreaCode     string    `json:"area_code"`
	PositionName string    `json:"position_name"`
	RowNumber    int       `json:"row_number"`
	ColumnNumber int       `json:"column_number"`
	Name         string    `json:"name"`
	Notes        string    `json:"notes"`
	SpiritCount  int       `json:"spirit_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type Spirit struct {
	ID           string     `json:"id"`
	TabletID     string     `json:"tablet_id"`
	HouseID      string     `json:"house_id"`
	HouseName    string     `json:"house_name"`
	AreaID       string     `json:"area_id"`
	AreaCode     string     `json:"area_code"`
	PositionID   string     `json:"position_id"`
	PositionName string     `json:"position_name"`
	TabletName   string     `json:"tablet_name"`
	FullName     string     `json:"full_name"`
	DharmaName   string     `json:"dharma_name"`
	BirthYear    string     `json:"birth_year"`
	DeathYear    string     `json:"death_year"`
	Age          string     `json:"age"`
	ImageURL     string     `json:"image_url"`
	BurialPlace  string     `json:"burial_place"`
	Sender       string     `json:"sender"`
	SentMonth    string     `json:"sent_month"`
	Notes        string     `json:"notes"`
	HasUrn       bool       `json:"has_urn"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}
type HouseInput struct {
	Name, Address, Notes string
	Active               bool
}
type AreaInput struct{ HouseID, Code, Name, Notes string }
type PositionInput struct {
	AreaID                  string
	RowNumber, ColumnNumber int
	Notes                   string
}
type TabletInput struct {
	PositionID, Name, Notes string
	Spirits                 []SpiritInput
	ExistingSpiritIDs       []string
}
type SpiritInput struct {
	ID, HouseID, TabletID, FullName, DharmaName, BirthYear, DeathYear, Age, ImageURL, BurialPlace, Sender, SentMonth, Notes string
	HasUrn                                                                                                                  bool
}
type SearchOptions struct {
	Query, HouseID, AreaID, PositionID, TabletID string
	Limit, Offset                                int
	Unplaced                                     bool
	PlacementStatus, UrnStatus                   string
}
type PositionSearchOptions struct {
	HouseID, Query string
	Limit          int
}
type SpiritImportIssue struct {
	RowNumber int    `json:"row_number"`
	Message   string `json:"message"`
}
type SpiritImportPreview struct {
	TotalRows           int                 `json:"total_rows"`
	ValidRows           int                 `json:"valid_rows"`
	InvalidRows         int                 `json:"invalid_rows"`
	CreateAreaCount     int                 `json:"create_area_count"`
	CreatePositionCount int                 `json:"create_position_count"`
	CreateTabletCount   int                 `json:"create_tablet_count"`
	CreateSpiritCount   int                 `json:"create_spirit_count"`
	Errors              []SpiritImportIssue `json:"errors"`
}
type SpiritImportResult struct {
	CreatedAreaCount     int `json:"created_area_count"`
	CreatedPositionCount int `json:"created_position_count"`
	CreatedTabletCount   int `json:"created_tablet_count"`
	CreatedSpiritCount   int `json:"created_spirit_count"`
}
