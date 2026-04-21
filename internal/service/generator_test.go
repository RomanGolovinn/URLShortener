package service

import (
	"strings"
	"testing"
)

func TestRandomGenerator_Generate(t *testing.T) {
	const lenght = 10
	gen := NewRandomGenerator(lenght)
	t.Run("correct length", func(t *testing.T) {
		short, err := gen.Generate()
		if err != nil {
			t.Errorf("неизвестная ошибка %v", err)
		}
		if len(short) != lenght {
			t.Errorf("длина короткой ссылки не равно %d, текущая длинна: %d", lenght, len(short))
		}
	})

	t.Run("randomness sanity check", func(t *testing.T) {
		code1, _ := gen.Generate()
		code2, _ := gen.Generate()

		if code1 == code2 {
			t.Errorf("генератор выдал две одинаковые строки подряд: %s", code1)
		}
	})

	t.Run("valid characters", func(t *testing.T) {
		code, _ := gen.Generate()
		validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

		for _, char := range code {
			if !strings.ContainsRune(validChars, char) {
				t.Errorf("найден недопустимый символ: %c", char)
			}
		}
	})
}
