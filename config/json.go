package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// loadJSONFile reads a JSON file and flattens it into a map of uppercase
// underscore-separated keys to string values. Nested objects are flattened:
// {"db": {"host": "localhost"}} becomes {"DB_HOST": "localhost"}.
func loadJSONFile(path string, required bool) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !required {
			return nil, nil
		}
		return nil, errors.BadRequest("config: file not found: " + path)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("config: invalid JSON in %s: %v", path, err))
	}

	result := make(map[string]string)
	flattenJSON("", raw, result)
	return result, nil
}

// flattenJSON recursively flattens a JSON object into uppercase
// underscore-separated keys.
func flattenJSON(prefix string, m map[string]any, result map[string]string) {
	for k, v := range m {
		key := strings.ToUpper(k)
		if prefix != "" {
			key = prefix + "_" + key
		}

		switch val := v.(type) {
		case map[string]any:
			flattenJSON(key, val, result)
		case []any:
			// Convert array to comma-separated string.
			parts := make([]string, 0, len(val))
			for _, item := range val {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
			result[key] = strings.Join(parts, ",")
		case float64:
			// JSON numbers are float64; format as int if no fractional part.
			if val == float64(int64(val)) {
				result[key] = strconv.FormatInt(int64(val), 10)
			} else {
				result[key] = strconv.FormatFloat(val, 'f', -1, 64)
			}
		case bool:
			result[key] = strconv.FormatBool(val)
		case string:
			result[key] = val
		case nil:
			// skip nil values
		default:
			result[key] = fmt.Sprintf("%v", val)
		}
	}
}
