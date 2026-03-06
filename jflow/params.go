package jflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

func (pl *Plan) GetParams(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 1 {
		panic(pl.VM.NewGoError(fmt.Errorf("jflow.GetParams requires exactly one schema argument")))
	}

	schema := map[string]any{}
	if err := pl.VM.ExportTo(call.Arguments[0], &schema); err != nil {
		panic(pl.VM.NewGoError(fmt.Errorf("jflow.GetParams schema must be an object: %w", err)))
	}

	validated, err := pl.validateParamsAgainstSchema(schema, pl.Parameters, "")
	if err != nil {
		panic(pl.VM.NewGoError(err))
	}

	return pl.VM.ToValue(validated)
}

func (pl *Plan) validateParamsAgainstSchema(schema map[string]any, params map[string]any, prefix string) (map[string]any, error) {
	out := map[string]any{}

	for key, schemaValue := range schema {
		fullPath := key
		if prefix != "" {
			fullPath = prefix + "." + key
		}

		rawValue, ok := params[key]
		if !ok {
			return nil, fmt.Errorf("missing required parameter %q", fullPath)
		}

		switch typedSchema := schemaValue.(type) {
		case string:
			normalizedValue, err := pl.coerceParamByType(fullPath, typedSchema, rawValue)
			if err != nil {
				return nil, err
			}
			out[key] = normalizedValue

		case map[string]any:
			nestedParams, ok := rawValue.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("parameter %q must be an object", fullPath)
			}
			normalizedNested, err := pl.validateParamsAgainstSchema(typedSchema, nestedParams, fullPath)
			if err != nil {
				return nil, err
			}
			out[key] = normalizedNested

		default:
			return nil, fmt.Errorf("invalid schema type at %q: expected string or object", fullPath)
		}
	}

	return out, nil
}

func (pl *Plan) coerceParamByType(path string, schemaType string, value any) (any, error) {
	switch strings.ToLower(strings.TrimSpace(schemaType)) {
	case "any":
		return value, nil
	case "string":
		if _, ok := value.(string); ok {
			return value, nil
		}
		return nil, fmt.Errorf("parameter %q must be a string", path)
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return value, nil
		default:
			return nil, fmt.Errorf("parameter %q must be a number", path)
		}
	case "boolean", "bool":
		if _, ok := value.(bool); ok {
			return value, nil
		}
		return nil, fmt.Errorf("parameter %q must be a boolean", path)
	case "array":
		if _, ok := value.([]any); ok {
			return value, nil
		}
		return nil, fmt.Errorf("parameter %q must be an array", path)
	case "object":
		if _, ok := value.(map[string]any); ok {
			return value, nil
		}
		return nil, fmt.Errorf("parameter %q must be an object", path)
	case "file":
		return pl.normalizeFileParam(path, value)
	default:
		return nil, fmt.Errorf("unknown schema type %q at %q", schemaType, path)
	}
}

func (pl *Plan) normalizeFileParam(path string, value any) (any, error) {
	if valueStr, ok := value.(string); ok {
		file := &File{Type: inferFileType(valueStr), Path: valueStr}
		if file.Type != FileTypeLocal {
			return file.Path, nil
		}

		normalizedPath := file.Path
		if !filepath.IsAbs(normalizedPath) {
			normalizedPath = filepath.Join(filepath.Dir(pl.Path), normalizedPath)
		}

		absPath, err := filepath.Abs(normalizedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve file parameter %q: %w", path, err)
		}

		return absPath, nil
	}

	file := fileFromValue(value)
	if file == nil {
		return nil, fmt.Errorf("parameter %q must be a file path string or file-like object", path)
	}

	if file.Type != FileTypeLocal {
		return file.Path, nil
	}

	normalizedPath := file.Path
	if !filepath.IsAbs(normalizedPath) {
		normalizedPath = filepath.Join(filepath.Dir(pl.Path), normalizedPath)
	}

	absPath, err := filepath.Abs(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file parameter %q: %w", path, err)
	}

	return absPath, nil
}
