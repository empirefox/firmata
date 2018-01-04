package firmata

import (
	"context"
	"errors"
	"io"
)

var (
	ErrReuse = errors.New("Firmata cannot be reused")
)

func (f *Firmata) Dial(ctx context.Context, c io.ReadWriteCloser) (err error) {
	if f.c != nil {
		return ErrReuse
	}
	f.c = c

	f.connecting.Store(true)
	connectError := make(chan error, 1)
	f.connectedCh = make(chan struct{}, 1)

	err = f.reportVersion()
	if err != nil {
		return err
	}

	go func() {
		for {
			if !f.Connecting() {
				return
			}
			if e := f.unmashal(); e != nil {
				connectError <- e
				return
			}
		}
	}()

	select {
	case <-f.connectedCh:
	case err = <-connectError:
	case <-ctx.Done():
		err = ctx.Err()
	}
	if err != nil {
		return err
	}

	if f.OnConnected != nil {
		f.OnConnected()
	}

	f.closed = make(chan struct{}, 1)
	go func() {
		for {
			if !f.Connected() {
				break
			}

			if err := f.unmashal(); err != nil && f.OnError != nil {
				f.OnError(err)
			}
		}
		close(f.closed)
	}()

	return
}
