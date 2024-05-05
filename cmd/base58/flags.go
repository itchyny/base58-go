package main

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

func parseFlags(args []string, opts any) ([]string, error) {
	rest := make([]string, 0, len(args))
	val := reflect.ValueOf(opts).Elem()
	typ := val.Type()
	longToValue := map[string]reflect.Value{}
	shortToValue := map[string]reflect.Value{}
	valueToChoices := map[reflect.Value][]string{}
	for i, l := 0, val.NumField(); i < l; i++ {
		if flag, ok := typ.Field(i).Tag.Lookup("long"); ok {
			longToValue[flag] = val.Field(i)
		}
		if flag, ok := typ.Field(i).Tag.Lookup("short"); ok {
			shortToValue[flag] = val.Field(i)
		}
		if def, ok := typ.Field(i).Tag.Lookup("default"); ok {
			switch val := val.Field(i); val.Kind() {
			case reflect.String:
				val.SetString(def)
			case reflect.Slice:
				val.Set(reflect.Append(val, reflect.ValueOf(def)))
			case reflect.Ptr:
				val.Set(reflect.New(val.Type().Elem()))
				if unmarshaler, ok := val.Interface().(interface {
					UnmarshalFlag(string) error
				}); ok {
					if err := unmarshaler.UnmarshalFlag(def); err != nil {
						return nil, fmt.Errorf(
							"invalid default value %q for flag `%s': %w",
							def, typ.Field(i).Tag.Get("long"), err)
					}
				}
			}
		}
		if choices, ok := typ.Field(i).Tag.Lookup("choices"); ok {
			valueToChoices[val.Field(i)] = strings.Split(choices, ",")
		}
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		var (
			val       reflect.Value
			ok        bool
			shortopts string
		)
		if arg == "--" {
			rest = append(rest, args[i+1:]...)
			break
		}
		if strings.HasPrefix(arg, "--") {
			if val, ok = longToValue[arg[2:]]; !ok {
				if j := strings.IndexByte(arg, '='); j >= 0 {
					if val, ok = longToValue[arg[2:j]]; ok {
						if val.Kind() == reflect.Bool {
							return nil, fmt.Errorf("boolean flag `%s' cannot have an argument", arg[:j])
						}
						args[i] = arg[j+1:]
						arg = arg[:j]
						i--
					}
				}
				if !ok {
					return nil, fmt.Errorf("unknown flag `%s'", arg)
				}
			}
		} else if len(arg) > 1 && arg[0] == '-' {
			var skip bool
			for i := 1; i < len(arg); i++ {
				opt := arg[i : i+1]
				if val, ok = shortToValue[opt]; ok {
					if val.Kind() != reflect.Bool {
						break
					}
				} else if !("A" <= opt && opt <= "Z" || "a" <= opt && opt <= "z") {
					skip = true
					break
				}
			}
			if !skip && (len(arg) > 2 || !ok) {
				shortopts = arg[1:]
				goto L
			}
		}
		if !ok {
			rest = append(rest, arg)
			continue
		}
	S:
		switch val.Kind() {
		case reflect.Bool:
			val.SetBool(true)
		case reflect.String:
			if i++; i >= len(args) {
				return nil, fmt.Errorf("expected argument for flag `%s'", arg)
			}
			val.SetString(args[i])
		case reflect.Slice:
			if i++; i >= len(args) {
				return nil, fmt.Errorf("expected argument for flag `%s'", arg)
			}
			val.Set(reflect.Append(val, reflect.ValueOf(args[i])))
		case reflect.Ptr:
			if i++; i >= len(args) {
				return nil, fmt.Errorf("expected argument for flag `%s'", arg)
			}
			if choices, ok := valueToChoices[val]; ok {
				if !slices.Contains(choices, args[i]) {
					return nil, fmt.Errorf(
						"invalid argument for flag `%s': expected one of [%s] but got %s",
						arg, strings.Join(choices, ", "), args[i])
				}
			}
			val.Set(reflect.New(val.Type().Elem()))
			if unmarshaler, ok := val.Interface().(interface {
				UnmarshalFlag(string) error
			}); ok {
				if err := unmarshaler.UnmarshalFlag(args[i]); err != nil {
					return nil, fmt.Errorf("invalid argument for flag `%s': %w", arg, err)
				}
			}
		}
	L:
		if shortopts != "" {
			opt := shortopts[:1]
			if val, ok = shortToValue[opt]; !ok {
				return nil, fmt.Errorf("unknown flag `%s'", opt)
			}
			if val.Kind() != reflect.Bool && len(shortopts) > 1 {
				if shortopts[1] == '=' {
					args[i] = shortopts[2:]
				} else {
					args[i] = shortopts[1:]
				}
				i--
				shortopts = ""
			} else {
				shortopts = shortopts[1:]
			}
			arg = "-" + opt
			goto S
		}
	}
	return rest, nil
}

func formatFlags(opts any) string {
	val := reflect.ValueOf(opts).Elem()
	typ := val.Type()
	var sb strings.Builder
	sb.WriteString("Command Options:\n")
	for i, l := 0, typ.NumField(); i < l; i++ {
		tag := typ.Field(i).Tag
		if i == l-1 {
			sb.WriteString("\nHelp Option:\n")
		}
		sb.WriteString("  ")
		var short bool
		if flag, ok := tag.Lookup("short"); ok {
			sb.WriteString("-")
			sb.WriteString(flag)
			short = true
		} else {
			sb.WriteString("  ")
		}
		m := sb.Len()
		if flag, ok := tag.Lookup("long"); ok {
			if short {
				sb.WriteString(", ")
			} else {
				sb.WriteString("  ")
			}
			sb.WriteString("--")
			sb.WriteString(flag)
			if val.Field(i).Kind() == reflect.Bool {
				sb.WriteString(" ")
			} else {
				sb.WriteString("=")
			}
		} else {
			sb.WriteString("=")
		}
		if choices, ok := tag.Lookup("choices"); ok {
			sb.WriteString("[")
			sb.WriteString(strings.ReplaceAll(choices, ",", "|"))
			sb.WriteString("] ")
		}
		sb.WriteString(strings.Repeat(" ", 37-sb.Len()+m))
		sb.WriteString(tag.Get("description"))
		if def, ok := typ.Field(i).Tag.Lookup("default"); ok {
			sb.WriteString(" (default: ")
			sb.WriteString(def)
			sb.WriteString(")")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
