package config

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"reflect"
)

type Validator interface {
	Validate() error
}

func LoadConfig(v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("must provide a pointer")
	}
	t = t.Elem()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if e, ok := f.Tag.Lookup("env"); ok {
			if err := viper.BindEnv(f.Name, e); err != nil {
				return err // rip
			}
		}

		if e, ok := f.Tag.Lookup("def"); ok {
			viper.SetDefault(f.Name, e)
		}
	}
	viper.AutomaticEnv()

	if err := viper.Unmarshal(v, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			decodeJson,
		),
	)); err != nil {
		return err
	}

	if impl, ok := v.(Validator); ok {
		if err := impl.Validate(); err != nil {
			return fmt.Errorf("failed to validate config: %w", err)
		}
	}

	return nil
}

func decodeJson(rf reflect.Type, rt reflect.Type, data interface{}) (interface{}, error) {
	if rf.Kind() != reflect.String {
		return data, nil
	}
	if rt.Kind() != reflect.Struct && rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice && rt.Kind() != reflect.Map {
		return data, nil
	}

	raw, ok := data.(string)
	if !ok {
		return data, nil
	}

	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Map {
		var out interface{}
		err := json.Unmarshal([]byte(raw), &out)
		return out, err
	} else {
		out := reflect.New(rt).Elem().Interface()
		err := json.Unmarshal([]byte(raw), &out)
		return out, err
	}
}
