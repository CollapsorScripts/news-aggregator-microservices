package config

import "testing"

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Инициализация env файла",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
		})
		t.Logf("Лог: %+v", Cfg)
	}
}
