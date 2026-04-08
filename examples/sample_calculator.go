package examples

// Calculator provides basic arithmetic operations
type Calculator struct {
	memory float64
}

// NewCalculator creates a new calculator instance
func NewCalculator() *Calculator {
	return &Calculator{memory: 0}
}

// Add adds two numbers
func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	c.memory = result
	return result
}

// Subtract subtracts b from a
func (c *Calculator) Subtract(a, b float64) float64 {
	result := a - b
	c.memory = result
	return result
}

// Multiply multiplies two numbers
func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	c.memory = result
	return result
}

// Divide divides a by b
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	result := a / b
	c.memory = result
	return result, nil
}

// GetMemory returns the last calculated value
func (c *Calculator) GetMemory() float64 {
	return c.memory
}

// ClearMemory resets the memory to zero
func (c *Calculator) ClearMemory() {
	c.memory = 0
}

// ErrDivisionByZero is returned when attempting to divide by zero
var ErrDivisionByZero = &CalculatorError{message: "division by zero"}

// CalculatorError represents a calculator error
type CalculatorError struct {
	message string
}

func (e *CalculatorError) Error() string {
	return e.message
}

// Made with Bob
