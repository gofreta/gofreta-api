package models

import (
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestCollectionForm_ResolveModel(t *testing.T) {
	testScenarios := []struct {
		Model      *Collection
		Title      string
		Name       string
		CreateHook string
		UpdateHook string
		DeleteHook string
		Fields     []CollectionField
	}{
		{nil, "test", "test", "test", "test", "test", []CollectionField{{Key: "duplicate_key", Type: FieldTypePlain, Label: "test", Required: true, Unique: false, Multilingual: true, Default: nil, Meta: nil}}},
		{
			&Collection{
				ID:         bson.ObjectIdHex("507f191e810c19729de860ea"),
				Title:      "",
				Name:       "",
				CreateHook: "",
				UpdateHook: "",
				DeleteHook: "",
				Fields:     nil,
				Created:    1518773370,
				Modified:   1518773370,
			},
			"Test title",
			"test_name",
			"test",
			"test",
			"test",
			[]CollectionField{{Key: "duplicate_key", Type: FieldTypePlain, Label: "test", Required: true, Unique: false, Multilingual: true, Default: nil, Meta: nil}},
		},
	}

	for _, scenario := range testScenarios {
		form := &CollectionForm{
			Model:      scenario.Model,
			Title:      scenario.Title,
			Name:       scenario.Name,
			CreateHook: scenario.CreateHook,
			UpdateHook: scenario.UpdateHook,
			DeleteHook: scenario.DeleteHook,
			Fields:     scenario.Fields,
		}

		resolvedModel := form.ResolveModel()

		if resolvedModel == nil {
			t.Fatal("Expected Collection model pointer, got nil")
		}

		if resolvedModel.Title != scenario.Title {
			t.Errorf("Expected resolved model Title to be %s, got %s", scenario.Title, resolvedModel.Title)
		}

		if resolvedModel.Name != scenario.Name {
			t.Errorf("Expected resolved model Name to be %s, got %s", scenario.Name, resolvedModel.Name)
		}

		if resolvedModel.CreateHook != scenario.CreateHook {
			t.Errorf("Expected resolved model CreateHook to be %s, got %s", scenario.CreateHook, resolvedModel.CreateHook)
		}

		if resolvedModel.UpdateHook != scenario.UpdateHook {
			t.Errorf("Expected resolved model UpdateHook to be %s, got %s", scenario.UpdateHook, resolvedModel.UpdateHook)
		}

		if resolvedModel.DeleteHook != scenario.DeleteHook {
			t.Errorf("Expected resolved model DeleteHook to be %s, got %s", scenario.DeleteHook, resolvedModel.DeleteHook)
		}

		for _, field := range resolvedModel.Fields {
			exist := false
			for _, eField := range scenario.Fields {
				if field.Key == eField.Key {
					exist = true
					break
				}
			}
			if !exist {
				t.Errorf("Unexpected field %v", field)
			}
		}

		if scenario.Model == nil { // new
			if resolvedModel.ID.Hex() == "" {
				t.Error("Expected resolved model id to be set")
			}

			if resolvedModel.Created == 0 {
				t.Error("Expected resolved model created timestamp to be set")
			}

			if resolvedModel.Modified != resolvedModel.Created {
				t.Error("Expected modified and created to be equal")
			}
		} else { // update
			if resolvedModel.ID.Hex() != scenario.Model.ID.Hex() {
				t.Errorf("Expected %s id, got %s", scenario.Model.ID.Hex(), resolvedModel.ID.Hex())
			}

			if resolvedModel.Created != scenario.Model.Created {
				t.Errorf("Expected %d created, got %d", scenario.Model.Created, resolvedModel.Created)
			}

			if resolvedModel.Modified <= scenario.Model.Modified {
				t.Error("Expected modified to be updated")
			}
		}
	}
}

func TestCollectionForm_Validate(t *testing.T) {
	// empty form
	f1 := &CollectionForm{}

	// invalid populated form
	f2 := &CollectionForm{
		Title: "Test title",
		Name:  "Invalid name 123",
		Fields: []CollectionField{
			{Key: "duplicate_key", Type: FieldTypePlain, Label: "test", Required: true, Unique: false, Multilingual: true, Default: nil, Meta: nil},
			{Key: "duplicate_key", Type: "invalid", Label: "test", Required: true, Unique: false, Multilingual: true, Default: nil, Meta: nil},
		},
		CreateHook: "invalid_url",
		UpdateHook: "invalid_url",
		DeleteHook: "invalid_url",
	}

	// valid populated form
	f3 := &CollectionForm{
		Title: "Test title",
		Name:  "test",
		Fields: []CollectionField{
			{Key: "key1", Type: FieldTypePlain, Label: "test", Required: true, Unique: false, Multilingual: true, Default: nil, Meta: nil},
			{Key: "key2", Type: FieldTypePlain, Label: "test", Required: true, Unique: false, Multilingual: true, Default: nil, Meta: nil},
		},
		CreateHook: "",
		UpdateHook: "http://test.dev/update",
		DeleteHook: "http://test.dev/delete",
	}

	testScenarios := []TestValidateScenario{
		{f1, []string{"title", "name", "fields"}},
		{f2, []string{"name", "create_hook", "update_hook", "delete_hook", "fields"}},
		{f3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestUniqueFieldKey(t *testing.T) {
	// --- invalid fields
	fields1 := []CollectionField{
		{Key: "key1"},
		{Key: "key1"},
	}

	result1 := uniqueFieldKey(fields1)

	if result1 == nil {
		t.Error("Expected error, got nil")
	}

	// --- valid fields
	fields2 := []CollectionField{
		{Key: "key1"},
		{Key: "key2"},
	}

	result2 := uniqueFieldKey(fields2)

	if result2 != nil {
		t.Error("Expected nil, got error - ", result2)
	}
}

func TestCollectionField_Validate(t *testing.T) {
	// empty model
	m1 := &CollectionField{}

	// invalid populated model
	m2 := &CollectionField{
		Key:          "Invalid key 123",
		Type:         "invalid",
		Label:        "More than 255 characters...Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book.",
		Required:     false,
		Unique:       true,
		Multilingual: true,
		Default:      nil,
		Meta:         nil,
	}

	// invalid populated model 2 (meta validation call test)
	m3 := &CollectionField{
		Key:          "key",
		Type:         FieldTypeRelation,
		Label:        "Test",
		Required:     false,
		Unique:       true,
		Multilingual: true,
		Default:      nil,
		Meta:         nil,
	}

	// valid populated model
	m4 := &CollectionField{
		Key:          "key",
		Type:         FieldTypePlain,
		Label:        "Test",
		Required:     false,
		Unique:       true,
		Multilingual: true,
		Default:      nil,
		Meta:         nil,
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"key", "type", "label"}},
		{m2, []string{"key", "type", "label"}},
		{m3, []string{"meta"}},
		{m4, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestCollectionField_CastValue(t *testing.T) {
	fieldPlain := CollectionField{Type: FieldTypePlain}
	fieldSwitch := CollectionField{Type: FieldTypeSwitch}
	fieldChecklist := CollectionField{Type: FieldTypeChecklist}
	fieldSelect := CollectionField{Type: FieldTypeSelect}
	fieldDate := CollectionField{Type: FieldTypeDate}
	fieldEditor := CollectionField{Type: FieldTypeEditor}
	fieldMedia := CollectionField{Type: FieldTypeMedia}
	fieldRelation := CollectionField{Type: FieldTypeRelation}

	testScenarios := []struct {
		Field    CollectionField
		Data     interface{}
		Expected interface{}
	}{
		// plain
		{fieldPlain, nil, ""},
		{fieldPlain, 123, ""},
		{fieldPlain, "test", "test"},
		// switch
		{fieldSwitch, nil, false},
		{fieldSwitch, 123, false},
		{fieldSwitch, "test", false},
		{fieldSwitch, false, false},
		{fieldSwitch, true, true},
		// checklist
		{fieldChecklist, nil, []string{}},
		{fieldChecklist, "test", []string{}},
		{fieldChecklist, []interface{}{"test"}, []string{"test"}},
		{fieldChecklist, []string{"test", "test1"}, []string{"test", "test1"}},
		// select
		{fieldSelect, nil, ""},
		{fieldSelect, 123, ""},
		{fieldSelect, "test", "test"},
		// date
		{fieldDate, nil, nil},
		{fieldDate, "1.5", 0},
		{fieldDate, "3", 3},
		{fieldDate, 123, 123},
		{fieldDate, 1.5, 1},
		// editor
		{fieldEditor, nil, ""},
		{fieldEditor, 123, ""},
		{fieldEditor, "test", "test"},
		// media
		{fieldMedia, nil, []bson.ObjectId{}},
		{fieldMedia, "test", []bson.ObjectId{}},
		{fieldMedia, []interface{}{"test"}, []bson.ObjectId{}},
		{fieldMedia, []string{"507f191e810c19729de860ea"}, []bson.ObjectId{}},
		{fieldMedia, []interface{}{"507f191e810c19729de860ea", ""}, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}},
		{fieldMedia, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}},
		// relation
		{fieldRelation, nil, []bson.ObjectId{}},
		{fieldRelation, "test", []bson.ObjectId{}},
		{fieldRelation, []interface{}{"test"}, []bson.ObjectId{}},
		{fieldRelation, []string{"507f191e810c19729de860ea"}, []bson.ObjectId{}},
		{fieldRelation, []interface{}{"507f191e810c19729de860ea", ""}, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}},
		{fieldRelation, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}},
	}

	for _, scenario := range testScenarios {
		result := scenario.Field.CastValue(scenario.Data)
		hasError := false

		switch scenario.Field.Type {
		case FieldTypePlain, FieldTypeSelect, FieldTypeEditor:
			hasError = result.(string) != scenario.Expected.(string)
		case FieldTypeDate:
			if result != nil && scenario.Expected != nil {
				hasError = (result.(int) != scenario.Expected.(int))
			} else {
				hasError = result != scenario.Expected
			}
		case FieldTypeMedia, FieldTypeRelation:
			ids, _ := result.([]bson.ObjectId)

			for _, v := range ids {
				exist := false
				for _, eV := range scenario.Expected.([]bson.ObjectId) {
					if v.Hex() == eV.Hex() {
						exist = true
						break
					}
				}
				if !exist {
					hasError = true
					break
				}
			}
		case FieldTypeChecklist:
			items, _ := result.([]string)

			for _, v := range items {
				exist := false
				for _, eV := range scenario.Expected.([]string) {
					if v == eV {
						exist = true
						break
					}
				}
				if !exist {
					hasError = true
					break
				}
			}
		case FieldTypeSwitch:
			hasError = result.(bool) != scenario.Expected.(bool)
		default:
			hasError = result != scenario.Expected
		}

		if hasError {
			t.Errorf("Expected %v, got %v (scenario: %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestCollectionField_IsEmptyValue(t *testing.T) {
	fieldPlain := CollectionField{Type: FieldTypePlain}
	fieldSwitch := CollectionField{Type: FieldTypeSwitch}
	fieldChecklist := CollectionField{Type: FieldTypeChecklist}
	fieldSelect := CollectionField{Type: FieldTypeSelect}
	fieldDate := CollectionField{Type: FieldTypeDate}
	fieldEditor := CollectionField{Type: FieldTypeEditor}
	fieldMedia := CollectionField{Type: FieldTypeMedia}
	fieldRelation := CollectionField{Type: FieldTypeRelation}

	testScenarios := []struct {
		Field    CollectionField
		Data     interface{}
		Expected bool
	}{
		// plain
		{fieldPlain, 123, true},
		{fieldPlain, nil, true},
		{fieldPlain, "", true},
		{fieldPlain, "123", false},
		// switch
		{fieldSwitch, nil, false},
		{fieldSwitch, 123, false},
		{fieldSwitch, "123", false},
		{fieldSwitch, true, false},
		{fieldSwitch, false, false},
		// checklist
		{fieldChecklist, nil, true},
		{fieldChecklist, "1232", true},
		{fieldChecklist, true, true},
		{fieldChecklist, []string{}, true},
		{fieldChecklist, []string{"", ""}, true},
		{fieldChecklist, []string{"test"}, false},
		// select
		{fieldSelect, 123, true},
		{fieldSelect, nil, true},
		{fieldSelect, "", true},
		{fieldSelect, "123", false},
		// date
		{fieldDate, nil, true},
		{fieldDate, "1.6", true},
		{fieldDate, "123", false},
		{fieldDate, 123, false},
		{fieldDate, 1.5, false},
		// editor
		{fieldEditor, 123, true},
		{fieldEditor, nil, true},
		{fieldEditor, "", true},
		{fieldEditor, "test", false},
		// media
		{fieldMedia, nil, true},
		{fieldMedia, 123, true},
		{fieldMedia, "test", true},
		{fieldMedia, []string{}, true},
		{fieldMedia, []string{""}, true},
		{fieldMedia, []bson.ObjectId{}, true},
		{fieldMedia, []interface{}{}, true},
		{fieldMedia, []interface{}{"507f191e810c19729de860ea"}, false},
		{fieldMedia, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}, false},
		// relation
		{fieldRelation, nil, true},
		{fieldRelation, 123, true},
		{fieldRelation, "test", true},
		{fieldRelation, []string{}, true},
		{fieldRelation, []string{""}, true},
		{fieldRelation, []bson.ObjectId{}, true},
		{fieldRelation, []interface{}{}, true},
		{fieldRelation, []interface{}{"507f191e810c19729de860ea"}, false},
		{fieldRelation, []bson.ObjectId{bson.ObjectIdHex("507f191e810c19729de860ea")}, false},
	}

	for _, scenario := range testScenarios {
		result := scenario.Field.IsEmptyValue(scenario.Data)

		if result != scenario.Expected {
			t.Errorf("Expected %t, got %t (scenario: %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestMetaPlain_Validate(t *testing.T) {
	m1 := &MetaPlain{} // no meta

	testScenarios := []TestValidateScenario{
		{m1, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaSwitch_Validate(t *testing.T) {
	m1 := &MetaSwitch{} // no meta

	testScenarios := []TestValidateScenario{
		{m1, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaChecklist_Validate(t *testing.T) {
	// empty model
	m1 := &MetaChecklist{}

	// invalid populated model
	m2 := &MetaChecklist{
		Options: []MetaChecklistOption{
			{Name: "name"},
			{Name: "name", Value: "value"},
		},
	}

	// valid populated model
	m3 := &MetaChecklist{
		Options: []MetaChecklistOption{
			{Name: "name1", Value: "value1"},
			{Name: "name2", Value: "value2"},
		},
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"options"}},
		{m2, []string{"options"}},
		{m3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaChecklistOption_Validate(t *testing.T) {
	// empty model
	m1 := &MetaChecklistOption{}

	// valid populated model
	m2 := &MetaChecklistOption{
		Name:  "test",
		Value: "test",
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"name", "value"}},
		{m2, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaSelect_Validate(t *testing.T) {
	// empty model
	m1 := &MetaSelect{}

	// invalid populated model
	m2 := &MetaSelect{
		Options: []MetaSelectOption{
			{Name: "name"},
			{Name: "name", Value: "value"},
		},
	}

	// valid populated model
	m3 := &MetaSelect{
		Options: []MetaSelectOption{
			{Name: "name1", Value: "value1"},
			{Name: "name2", Value: "value2"},
		},
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"options"}},
		{m2, []string{"options"}},
		{m3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaSelectOption_Validate(t *testing.T) {
	// empty model
	m1 := &MetaSelectOption{}

	// valid populated model
	m2 := &MetaSelectOption{
		Name:  "test",
		Value: "test",
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"name", "value"}},
		{m2, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaDate_Validate(t *testing.T) {
	// empty model
	m1 := &MetaDate{}

	// invalid populated model
	m2 := &MetaDate{
		Format: "y-m-d",
		Mode:   "invalid",
	}

	// valid populated model
	m3 := &MetaDate{
		Format: "y-m-d",
		Mode:   MetaDateModeDateTime,
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"format", "mode"}},
		{m2, []string{"mode"}},
		{m3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaEditor_Validate(t *testing.T) {
	// empty model
	m1 := &MetaEditor{}

	// invalid populated model
	m2 := &MetaEditor{
		Mode: "invalid",
	}

	// valid populated model
	m3 := &MetaEditor{
		Mode: MetaEditorModeRich,
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"mode"}},
		{m2, []string{"mode"}},
		{m3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaMedia_Validate(t *testing.T) {
	// empty model
	m1 := &MetaMedia{}

	// populated model
	m2 := &MetaMedia{
		Max: 1,
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{}},
		{m2, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMetaRelation_Validate(t *testing.T) {
	// empty model
	m1 := &MetaRelation{}

	// valid populated model
	m2 := &MetaRelation{
		Max:          1,
		CollectionID: bson.ObjectIdHex("507f191e810c19729de860ea"),
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"collection_id"}},
		{m2, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestNewMetaPlain(t *testing.T) {
	testData := []interface{}{
		nil,
		struct {
			Data1 int
			Data2 string
		}{123, "test"},
	}

	for _, data := range testData {
		meta, err := NewMetaPlain(data)

		if err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		if meta == nil {
			t.Error("Expected MetaPlain instance, got nil")
		}
	}
}

func TestNewMetaSwitch(t *testing.T) {
	testData := []interface{}{
		nil,
		struct {
			Data1 int
			Data2 string
		}{123, "test"},
	}

	for _, data := range testData {
		meta, err := NewMetaSwitch(data)

		if err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		if meta == nil {
			t.Error("Expected MetaSwitch instance, got nil")
		}
	}
}

func TestNewMetaChecklist(t *testing.T) {
	type OptionStruct struct {
		Name  interface{} `json:"name"`
		Value interface{} `json:"value"`
	}

	data1 := struct {
		Options []OptionStruct `json:"options"`
	}{
		Options: []OptionStruct{
			{"name", true},
			{123, false},
		},
	}

	expectedOptions1 := []MetaChecklistOption{
		{"name", ""},
		{"", ""},
	}

	data2 := struct {
		Options []OptionStruct `json:"options"`
	}{
		Options: []OptionStruct{
			{"name1", "val1"},
			{"name2", "val2"},
		},
	}

	expectedOptions2 := []MetaChecklistOption{
		{"name1", "val1"},
		{"name2", "val2"},
	}

	// ---

	testScenarios := []struct {
		HasError        bool
		Data            interface{}
		ExpectedOptions []MetaChecklistOption
	}{
		{true, data1, expectedOptions1},
		{false, data2, expectedOptions2},
	}

	for _, scenario := range testScenarios {
		meta, err := NewMetaChecklist(scenario.Data)

		if scenario.HasError && err == nil {
			t.Error("Expected error, got nil")
		} else if !scenario.HasError && err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		for _, option := range meta.Options {
			exist := false

			for _, eOption := range scenario.ExpectedOptions {
				if option.Name == eOption.Name && option.Value == eOption.Value {
					exist = true
					break
				}
			}

			if !exist {
				t.Errorf("Option %v is not expected in %v", option, scenario.ExpectedOptions)
			}
		}
	}
}

func TestNewMetaSelect(t *testing.T) {
	type OptionStruct struct {
		Name  interface{} `json:"name"`
		Value interface{} `json:"value"`
	}

	data1 := struct {
		Options []OptionStruct `json:"options"`
	}{
		Options: []OptionStruct{
			{"name", true},
			{123, false},
		},
	}

	expectedOptions1 := []MetaSelectOption{
		{"name", ""},
		{"", ""},
	}

	data2 := struct {
		Options []OptionStruct `json:"options"`
	}{
		Options: []OptionStruct{
			{"name1", "val1"},
			{"name2", "val2"},
		},
	}

	expectedOptions2 := []MetaSelectOption{
		{"name1", "val1"},
		{"name2", "val2"},
	}

	// ---

	testScenarios := []struct {
		HasError        bool
		Data            interface{}
		ExpectedOptions []MetaSelectOption
	}{
		{true, data1, expectedOptions1},
		{false, data2, expectedOptions2},
	}

	for _, scenario := range testScenarios {
		meta, err := NewMetaSelect(scenario.Data)

		if scenario.HasError && err == nil {
			t.Error("Expected error, got nil")
		} else if !scenario.HasError && err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		for _, option := range meta.Options {
			exist := false

			for _, eOption := range scenario.ExpectedOptions {
				if option.Name == eOption.Name && option.Value == eOption.Value {
					exist = true
					break
				}
			}

			if !exist {
				t.Errorf("Option %v is not expected in %v", option, scenario.ExpectedOptions)
			}
		}
	}
}

func TestNewMetaDate(t *testing.T) {
	type MarshalStruct struct {
		Format interface{} `json:"format"`
		Mode   interface{} `json:"mode"`
	}

	testScenarios := []struct {
		HasError       bool
		Data           interface{}
		ExpectedFormat string
		ExpectedMode   string
	}{
		{true, MarshalStruct{123, true}, "Y-m-d H:i:s", MetaDateModeDateTime},
		{true, MarshalStruct{"test_format", 456}, "test_format", MetaDateModeDateTime},
		{true, MarshalStruct{false, "test_mode"}, "Y-m-d H:i:s", "test_mode"},
		{false, MarshalStruct{"test_format", "test_mode"}, "test_format", "test_mode"},
	}

	for _, scenario := range testScenarios {
		meta, err := NewMetaDate(scenario.Data)

		if scenario.HasError && err == nil {
			t.Error("Expected error, got nil")
		} else if !scenario.HasError && err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		if meta.Format != scenario.ExpectedFormat {
			t.Errorf("Expected %s format, got %s", scenario.ExpectedFormat, meta.Format)
		}

		if meta.Mode != scenario.ExpectedMode {
			t.Errorf("Expected %s mode, got %s", scenario.ExpectedMode, meta.Mode)
		}
	}
}

func TestNewMetaEditor(t *testing.T) {
	type MarshalStruct struct {
		Mode interface{} `json:"mode"`
	}

	testScenarios := []struct {
		HasError     bool
		Data         interface{}
		ExpectedMode string
	}{
		{true, MarshalStruct{123}, MetaEditorModeSimple},
		{true, MarshalStruct{true}, MetaEditorModeSimple},
		{false, MarshalStruct{"test"}, "test"},
		{false, MarshalStruct{MetaEditorModeRich}, MetaEditorModeRich},
	}

	for _, scenario := range testScenarios {
		meta, err := NewMetaEditor(scenario.Data)

		if scenario.HasError && err == nil {
			t.Error("Expected error, got nil")
		} else if !scenario.HasError && err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		if meta.Mode != scenario.ExpectedMode {
			t.Errorf("Expected %s mode, got %s", scenario.ExpectedMode, meta.Mode)
		}
	}
}

func TestNewMetaMedia(t *testing.T) {
	type MarshalStruct struct {
		Max interface{} `json:"max"`
	}

	testScenarios := []struct {
		HasError    bool
		Data        interface{}
		ExpectedMax uint8
	}{
		{true, MarshalStruct{"test"}, 0},
		{true, MarshalStruct{true}, 0},
		{true, MarshalStruct{1.5}, 0},
		{false, nil, 0},
		{false, MarshalStruct{123}, 123},
	}

	for _, scenario := range testScenarios {
		meta, err := NewMetaMedia(scenario.Data)

		if scenario.HasError && err == nil {
			t.Error("Expected error, got nil")
		} else if !scenario.HasError && err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		if meta.Max != scenario.ExpectedMax {
			t.Errorf("Expected %d max, got %d", scenario.ExpectedMax, meta.Max)
		}
	}
}

func TestNewMetaRelation(t *testing.T) {
	type MarshalStruct struct {
		Max          interface{} `json:"max"`
		CollectionID interface{} `json:"collection_id"`
	}

	testScenarios := []struct {
		HasError             bool
		Data                 interface{}
		ExpectedMax          uint8
		ExpectedCollectionID bson.ObjectId
	}{
		{true, MarshalStruct{"test", 123}, 0, ""},
		{true, MarshalStruct{true, "test"}, 0, ""},
		{true, MarshalStruct{1.5, false}, 0, ""},
		{false, nil, 0, ""},
		{false, MarshalStruct{123, "507f191e810c19729de850ea"}, 123, bson.ObjectIdHex("507f191e810c19729de850ea")},
		{false, MarshalStruct{123, bson.ObjectIdHex("507f191e810c19729de860ea")}, 123, bson.ObjectIdHex("507f191e810c19729de860ea")},
	}

	for _, scenario := range testScenarios {
		meta, err := NewMetaRelation(scenario.Data)

		if scenario.HasError && err == nil {
			t.Error("Expected error, got nil")
		} else if !scenario.HasError && err != nil {
			t.Error("Expected nil, got error - ", err)
		}

		if meta.Max != scenario.ExpectedMax {
			t.Errorf("Expected %d max, got %d", scenario.ExpectedMax, meta.Max)
		}

		if meta.CollectionID.Hex() != scenario.ExpectedCollectionID.Hex() {
			t.Errorf("Expected %s collection, got %s", scenario.ExpectedCollectionID.Hex(), meta.CollectionID.Hex())
		}
	}
}
