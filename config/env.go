package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/KARTIKrocks/apikit/errors"
)

// loadEnvFile reads a .env file and returns parsed key-value pairs as a map.
// It does not modify the process environment. Lines starting with # and
// blank lines are skipped. Values may be optionally quoted with single or
// double quotes, which are stripped.
func loadEnvFile(path string, required bool) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && !required {
			return nil, nil
		}
		return nil, errors.BadRequest("config: file not found: " + path)
	}
	defer func() { _ = f.Close() }()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := parseEnvLine(line)
		if !ok {
			continue
		}

		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

// parseEnvLine splits a line on the first '=' and returns key, value, ok.
// It trims surrounding quotes from the value.
func parseEnvLine(line string) (key, value string, ok bool) {
	// Split on first '='.
	var rawVal string
	key, rawVal, ok = strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(rawVal)

	// Strip inline comment (simple: only if not inside quotes).
	if !isQuoted(value) {
		if ci := strings.IndexByte(value, '#'); ci >= 0 {
			value = strings.TrimSpace(value[:ci])
		}
	}

	// Strip surrounding quotes.
	value = stripQuotes(value)

	return key, value, key != ""
}

// isQuoted returns true if s starts and ends with matching quotes.
func isQuoted(s string) bool {
	if len(s) < 2 {
		return false
	}
	return (s[0] == '"' && s[len(s)-1] == '"') ||
		(s[0] == '\'' && s[len(s)-1] == '\'')
}

// stripQuotes removes surrounding single or double quotes.
func stripQuotes(s string) string {
	if isQuoted(s) {
		return s[1 : len(s)-1]
	}
	return s
}

// lookupEnv looks up an environment variable with optional prefix.
// Returns the value and whether it was found.
func lookupEnv(key, prefix string) (string, bool) {
	if prefix != "" {
		key = prefix + "_" + key
	}
	return os.LookupEnv(key)
}

// lookupEnvFile looks up a key in the .env file map with optional prefix.
// Real environment variables take precedence: if the key exists in os env,
// the .env value is skipped (returns not-found so the caller falls through
// to the env-var lookup that already ran).
func lookupEnvFile(key, prefix string, envFileValues map[string]string) (string, bool) {
	if envFileValues == nil {
		return "", false
	}
	if prefix != "" {
		key = prefix + "_" + key
	}
	// Only return .env value if the real env doesn't have this key.
	if _, inEnv := os.LookupEnv(key); inEnv {
		return "", false
	}
	val, ok := envFileValues[key]
	return val, ok
}
