package memorial

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }
func accessSQL(a Actor) (string, []any) {
	if a.Role == "admin" || a.AllHouses {
		return "TRUE", []any{a.Role}
	}
	return "EXISTS (SELECT 1 FROM spirit_house_members hm WHERE hm.house_id=h.id AND hm.user_id=$2)", []any{a.Role, a.ID}
}

func (s *PostgresStore) ListHouses(ctx context.Context, a Actor) ([]House, error) {
	where, args := accessSQL(a)
	rows, e := s.pool.Query(ctx, `SELECT h.id,h.name,h.address,h.notes,h.active,CASE WHEN $1='admin' THEN 'admin' ELSE $1 END,h.created_at,h.updated_at FROM spirit_houses h WHERE `+where+` ORDER BY h.active DESC,h.name`, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []House{}
	for rows.Next() {
		var v House
		if e = rows.Scan(&v.ID, &v.Name, &v.Address, &v.Notes, &v.Active, &v.AccessRole, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *PostgresStore) CreateHouse(ctx context.Context, v House) (House, error) {
	e := s.pool.QueryRow(ctx, `INSERT INTO spirit_houses(id,name,address,notes,active,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,name,address,notes,active,created_at,updated_at`, v.ID, v.Name, v.Address, v.Notes, v.Active, v.CreatedAt, v.UpdatedAt).Scan(&v.ID, &v.Name, &v.Address, &v.Notes, &v.Active, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func (s *PostgresStore) UpdateHouse(ctx context.Context, v House) (House, error) {
	e := s.pool.QueryRow(ctx, `UPDATE spirit_houses SET name=$2,address=$3,notes=$4,active=$5,updated_at=$6 WHERE id=$1 RETURNING id,name,address,notes,active,created_at,updated_at`, v.ID, v.Name, v.Address, v.Notes, v.Active, v.UpdatedAt).Scan(&v.ID, &v.Name, &v.Address, &v.Notes, &v.Active, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func (s *PostgresStore) DeleteHouse(ctx context.Context, id string) error {
	r, e := s.pool.Exec(ctx, `DELETE FROM spirit_houses WHERE id=$1`, id)
	if e != nil {
		return e
	}
	if r.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *PostgresStore) AccessRole(ctx context.Context, a Actor, id string) (string, error) {
	if a.Role == "admin" || a.AllHouses {
		var ok bool
		e := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM spirit_houses WHERE id=$1)`, id).Scan(&ok)
		if e != nil {
			return "", e
		}
		if !ok {
			return "", ErrNotFound
		}
		return a.Role, nil
	}
	var allowed bool
	e := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM spirit_house_members WHERE house_id=$1 AND user_id=$2)`, id, a.ID).Scan(&allowed)
	if errors.Is(e, pgx.ErrNoRows) {
		return "", ErrForbidden
	}
	if e != nil {
		return "", e
	}
	if !allowed {
		return "", ErrForbidden
	}
	return a.Role, nil
}
func (s *PostgresStore) ListAreas(ctx context.Context, a Actor, h string) ([]Area, error) {
	rows, e := s.pool.Query(ctx, `SELECT a.id,a.house_id,a.code,a.name,a.notes,COUNT(DISTINCT p.id),COUNT(DISTINCT t.id),COUNT(s.id),a.created_at,a.updated_at FROM memorial_areas a LEFT JOIN memorial_positions p ON p.area_id=a.id LEFT JOIN memorial_tablets t ON t.position_id=p.id LEFT JOIN spirits s ON s.tablet_id=t.id WHERE a.house_id=$1 GROUP BY a.id ORDER BY a.code`, h)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Area{}
	for rows.Next() {
		var v Area
		if e = rows.Scan(&v.ID, &v.HouseID, &v.Code, &v.Name, &v.Notes, &v.PositionCount, &v.TabletCount, &v.SpiritCount, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *PostgresStore) CreateArea(ctx context.Context, v Area) (Area, error) {
	e := s.pool.QueryRow(ctx, `INSERT INTO memorial_areas(id,house_id,code,name,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id,house_id,code,name,notes,created_at,updated_at`, v.ID, v.HouseID, v.Code, v.Name, v.Notes, v.CreatedAt, v.UpdatedAt).Scan(&v.ID, &v.HouseID, &v.Code, &v.Name, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func (s *PostgresStore) AreaCode(ctx context.Context, id string) (string, error) {
	var code string
	e := s.pool.QueryRow(ctx, `SELECT code FROM memorial_areas WHERE id=$1`, id).Scan(&code)
	return code, mapErr(e)
}
func (s *PostgresStore) ListPositions(ctx context.Context, a Actor, area string) ([]Position, error) {
	rows, e := s.pool.Query(ctx, `SELECT p.id,p.area_id,h.id,h.name,a.code,p.row_number,p.column_number,p.name,p.notes,COUNT(DISTINCT t.id),COUNT(s.id),p.created_at,p.updated_at FROM memorial_positions p JOIN memorial_areas a ON a.id=p.area_id JOIN spirit_houses h ON h.id=a.house_id LEFT JOIN memorial_tablets t ON t.position_id=p.id LEFT JOIN spirits s ON s.tablet_id=t.id WHERE p.area_id=$1 GROUP BY p.id,a.id,h.id ORDER BY p.row_number,p.column_number,p.name`, area)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Position{}
	for rows.Next() {
		var v Position
		if e = rows.Scan(&v.ID, &v.AreaID, &v.HouseID, &v.HouseName, &v.AreaCode, &v.RowNumber, &v.ColumnNumber, &v.Name, &v.Notes, &v.TabletCount, &v.SpiritCount, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *PostgresStore) SearchPositions(ctx context.Context, _ Actor, o PositionSearchOptions) ([]Position, error) {
	rows, e := s.pool.Query(ctx, `SELECT p.id,p.area_id,h.id,h.name,a.code,p.row_number,p.column_number,p.name,p.notes,COUNT(DISTINCT t.id),COUNT(s.id),p.created_at,p.updated_at
		FROM memorial_positions p
		JOIN memorial_areas a ON a.id=p.area_id
		JOIN spirit_houses h ON h.id=a.house_id
		LEFT JOIN memorial_tablets t ON t.position_id=p.id
		LEFT JOIN spirits s ON s.tablet_id=t.id
		WHERE a.house_id=$1 AND ($2='' OR replace(unaccent(lower(concat_ws(' ',p.name,p.notes,a.code,h.name,p.row_number::text,p.column_number::text))),'-','') LIKE '%'||replace(unaccent(lower($2)),'-','')||'%')
		GROUP BY p.id,a.id,h.id
		ORDER BY CASE WHEN replace(lower(p.name),'-','')=replace(lower($2),'-','') THEN 0 ELSE 1 END,p.row_number,p.column_number,p.name
		LIMIT $3`, o.HouseID, o.Query, o.Limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Position{}
	for rows.Next() {
		var v Position
		if e = rows.Scan(&v.ID, &v.AreaID, &v.HouseID, &v.HouseName, &v.AreaCode, &v.RowNumber, &v.ColumnNumber, &v.Name, &v.Notes, &v.TabletCount, &v.SpiritCount, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *PostgresStore) CreatePosition(ctx context.Context, v Position) (Position, error) {
	e := s.pool.QueryRow(ctx, `INSERT INTO memorial_positions(id,area_id,name,row_number,column_number,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,area_id,name,row_number,column_number,notes,created_at,updated_at`, v.ID, v.AreaID, v.Name, v.RowNumber, v.ColumnNumber, v.Notes, v.CreatedAt, v.UpdatedAt).Scan(&v.ID, &v.AreaID, &v.Name, &v.RowNumber, &v.ColumnNumber, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func (s *PostgresStore) CreatePositions(ctx context.Context, positions []Position) ([]Position, error) {
	data, err := json.Marshal(positions)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `INSERT INTO memorial_positions(id,area_id,name,row_number,column_number,notes,created_at,updated_at)
		SELECT x.id,x.area_id,x.name,x.row_number,x.column_number,x.notes,x.created_at,x.updated_at
		FROM jsonb_to_recordset($1::jsonb) AS x(id text,area_id text,name text,row_number integer,column_number integer,notes text,created_at timestamptz,updated_at timestamptz)
		ON CONFLICT DO NOTHING
		RETURNING id,area_id,name,row_number,column_number,notes,created_at,updated_at`, string(data))
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	created := make([]Position, 0, len(positions))
	for rows.Next() {
		var v Position
		if err = rows.Scan(&v.ID, &v.AreaID, &v.Name, &v.RowNumber, &v.ColumnNumber, &v.Notes, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		created = append(created, v)
	}
	return created, rows.Err()
}
func (s *PostgresStore) UpdatePosition(ctx context.Context, v Position) (Position, error) {
	e := s.pool.QueryRow(ctx, `UPDATE memorial_positions SET area_id=$2,name=$3,row_number=$4,column_number=$5,notes=$6,updated_at=$7 WHERE id=$1 RETURNING id,area_id,name,row_number,column_number,notes,created_at,updated_at`, v.ID, v.AreaID, v.Name, v.RowNumber, v.ColumnNumber, v.Notes, v.UpdatedAt).Scan(&v.ID, &v.AreaID, &v.Name, &v.RowNumber, &v.ColumnNumber, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}

func (s *PostgresStore) ListOccupancyPositions(ctx context.Context, _ Actor, houseID string) ([]Position, int, error) {
	rows, err := s.pool.Query(ctx, `WITH position_stats AS (
		SELECT t.position_id,COUNT(DISTINCT t.id) AS tablet_count,COUNT(s.id) AS spirit_count
		FROM memorial_tablets t
		LEFT JOIN spirits s ON s.tablet_id=t.id
		GROUP BY t.position_id
	)
		SELECT p.id,p.area_id,h.id,h.name,a.code,p.row_number,p.column_number,p.name,p.notes,COALESCE(ps.tablet_count,0),COALESCE(ps.spirit_count,0),p.created_at,p.updated_at
		FROM memorial_positions p
		JOIN memorial_areas a ON a.id=p.area_id
		JOIN spirit_houses h ON h.id=a.house_id
		LEFT JOIN position_stats ps ON ps.position_id=p.id
		WHERE h.id=$1
		ORDER BY a.code,p.row_number,p.column_number,p.name`, houseID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	positions := []Position{}
	for rows.Next() {
		var v Position
		if err = rows.Scan(&v.ID, &v.AreaID, &v.HouseID, &v.HouseName, &v.AreaCode, &v.RowNumber, &v.ColumnNumber, &v.Name, &v.Notes, &v.TabletCount, &v.SpiritCount, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		positions = append(positions, v)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	var unplaced int
	if err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM spirits WHERE house_id=$1 AND tablet_id IS NULL`, houseID).Scan(&unplaced); err != nil {
		return nil, 0, err
	}
	return positions, unplaced, nil
}
func (s *PostgresStore) ListTablets(ctx context.Context, a Actor, position string) ([]Tablet, error) {
	rows, e := s.pool.Query(ctx, `SELECT t.id,t.position_id,h.id,h.name,a.id,a.code,p.name,p.row_number,p.column_number,t.name,t.notes,COUNT(s.id),t.created_at,t.updated_at FROM memorial_tablets t JOIN memorial_positions p ON p.id=t.position_id JOIN memorial_areas a ON a.id=p.area_id JOIN spirit_houses h ON h.id=a.house_id LEFT JOIN spirits s ON s.tablet_id=t.id WHERE t.position_id=$1 GROUP BY t.id,p.id,a.id,h.id ORDER BY t.name`, position)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Tablet{}
	for rows.Next() {
		var v Tablet
		if e = rows.Scan(&v.ID, &v.PositionID, &v.HouseID, &v.HouseName, &v.AreaID, &v.AreaCode, &v.PositionName, &v.RowNumber, &v.ColumnNumber, &v.Name, &v.Notes, &v.SpiritCount, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *PostgresStore) CreateTablet(ctx context.Context, v Tablet) (Tablet, error) {
	e := s.pool.QueryRow(ctx, `INSERT INTO memorial_tablets(id,position_id,name,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,position_id,name,notes,created_at,updated_at`, v.ID, v.PositionID, v.Name, v.Notes, v.CreatedAt, v.UpdatedAt).Scan(&v.ID, &v.PositionID, &v.Name, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func (s *PostgresStore) CreateTabletWithSpirits(ctx context.Context, v Tablet, spirits []Spirit, existingSpiritIDs []string, houseID string) (Tablet, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Tablet{}, fmt.Errorf("begin create tablet: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err = tx.QueryRow(ctx, `INSERT INTO memorial_tablets(id,position_id,name,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,position_id,name,notes,created_at,updated_at`, v.ID, v.PositionID, v.Name, v.Notes, v.CreatedAt, v.UpdatedAt).Scan(&v.ID, &v.PositionID, &v.Name, &v.Notes, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return Tablet{}, mapErr(err)
	}
	if len(existingSpiritIDs) > 0 {
		result, updateErr := tx.Exec(ctx, `UPDATE spirits SET tablet_id=$1,updated_at=$2 WHERE house_id=$3 AND tablet_id IS NULL AND id=ANY($4::text[])`, v.ID, v.UpdatedAt, houseID, existingSpiritIDs)
		if updateErr != nil {
			return Tablet{}, mapErr(updateErr)
		}
		if result.RowsAffected() != int64(len(existingSpiritIDs)) {
			return Tablet{}, fmt.Errorf("%w: one or more spirits are no longer unplaced", ErrConflict)
		}
	}
	if len(spirits) > 0 {
		data, marshalErr := json.Marshal(spirits)
		if marshalErr != nil {
			return Tablet{}, fmt.Errorf("encode spirits: %w", marshalErr)
		}
		_, err = tx.Exec(ctx, `INSERT INTO spirits(id,house_id,tablet_id,full_name,dharma_name,birth_year,death_year,age,image_url,burial_place,sender,sent_month,notes,created_at,updated_at)
			SELECT x.id,x.house_id,$1,x.full_name,x.dharma_name,x.birth_year,x.death_year,x.age,x.image_url,x.burial_place,x.sender,x.sent_month,x.notes,x.created_at,x.updated_at
			FROM jsonb_to_recordset($2::jsonb) AS x(id text,house_id text,full_name text,dharma_name text,birth_year text,death_year text,age text,image_url text,burial_place text,sender text,sent_month text,notes text,created_at timestamptz,updated_at timestamptz)`, v.ID, string(data))
		if err != nil {
			return Tablet{}, mapErr(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Tablet{}, fmt.Errorf("commit create tablet: %w", err)
	}
	v.SpiritCount = len(spirits) + len(existingSpiritIDs)
	return v, nil
}
func (s *PostgresStore) UpdateTabletWithSpirits(ctx context.Context, v Tablet, spirits []Spirit) (Tablet, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Tablet{}, fmt.Errorf("begin update tablet: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err = tx.QueryRow(ctx, `UPDATE memorial_tablets SET position_id=$2,name=$3,notes=$4,updated_at=$5 WHERE id=$1 RETURNING id,position_id,name,notes,created_at,updated_at`, v.ID, v.PositionID, v.Name, v.Notes, v.UpdatedAt).Scan(&v.ID, &v.PositionID, &v.Name, &v.Notes, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return Tablet{}, mapErr(err)
	}
	kept := make([]string, 0, len(spirits))
	existingIDs := make([]string, 0, len(spirits))
	for i := range spirits {
		kept = append(kept, spirits[i].ID)
		if spirits[i].CreatedAt.IsZero() {
			existingIDs = append(existingIDs, spirits[i].ID)
			spirits[i].CreatedAt = spirits[i].UpdatedAt
		}
	}
	var validExisting bool
	if err = tx.QueryRow(ctx, `SELECT COUNT(*)=cardinality($2::text[]) FROM spirits WHERE tablet_id=$1 AND id=ANY($2::text[])`, v.ID, existingIDs).Scan(&validExisting); err != nil {
		return Tablet{}, err
	}
	if !validExisting {
		return Tablet{}, fmt.Errorf("%w: spirit does not belong to tablet", ErrInvalidInput)
	}
	data, err := json.Marshal(spirits)
	if err != nil {
		return Tablet{}, fmt.Errorf("encode spirits: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO spirits(id,house_id,tablet_id,full_name,dharma_name,birth_year,death_year,age,image_url,burial_place,sender,sent_month,notes,created_at,updated_at)
		SELECT x.id,x.house_id,$1,x.full_name,x.dharma_name,x.birth_year,x.death_year,x.age,x.image_url,x.burial_place,x.sender,x.sent_month,x.notes,x.created_at,x.updated_at
		FROM jsonb_to_recordset($2::jsonb) AS x(id text,house_id text,full_name text,dharma_name text,birth_year text,death_year text,age text,image_url text,burial_place text,sender text,sent_month text,notes text,created_at timestamptz,updated_at timestamptz)
		ON CONFLICT(id) DO UPDATE SET full_name=EXCLUDED.full_name,dharma_name=EXCLUDED.dharma_name,birth_year=EXCLUDED.birth_year,death_year=EXCLUDED.death_year,age=EXCLUDED.age,image_url=EXCLUDED.image_url,burial_place=EXCLUDED.burial_place,sender=EXCLUDED.sender,sent_month=EXCLUDED.sent_month,notes=EXCLUDED.notes,updated_at=EXCLUDED.updated_at
		WHERE spirits.tablet_id=$1`, v.ID, string(data))
	if err != nil {
		return Tablet{}, mapErr(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM spirits WHERE tablet_id=$1 AND NOT (id=ANY($2::text[]))`, v.ID, kept); err != nil {
		return Tablet{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Tablet{}, fmt.Errorf("commit update tablet: %w", err)
	}
	v.SpiritCount = len(spirits)
	return v, nil
}

const spiritCols = `s.id,COALESCE(s.tablet_id,''),h.id,h.name,COALESCE(a.id,''),COALESCE(a.code,''),COALESCE(p.id,''),COALESCE(p.name,''),COALESCE(t.name,''),s.full_name,s.dharma_name,s.birth_year,s.death_year,s.age,s.image_url,s.burial_place,s.sender,s.sent_month,s.notes,s.created_at,s.updated_at`

func (s *PostgresStore) ListSpirits(ctx context.Context, a Actor, o SearchOptions) ([]Spirit, int, error) {
	access := "($1='admin' OR $3 OR EXISTS(SELECT 1 FROM spirit_house_members hm WHERE hm.house_id=h.id AND hm.user_id=$2))"
	filter := access + ` AND ($4='' OR h.id=$4) AND ($5='' OR a.id=$5) AND ($6='' OR p.id=$6) AND ($7='' OR t.id=$7) AND ($8='' OR replace(unaccent(lower(concat_ws(' ',s.full_name,s.dharma_name,s.birth_year,s.death_year,s.age,s.burial_place,s.sender,s.sent_month,s.notes,p.name,t.name,a.code,h.name))),'-','') LIKE '%'||replace(unaccent(lower($8)),'-','')||'%') AND (NOT $9 OR s.tablet_id IS NULL)`
	args := []any{a.Role, a.ID, a.AllHouses, o.HouseID, o.AreaID, o.PositionID, o.TabletID, o.Query, o.Unplaced}
	var total int
	if e := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM spirits s JOIN spirit_houses h ON h.id=s.house_id LEFT JOIN memorial_tablets t ON t.id=s.tablet_id LEFT JOIN memorial_positions p ON p.id=t.position_id LEFT JOIN memorial_areas a ON a.id=p.area_id WHERE `+filter, args...).Scan(&total); e != nil {
		return nil, 0, e
	}
	rows, e := s.pool.Query(ctx, `SELECT `+spiritCols+` FROM spirits s JOIN spirit_houses h ON h.id=s.house_id LEFT JOIN memorial_tablets t ON t.id=s.tablet_id LEFT JOIN memorial_positions p ON p.id=t.position_id LEFT JOIN memorial_areas a ON a.id=p.area_id WHERE `+filter+` ORDER BY s.full_name,s.id LIMIT $10 OFFSET $11`, append(args, o.Limit, o.Offset)...)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	out := []Spirit{}
	for rows.Next() {
		v, e := scanSpirit(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}
func (s *PostgresStore) GetSpirit(ctx context.Context, a Actor, id string) (Spirit, error) {
	return scanSpirit(s.pool.QueryRow(ctx, `SELECT `+spiritCols+` FROM spirits s JOIN spirit_houses h ON h.id=s.house_id LEFT JOIN memorial_tablets t ON t.id=s.tablet_id LEFT JOIN memorial_positions p ON p.id=t.position_id LEFT JOIN memorial_areas a ON a.id=p.area_id WHERE s.id=$1 AND ($2='admin' OR $4 OR EXISTS(SELECT 1 FROM spirit_house_members hm WHERE hm.house_id=h.id AND hm.user_id=$3))`, id, a.Role, a.ID, a.AllHouses))
}
func (s *PostgresStore) CreateSpirit(ctx context.Context, v Spirit) (Spirit, error) {
	e := s.pool.QueryRow(ctx, `INSERT INTO spirits(id,house_id,tablet_id,full_name,dharma_name,birth_year,death_year,age,image_url,burial_place,sender,sent_month,notes,created_at,updated_at) VALUES($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id,created_at,updated_at`, v.ID, v.HouseID, v.TabletID, v.FullName, v.DharmaName, v.BirthYear, v.DeathYear, v.Age, v.ImageURL, v.BurialPlace, v.Sender, v.SentMonth, v.Notes, v.CreatedAt, v.UpdatedAt).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func (s *PostgresStore) CreateSpirits(ctx context.Context, items []Spirit) ([]Spirit, error) {
	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO spirits(id,house_id,tablet_id,full_name,dharma_name,birth_year,death_year,age,image_url,burial_place,sender,sent_month,notes,created_at,updated_at)
		SELECT x.id,x.house_id,NULLIF(x.tablet_id,''),x.full_name,x.dharma_name,x.birth_year,x.death_year,x.age,x.image_url,x.burial_place,x.sender,x.sent_month,x.notes,x.created_at,x.updated_at
		FROM jsonb_to_recordset($1::jsonb) AS x(id text,house_id text,tablet_id text,full_name text,dharma_name text,birth_year text,death_year text,age text,image_url text,burial_place text,sender text,sent_month text,notes text,created_at timestamptz,updated_at timestamptz)`, string(data))
	return items, mapErr(err)
}
func (s *PostgresStore) UpdateSpirit(ctx context.Context, v Spirit) (Spirit, error) {
	e := s.pool.QueryRow(ctx, `UPDATE spirits SET house_id=$2,tablet_id=NULLIF($3,''),full_name=$4,dharma_name=$5,birth_year=$6,death_year=$7,age=$8,image_url=$9,burial_place=$10,sender=$11,sent_month=$12,notes=$13,updated_at=$14 WHERE id=$1 RETURNING created_at,updated_at`, v.ID, v.HouseID, v.TabletID, v.FullName, v.DharmaName, v.BirthYear, v.DeathYear, v.Age, v.ImageURL, v.BurialPlace, v.Sender, v.SentMonth, v.Notes, v.UpdatedAt).Scan(&v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func (s *PostgresStore) PatchSpirit(ctx context.Context, id, field, value string, updatedAt time.Time) error {
	columns := map[string]string{"full_name": "full_name", "dharma_name": "dharma_name", "birth_year": "birth_year", "death_year": "death_year", "age": "age", "burial_place": "burial_place", "sender": "sender", "sent_month": "sent_month", "notes": "notes"}
	column, ok := columns[field]
	if !ok {
		return ErrInvalidInput
	}
	result, err := s.pool.Exec(ctx, `UPDATE spirits SET `+column+`=$2,updated_at=$3 WHERE id=$1`, id, value, updatedAt)
	if err != nil {
		return mapErr(err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *PostgresStore) DeleteSpirit(ctx context.Context, id string) error {
	r, e := s.pool.Exec(ctx, `DELETE FROM spirits WHERE id=$1`, id)
	if e != nil {
		return e
	}
	if r.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *PostgresStore) HouseIDForArea(ctx context.Context, id string) (string, error) {
	var h string
	e := s.pool.QueryRow(ctx, `SELECT house_id FROM memorial_areas WHERE id=$1`, id).Scan(&h)
	return h, mapErr(e)
}
func (s *PostgresStore) HouseIDForTablet(ctx context.Context, id string) (string, error) {
	var h string
	e := s.pool.QueryRow(ctx, `SELECT a.house_id FROM memorial_tablets t JOIN memorial_positions p ON p.id=t.position_id JOIN memorial_areas a ON a.id=p.area_id WHERE t.id=$1`, id).Scan(&h)
	return h, mapErr(e)
}
func (s *PostgresStore) HouseIDForPosition(ctx context.Context, id string) (string, error) {
	var h string
	e := s.pool.QueryRow(ctx, `SELECT a.house_id FROM memorial_positions p JOIN memorial_areas a ON a.id=p.area_id WHERE p.id=$1`, id).Scan(&h)
	return h, mapErr(e)
}
func (s *PostgresStore) HouseIDForSpirit(ctx context.Context, id string) (string, error) {
	var h string
	e := s.pool.QueryRow(ctx, `SELECT house_id FROM spirits WHERE id=$1`, id).Scan(&h)
	return h, mapErr(e)
}

type scanner interface{ Scan(...any) error }

func scanSpirit(r scanner) (Spirit, error) {
	var v Spirit
	e := r.Scan(&v.ID, &v.TabletID, &v.HouseID, &v.HouseName, &v.AreaID, &v.AreaCode, &v.PositionID, &v.PositionName, &v.TabletName, &v.FullName, &v.DharmaName, &v.BirthYear, &v.DeathYear, &v.Age, &v.ImageURL, &v.BurialPlace, &v.Sender, &v.SentMonth, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	return v, mapErr(e)
}
func mapErr(e error) error {
	if errors.Is(e, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var p *pgconn.PgError
	if errors.As(e, &p) && (p.Code == "23505" || p.Code == "23503") {
		return fmt.Errorf("%w: %s", ErrConflict, p.ConstraintName)
	}
	return e
}
