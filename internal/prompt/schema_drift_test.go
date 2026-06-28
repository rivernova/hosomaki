// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"encoding/json"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// the field names declared in each schema skeleton sent to the LLM
// must match the json tags of the struct the code decodes the response into

var schemaBarewordScalar = regexp.MustCompile(`:\s*([A-Za-z_][A-Za-z0-9_]*)`)

func schemaSkeletonToJSON(schema string) string {
	return schemaBarewordScalar.ReplaceAllStringFunc(schema, func(match string) string {
		word := strings.TrimSpace(match[strings.IndexByte(match, ':')+1:])
		switch word {
		case "true", "false", "null":
			return match
		default:
			return ": 0"
		}
	})
}

func structJSONFields(t reflect.Type) []string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	names := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name := jsonFieldName(field)
		if name == "-" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	name := strings.SplitN(tag, ",", 2)[0]
	if name == "" {
		return field.Name
	}
	return name
}

func assertSchemaMatchesType(t *testing.T, path string, schemaValue any, target reflect.Type) {
	t.Helper()
	for target.Kind() == reflect.Pointer {
		target = target.Elem()
	}

	switch target.Kind() {
	case reflect.Struct:
		obj, ok := schemaValue.(map[string]any)
		if !ok {
			t.Errorf("%s: schema declares a non-object where struct %s is expected", path, target.Name())
			return
		}

		schemaKeys := make([]string, 0, len(obj))
		for k := range obj {
			schemaKeys = append(schemaKeys, k)
		}
		sort.Strings(schemaKeys)

		structKeys := structJSONFields(target)
		if !reflect.DeepEqual(schemaKeys, structKeys) {
			t.Errorf("%s: field-name drift\n  schema: %v\n  struct: %v", path, schemaKeys, structKeys)
		}

		for i := range target.NumField() {
			field := target.Field(i)
			if !field.IsExported() {
				continue
			}
			name := jsonFieldName(field)
			if name == "-" {
				continue
			}
			nested, present := obj[name]
			if !present {
				continue
			}
			assertSchemaMatchesType(t, path+"."+name, nested, field.Type)
		}

	case reflect.Slice:
		arr, ok := schemaValue.([]any)
		if !ok {
			t.Errorf("%s: schema declares a non-array where slice is expected", path)
			return
		}
		if len(arr) == 0 {
			return
		}
		assertSchemaMatchesType(t, path+"[0]", arr[0], target.Elem())

	default:
		return
	}
}

func TestSchemaMatchesResultStruct(t *testing.T) {
	cases := []struct {
		name   string
		schema string
		target reflect.Type
	}{
		{"explain", SchemaExplain, reflect.TypeFor[ExplainResult]()},
		{"watch", SchemaWatch, reflect.TypeFor[WatchResult]()},
		{"doctor-full", SchemaDoctorFull, reflect.TypeFor[DoctorResult]()},
		{"doctor-brief", SchemaDoctorBrief, reflect.TypeFor[DoctorBriefResult]()},
		{"status-full", SchemaStatusFull, reflect.TypeFor[StatusResult]()},
		{"status-brief", SchemaStatusBrief, reflect.TypeFor[StatusBriefResult]()},
		{"audit", SchemaAudit, reflect.TypeFor[AuditResult]()},
		{"why", SchemaWhy, reflect.TypeFor[WhyResult]()},
		{"ports", SchemaPorts, reflect.TypeFor[PortsResult]()},
		{"timers", SchemaTimers, reflect.TypeFor[TimersResult]()},
		{"crons", SchemaCrons, reflect.TypeFor[CronsResult]()},
		{"mounts", SchemaMounts, reflect.TypeFor[MountsResult]()},
		{"updates", SchemaUpdates, reflect.TypeFor[UpdatesResult]()},
		{"history", SchemaHistory, reflect.TypeFor[HistoryResult]()},
		{"firewall", SchemaFirewall, reflect.TypeFor[FirewallResult]()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var skeleton any
			if err := json.Unmarshal([]byte(schemaSkeletonToJSON(tc.schema)), &skeleton); err != nil {
				t.Fatalf("schema %q is not parseable as a skeleton: %v", tc.name, err)
			}
			assertSchemaMatchesType(t, tc.name, skeleton, tc.target)
		})
	}
}
