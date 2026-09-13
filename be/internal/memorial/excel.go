package memorial

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	maxSpiritImportRows   = 5000
	maxSpiritImportErrors = 100
	maxImportNameLength   = 200
	maxImportTextLength   = 1000
)

var (
	spiritImportHeaders = []string{
		"Họ tên",
		"Pháp danh",
		"Năm sinh",
		"Năm mất",
		"Tuổi",
		"Ảnh URL",
		"Nơi an táng",
		"Người gửi",
		"Tháng gửi",
		"Ghi chú",
		"Vị trí",
		"Bài vị",
	}
	positionPattern = regexp.MustCompile(`^(\d+)([A-Z]+)-(\d+)$`)
	yearPattern     = regexp.MustCompile(`^\d{4}$`)
	monthPattern    = regexp.MustCompile(`^(0[1-9]|1[0-2])/(\d{4})$`)
)

type spiritImportRow struct {
	rowNumber   int
	fullName    string
	dharmaName  string
	birthYear   string
	deathYear   string
	age         string
	imageURL    string
	burialPlace string
	sender      string
	sentMonth   string
	notes       string
	position    string
	tablet      string
}

type spiritImportPlanRow struct {
	input        SpiritInput
	areaCode     string
	rowNumber    int
	columnNumber int
	hasPosition  bool
	tabletName   string
}

type spiritImportPlan struct {
	preview SpiritImportPreview
	rows    []spiritImportPlanRow
}

type memorialLookup struct {
	areasByCode    map[string]Area
	positionsByKey map[string]Position
	tabletsByKey   map[string]Tablet
}

func (s *Service) SpiritImportTemplate() ([]byte, error) {
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	sheet := file.GetSheetName(file.GetActiveSheetIndex())
	for index, header := range spiritImportHeaders {
		cell, err := excelize.CoordinatesToCellName(index+1, 1)
		if err != nil {
			return nil, err
		}
		if err = file.SetCellValue(sheet, cell, header); err != nil {
			return nil, err
		}
	}
	sample := []string{
		"Nguyễn Văn A",
		"Thiện Tâm",
		"1948",
		"2025",
		"78",
		"https://example.com/huong-linh-a.jpg",
		"Nghĩa trang Hoa Viên",
		"Gia đình Nguyễn Văn A",
		"08/2026",
		"Thông tin mẫu",
		"38D-10",
		"",
	}
	for index, value := range sample {
		cell, err := excelize.CoordinatesToCellName(index+1, 2)
		if err != nil {
			return nil, err
		}
		if err = file.SetCellValue(sheet, cell, value); err != nil {
			return nil, err
		}
	}
	if err := file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(sheet, "A", "L", 18); err != nil {
		return nil, err
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *Service) PreviewSpiritImport(ctx context.Context, actor Actor, houseID string, raw []byte) (SpiritImportPreview, error) {
	plan, err := s.planSpiritImport(ctx, actor, houseID, raw)
	if err != nil {
		return SpiritImportPreview{}, err
	}
	return plan.preview, nil
}

func (s *Service) ImportSpirits(ctx context.Context, actor Actor, houseID string, raw []byte) (SpiritImportResult, error) {
	plan, err := s.planSpiritImport(ctx, actor, houseID, raw)
	if err != nil {
		return SpiritImportResult{}, err
	}
	if plan.preview.InvalidRows > 0 {
		return SpiritImportResult{}, fmt.Errorf("%w: %s", ErrInvalidInput, formatSpiritImportErrors(plan.preview))
	}
	if optimized, ok := s.store.(interface {
		ImportSpiritsAtomic(context.Context, string, []spiritImportPlanRow, time.Time) (SpiritImportResult, error)
	}); ok {
		return optimized.ImportSpiritsAtomic(ctx, houseID, plan.rows, s.now().UTC())
	}
	lookup, err := s.loadMemorialLookup(ctx, actor, houseID)
	if err != nil {
		return SpiritImportResult{}, err
	}
	result := SpiritImportResult{}
	for _, row := range plan.rows {
		if row.hasPosition {
			if _, created, err := s.ensureImportArea(ctx, actor, houseID, row.areaCode, &lookup); err != nil {
				return SpiritImportResult{}, err
			} else if created {
				result.CreatedAreaCount++
			}
		}
	}
	positionsByArea := map[string][]Position{}
	for _, row := range plan.rows {
		if row.hasPosition {
			area := lookup.areasByCode[row.areaCode]
			key := memorialPositionKey(area.ID, row.rowNumber, row.columnNumber)
			if _, ok := lookup.positionsByKey[key]; !ok {
				lookup.positionsByKey[key] = Position{ID: newID("position"), AreaID: area.ID, RowNumber: row.rowNumber, ColumnNumber: row.columnNumber, Name: fmt.Sprintf("%d%s-%d", row.columnNumber, row.areaCode, row.rowNumber), CreatedAt: s.now().UTC(), UpdatedAt: s.now().UTC()}
				positionsByArea[area.ID] = append(positionsByArea[area.ID], lookup.positionsByKey[key])
			}
		}
	}
	for _, positions := range positionsByArea {
		created, err := s.store.CreatePositions(ctx, positions)
		if err != nil {
			return SpiritImportResult{}, err
		}
		result.CreatedPositionCount += len(created)
	}
	lookup, err = s.loadMemorialLookup(ctx, actor, houseID)
	if err != nil {
		return SpiritImportResult{}, err
	}
	newTablets := []Tablet{}
	for _, row := range plan.rows {
		if row.hasPosition {
			area := lookup.areasByCode[row.areaCode]
			position := lookup.positionsByKey[memorialPositionKey(area.ID, row.rowNumber, row.columnNumber)]
			key := memorialTabletKey(position.ID, row.tabletName)
			if _, ok := lookup.tabletsByKey[key]; !ok {
				t := Tablet{ID: newID("tablet"), PositionID: position.ID, Name: row.tabletName, CreatedAt: s.now().UTC(), UpdatedAt: s.now().UTC()}
				lookup.tabletsByKey[key] = t
				newTablets = append(newTablets, t)
			}
		}
	}
	if len(newTablets) > 0 {
		created, err := s.store.CreateTablets(ctx, newTablets)
		if err != nil {
			return SpiritImportResult{}, err
		}
		result.CreatedTabletCount = len(created)
	}
	lookup, err = s.loadMemorialLookup(ctx, actor, houseID)
	if err != nil {
		return SpiritImportResult{}, err
	}
	inputs := make([]SpiritInput, 0, len(plan.rows))
	for _, row := range plan.rows {
		input := row.input
		if row.hasPosition {
			area := lookup.areasByCode[row.areaCode]
			position := lookup.positionsByKey[memorialPositionKey(area.ID, row.rowNumber, row.columnNumber)]
			tablet := lookup.tabletsByKey[memorialTabletKey(position.ID, row.tabletName)]
			input.TabletID = tablet.ID
		}
		inputs = append(inputs, input)
	}
	for start := 0; start < len(inputs); start += 500 {
		end := start + 500
		if end > len(inputs) {
			end = len(inputs)
		}
		created, createErr := s.CreateSpirits(ctx, actor, inputs[start:end])
		if createErr != nil {
			return SpiritImportResult{}, createErr
		}
		result.CreatedSpiritCount += len(created)
	}
	return result, nil
}

func (s *Service) ExportSpirits(ctx context.Context, actor Actor, scope string, options SearchOptions) ([]byte, error) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "current"
	}
	if scope != "current" && scope != "all" {
		return nil, fmt.Errorf("%w: scope must be current or all", ErrInvalidInput)
	}
	exportOptions := options
	if scope == "all" {
		exportOptions = SearchOptions{}
	}
	items, err := s.collectSpirits(ctx, actor, exportOptions)
	if err != nil {
		return nil, err
	}
	sortSpiritsForExport(items)
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	sheet := file.GetSheetName(file.GetActiveSheetIndex())
	headers := []string{
		"Họ tên",
		"Pháp danh",
		"Năm sinh",
		"Năm mất",
		"Tuổi",
		"Nhà Linh",
		"Khu vực",
		"Vị trí",
		"Bài vị",
		"Nơi an táng",
		"Người gửi",
		"Tháng gửi",
		"Ghi chú",
		"Ảnh URL",
		"Ngày tạo",
		"Cập nhật",
	}
	for index, header := range headers {
		cell, cellErr := excelize.CoordinatesToCellName(index+1, 1)
		if cellErr != nil {
			return nil, cellErr
		}
		if err = file.SetCellValue(sheet, cell, header); err != nil {
			return nil, err
		}
	}
	for index, item := range items {
		rowIndex := index + 2
		values := []any{
			item.FullName,
			item.DharmaName,
			item.BirthYear,
			item.DeathYear,
			item.Age,
			item.HouseName,
			item.AreaCode,
			item.PositionName,
			item.TabletName,
			item.BurialPlace,
			item.Sender,
			item.SentMonth,
			item.Notes,
			item.ImageURL,
			item.CreatedAt.Format("2006-01-02 15:04:05"),
			item.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		for columnIndex, value := range values {
			cell, cellErr := excelize.CoordinatesToCellName(columnIndex+1, rowIndex)
			if cellErr != nil {
				return nil, cellErr
			}
			if err = file.SetCellValue(sheet, cell, value); err != nil {
				return nil, err
			}
		}
	}
	if err := file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(sheet, "A", "N", 18); err != nil {
		return nil, err
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *Service) collectSpirits(ctx context.Context, actor Actor, options SearchOptions) ([]Spirit, error) {
	options.Limit = 500
	options.Offset = 0
	items := make([]Spirit, 0, 500)
	for {
		page, total, err := s.ListSpirits(ctx, actor, options)
		if err != nil {
			return nil, err
		}
		items = append(items, page...)
		options.Offset += len(page)
		if len(page) == 0 || options.Offset >= total {
			break
		}
	}
	return items, nil
}

func (s *Service) planSpiritImport(ctx context.Context, actor Actor, houseID string, raw []byte) (spiritImportPlan, error) {
	houseID = strings.TrimSpace(houseID)
	if houseID == "" {
		return spiritImportPlan{}, fmt.Errorf("%w: house_id is required", ErrInvalidInput)
	}
	if err := s.requireWrite(ctx, actor, houseID); err != nil {
		return spiritImportPlan{}, err
	}
	rows, err := parseSpiritImportRows(raw)
	if err != nil {
		return spiritImportPlan{}, err
	}
	if len(rows) > maxSpiritImportRows {
		return spiritImportPlan{}, fmt.Errorf("%w: import supports up to %d rows", ErrInvalidInput, maxSpiritImportRows)
	}
	lookup, err := s.loadMemorialLookup(ctx, actor, houseID)
	if err != nil {
		return spiritImportPlan{}, err
	}
	existingKeys, err := s.existingSpiritImportKeys(ctx, actor, houseID)
	if err != nil {
		return spiritImportPlan{}, err
	}
	preview := SpiritImportPreview{TotalRows: len(rows), Errors: []SpiritImportIssue{}}
	plannedRows := make([]spiritImportPlanRow, 0, len(rows))
	seenKeys := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		planned, rowErr := buildSpiritImportPlanRow(houseID, row)
		if rowErr != nil {
			preview.InvalidRows++
			appendSpiritImportError(&preview, row.rowNumber, rowErr.Error())
			continue
		}
		key := spiritImportIdentity(planned)
		if _, exists := seenKeys[key]; exists {
			preview.InvalidRows++
			appendSpiritImportError(&preview, row.rowNumber, "bản ghi trùng với một dòng khác trong file")
			continue
		}
		seenKeys[key] = struct{}{}
		if _, exists := existingKeys[key]; exists {
			preview.InvalidRows++
			appendSpiritImportError(&preview, row.rowNumber, "bản ghi đã tồn tại trong Nhà Linh này")
			continue
		}
		if planned.hasPosition {
			if _, exists := lookup.areasByCode[planned.areaCode]; !exists {
				lookup.areasByCode[planned.areaCode] = Area{ID: "preview-area-" + planned.areaCode, HouseID: houseID, Code: planned.areaCode}
				preview.CreateAreaCount++
			}
			area := lookup.areasByCode[planned.areaCode]
			positionKey := memorialPositionKey(area.ID, planned.rowNumber, planned.columnNumber)
			if _, exists := lookup.positionsByKey[positionKey]; !exists {
				position := Position{
					ID:           "preview-position-" + positionKey,
					AreaID:       area.ID,
					RowNumber:    planned.rowNumber,
					ColumnNumber: planned.columnNumber,
					Name:         fmt.Sprintf("%d%s-%d", planned.columnNumber, planned.areaCode, planned.rowNumber),
				}
				lookup.positionsByKey[positionKey] = position
				preview.CreatePositionCount++
			}
			position := lookup.positionsByKey[positionKey]
			tabletKey := memorialTabletKey(position.ID, planned.tabletName)
			if _, exists := lookup.tabletsByKey[tabletKey]; !exists {
				lookup.tabletsByKey[tabletKey] = Tablet{ID: "preview-tablet-" + strconv.Itoa(row.rowNumber), PositionID: position.ID, Name: planned.tabletName}
				preview.CreateTabletCount++
			}
		}
		preview.ValidRows++
		preview.CreateSpiritCount++
		plannedRows = append(plannedRows, planned)
	}
	return spiritImportPlan{preview: preview, rows: plannedRows}, nil
}

func buildSpiritImportPlanRow(houseID string, row spiritImportRow) (spiritImportPlanRow, error) {
	fullName := strings.TrimSpace(row.fullName)
	if fullName == "" {
		return spiritImportPlanRow{}, fmt.Errorf("họ tên là bắt buộc")
	}
	if err := validateSpiritImportValues(row); err != nil {
		return spiritImportPlanRow{}, err
	}
	tabletName := strings.TrimSpace(row.tablet)
	if tabletName == "" {
		tabletName = fullName
	}
	planned := spiritImportPlanRow{
		input: SpiritInput{
			HouseID:     houseID,
			FullName:    fullName,
			DharmaName:  strings.TrimSpace(row.dharmaName),
			BirthYear:   strings.TrimSpace(row.birthYear),
			DeathYear:   strings.TrimSpace(row.deathYear),
			Age:         strings.TrimSpace(row.age),
			ImageURL:    strings.TrimSpace(row.imageURL),
			BurialPlace: strings.TrimSpace(row.burialPlace),
			Sender:      strings.TrimSpace(row.sender),
			SentMonth:   strings.TrimSpace(row.sentMonth),
			Notes:       strings.TrimSpace(row.notes),
		},
		tabletName: tabletName,
	}
	if strings.TrimSpace(row.position) == "" {
		if strings.TrimSpace(row.tablet) != "" {
			return spiritImportPlanRow{}, fmt.Errorf("không thể gán bài vị khi vị trí đang để trống")
		}
		return planned, nil
	}
	rowNumber, areaCode, columnNumber, err := parsePositionReference(row.position)
	if err != nil {
		return spiritImportPlanRow{}, fmt.Errorf("dòng %d: %v", row.rowNumber, err)
	}
	planned.hasPosition = true
	planned.rowNumber = rowNumber
	planned.areaCode = areaCode
	planned.columnNumber = columnNumber
	return planned, nil
}

func (s *Service) loadMemorialLookup(ctx context.Context, actor Actor, houseID string) (memorialLookup, error) {
	if optimized, ok := s.store.(interface {
		ImportLookup(context.Context, string) ([]Area, []Position, []Tablet, error)
	}); ok {
		areas, positions, tablets, err := optimized.ImportLookup(ctx, houseID)
		if err != nil {
			return memorialLookup{}, err
		}
		lookup := memorialLookup{areasByCode: map[string]Area{}, positionsByKey: map[string]Position{}, tabletsByKey: map[string]Tablet{}}
		for _, area := range areas {
			lookup.areasByCode[strings.ToUpper(strings.TrimSpace(area.Code))] = area
		}
		for _, position := range positions {
			lookup.positionsByKey[memorialPositionKey(position.AreaID, position.RowNumber, position.ColumnNumber)] = position
		}
		for _, tablet := range tablets {
			lookup.tabletsByKey[memorialTabletKey(tablet.PositionID, tablet.Name)] = tablet
		}
		return lookup, nil
	}
	areas, err := s.ListAreas(ctx, actor, houseID)
	if err != nil {
		return memorialLookup{}, err
	}
	lookup := memorialLookup{
		areasByCode:    make(map[string]Area, len(areas)),
		positionsByKey: map[string]Position{},
		tabletsByKey:   map[string]Tablet{},
	}
	for _, area := range areas {
		lookup.areasByCode[strings.ToUpper(strings.TrimSpace(area.Code))] = area
		positions, listErr := s.ListPositions(ctx, actor, area.ID)
		if listErr != nil {
			return memorialLookup{}, listErr
		}
		for _, position := range positions {
			lookup.positionsByKey[memorialPositionKey(area.ID, position.RowNumber, position.ColumnNumber)] = position
			tablets, tabletErr := s.ListTablets(ctx, actor, position.ID)
			if tabletErr != nil {
				return memorialLookup{}, tabletErr
			}
			for _, tablet := range tablets {
				lookup.tabletsByKey[memorialTabletKey(position.ID, tablet.Name)] = tablet
			}
		}
	}
	return lookup, nil
}

func (s *Service) ensureImportArea(ctx context.Context, actor Actor, houseID, areaCode string, lookup *memorialLookup) (Area, bool, error) {
	if area, exists := lookup.areasByCode[areaCode]; exists {
		return area, false, nil
	}
	area, err := s.CreateArea(ctx, actor, AreaInput{HouseID: houseID, Code: areaCode})
	if err != nil {
		if !Is(err, ErrConflict) {
			return Area{}, false, err
		}
		areas, listErr := s.ListAreas(ctx, actor, houseID)
		if listErr != nil {
			return Area{}, false, listErr
		}
		for _, existing := range areas {
			lookup.areasByCode[strings.ToUpper(strings.TrimSpace(existing.Code))] = existing
		}
		if existing, exists := lookup.areasByCode[areaCode]; exists {
			return existing, false, nil
		}
		return Area{}, false, err
	}
	lookup.areasByCode[areaCode] = area
	return area, true, nil
}

func (s *Service) ensureImportPosition(ctx context.Context, actor Actor, area Area, rowNumber, columnNumber int, lookup *memorialLookup) (Position, bool, error) {
	key := memorialPositionKey(area.ID, rowNumber, columnNumber)
	if position, exists := lookup.positionsByKey[key]; exists {
		return position, false, nil
	}
	position, err := s.CreatePosition(ctx, actor, PositionInput{AreaID: area.ID, RowNumber: rowNumber, ColumnNumber: columnNumber})
	if err != nil {
		if !Is(err, ErrConflict) {
			return Position{}, false, err
		}
		positions, listErr := s.ListPositions(ctx, actor, area.ID)
		if listErr != nil {
			return Position{}, false, listErr
		}
		for _, existing := range positions {
			lookup.positionsByKey[memorialPositionKey(area.ID, existing.RowNumber, existing.ColumnNumber)] = existing
		}
		if existing, exists := lookup.positionsByKey[key]; exists {
			return existing, false, nil
		}
		return Position{}, false, err
	}
	lookup.positionsByKey[key] = position
	return position, true, nil
}

func (s *Service) ensureImportTablet(ctx context.Context, position Position, name string, lookup *memorialLookup) (Tablet, bool, error) {
	key := memorialTabletKey(position.ID, name)
	if tablet, exists := lookup.tabletsByKey[key]; exists {
		return tablet, false, nil
	}
	now := s.now().UTC()
	tablet, err := s.store.CreateTablet(ctx, Tablet{
		ID:         newID("tablet"),
		PositionID: position.ID,
		Name:       strings.TrimSpace(name),
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		if !Is(err, ErrConflict) {
			return Tablet{}, false, err
		}
		tablets, listErr := s.store.ListTablets(ctx, Actor{Role: "admin", AllHouses: true}, position.ID)
		if listErr != nil {
			return Tablet{}, false, listErr
		}
		for _, existing := range tablets {
			lookup.tabletsByKey[memorialTabletKey(position.ID, existing.Name)] = existing
		}
		if existing, exists := lookup.tabletsByKey[key]; exists {
			return existing, false, nil
		}
		return Tablet{}, false, err
	}
	lookup.tabletsByKey[key] = tablet
	return tablet, true, nil
}

func parseSpiritImportRows(raw []byte) ([]spiritImportRow, error) {
	file, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: file excel không hợp lệ", ErrInvalidInput)
	}
	defer func() { _ = file.Close() }()
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("%w: file excel không có sheet dữ liệu", ErrInvalidInput)
	}
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("%w: không đọc được sheet dữ liệu", ErrInvalidInput)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: file excel không có dữ liệu", ErrInvalidInput)
	}
	headerIndex := make(map[string]int, len(rows[0]))
	for index, value := range rows[0] {
		headerIndex[fold(strings.TrimSpace(value))] = index
	}
	requiredHeaders := []string{"Họ tên", "Vị trí", "Bài vị"}
	for _, header := range requiredHeaders {
		if _, ok := headerIndex[fold(header)]; !ok {
			return nil, fmt.Errorf("%w: thiếu cột %q trong file excel", ErrInvalidInput, header)
		}
	}
	get := func(row []string, header string) string {
		index, ok := headerIndex[fold(header)]
		if !ok || index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}
	items := make([]spiritImportRow, 0, len(rows)-1)
	for index := 1; index < len(rows); index++ {
		row := rows[index]
		item := spiritImportRow{
			rowNumber:   index + 1,
			fullName:    get(row, "Họ tên"),
			dharmaName:  get(row, "Pháp danh"),
			birthYear:   get(row, "Năm sinh"),
			deathYear:   get(row, "Năm mất"),
			age:         get(row, "Tuổi"),
			imageURL:    get(row, "Ảnh URL"),
			burialPlace: get(row, "Nơi an táng"),
			sender:      get(row, "Người gửi"),
			sentMonth:   get(row, "Tháng gửi"),
			notes:       get(row, "Ghi chú"),
			position:    get(row, "Vị trí"),
			tablet:      get(row, "Bài vị"),
		}
		if isSpiritImportRowEmpty(item) {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: file excel không có dòng dữ liệu", ErrInvalidInput)
	}
	return items, nil
}

func validateSpiritImportValues(row spiritImportRow) error {
	if err := validateImportText("họ tên", row.fullName, maxImportNameLength); err != nil {
		return err
	}
	for _, field := range []struct{ label, value string }{
		{"pháp danh", row.dharmaName},
		{"ảnh URL", row.imageURL},
		{"nơi an táng", row.burialPlace},
		{"người gửi", row.sender},
		{"ghi chú", row.notes},
		{"bài vị", row.tablet},
	} {
		if err := validateImportText(field.label, field.value, maxImportTextLength); err != nil {
			return err
		}
	}
	birthYear, err := validateImportYear("năm sinh", row.birthYear)
	if err != nil {
		return err
	}
	deathYear, err := validateImportYear("năm mất", row.deathYear)
	if err != nil {
		return err
	}
	if birthYear > 0 && deathYear > 0 && birthYear > deathYear {
		return fmt.Errorf("năm sinh không được lớn hơn năm mất")
	}
	if age := strings.TrimSpace(row.age); age != "" {
		value, parseErr := strconv.Atoi(age)
		if parseErr != nil || value < 0 || value > 200 {
			return fmt.Errorf("tuổi phải là số nguyên từ 0 đến 200")
		}
	}
	if month := strings.TrimSpace(row.sentMonth); month != "" && !monthPattern.MatchString(month) {
		return fmt.Errorf("tháng gửi phải theo định dạng MM/YYYY")
	}
	if imageURL := strings.TrimSpace(row.imageURL); imageURL != "" {
		parsed, parseErr := url.ParseRequestURI(imageURL)
		if parseErr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("ảnh URL phải là địa chỉ http hoặc https hợp lệ")
		}
	}
	return nil
}

func validateImportText(label, value string, maxLength int) error {
	if len([]rune(strings.TrimSpace(value))) > maxLength {
		return fmt.Errorf("%s không được vượt quá %d ký tự", label, maxLength)
	}
	return nil
}

func validateImportYear(label, raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}
	if !yearPattern.MatchString(value) {
		return 0, fmt.Errorf("%s phải gồm đúng 4 chữ số", label)
	}
	year, err := strconv.Atoi(value)
	if err != nil || year < 1 {
		return 0, fmt.Errorf("%s không hợp lệ", label)
	}
	return year, nil
}

func spiritImportIdentity(row spiritImportPlanRow) string {
	position, tablet := "", ""
	if row.hasPosition {
		position = fmt.Sprintf("%d%s-%d", row.columnNumber, row.areaCode, row.rowNumber)
		tablet = row.tabletName
	}
	return normalizedImportIdentity(row.input.FullName, row.input.DharmaName, row.input.BirthYear, row.input.DeathYear, position, tablet)
}

func normalizedImportIdentity(fullName, dharmaName, birthYear, deathYear, position, tablet string) string {
	parts := []string{fullName, dharmaName, birthYear, deathYear, position, tablet}
	for index, part := range parts {
		parts[index] = fold(strings.TrimSpace(part))
	}
	return strings.Join(parts, "|")
}

func (s *Service) existingSpiritImportKeys(ctx context.Context, actor Actor, houseID string) (map[string]struct{}, error) {
	keys := map[string]struct{}{}
	options := SearchOptions{HouseID: houseID, Limit: 500}
	for {
		items, total, err := s.ListSpirits(ctx, actor, options)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			keys[normalizedImportIdentity(item.FullName, item.DharmaName, item.BirthYear, item.DeathYear, item.PositionName, item.TabletName)] = struct{}{}
		}
		options.Offset += len(items)
		if len(items) == 0 || options.Offset >= total {
			return keys, nil
		}
	}
}

func isSpiritImportRowEmpty(row spiritImportRow) bool {
	return strings.TrimSpace(row.fullName) == "" &&
		strings.TrimSpace(row.dharmaName) == "" &&
		strings.TrimSpace(row.birthYear) == "" &&
		strings.TrimSpace(row.deathYear) == "" &&
		strings.TrimSpace(row.age) == "" &&
		strings.TrimSpace(row.imageURL) == "" &&
		strings.TrimSpace(row.burialPlace) == "" &&
		strings.TrimSpace(row.sender) == "" &&
		strings.TrimSpace(row.sentMonth) == "" &&
		strings.TrimSpace(row.notes) == "" &&
		strings.TrimSpace(row.position) == "" &&
		strings.TrimSpace(row.tablet) == ""
}

func parsePositionReference(raw string) (int, string, int, error) {
	value := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(raw), " ", ""))
	matches := positionPattern.FindStringSubmatch(value)
	if len(matches) != 4 {
		return 0, "", 0, fmt.Errorf("vị trí %q không đúng định dạng 38D-10", raw)
	}
	columnNumber, err := strconv.Atoi(matches[1])
	if err != nil || columnNumber < 1 {
		return 0, "", 0, fmt.Errorf("cột trong vị trí %q không hợp lệ", raw)
	}
	rowNumber, err := strconv.Atoi(matches[3])
	if err != nil || rowNumber < 1 {
		return 0, "", 0, fmt.Errorf("hàng trong vị trí %q không hợp lệ", raw)
	}
	return rowNumber, matches[2], columnNumber, nil
}

func memorialPositionKey(areaID string, rowNumber, columnNumber int) string {
	return areaID + "|" + strconv.Itoa(rowNumber) + "|" + strconv.Itoa(columnNumber)
}

func memorialTabletKey(positionID, name string) string {
	return positionID + "|" + fold(strings.TrimSpace(name))
}

func appendSpiritImportError(preview *SpiritImportPreview, rowNumber int, message string) {
	if len(preview.Errors) >= maxSpiritImportErrors {
		return
	}
	preview.Errors = append(preview.Errors, SpiritImportIssue{RowNumber: rowNumber, Message: message})
}

func formatSpiritImportErrors(preview SpiritImportPreview) string {
	if preview.InvalidRows == 0 {
		return "không có dữ liệu hợp lệ để import"
	}
	parts := make([]string, 0, len(preview.Errors))
	for _, issue := range preview.Errors {
		parts = append(parts, fmt.Sprintf("dòng %d: %s", issue.RowNumber, issue.Message))
	}
	suffix := ""
	if preview.InvalidRows > len(preview.Errors) {
		suffix = fmt.Sprintf(" và %d lỗi khác", preview.InvalidRows-len(preview.Errors))
	}
	return fmt.Sprintf("file import có %d dòng lỗi: %s%s", preview.InvalidRows, strings.Join(parts, "; "), suffix)
}

func sortSpiritsForExport(items []Spirit) {
	sort.Slice(items, func(i, j int) bool {
		left := []string{items[i].HouseName, items[i].AreaCode, items[i].PositionName, items[i].TabletName, items[i].FullName, items[i].ID}
		right := []string{items[j].HouseName, items[j].AreaCode, items[j].PositionName, items[j].TabletName, items[j].FullName, items[j].ID}
		for index := range left {
			compared := strings.Compare(left[index], right[index])
			if compared != 0 {
				return compared < 0
			}
		}
		return false
	})
}
