// Package config is contract C1: the typed, validated configuration of the
// service. One reflection walk over the struct tags of Config drives loading,
// validation and the generated reference (docs/config-reference.md).
//
// Precedence, lowest to highest: the "default" tag, the JSON file named by
// IAM_CONFIG_FILE (decoded with jsonx.DecodeStrict), then the environment.
// An empty environment variable counts as unset: the file value or the
// default still applies, because getenv cannot tell empty from absent.
// The absolute prohibitions of ADR 0012 and the critical alerts are evaluated
// here; a prohibition makes Load fail, an alert is returned as a Finding.
package config
