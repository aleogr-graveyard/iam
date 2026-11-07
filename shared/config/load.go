package config

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aleogr/iam/shared/jsonx"
)

// field is one configuration leaf discovered through the struct tags.
type field struct {
	env      string
	key      string // JSON file key
	def      string
	doc      string
	typeName string
	secret   bool
	required bool
	nofile   bool
	value    reflect.Value
}

// fieldsOf walks v (a struct) and returns its tagged leaves in declaration order.
func fieldsOf(v reflect.Value) []field {
	var out []field
	t := v.Type()
	for i := range t.NumField() {
		sf := t.Field(i)
		if sf.Type.Kind() == reflect.Struct && sf.Type != reflect.TypeFor[time.Time]() {
			out = append(out, fieldsOf(v.Field(i))...)
			continue
		}
		env := sf.Tag.Get("env")
		if env == "" {
			continue
		}
		out = append(out, field{
			env:      env,
			key:      strings.ToLower(strings.TrimPrefix(env, "IAM_")),
			def:      sf.Tag.Get("default"),
			doc:      sf.Tag.Get("doc"),
			typeName: typeName(sf.Type),
			secret:   sf.Tag.Get("secret") == "true",
			required: sf.Tag.Get("required") == "true",
			nofile:   sf.Tag.Get("nofile") == "true",
			value:    v.Field(i),
		})
	}
	return out
}

func typeName(t reflect.Type) string {
	switch {
	case t == reflect.TypeFor[time.Duration]():
		return "duration"
	case t.Kind() == reflect.String:
		return "string"
	default:
		return "integer"
	}
}

// set parses s into the field. The error never contains s.
func (f field) set(s string) error {
	switch f.value.Kind() {
	case reflect.String:
		f.value.SetString(s)
	case reflect.Int64: // time.Duration is the only int64 field
		d, err := time.ParseDuration(s)
		if err != nil {
			return errors.New("must be a duration such as 10s")
		}
		f.value.SetInt(int64(d))
	case reflect.Uint32:
		n, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return errors.New("must be a non-negative integer")
		}
		f.value.SetUint(n)
	default:
		return errors.New("unsupported field type")
	}
	return nil
}

// Load builds the configuration. It returns every problem at once: a single
// error naming each invalid field and each prohibited setting. Findings below
// Prohibited are returned for the caller to log. Errors never contain field
// values, so secrets cannot leak through them.
func Load(getenv func(string) string, readFile func(string) ([]byte, error)) (Config, []Finding, error) {
	var cfg Config
	fs := fieldsOf(reflect.ValueOf(&cfg).Elem())
	var problems []string
	problems = append(problems, applyDefaults(fs)...)
	if path := getenv("IAM_CONFIG_FILE"); path != "" {
		problems = append(problems, applyFile(fs, path, readFile)...)
	}
	problems = append(problems, applyEnv(fs, getenv)...)

	findings, vp := validate(&cfg)
	problems = append(problems, vp...)
	for _, fd := range findings {
		if fd.Severity == Prohibited {
			problems = append(problems, fmt.Sprintf("%s: %s (%s)", fd.Code, fd.Message, fd.Reference))
		}
	}
	if len(problems) > 0 {
		return Config{}, findings, errors.New("invalid configuration:\n  " + strings.Join(problems, "\n  "))
	}
	return cfg, findings, nil
}

func applyDefaults(fs []field) []string {
	var problems []string
	for _, f := range fs {
		if f.def == "" {
			continue
		}
		if err := f.set(f.def); err != nil {
			problems = append(problems, f.env+": "+err.Error())
		}
	}
	return problems
}

func applyEnv(fs []field, getenv func(string) string) []string {
	var problems []string
	for _, f := range fs {
		v := getenv(f.env)
		if v == "" {
			continue
		}
		if err := f.set(v); err != nil {
			problems = append(problems, f.env+": "+err.Error())
		}
	}
	return problems
}

// applyFile decodes the JSON file and applies its keys. Keys are the lower-case
// environment names without the IAM_ prefix, e.g. "http_addr".
func applyFile(fs []field, path string, readFile func(string) ([]byte, error)) []string {
	raw, err := readFile(path)
	if err != nil {
		return []string{"IAM_CONFIG_FILE: cannot read the file"}
	}
	var doc map[string]any
	if err := jsonx.DecodeStrict(bytes.NewReader(raw), &doc, jsonx.DefaultLimits); err != nil {
		return []string{"IAM_CONFIG_FILE: " + err.Error()}
	}
	byKey := make(map[string]field, len(fs))
	for _, f := range fs {
		if !f.nofile {
			byKey[f.key] = f
		}
	}
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var problems []string
	for _, k := range keys {
		f, ok := byKey[k]
		if !ok {
			problems = append(problems, fmt.Sprintf("config file: unknown key %q", truncate(k, 64)))
			continue
		}
		s, ok := fileValue(doc[k])
		if !ok {
			problems = append(problems, fmt.Sprintf("config file key %q: must be a string or a number", k))
			continue
		}
		if err := f.set(s); err != nil {
			problems = append(problems, fmt.Sprintf("config file key %q: %s", k, err))
		}
	}
	return problems
}

func fileValue(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	}
	return "", false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
