package examples

import (
	"testing"
)

// Example of auto-generated test by Intelligent Test Orchestrator

// TestNewCalculator_HappyPath tests calculator initialization
func TestNewCalculator_HappyPath(t *testing.T) {
	// Arrange & Act
	calc := NewCalculator()

	// Assert
	if calc == nil {
		t.Error("Expected non-nil calculator")
	}
	if calc.GetMemory() != 0 {
		t.Errorf("Expected initial memory to be 0, got %f", calc.GetMemory())
	}
}

// TestAdd_HappyPath tests addition with valid inputs
func TestAdd_HappyPath(t *testing.T) {
	// Arrange
	calc := NewCalculator()

	// Act
	result := calc.Add(5, 3)

	// Assert
	if result != 8 {
		t.Errorf("Expected 8, got %f", result)
	}
	if calc.GetMemory() != 8 {
		t.Errorf("Expected memory to be 8, got %f", calc.GetMemory())
	}
}

// TestAdd_EdgeCases tests addition with edge cases
func TestAdd_EdgeCases(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"Zero values", 0, 0, 0},
		{"Negative numbers", -5, -3, -8},
		{"Mixed signs", 5, -3, 2},
		{"Large numbers", 1e10, 1e10, 2e10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%f, %f) = %f; want %f", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestSubtract_HappyPath tests subtraction with valid inputs
func TestSubtract_HappyPath(t *testing.T) {
	// Arrange
	calc := NewCalculator()

	// Act
	result := calc.Subtract(10, 3)

	// Assert
	if result != 7 {
		t.Errorf("Expected 7, got %f", result)
	}
}

// TestMultiply_HappyPath tests multiplication with valid inputs
func TestMultiply_HappyPath(t *testing.T) {
	// Arrange
	calc := NewCalculator()

	// Act
	result := calc.Multiply(4, 5)

	// Assert
	if result != 20 {
		t.Errorf("Expected 20, got %f", result)
	}
}

// TestDivide_HappyPath tests division with valid inputs
func TestDivide_HappyPath(t *testing.T) {
	// Arrange
	calc := NewCalculator()

	// Act
	result, err := calc.Divide(10, 2)

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("Expected 5, got %f", result)
	}
}

// TestDivide_ErrorHandling tests division by zero
func TestDivide_ErrorHandling(t *testing.T) {
	// Arrange
	calc := NewCalculator()

	// Act
	result, err := calc.Divide(10, 0)

	// Assert
	if err == nil {
		t.Error("Expected error for division by zero")
	}
	if result != 0 {
		t.Errorf("Expected 0 for error case, got %f", result)
	}
	if err != ErrDivisionByZero {
		t.Errorf("Expected ErrDivisionByZero, got %v", err)
	}
}

// TestGetMemory tests memory retrieval
func TestGetMemory(t *testing.T) {
	// Arrange
	calc := NewCalculator()
	calc.Add(5, 3)

	// Act
	memory := calc.GetMemory()

	// Assert
	if memory != 8 {
		t.Errorf("Expected memory to be 8, got %f", memory)
	}
}

// TestClearMemory tests memory clearing
func TestClearMemory(t *testing.T) {
	// Arrange
	calc := NewCalculator()
	calc.Add(5, 3)

	// Act
	calc.ClearMemory()

	// Assert
	if calc.GetMemory() != 0 {
		t.Errorf("Expected memory to be 0 after clear, got %f", calc.GetMemory())
	}
}

// TestCalculator_Integration tests multiple operations in sequence
func TestCalculator_Integration(t *testing.T) {
	// Arrange
	calc := NewCalculator()

	// Act & Assert - Chain of operations
	result := calc.Add(10, 5)
	if result != 15 {
		t.Errorf("Step 1: Expected 15, got %f", result)
	}

	result = calc.Multiply(result, 2)
	if result != 30 {
		t.Errorf("Step 2: Expected 30, got %f", result)
	}

	result = calc.Subtract(result, 10)
	if result != 20 {
		t.Errorf("Step 3: Expected 20, got %f", result)
	}

	result, err := calc.Divide(result, 4)
	if err != nil {
		t.Errorf("Step 4: Unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("Step 4: Expected 5, got %f", result)
	}

	// Verify memory
	if calc.GetMemory() != 5 {
		t.Errorf("Final memory: Expected 5, got %f", calc.GetMemory())
	}
}

// Benchmark tests
func BenchmarkAdd(b *testing.B) {
	calc := NewCalculator()
	for i := 0; i < b.N; i++ {
		calc.Add(5, 3)
	}
}

func BenchmarkDivide(b *testing.B) {
	calc := NewCalculator()
	for i := 0; i < b.N; i++ {
		calc.Divide(10, 2)
	}
}

// Made with Bob
