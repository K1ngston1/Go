package math

import "testing"

// ------------------ Код функцій ------------------

// Add повертає суму двох чисел
func Add(a, b int) int {
	return a + b
}

// LongOperation - приклад тривалої операції
func LongOperation(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}

// ------------------ Табличні тести з підтестами ------------------
func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"both positive", 2, 3, 5},
		{"both negative", -2, -3, -5},
		{"positive + negative", 5, -3, 2},
		{"zero + number", 0, 7, 7},
		{"both zero", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ------------------ Benchmark тест ------------------
func BenchmarkLongOperation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LongOperation(1000000) // тривала операція
	}
}

// ------------------ Fuzz-тест ------------------
func FuzzAdd(f *testing.F) {
	// Стартові значення
	f.Add(1, 2)
	f.Add(-5, 3)

	f.Fuzz(func(t *testing.T, a, b int) {
		result := Add(a, b)
		expected := a + b
		if result != expected {
			t.Errorf("Add(%d, %d) = %d; want %d", a, b, result, expected)
		}
	})
}
