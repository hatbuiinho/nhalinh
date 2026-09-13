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
	ID                    string    `json:"id"`
	AreaID                string    `json:"area_id"`
	HouseID               string    `json:"house_id"`
	HouseName             string    `json:"house_name"`
	AreaCode              string    `json:"area_code"`
	RowNumber             int       `json:"row_number"`
	ColumnNumber          int       `json:"column_number"`
	Name                  string    `json:"name"`
	Notes                 string    `json:"notes"`
	TabletCount           int       `json:"tablet_count"`
	SpiritCount           int       `json:"spirit_count"`
	SpiritNames           []string  `json:"spirit_names"`
	TabletStatuses        []string  `json:"tablet_statuses"`
	TabletTypes           []string  `json:"tablet_types"`
	SingleSpiritName      string    `json:"single_spirit_name"`
	SingleSpiritBirthYear string    `json:"single_spirit_birth_year"`
	SingleSpiritDeathYear string    `json:"single_spirit_death_year"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
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
	ID                   string    `json:"id"`
	PositionID           string    `json:"position_id"`
	HouseID              string    `json:"house_id"`
	HouseName            string    `json:"house_name"`
	AreaID               string    `json:"area_id"`
	AreaCode             string    `json:"area_code"`
	PositionName         string    `json:"position_name"`
	RowNumber            int       `json:"row_number"`
	ColumnNumber         int       `json:"column_number"`
	Name                 string    `json:"name"`
	Code                 string    `json:"code"`
	RegisteredAt         string    `json:"registered_at"`
	EnshrinedAt          string    `json:"enshrined_at"`
	EnteredWorshipAreaAt string    `json:"entered_worship_area_at"`
	ImageURL             string    `json:"image_url"`
	Sender               string    `json:"sender"`
	Notes                string    `json:"notes"`
	Status               string    `json:"status"`
	Type                 string    `json:"type"`
	SpiritCount          int       `json:"spirit_count"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
type Spirit struct {
	ID                   string     `json:"id"`
	TabletID             string     `json:"tablet_id"`
	HouseID              string     `json:"house_id"`
	HouseName            string     `json:"house_name"`
	AreaID               string     `json:"area_id"`
	AreaCode             string     `json:"area_code"`
	PositionID           string     `json:"position_id"`
	PositionName         string     `json:"position_name"`
	TabletName           string     `json:"tablet_name"`
	TabletCode           string     `json:"tablet_code"`
	TabletImageURL       string     `json:"tablet_image_url"`
	TabletRegisteredAt   string     `json:"tablet_registered_at"`
	TabletSender         string     `json:"tablet_sender"`
	TabletType           string     `json:"tablet_type"`
	TabletStatus         string     `json:"tablet_status"`
	FullName             string     `json:"full_name"`
	DharmaName           string     `json:"dharma_name"`
	FamiliarName         string     `json:"familiar_name"`
	Gender               string     `json:"gender"`
	BirthDate            string     `json:"birth_date"`
	DeathDate            string     `json:"death_date"`
	BirthLunar           string     `json:"birth_lunar"`
	DeathLunar           string     `json:"death_lunar"`
	BirthYear            string     `json:"birth_year"`
	DeathYear            string     `json:"death_year"`
	Status               string     `json:"status"`
	EnteredWorshipAreaAt string     `json:"entered_worship_area_at"`
	EnshrinedAt          string     `json:"enshrined_at"`
	Age                  string     `json:"age"`
	ImageURL             string     `json:"image_url"`
	BurialPlace          string     `json:"burial_place"`
	Sender               string     `json:"sender"`
	SentMonth            string     `json:"sent_month"`
	Notes                string     `json:"notes"`
	HasUrn               bool       `json:"has_urn"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"-"`
}
type SpiritPositionHistory struct {
	ID             string    `json:"id"`
	SpiritID       string    `json:"spirit_id"`
	FromTabletID   string    `json:"from_tablet_id"`
	ToTabletID     string    `json:"to_tablet_id"`
	FromPositionID string    `json:"from_position_id"`
	ToPositionID   string    `json:"to_position_id"`
	ChangeType     string    `json:"change_type"`
	ChangedAt      time.Time `json:"changed_at"`
	Notes          string    `json:"notes"`
	PerformedBy    string    `json:"performed_by"`
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
	PositionID, Name, Code, RegisteredAt, EnshrinedAt, EnteredWorshipAreaAt, ImageURL, Sender, Notes, Status, Type string
	Spirits                                                                                                        []SpiritInput
	ExistingSpiritIDs                                                                                              []string
}
type SpiritInput struct {
	ID, HouseID, TabletID, FullName, DharmaName, FamiliarName, Gender, BirthDate, DeathDate, BirthLunar, DeathLunar, BirthYear, DeathYear, Status, EnteredWorshipAreaAt, EnshrinedAt, Age, ImageURL, BurialPlace, Sender, SentMonth, Notes string
	HasUrn                                                                                                                                                                                                                                 bool
}
type SearchOptions struct {
	Query, HouseID, AreaID, PositionID, TabletID string
	Limit, Offset                                int
	Unplaced                                     bool
	GroupByTablet                                bool
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
type Urn struct {
	ID              string    `json:"id"`
	SpiritID        string    `json:"spirit_id"`
	HouseID         string    `json:"house_id"`
	FullName        string    `json:"full_name"`
	DharmaName      string    `json:"dharma_name"`
	BirthYear       string    `json:"birth_year"`
	DeathYear       string    `json:"death_year"`
	Code            string    `json:"code"`
	UrnType         string    `json:"urn_type"`
	Material        string    `json:"material"`
	Color           string    `json:"color"`
	Dimensions      string    `json:"dimensions"`
	StorageLocation string    `json:"storage_location"`
	InstalledAt     string    `json:"installed_at"`
	MovedAt         string    `json:"moved_at"`
	ImageURL        string    `json:"image_url"`
	Status          string    `json:"status"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type UrnHistory struct {
	ID              string    `json:"id"`
	UrnID           string    `json:"urn_id"`
	Status          string    `json:"status"`
	StorageLocation string    `json:"storage_location"`
	ChangedAt       string    `json:"changed_at"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
}
type UrnInput struct{ SpiritID, Code, UrnType, Material, Color, Dimensions, StorageLocation, InstalledAt, MovedAt, ImageURL, Status, Notes string }
