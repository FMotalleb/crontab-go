package template

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

func buildFuncMap() template.FuncMap {
	result := template.FuncMap{
		"env": os.Getenv,
		"b64enc": func(s string) string {
			return base64.StdEncoding.EncodeToString([]byte(s))
		},
		"sum": func(a int, b int) int {
			return a + b
		},
		"b64dec":    b64dec,
		"toUpper":   strings.ToUpper,
		"toLower":   strings.ToLower,
		"trim":      strings.TrimSpace,
		"join":      strings.Join,
		"replace":   strings.ReplaceAll,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"contains":  strings.Contains,
		"toJSON":    toJSON,
		"fromJSON":  fromJSON,
		"itoa":      strconv.Itoa,
		"toInt":     toInt,
		"atoi":      strconv.Atoi,
		"atob":      atob,
		"read":      readFile,
	}

	return result
}

func EvaluateTemplate(text string, vars any) (string, error) {
	templateObj := template.New("template")

	templateObj = templateObj.Funcs(buildFuncMap())

	templateObj, err := templateObj.Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	output := bytes.NewBufferString("")
	err = templateObj.Execute(output, vars)
	if err != nil {
		return "", fmt.Errorf("failed to execute template using vars snapshot: %w", err)
	}
	return output.String(), nil
}

func toJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func fromJSON(s string) map[string]any {
	var result map[string]any
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		panic(err)
	}
	return result
}

func b64dec(s string) string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func atob(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true":
		return true, nil
	case "0", "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(val).Int())
	case uint, uint8, uint16, uint32, uint64:
		uval := reflect.ValueOf(val).Uint()
		if uval > uint64(^uint(0)>>1) {
			panic(fmt.Errorf("integer overflow: value %d exceeds int range", uval))
		}
		return int(uval)
	case float32:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return int(f)
		}
		panic(fmt.Errorf("cannot convert string to int: %s", val))
	default:
		panic(fmt.Errorf("unsupported type: %T", val))
	}
}

func readFile(path string) string {
	res, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("failed to read file %s: %w", path, err))
	}
	return string(res)
}
