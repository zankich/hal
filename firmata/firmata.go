package firmata

import (
	"fmt"
	"io"

	"github.com/tarm/serial"

	"golang.org/x/exp/io/i2c/driver"
)

const (
	START_SYSEX     = 0xF0
	I2C_CONFIG      = 0x78
	I2C_REQUEST     = 0x76
	I2C_REPLY       = 0x77
	END_SYSEX       = 0xF7
	REPORT_FIRMWARE = 0x79
)

type Firmata struct {
	Conn io.ReadWriteCloser
}

type firmataConn struct {
	addr int
	f    io.ReadWriteCloser
}

func New(address string, baud int) *Firmata {
	s, err := serial.OpenPort(&serial.Config{Name: address, Baud: baud})
	if err != nil {
		panic(err)
	}

	for {
		buf := make([]byte, 1)
		_, err := s.Read(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}
		break
	}

	return &Firmata{Conn: s}
}

func (f *Firmata) Open(addr int, tenbit bool) (driver.Conn, error) {
	fc := &firmataConn{addr: addr, f: f.Conn}
	fc.addr = addr

	payload := []byte{START_SYSEX, I2C_CONFIG, 0x00, 0x00, END_SYSEX}

	if _, err := fc.f.Write(payload); err != nil {
		return nil, err
	}

	return fc, nil
}

func (f *firmataConn) Tx(w, r []byte) error {
	if w != nil {
		payload := []byte{START_SYSEX, I2C_REQUEST, byte(f.addr), 0x00}

		for _, b := range w {
			payload = append(payload, b&0x7F)
			payload = append(payload, byte(int(b>>7)&0x7F))
		}

		payload = append(payload, END_SYSEX)

		if _, err := f.f.Write(payload); err != nil {
			return err
		}
	}

	if r != nil {
		payload := []byte{START_SYSEX, I2C_REQUEST, byte(f.addr), 0x01 << 3}

		for _, b := range w {
			payload = append(payload, b&0x7F)
			payload = append(payload, byte(int(b>>7)&0x7F))
		}

		payload = append(payload, END_SYSEX)

		if _, err := f.f.Write(payload); err != nil {
			return err
		}

		i2cReply := []byte{}
		for {
			buf, err := f.readByte()
			if err != nil {
				return err
			}

			if buf == START_SYSEX {
				for {
					b, err := f.readByte()
					if err != nil {
						return err
					}

					if b == I2C_REPLY {
						for {
							irb, err := f.readByte()
							if err != nil {
								return err
							}
							if irb == END_SYSEX {
								ret := parseI2cReply(i2cReply)
								if int(ret.address) == f.addr {
									copy(r, ret.data)
								}
								return nil
							}
							i2cReply = append(i2cReply, irb)
						}
					}
				}
			}
		}
	}

	return nil
}

type i2cReply struct {
	address  byte
	register byte
	data     []byte
}

func parseI2cReply(b []byte) i2cReply {
	r := i2cReply{
		address:  (b[0] & 0x7F) | ((b[1] & 0x7F) << 7),
		register: (b[2] & 0x7F) | ((b[3] & 0x7F) << 7),
	}

	for i := 4; i < len(b); i += 2 {
		r.data = append(r.data, b[i]|b[i+1]<<7)
	}
	return r
}

func (*firmataConn) Close() error {
	panic("not implemented")
}

func (f *firmataConn) readByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := f.f.Read(buf)
	if err != nil {
		return 0x00, err
	}

	return buf[0], nil
}
