package config

import (
	"reflect"
	"testing"
)

func TestGetControllerConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetControllerConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetControllerConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadConfig(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadConfig(tt.args.filePath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
