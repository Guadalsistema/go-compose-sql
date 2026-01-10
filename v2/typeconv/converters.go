package typeconv

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// Common converter functions that can be reused across dialects

// StringToTime converts a string to time.Time
// Tries multiple common formats
func StringToTime(source interface{}) (interface{}, error) {
	s, ok := source.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", source)
	}

	// Try common timestamp formats
	formats := []string{
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:04:05.999999999Z07:00",
		time.RFC3339,
		time.RFC3339Nano,
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("cannot parse time string %q: %w", s, lastErr)
}

// StringToNullTime converts a string to sql.NullTime
func StringToNullTime(source interface{}) (interface{}, error) {
	if source == nil {
		return sql.NullTime{Valid: false}, nil
	}

	t, err := StringToTime(source)
	if err != nil {
		return sql.NullTime{Valid: false}, err
	}

	return sql.NullTime{Time: t.(time.Time), Valid: true}, nil
}

// Int64ToTime converts Unix timestamp (int64) to time.Time
func Int64ToTime(source interface{}) (interface{}, error) {
	i, ok := source.(int64)
	if !ok {
		return nil, fmt.Errorf("expected int64, got %T", source)
	}

	return time.Unix(i, 0), nil
}

// Int64ToNullTime converts Unix timestamp (int64) to sql.NullTime
func Int64ToNullTime(source interface{}) (interface{}, error) {
	if source == nil {
		return sql.NullTime{Valid: false}, nil
	}

	t, err := Int64ToTime(source)
	if err != nil {
		return sql.NullTime{Valid: false}, err
	}

	return sql.NullTime{Time: t.(time.Time), Valid: true}, nil
}

// TimeToTime is a pass-through converter (for completeness)
func TimeToTime(source interface{}) (interface{}, error) {
	t, ok := source.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expected time.Time, got %T", source)
	}
	return t, nil
}

// TimeToNullTime converts time.Time to sql.NullTime
func TimeToNullTime(source interface{}) (interface{}, error) {
	if source == nil {
		return sql.NullTime{Valid: false}, nil
	}

	t, ok := source.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expected time.Time, got %T", source)
	}

	return sql.NullTime{Time: t, Valid: true}, nil
}

// DefaultTimeConverter handles multiple source types for time.Time target
func DefaultTimeConverter(source interface{}) (interface{}, error) {
	switch v := source.(type) {
	case time.Time:
		return v, nil
	case string:
		return StringToTime(v)
	case int64:
		return Int64ToTime(v)
	case []byte:
		return StringToTime(string(v))
	default:
		return nil, fmt.Errorf("cannot convert %T to time.Time", source)
	}
}

// DefaultNullTimeConverter handles multiple source types for sql.NullTime target
func DefaultNullTimeConverter(source interface{}) (interface{}, error) {
	if source == nil {
		return sql.NullTime{Valid: false}, nil
	}

	switch v := source.(type) {
	case time.Time:
		return sql.NullTime{Time: v, Valid: true}, nil
	case string:
		t, err := StringToTime(v)
		if err != nil {
			return sql.NullTime{Valid: false}, err
		}
		return sql.NullTime{Time: t.(time.Time), Valid: true}, nil
	case int64:
		t, err := Int64ToTime(v)
		if err != nil {
			return sql.NullTime{Valid: false}, err
		}
		return sql.NullTime{Time: t.(time.Time), Valid: true}, nil
	case []byte:
		t, err := StringToTime(string(v))
		if err != nil {
			return sql.NullTime{Valid: false}, err
		}
		return sql.NullTime{Time: t.(time.Time), Valid: true}, nil
	case sql.NullTime:
		return v, nil
	default:
		return sql.NullTime{Valid: false}, fmt.Errorf("cannot convert %T to sql.NullTime", source)
	}
}

// Int64ToBool converts SQLite integer (0/1) to bool
func Int64ToBool(source interface{}) (interface{}, error) {
	i, ok := source.(int64)
	if !ok {
		return nil, fmt.Errorf("expected int64, got %T", source)
	}
	return i != 0, nil
}

// StringToBool converts string to bool
func StringToBool(source interface{}) (interface{}, error) {
	s, ok := source.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", source)
	}

	return strconv.ParseBool(s)
}

// DefaultBoolConverter handles multiple source types for bool target
func DefaultBoolConverter(source interface{}) (interface{}, error) {
	switch v := source.(type) {
	case bool:
		return v, nil
	case int64:
		return Int64ToBool(v)
	case string:
		return StringToBool(v)
	case []byte:
		return StringToBool(string(v))
	default:
		return nil, fmt.Errorf("cannot convert %T to bool", source)
	}
}
