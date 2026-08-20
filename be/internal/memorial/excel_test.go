package memorial

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestPreviewAndImportSpiritsAutoCreateMemorialStructure(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), func() time.Time {
		return time.Date(2026, 8, 20, 8, 30, 0, 0, time.UTC)
	})
	admin := Actor{ID: "admin", Role: "admin"}
	house, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh Import"})
	if err != nil {
		t.Fatal(err)
	}
	file := mustSpiritImportWorkbook(t,
		[]string{"Nguyễn Văn Một", "", "", "", "", "", "", "", "", "", "38D-10", ""},
		[]string{"Nguyễn Văn Hai", "Thiện An", "", "", "", "", "", "", "", "", "38D-10", "Gia tiên họ Nguyễn"},
		[]string{"Nguyễn Văn Ba", "", "", "", "", "", "", "", "", "", "", ""},
	)

	preview, err := service.PreviewSpiritImport(ctx, admin, house.ID, file)
	if err != nil {
		t.Fatal(err)
	}
	if preview.TotalRows != 3 || preview.ValidRows != 3 || preview.InvalidRows != 0 {
		t.Fatalf("unexpected preview counts: %#v", preview)
	}
	if preview.CreateAreaCount != 1 || preview.CreatePositionCount != 1 || preview.CreateTabletCount != 2 || preview.CreateSpiritCount != 3 {
		t.Fatalf("unexpected preview create counters: %#v", preview)
	}

	result, err := service.ImportSpirits(ctx, admin, house.ID, file)
	if err != nil {
		t.Fatal(err)
	}
	if result.CreatedAreaCount != 1 || result.CreatedPositionCount != 1 || result.CreatedTabletCount != 2 || result.CreatedSpiritCount != 3 {
		t.Fatalf("unexpected import result: %#v", result)
	}

	areas, err := service.ListAreas(ctx, admin, house.ID)
	if err != nil || len(areas) != 1 || areas[0].Code != "D" {
		t.Fatalf("unexpected areas after import: %#v err=%v", areas, err)
	}
	positions, err := service.ListPositions(ctx, admin, areas[0].ID)
	if err != nil || len(positions) != 1 || positions[0].Name != "38D-10" {
		t.Fatalf("unexpected positions after import: %#v err=%v", positions, err)
	}
	tablets, err := service.ListTablets(ctx, admin, positions[0].ID)
	if err != nil || len(tablets) != 2 {
		t.Fatalf("unexpected tablets after import: %#v err=%v", tablets, err)
	}
	spirits, total, err := service.ListSpirits(ctx, admin, SearchOptions{HouseID: house.ID, Limit: 20})
	if err != nil || total != 3 || len(spirits) != 3 {
		t.Fatalf("unexpected spirits after import: total=%d items=%#v err=%v", total, spirits, err)
	}
	unplaced, total, err := service.ListSpirits(ctx, admin, SearchOptions{HouseID: house.ID, Unplaced: true, Limit: 20})
	if err != nil || total != 1 || len(unplaced) != 1 || unplaced[0].FullName != "Nguyễn Văn Ba" {
		t.Fatalf("unexpected unplaced spirits after import: total=%d items=%#v err=%v", total, unplaced, err)
	}
}

func TestPreviewSpiritImportRejectsInvalidRows(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), time.Now)
	admin := Actor{ID: "admin", Role: "admin"}
	house, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh Lỗi"})
	if err != nil {
		t.Fatal(err)
	}
	file := mustSpiritImportWorkbook(t,
		[]string{"", "", "", "", "", "", "", "", "", "", "38D-10", ""},
		[]string{"Nguyễn Văn Lỗi", "", "", "", "", "", "", "", "", "", "", "Bài vị chưa có vị trí"},
	)

	preview, err := service.PreviewSpiritImport(ctx, admin, house.ID, file)
	if err != nil {
		t.Fatal(err)
	}
	if preview.ValidRows != 0 || preview.InvalidRows != 2 {
		t.Fatalf("unexpected invalid preview summary: %#v", preview)
	}
	if len(preview.Errors) != 2 {
		t.Fatalf("expected two preview errors, got %#v", preview.Errors)
	}
	if _, err = service.ImportSpirits(ctx, admin, house.ID, file); err == nil {
		t.Fatal("expected import to fail for invalid rows")
	}
}

func TestExportSpiritsRespectsScope(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore(), func() time.Time {
		return time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	})
	admin := Actor{ID: "admin", Role: "admin"}
	houseA, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh A"})
	if err != nil {
		t.Fatal(err)
	}
	houseB, err := service.CreateHouse(ctx, admin, HouseInput{Name: "Nhà Linh B"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.CreateSpirit(ctx, admin, SpiritInput{HouseID: houseA.ID, FullName: "Hương linh A"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.CreateSpirit(ctx, admin, SpiritInput{HouseID: houseB.ID, FullName: "Hương linh B"}); err != nil {
		t.Fatal(err)
	}

	currentData, err := service.ExportSpirits(ctx, admin, "current", SearchOptions{HouseID: houseA.ID})
	if err != nil {
		t.Fatal(err)
	}
	currentRows := readWorkbookRows(t, currentData)
	if len(currentRows) != 2 || currentRows[1][0] != "Hương linh A" {
		t.Fatalf("unexpected current export rows: %#v", currentRows)
	}

	allData, err := service.ExportSpirits(ctx, admin, "all", SearchOptions{HouseID: houseA.ID})
	if err != nil {
		t.Fatal(err)
	}
	allRows := readWorkbookRows(t, allData)
	if len(allRows) != 3 {
		t.Fatalf("unexpected all export rows: %#v", allRows)
	}
}

func mustSpiritImportWorkbook(t *testing.T, rows ...[]string) []byte {
	t.Helper()
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	sheet := file.GetSheetName(file.GetActiveSheetIndex())
	for column, header := range spiritImportHeaders {
		cell, err := excelize.CoordinatesToCellName(column+1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if err = file.SetCellValue(sheet, cell, header); err != nil {
			t.Fatal(err)
		}
	}
	for rowIndex, row := range rows {
		for columnIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+2)
			if err != nil {
				t.Fatal(err)
			}
			if err = file.SetCellValue(sheet, cell, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func readWorkbookRows(t *testing.T, raw []byte) [][]string {
	t.Helper()
	file, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		t.Fatal("expected workbook sheet")
	}
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		t.Fatal(err)
	}
	return rows
}
