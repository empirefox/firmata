package dial

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

// Dial accepts p like:
// tcp://x.x.x.x:xxx?timeout=2s&keep_alive=1 or
// serial:///dev/ttyUSB0?baud=4800&size=8&parity=N&stop_bit=1&timeout=2s or
// serial:///dev/serial/by-path/pci-0000:00:1a.0-usb-0:1.2:1.0-port0?baud=4800
// by `udevadm info /dev/ttyUSB0`
func Dial(p string) (io.ReadWriteCloser, error) {
	u, err := parseDialAddr(p)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "tcp":
		c, err := toTCP(u)
		if err != nil {
			return nil, err
		}

		d := net.Dialer{Timeout: c.timeout, KeepAlive: c.keepAlive}
		return d.Dial("tcp", c.addr)
	case "serial":
		c, err := toSerial(u)
		if err != nil {
			return nil, err
		}
		return serial.OpenPort(c)
	default:
		return nil, fmt.Errorf("schema not support: %s", u.Scheme)
	}
}

func parseDialAddr(p string) (*url.URL, error) {
	return url.ParseRequestURI(p)
}

type tcpConfig struct {
	addr      string
	timeout   time.Duration
	keepAlive time.Duration
}

func toTCP(u *url.URL) (*tcpConfig, error) {
	if u.Host == "" {
		return nil, fmt.Errorf("empty tcp host: %s", u)
	}
	if u.Path != "" {
		return nil, fmt.Errorf("tcp cannot contain path: %s", u)
	}
	c := tcpConfig{
		addr:      u.Host,
		timeout:   0,
		keepAlive: 0,
	}

	var err error
	q := u.Query()

	v := q.Get("timeout")
	if v != "" {
		c.timeout, err = time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %v", u)
		}
	}

	v = q.Get("keep_alive")
	if v != "" {
		c.keepAlive, err = time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid keep_alive: %v", u)
		}
	}

	return &c, nil
}

func toSerial(u *url.URL) (c *serial.Config, err error) {
	if u.Host != "" {
		return nil, fmt.Errorf("serial cannot contain host: %s", u)
	}

	c = &serial.Config{
		Name: u.Path,
	}
	q := u.Query()
	v := q.Get("timeout")
	if v != "" {
		c.ReadTimeout, err = time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %v", u)
		}
	}

	v = q.Get("baud")
	if v != "" {
		c.Baud, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid baud: %v", u)
		}
	}

	var parsed int
	v = q.Get("size")
	if v != "" {
		parsed, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid size: %v", u)
		}
		c.Size = byte(parsed)
	}

	v = q.Get("stop_bit")
	if v != "" {
		parsed, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid stop_bit: %v", u)
		}
		c.StopBits = serial.StopBits(parsed)
	}

	v = q.Get("parity")
	if v != "" {
		if len(v) != 1 {
			return nil, fmt.Errorf("invalid parity: %v", u)
		}
		c.Parity = serial.Parity(v[0])
	}

	return c, nil
}
