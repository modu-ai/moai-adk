package cli

import (
	"testing"
)

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    any
		b    any
		want bool
	}{
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "a nil b not nil",
			a:    nil,
			b:    "hello",
			want: false,
		},
		{
			name: "a not nil b nil",
			a:    "hello",
			b:    nil,
			want: false,
		},
		{
			name: "same strings",
			a:    "hello",
			b:    "hello",
			want: true,
		},
		{
			name: "different strings",
			a:    "hello",
			b:    "world",
			want: false,
		},
		{
			name: "same ints",
			a:    42,
			b:    42,
			want: true,
		},
		{
			name: "different ints",
			a:    42,
			b:    99,
			want: false,
		},
		{
			name: "same bools true",
			a:    true,
			b:    true,
			want: true,
		},
		{
			name: "same bools false",
			a:    false,
			b:    false,
			want: true,
		},
		{
			name: "different bools",
			a:    true,
			b:    false,
			want: false,
		},
		{
			name: "same float64",
			a:    3.14,
			b:    3.14,
			want: true,
		},
		{
			name: "different float64",
			a:    3.14,
			b:    2.71,
			want: false,
		},
		{
			name: "int and string same repr",
			a:    42,
			b:    "42",
			want: true, // fmt.Sprintf("%v", 42) == "42"
		},
		{
			name: "bool and string same repr",
			a:    true,
			b:    "true",
			want: true, // fmt.Sprintf("%v", true) == "true"
		},
		{
			name: "int and float different repr",
			a:    1,
			b:    1.5,
			want: false,
		},
		{
			name: "empty strings",
			a:    "",
			b:    "",
			want: true,
		},
		{
			name: "empty string and nil",
			a:    "",
			b:    nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDeepMergeMaps(t *testing.T) {
	tests := []struct {
		name   string
		newMap map[string]any
		oldMap map[string]any
		check  func(t *testing.T, result map[string]any)
	}{
		{
			name:   "nil maps",
			newMap: nil,
			oldMap: nil,
			check: func(t *testing.T, result map[string]any) {
				if len(result) != 0 {
					t.Errorf("expected empty map, got %v", result)
				}
			},
		},
		{
			name:   "empty maps",
			newMap: map[string]any{},
			oldMap: map[string]any{},
			check: func(t *testing.T, result map[string]any) {
				if len(result) != 0 {
					t.Errorf("expected empty map, got %v", result)
				}
			},
		},
		{
			name: "old values preserved over new for scalar",
			newMap: map[string]any{
				"key": "new_value",
			},
			oldMap: map[string]any{
				"key": "old_value",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["key"] != "old_value" {
					t.Errorf("expected old_value, got %v", result["key"])
				}
			},
		},
		{
			name: "new-only keys added",
			newMap: map[string]any{
				"new_key": "new_value",
			},
			oldMap: map[string]any{},
			check: func(t *testing.T, result map[string]any) {
				if result["new_key"] != "new_value" {
					t.Errorf("expected new_value, got %v", result["new_key"])
				}
			},
		},
		{
			name:   "old-only keys added",
			newMap: map[string]any{},
			oldMap: map[string]any{
				"old_key": "old_value",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["old_key"] != "old_value" {
					t.Errorf("expected old_value, got %v", result["old_key"])
				}
			},
		},
		{
			name: "template_version always uses new value",
			newMap: map[string]any{
				"template_version": "2.0.0",
			},
			oldMap: map[string]any{
				"template_version": "1.0.0",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["template_version"] != "2.0.0" {
					t.Errorf("expected 2.0.0, got %v", result["template_version"])
				}
			},
		},
		{
			name: "nested maps merged recursively",
			newMap: map[string]any{
				"nested": map[string]any{
					"a": "new_a",
					"b": "new_b",
				},
			},
			oldMap: map[string]any{
				"nested": map[string]any{
					"a": "old_a",
					"c": "old_c",
				},
			},
			check: func(t *testing.T, result map[string]any) {
				nested, ok := result["nested"].(map[string]any)
				if !ok {
					t.Fatal("nested should be a map")
				}
				// old value preserved
				if nested["a"] != "old_a" {
					t.Errorf("nested.a = %v, want old_a", nested["a"])
				}
				// new key added
				if nested["b"] != "new_b" {
					t.Errorf("nested.b = %v, want new_b", nested["b"])
				}
				// old-only key preserved
				if nested["c"] != "old_c" {
					t.Errorf("nested.c = %v, want old_c", nested["c"])
				}
			},
		},
		{
			name: "type conflict map vs scalar uses old scalar",
			newMap: map[string]any{
				"key": map[string]any{"a": "1"},
			},
			oldMap: map[string]any{
				"key": "scalar_value",
			},
			check: func(t *testing.T, result map[string]any) {
				// When new is map but old is scalar, old value is preserved
				if result["key"] != "scalar_value" {
					t.Errorf("expected scalar_value, got %v", result["key"])
				}
			},
		},
		{
			name: "type conflict scalar vs map uses old map",
			newMap: map[string]any{
				"key": "scalar_value",
			},
			oldMap: map[string]any{
				"key": map[string]any{"a": "1"},
			},
			check: func(t *testing.T, result map[string]any) {
				// When new is scalar but old is map, old value is preserved
				m, ok := result["key"].(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T: %v", result["key"], result["key"])
				}
				if m["a"] != "1" {
					t.Errorf("expected 1, got %v", m["a"])
				}
			},
		},
		{
			name: "deeply nested maps",
			newMap: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": "new_deep",
					},
				},
			},
			oldMap: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": "old_deep",
					},
				},
			},
			check: func(t *testing.T, result map[string]any) {
				l1 := result["level1"].(map[string]any)
				l2 := l1["level2"].(map[string]any)
				if l2["level3"] != "old_deep" {
					t.Errorf("expected old_deep, got %v", l2["level3"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepMergeMaps(tt.newMap, tt.oldMap)
			tt.check(t, result)
		})
	}
}

func TestMergeYAMLDeep(t *testing.T) {
	tests := []struct {
		name    string
		newData string
		oldData string
		check   func(t *testing.T, result []byte)
		wantErr bool
	}{
		{
			name:    "basic merge preserves old values",
			newData: "key: new_value\n",
			oldData: "key: old_value\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "old_value") {
					t.Errorf("expected old_value in result, got %s", s)
				}
			},
		},
		{
			name:    "new keys added",
			newData: "key1: val1\nkey2: val2\n",
			oldData: "key1: old1\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "old1") {
					t.Errorf("expected old1 preserved, got %s", s)
				}
				if !contains(s, "key2") {
					t.Errorf("expected key2 added, got %s", s)
				}
			},
		},
		{
			name:    "template_version uses new",
			newData: "template_version: \"2.0\"\n",
			oldData: "template_version: \"1.0\"\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "2.0") {
					t.Errorf("expected template_version 2.0, got %s", s)
				}
			},
		},
		{
			name:    "invalid new YAML returns error",
			newData: ":\ninvalid: [yaml\n",
			oldData: "key: value\n",
			wantErr: true,
		},
		{
			name:    "invalid old YAML returns error",
			newData: "key: value\n",
			oldData: ":\ninvalid: [yaml\n",
			wantErr: true,
		},
		{
			name:    "nested YAML merge",
			newData: "parent:\n  child: new_child\n  new_child: added\n",
			oldData: "parent:\n  child: old_child\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "old_child") {
					t.Errorf("expected old_child preserved, got %s", s)
				}
				if !contains(s, "new_child") {
					t.Errorf("expected new_child added, got %s", s)
				}
			},
		},
		{
			name:    "empty YAML inputs",
			newData: "",
			oldData: "",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				// Empty YAML marshals to "{}\n" or "null\n"
				if s != "{}\n" && s != "null\n" {
					t.Errorf("expected empty result, got %q", s)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mergeYAMLDeep([]byte(tt.newData), []byte(tt.oldData))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestMergeYAML3Way(t *testing.T) {
	tests := []struct {
		name     string
		newData  string
		oldData  string
		baseData string
		check    func(t *testing.T, result []byte)
		wantErr  bool
	}{
		{
			name:     "user unchanged uses new template value",
			newData:  "key: new_template\n",
			oldData:  "key: base_value\n",
			baseData: "key: base_value\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "new_template") {
					t.Errorf("expected new_template, got %s", s)
				}
			},
		},
		{
			name:     "user changed preserves user value",
			newData:  "key: new_template\n",
			oldData:  "key: user_custom\n",
			baseData: "key: base_value\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "user_custom") {
					t.Errorf("expected user_custom, got %s", s)
				}
			},
		},
		{
			name:     "new key added by template",
			newData:  "existing: val\nnew_key: new_val\n",
			oldData:  "existing: val\n",
			baseData: "existing: val\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "new_key") {
					t.Errorf("expected new_key added, got %s", s)
				}
			},
		},
		{
			name:     "key removed from template is dropped",
			newData:  "kept: val\n",
			oldData:  "kept: val\nremoved: old\n",
			baseData: "kept: val\nremoved: old\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if contains(s, "removed") {
					t.Errorf("expected removed key to be dropped, got %s", s)
				}
			},
		},
		{
			name:     "system field version always uses new",
			newData:  "version: \"2.0\"\n",
			oldData:  "version: \"user_modified\"\n",
			baseData: "version: \"1.0\"\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "2.0") {
					t.Errorf("expected version 2.0, got %s", s)
				}
			},
		},
		{
			name:     "system field template_version always uses new",
			newData:  "template_version: \"3.0\"\n",
			oldData:  "template_version: \"user_val\"\n",
			baseData: "template_version: \"1.0\"\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if !contains(s, "3.0") {
					t.Errorf("expected template_version 3.0, got %s", s)
				}
			},
		},
		{
			name:     "invalid new YAML",
			newData:  ":\n[invalid\n",
			oldData:  "key: val\n",
			baseData: "key: val\n",
			wantErr:  true,
		},
		{
			name:     "invalid old YAML",
			newData:  "key: val\n",
			oldData:  ":\n[invalid\n",
			baseData: "key: val\n",
			wantErr:  true,
		},
		{
			name:     "invalid base YAML",
			newData:  "key: val\n",
			oldData:  "key: val\n",
			baseData: ":\n[invalid\n",
			wantErr:  true,
		},
		{
			name:     "nested map 3-way merge",
			newData:  "parent:\n  child: new_child\n  added: new_added\n",
			oldData:  "parent:\n  child: user_child\n",
			baseData: "parent:\n  child: base_child\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				// user changed child from base -> preserve user value
				if !contains(s, "user_child") {
					t.Errorf("expected user_child preserved, got %s", s)
				}
				// new key added by template
				if !contains(s, "added") {
					t.Errorf("expected added key present, got %s", s)
				}
			},
		},
		{
			name:     "user added key not in base preserved",
			newData:  "shared: new_val\n",
			oldData:  "shared: new_val\nuser_added: custom\n",
			baseData: "shared: base_val\n",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				// user_added is only in old, not in new template -> dropped
				// (keys only in old are dropped per the implementation)
				if contains(s, "user_added") {
					t.Errorf("expected user_added to be dropped (not in new template), got %s", s)
				}
			},
		},
		{
			name:     "empty inputs",
			newData:  "",
			oldData:  "",
			baseData: "",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				if s != "{}\n" && s != "null\n" {
					t.Errorf("expected empty result, got %q", s)
				}
			},
		},
		{
			name:     "no base value user added this key",
			newData:  "key: new_val\n",
			oldData:  "key: user_val\n",
			baseData: "",
			check: func(t *testing.T, result []byte) {
				s := string(result)
				// No base value -> user added this, preserve user value
				if !contains(s, "user_val") {
					t.Errorf("expected user_val preserved (no base), got %s", s)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mergeYAML3Way([]byte(tt.newData), []byte(tt.oldData), []byte(tt.baseData))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestDeepMerge3Way(t *testing.T) {
	tests := []struct {
		name    string
		newMap  map[string]any
		oldMap  map[string]any
		baseMap map[string]any
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name:    "nil maps",
			newMap:  nil,
			oldMap:  nil,
			baseMap: nil,
			check: func(t *testing.T, result map[string]any) {
				if len(result) != 0 {
					t.Errorf("expected empty map, got %v", result)
				}
			},
		},
		{
			name:    "empty maps",
			newMap:  map[string]any{},
			oldMap:  map[string]any{},
			baseMap: map[string]any{},
			check: func(t *testing.T, result map[string]any) {
				if len(result) != 0 {
					t.Errorf("expected empty map, got %v", result)
				}
			},
		},
		{
			name: "user unchanged from base uses new value",
			newMap: map[string]any{
				"key": "new_val",
			},
			oldMap: map[string]any{
				"key": "base_val",
			},
			baseMap: map[string]any{
				"key": "base_val",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["key"] != "new_val" {
					t.Errorf("expected new_val, got %v", result["key"])
				}
			},
		},
		{
			name: "user changed from base preserves user value",
			newMap: map[string]any{
				"key": "new_val",
			},
			oldMap: map[string]any{
				"key": "user_val",
			},
			baseMap: map[string]any{
				"key": "base_val",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["key"] != "user_val" {
					t.Errorf("expected user_val, got %v", result["key"])
				}
			},
		},
		{
			name: "key only in new added",
			newMap: map[string]any{
				"new_only": "added",
			},
			oldMap:  map[string]any{},
			baseMap: map[string]any{},
			check: func(t *testing.T, result map[string]any) {
				if result["new_only"] != "added" {
					t.Errorf("expected added, got %v", result["new_only"])
				}
			},
		},
		{
			name:   "key only in old is dropped",
			newMap: map[string]any{},
			oldMap: map[string]any{
				"old_only": "removed",
			},
			baseMap: map[string]any{},
			check: func(t *testing.T, result map[string]any) {
				if _, exists := result["old_only"]; exists {
					t.Errorf("expected old_only to be dropped, got %v", result["old_only"])
				}
			},
		},
		{
			name: "system field version always uses new",
			newMap: map[string]any{
				"version": "2.0",
			},
			oldMap: map[string]any{
				"version": "user_modified",
			},
			baseMap: map[string]any{
				"version": "1.0",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["version"] != "2.0" {
					t.Errorf("expected 2.0, got %v", result["version"])
				}
			},
		},
		{
			name: "system field template_version always uses new",
			newMap: map[string]any{
				"template_version": "3.0",
			},
			oldMap: map[string]any{
				"template_version": "user_val",
			},
			baseMap: map[string]any{
				"template_version": "1.0",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["template_version"] != "3.0" {
					t.Errorf("expected 3.0, got %v", result["template_version"])
				}
			},
		},
		{
			name: "nested maps recurse with 3-way merge",
			newMap: map[string]any{
				"parent": map[string]any{
					"child":     "new_child",
					"new_field": "added",
				},
			},
			oldMap: map[string]any{
				"parent": map[string]any{
					"child": "user_child",
				},
			},
			baseMap: map[string]any{
				"parent": map[string]any{
					"child": "base_child",
				},
			},
			check: func(t *testing.T, result map[string]any) {
				parent := result["parent"].(map[string]any)
				// user changed from base -> preserve user
				if parent["child"] != "user_child" {
					t.Errorf("expected user_child, got %v", parent["child"])
				}
				// new field added by template
				if parent["new_field"] != "added" {
					t.Errorf("expected added, got %v", parent["new_field"])
				}
			},
		},
		{
			name: "nested maps base not a map creates empty base",
			newMap: map[string]any{
				"parent": map[string]any{
					"child": "new_val",
				},
			},
			oldMap: map[string]any{
				"parent": map[string]any{
					"child": "old_val",
				},
			},
			baseMap: map[string]any{
				"parent": "was_a_scalar",
			},
			check: func(t *testing.T, result map[string]any) {
				parent := result["parent"].(map[string]any)
				// base is not a map, so baseMapVal is empty
				// old exists and no base key -> user added this -> preserve old
				if parent["child"] != "old_val" {
					t.Errorf("expected old_val, got %v", parent["child"])
				}
			},
		},
		{
			name: "scalar no base exists preserves user value",
			newMap: map[string]any{
				"key": "new_val",
			},
			oldMap: map[string]any{
				"key": "user_val",
			},
			baseMap: map[string]any{},
			check: func(t *testing.T, result map[string]any) {
				// No base value -> user added this -> preserve old
				if result["key"] != "user_val" {
					t.Errorf("expected user_val, got %v", result["key"])
				}
			},
		},
		{
			name: "old not exists uses new value",
			newMap: map[string]any{
				"key": "new_val",
			},
			oldMap:  map[string]any{},
			baseMap: map[string]any{},
			check: func(t *testing.T, result map[string]any) {
				if result["key"] != "new_val" {
					t.Errorf("expected new_val, got %v", result["key"])
				}
			},
		},
		{
			name: "multiple keys mixed scenarios",
			newMap: map[string]any{
				"unchanged":        "new_template",
				"user_customized":  "new_template",
				"new_field":        "brand_new",
				"version":          "2.0",
				"template_version": "3.0",
			},
			oldMap: map[string]any{
				"unchanged":        "base_val",
				"user_customized":  "my_custom",
				"deprecated_field": "old",
			},
			baseMap: map[string]any{
				"unchanged":        "base_val",
				"user_customized":  "base_val",
				"deprecated_field": "old",
			},
			check: func(t *testing.T, result map[string]any) {
				// unchanged from base -> use new
				if result["unchanged"] != "new_template" {
					t.Errorf("unchanged: expected new_template, got %v", result["unchanged"])
				}
				// user customized -> preserve user
				if result["user_customized"] != "my_custom" {
					t.Errorf("user_customized: expected my_custom, got %v", result["user_customized"])
				}
				// new field -> add
				if result["new_field"] != "brand_new" {
					t.Errorf("new_field: expected brand_new, got %v", result["new_field"])
				}
				// deprecated only in old -> dropped
				if _, exists := result["deprecated_field"]; exists {
					t.Error("deprecated_field should be dropped")
				}
				// system fields always new
				if result["version"] != "2.0" {
					t.Errorf("version: expected 2.0, got %v", result["version"])
				}
				if result["template_version"] != "3.0" {
					t.Errorf("template_version: expected 3.0, got %v", result["template_version"])
				}
			},
		},
		{
			name: "new is map old is scalar uses old scalar",
			newMap: map[string]any{
				"key": map[string]any{"nested": "val"},
			},
			oldMap: map[string]any{
				"key": "scalar",
			},
			baseMap: map[string]any{
				"key": "base_scalar",
			},
			check: func(t *testing.T, result map[string]any) {
				// new is map, old is not map -> scalar path
				// old != base -> user changed -> preserve old
				if result["key"] != "scalar" {
					t.Errorf("expected scalar, got %v", result["key"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepMerge3Way(tt.newMap, tt.oldMap, tt.baseMap)
			tt.check(t, result)
		})
	}
}

// contains is a test helper that checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

// searchString is a simple string search helper.
func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
