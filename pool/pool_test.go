package pool_test

import (
	"io"
	"testing"
)

func TestReaderFail(t *testing.T) {
	p := pool.New()

	var rwc io.ReadWriteCloser

	p.Write(rwc, []byte("msg"))

	p.Read(rwc, func(n int, msg []byte) {

	})

	p.CloseAll()

}
