package grovepi

import (
	"fmt"
	"strings"
	"time"

	gpiodriver "golang.org/x/exp/io/gpio/driver"
	"golang.org/x/exp/io/i2c"
	"golang.org/x/exp/io/i2c/driver"
)

const (
	addr = 0x04
	A0   = "A0"
	A1   = "A1"
	A2   = "A2"
	D2   = "D2"
	D3   = "D3"
	D4   = "D4"
	D5   = "D5"
	D6   = "D6"
	D7   = "D7"
	D8   = "D8"

	digitalRead  = 1
	digitalWrite = 2
	analogRead   = 3
	analogWrite  = 4
	pinMode      = 5
	dhtRead      = 40
)

var (
	pinMap = map[string]int{
		A0: 0,
		A1: 1,
		A2: 2,
		D2: 2,
		D3: 3,
		D4: 4,
		D5: 5,
		D6: 6,
		D7: 7,
		D8: 8,
	}
)

type GrovePI struct {
	Conn   *i2c.Device
	pinMap map[string]int
}

func New(i2co driver.Opener) (*GrovePI, error) {
	conn, err := i2c.Open(i2co, addr)
	if err != nil {
		return nil, err
	}

	return &GrovePI{Conn: conn}, nil
}

func (g *GrovePI) Open() (gpiodriver.Conn, error) {
	return g, nil
}

// Value returns the value of the pin. 0 for low values, 1 for high.
func (g *GrovePI) Value(pin string) (int, error) { return 0, nil }

// SetValue sets the value of the pin. 0 for low values, 1 for high.
func (g *GrovePI) SetValue(pin string, v int) error {
	if !strings.HasPrefix(pin, "D") {
		return fmt.Errorf("%s is an unsupported pin; only digial pins are supported", pin)
	}

	if err := g.Conn.Write([]byte{1, pinMode, byte(pinMap[pin]), 1, 0}); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	if err := g.Conn.Write([]byte{1, digitalWrite, byte(pinMap[pin]), byte(v), 0}); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	return nil
}

// SetDirection sets the direction of the pin.
func (g *GrovePI) SetDirection(pin string, dir gpiodriver.Direction) error { return nil }

// Map should map a virtual GPIO pin number to a physical pin number.
// This is also useful to configure driver implementations for boards
// with different GPIO pin layouts. E.g. GPIO 25 pin on a Raspberry Pi,
// can be represented by a different physical pin out on a different
// board.
func (g *GrovePI) Map(virtual string, physical int) {}

// Close closes the connection and frees the underlying resources.
func (g *GrovePI) Close() error { return nil }
