package memorial

import (
	"context"
	"time"
)

type Store interface {
	ListHouses(context.Context, Actor) ([]House, error)
	CreateHouse(context.Context, House) (House, error)
	UpdateHouse(context.Context, House) (House, error)
	DeleteHouse(context.Context, string) error
	AccessRole(context.Context, Actor, string) (string, error)
	ListAreas(context.Context, Actor, string) ([]Area, error)
	CreateArea(context.Context, Area) (Area, error)
	AreaCode(context.Context, string) (string, error)
	ListPositions(context.Context, Actor, string) ([]Position, error)
	ListOccupancyPositions(context.Context, Actor, string) ([]Position, int, error)
	SearchPositions(context.Context, Actor, PositionSearchOptions) ([]Position, error)
	CreatePosition(context.Context, Position) (Position, error)
	CreatePositions(context.Context, []Position) ([]Position, error)
	UpdatePosition(context.Context, Position) (Position, error)
	DeletePosition(context.Context, string, time.Time) error
	ListTablets(context.Context, Actor, string) ([]Tablet, error)
	ListUnplacedTablets(context.Context, Actor, string, string) ([]Tablet, error)
	MoveTablet(context.Context, string, string, time.Time) error
	CreateTablet(context.Context, Tablet) (Tablet, error)
	CreateTablets(context.Context, []Tablet) ([]Tablet, error)
	CreateTabletWithSpirits(context.Context, Tablet, []Spirit, []string, string) (Tablet, error)
	UpdateTabletWithSpirits(context.Context, Tablet, []Spirit) (Tablet, error)
	DeleteTablet(context.Context, string, bool, time.Time) error
	ListSpirits(context.Context, Actor, SearchOptions) ([]Spirit, int, error)
	GetSpirit(context.Context, Actor, string) (Spirit, error)
	CreateSpirit(context.Context, Spirit) (Spirit, error)
	CreateSpirits(context.Context, []Spirit) ([]Spirit, error)
	UpdateSpirit(context.Context, Spirit) (Spirit, error)
	PatchSpirit(context.Context, string, string, string, time.Time) error
	DeleteSpirit(context.Context, string) error
	BulkPatchSpirits(context.Context, []string, string, string, time.Time) error
	SoftDeleteSpirits(context.Context, []string, time.Time) error
	HouseIDForArea(context.Context, string) (string, error)
	HouseIDForPosition(context.Context, string) (string, error)
	HouseIDForTablet(context.Context, string) (string, error)
	HouseIDForSpirit(context.Context, string) (string, error)
	HouseIDsForSpirits(context.Context, []string) (map[string]string, error)
}
