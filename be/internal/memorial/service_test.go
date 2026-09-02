package memorial

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestHouseAccessAndMultipleSpiritsPerTablet(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), func() time.Time { return time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC) })
	admin := Actor{ID: "admin", Role: "admin"}
	viewer := Actor{ID: "viewer", Role: "viewer", AllHouses: true}
	house, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh 1"})
	if err != nil {
		t.Fatal(err)
	}
	area, err := service.CreateArea(ctx, admin, AreaInput{HouseID: house.ID, Code: "a"})
	if err != nil {
		t.Fatal(err)
	}
	position, err := service.CreatePosition(ctx, admin, PositionInput{AreaID: area.ID, RowNumber: 2, ColumnNumber: 1})
	if err != nil {
		t.Fatal(err)
	}
	if position.Name != "1A-2" {
		t.Fatalf("unexpected generated position: %s", position.Name)
	}
	positionMatches, err := service.SearchPositions(ctx, viewer, PositionSearchOptions{HouseID: house.ID, Query: "1a2", Limit: 20})
	if err != nil || len(positionMatches) != 1 || positionMatches[0].ID != position.ID {
		t.Fatalf("unexpected direct position search: items=%d err=%v", len(positionMatches), err)
	}
	position, err = service.UpdatePosition(ctx, admin, position.ID, PositionInput{AreaID: area.ID, RowNumber: 3, ColumnNumber: 2, Notes: "Đã sửa"})
	if err != nil || position.Name != "2A-3" || position.Notes != "Đã sửa" {
		t.Fatalf("unexpected updated position: %#v, err=%v", position, err)
	}
	if _, err = service.CreatePosition(ctx, admin, PositionInput{AreaID: area.ID, RowNumber: 4, ColumnNumber: 1}); err != nil {
		t.Fatal(err)
	}
	tablet, err := service.CreateTablet(ctx, admin, TabletInput{PositionID: position.ID, Name: "Bài vị 1", Spirits: []SpiritInput{{FullName: "Nguyễn Văn An"}, {FullName: "Trần Thị Bình"}}})
	if err != nil {
		t.Fatal(err)
	}
	if tablet.SpiritCount != 2 {
		t.Fatalf("expected two inline spirits, got %d", tablet.SpiritCount)
	}
	created, _, err := service.ListSpirits(ctx, admin, SearchOptions{TabletID: tablet.ID, Limit: 500})
	if err != nil || len(created) != 2 {
		t.Fatalf("unexpected created spirits: items=%d err=%v", len(created), err)
	}
	tablet, err = service.UpdateTablet(ctx, admin, tablet.ID, TabletInput{
		PositionID: position.ID,
		Name:       "Bài vị đã sửa",
		Notes:      "Ghi chú mới",
		Spirits: []SpiritInput{
			{ID: created[0].ID, FullName: "Nguyễn Văn An đã sửa"},
			{FullName: "Lê Văn Cường"},
		},
	})
	if err != nil || tablet.Name != "Bài vị đã sửa" || tablet.SpiritCount != 2 {
		t.Fatalf("unexpected updated tablet: %#v, err=%v", tablet, err)
	}
	updated, _, err := service.ListSpirits(ctx, admin, SearchOptions{TabletID: tablet.ID, Limit: 500})
	if err != nil || len(updated) != 2 {
		t.Fatalf("unexpected updated spirits: items=%d err=%v", len(updated), err)
	}
	names := map[string]bool{updated[0].FullName: true, updated[1].FullName: true}
	if !names["Nguyễn Văn An đã sửa"] || !names["Lê Văn Cường"] || names["Trần Thị Bình"] {
		t.Fatalf("unexpected synchronized spirit list: %#v", names)
	}
	items, total, err := service.ListSpirits(ctx, viewer, SearchOptions{HouseID: house.ID, Query: "2a3", Limit: 20})
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("unexpected search: total=%d items=%d err=%v", total, len(items), err)
	}
	if _, err = service.CreateSpirit(ctx, viewer, SpiritInput{TabletID: tablet.ID, FullName: "Không được phép"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer write should be forbidden, got %v", err)
	}
	unplaced, err := service.CreateSpirit(ctx, admin, SpiritInput{HouseID: house.ID, FullName: "Hương linh chưa xếp"})
	if err != nil || unplaced.TabletID != "" || unplaced.HouseID != house.ID {
		t.Fatalf("unexpected unplaced spirit: %#v err=%v", unplaced, err)
	}
	batch, err := service.CreateSpirits(ctx, admin, []SpiritInput{
		{HouseID: house.ID, FullName: "Hương linh batch 1"},
		{HouseID: house.ID, FullName: "Hương linh batch 2"},
	})
	if err != nil || len(batch) != 2 || batch[0].HouseID != house.ID || batch[1].HouseID != house.ID {
		t.Fatalf("unexpected spirit batch: %#v err=%v", batch, err)
	}
	patched, err := service.PatchSpirit(ctx, admin, batch[0].ID, "sender", "Người gửi mới")
	if err != nil || patched.Sender != "Người gửi mới" || patched.FullName != batch[0].FullName {
		t.Fatalf("unexpected patched spirit: %#v err=%v", patched, err)
	}
	if _, err = service.PatchSpirit(ctx, admin, batch[0].ID, "image_url", "forbidden"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("image_url patch should be rejected, got %v", err)
	}
	if err = service.BulkPatchSpirits(ctx, admin, []string{batch[0].ID, batch[1].ID}, "sender", "Gia đình chung"); err != nil {
		t.Fatal(err)
	}
	if err = service.BulkDeleteSpirits(ctx, admin, []string{batch[1].ID}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.GetSpirit(ctx, admin, batch[1].ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("soft-deleted spirit must be hidden, got %v", err)
	}
	occupancy, err := service.Occupancy(ctx, viewer, house.ID)
	if err != nil {
		t.Fatal(err)
	}
	if occupancy.Summary.PositionCount != 2 || occupancy.Summary.EmptyPositionCount != 1 || occupancy.Summary.TabletCount != 1 || occupancy.Summary.SpiritCount != 4 || occupancy.Summary.UnplacedSpiritCount != 2 {
		t.Fatalf("unexpected occupancy summary: %#v", occupancy.Summary)
	}
}

func TestCreatePositionsSkipsExistingCoordinates(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), func() time.Time { return time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC) })
	admin := Actor{ID: "admin", Role: "admin"}
	house, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh batch"})
	if err != nil {
		t.Fatal(err)
	}
	area, err := service.CreateArea(ctx, admin, AreaInput{HouseID: house.ID, Code: "B"})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreatePositions(ctx, admin, []PositionInput{
		{AreaID: area.ID, RowNumber: 1, ColumnNumber: 1},
		{AreaID: area.ID, RowNumber: 1, ColumnNumber: 2},
	})
	if err != nil || len(created) != 2 || created[0].Name != "1B-1" || created[1].Name != "2B-1" {
		t.Fatalf("unexpected batch positions: %#v err=%v", created, err)
	}
	created, err = service.CreatePositions(ctx, admin, []PositionInput{
		{AreaID: area.ID, RowNumber: 2, ColumnNumber: 1},
		{AreaID: area.ID, RowNumber: 1, ColumnNumber: 1},
		{AreaID: area.ID, RowNumber: 2, ColumnNumber: 1},
	})
	if err != nil || len(created) != 1 || created[0].Name != "1B-2" {
		t.Fatalf("expected one new position and duplicates skipped: %#v err=%v", created, err)
	}
	positions, err := service.ListPositions(ctx, admin, area.ID)
	if err != nil || len(positions) != 3 {
		t.Fatalf("unexpected positions after skipped duplicates: items=%d err=%v", len(positions), err)
	}
}

func TestBulkDeleteSpiritsHasNoSelectionLimit(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), time.Now)
	admin := Actor{ID: "admin", Role: "admin"}
	house, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh xóa hàng loạt"})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, 1000)
	for batch := 0; batch < 2; batch++ {
		inputs := make([]SpiritInput, 0, 500)
		for index := 0; index < 500; index++ {
			inputs = append(inputs, SpiritInput{HouseID: house.ID, FullName: fmt.Sprintf("Hương linh %d", batch*500+index)})
		}
		created, err := service.CreateSpirits(ctx, admin, inputs)
		if err != nil {
			t.Fatal(err)
		}
		for _, spirit := range created {
			ids = append(ids, spirit.ID)
		}
	}
	if err := service.BulkDeleteSpirits(ctx, admin, ids); err != nil {
		t.Fatalf("bulk delete 1000 spirits: %v", err)
	}
	items, total, err := service.ListSpirits(ctx, admin, SearchOptions{HouseID: house.ID, Limit: 20})
	if err != nil || total != 0 || len(items) != 0 {
		t.Fatalf("expected all spirits to be hidden, items=%d total=%d err=%v", len(items), total, err)
	}
}

func TestCreateTabletAttachesExistingUnplacedSpiritsAtomically(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), func() time.Time { return time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC) })
	admin := Actor{ID: "admin", Role: "admin"}
	house, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh gắn nhanh"})
	if err != nil {
		t.Fatal(err)
	}
	area, err := service.CreateArea(ctx, admin, AreaInput{HouseID: house.ID, Code: "A"})
	if err != nil {
		t.Fatal(err)
	}
	firstPosition, err := service.CreatePosition(ctx, admin, PositionInput{AreaID: area.ID, RowNumber: 1, ColumnNumber: 1})
	if err != nil {
		t.Fatal(err)
	}
	secondPosition, err := service.CreatePosition(ctx, admin, PositionInput{AreaID: area.ID, RowNumber: 1, ColumnNumber: 2})
	if err != nil {
		t.Fatal(err)
	}
	unplaced, err := service.CreateSpirit(ctx, admin, SpiritInput{HouseID: house.ID, FullName: "Nguyễn Văn Chưa Xếp"})
	if err != nil {
		t.Fatal(err)
	}
	matches, total, err := service.ListSpirits(ctx, admin, SearchOptions{HouseID: house.ID, Query: "chua xep", Unplaced: true, Limit: 20})
	if err != nil || total != 1 || len(matches) != 1 || matches[0].ID != unplaced.ID {
		t.Fatalf("unexpected unplaced search: total=%d items=%#v err=%v", total, matches, err)
	}
	tablet, err := service.CreateTablet(ctx, admin, TabletInput{
		PositionID:        firstPosition.ID,
		Name:              "Bài vị kết hợp",
		ExistingSpiritIDs: []string{unplaced.ID},
		Spirits:           []SpiritInput{{FullName: "Hương linh nhập mới"}},
	})
	if err != nil || tablet.SpiritCount != 2 {
		t.Fatalf("unexpected combined tablet: %#v err=%v", tablet, err)
	}
	attached, total, err := service.ListSpirits(ctx, admin, SearchOptions{TabletID: tablet.ID, Limit: 20})
	if err != nil || total != 2 || len(attached) != 2 {
		t.Fatalf("unexpected attached spirits: total=%d items=%#v err=%v", total, attached, err)
	}
	if _, err = service.CreateTablet(ctx, admin, TabletInput{PositionID: secondPosition.ID, Name: "Không được tạo", ExistingSpiritIDs: []string{unplaced.ID}}); !errors.Is(err, ErrConflict) {
		t.Fatalf("already attached spirit should conflict, got %v", err)
	}
	tablets, err := service.ListTablets(ctx, admin, secondPosition.ID)
	if err != nil || len(tablets) != 0 {
		t.Fatalf("conflicted transaction must not create tablet: %#v err=%v", tablets, err)
	}
}

func TestDeletePositionKeepsTabletsUnplacedAndTheyCanBeMoved(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), func() time.Time { return time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC) })
	admin := Actor{ID: "admin", Role: "admin"}
	house, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh chuyển bài vị"})
	if err != nil {
		t.Fatal(err)
	}
	area, err := service.CreateArea(ctx, admin, AreaInput{HouseID: house.ID, Code: "A"})
	if err != nil {
		t.Fatal(err)
	}
	from, err := service.CreatePosition(ctx, admin, PositionInput{AreaID: area.ID, RowNumber: 1, ColumnNumber: 1})
	if err != nil {
		t.Fatal(err)
	}
	to, err := service.CreatePosition(ctx, admin, PositionInput{AreaID: area.ID, RowNumber: 1, ColumnNumber: 2})
	if err != nil {
		t.Fatal(err)
	}
	tablet, err := service.CreateTablet(ctx, admin, TabletInput{PositionID: from.ID, Name: "Bài vị giữ lại", Spirits: []SpiritInput{{FullName: "Nguyễn Văn A"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.DeletePosition(ctx, admin, from.ID); err != nil {
		t.Fatal(err)
	}
	unplaced, err := service.ListUnplacedTablets(ctx, admin, house.ID, "giu lai")
	if err != nil || len(unplaced) != 1 || unplaced[0].ID != tablet.ID || unplaced[0].SpiritCount != 1 {
		t.Fatalf("unexpected unplaced tablets: %#v, err=%v", unplaced, err)
	}
	if err = service.MoveTablet(ctx, admin, tablet.ID, to.ID); err != nil {
		t.Fatal(err)
	}
	placed, err := service.ListTablets(ctx, admin, to.ID)
	if err != nil || len(placed) != 1 || placed[0].ID != tablet.ID {
		t.Fatalf("unexpected moved tablet: %#v, err=%v", placed, err)
	}
}

func TestEditorAndViewerRespectHouseScope(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	service := NewService(store, time.Now)
	admin := Actor{ID: "admin", Role: "admin"}
	houseA, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh A"})
	if err != nil {
		t.Fatal(err)
	}
	houseB, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh B"})
	if err != nil {
		t.Fatal(err)
	}
	store.members[houseA.ID] = map[string]bool{"editor": true, "viewer": true}

	editor := Actor{ID: "editor", Role: "editor"}
	if _, err = service.CreateArea(ctx, editor, AreaInput{HouseID: houseA.ID, Code: "A"}); err != nil {
		t.Fatalf("editor should write assigned house: %v", err)
	}
	if _, err = service.CreateArea(ctx, editor, AreaInput{HouseID: houseB.ID, Code: "B"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("editor should not access unassigned house: %v", err)
	}

	viewer := Actor{ID: "viewer", Role: "viewer"}
	if _, err = service.ListAreas(ctx, viewer, houseA.ID); err != nil {
		t.Fatalf("viewer should read assigned house: %v", err)
	}
	if _, err = service.CreateArea(ctx, viewer, AreaInput{HouseID: houseA.ID, Code: "C"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer write should be forbidden: %v", err)
	}
}
