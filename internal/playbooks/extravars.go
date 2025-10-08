package playbooks

import (
	"fmt"
	"strings"
)

// ParseExtraVars converts key=val strings into a map
func ParseExtraVars(kvs []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, kv := range kvs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid extra-vars entry: %s (expected key=val)", kv)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty key in extra-vars entry: %s", kv)
		}
		m[key] = val
	}
	return m, nil
}
