package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	var result strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		// 1. Обработка экранирования
		if char == '\\' {
			if i+1 >= len(runes) {
				return "", ErrInvalidString
			}
			escapedSymbol := runes[i+1]
			// Можно экранировать только цифру или сам слэш
			if !unicode.IsDigit(escapedSymbol) && escapedSymbol != '\\' {
				return "", ErrInvalidString
			}

			// Проверяем, есть ли за заэкранированным символом множитель
			count, skip, err := parseRepeatCount(runes, i+2)
			if err != nil {
				return "", err
			}

			result.WriteString(strings.Repeat(string(escapedSymbol), count))
			i += 1 + skip
			continue
		}

		// 2. Обычный символ (не цифра)
		if !unicode.IsDigit(char) {
			count, skip, err := parseRepeatCount(runes, i+1)
			if err != nil {
				return "", err
			}
			result.WriteString(strings.Repeat(string(char), count))
			i += skip
			continue
		}

		// 3. Если встретили цифру в начале или после другой цифры
		return "", ErrInvalidString
	}

	return result.String(), nil
}

// Вспомогательная функция для получения числа повторений
func parseRepeatCount(runes []rune, idx int) (int, int, error) {
	if idx >= len(runes) || !unicode.IsDigit(runes[idx]) {
		return 1, 0, nil
	}

	// Проверка на "число" (две цифры подряд)
	if idx+1 < len(runes) && unicode.IsDigit(runes[idx+1]) {
		return 0, 0, ErrInvalidString
	}

	val, _ := strconv.Atoi(string(runes[idx]))
	return val, 1, nil
}
