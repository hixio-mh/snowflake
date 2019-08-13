package proto

import (
	"net"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSnowflakeProto(t *testing.T) {
	Convey("Connection set up", t, func(ctx C) {

		client, server := net.Pipe()

		c := NewSnowflakeConn(client)
		s := NewSnowflakeConn(server)

		Convey("Create correct headers", func(ctx C) {
			var sent, received, wire []byte
			var wg sync.WaitGroup
			sent = []byte{'H', 'E', 'L', 'L', 'O'}
			wire = []byte{
				0x00, 0x00, 0x00, 0x00, //seq
				0x00, 0x00, 0x00, 0x00, //ack
				0x00, 0x05, //len
				'H', 'E', 'L', 'L', 'O',
			}
			received = make([]byte, len(wire), len(wire))

			wg.Add(2)
			go func() {
				n, err := c.Write(sent)
				ctx.So(n, ShouldEqual, len(sent))
				ctx.So(err, ShouldEqual, nil)
				ctx.So(c.seq, ShouldEqual, 5)
				wg.Done()
			}()

			go func() {
				n, err := s.Read(received)

				ctx.So(err, ShouldEqual, nil)
				ctx.So(n, ShouldEqual, len(sent))
				ctx.So(received[:n], ShouldResemble, sent)
				s.lock.Lock()
				ctx.So(s.ack, ShouldEqual, 5)
				s.lock.Unlock()
				wg.Done()
			}()

			wg.Wait()

			// Check that acknowledgement packet was written
			//n, err = s.Read(received)
			//So(err, ShouldEqual, nil)
			//So(n, ShouldEqual, 0)

		})

		Convey("Partial reads work correctly", func(ctx C) {
			var sent, received []byte
			var wg sync.WaitGroup
			sent = []byte{'H', 'E', 'L', 'L', 'O'}
			received = make([]byte, 3, 3)

			wg.Add(2)
			go func() {
				n, err := c.Write(sent)
				ctx.So(err, ShouldEqual, nil)
				ctx.So(n, ShouldEqual, 5)
				wg.Done()
			}()

			//Read in first part of message
			go func() {
				n, err := s.Read(received)

				ctx.So(err, ShouldEqual, nil)
				ctx.So(n, ShouldEqual, 3)
				ctx.So(received[:n], ShouldResemble, sent[:n])

				//Read in rest of message
				n2, err := s.Read(received)

				ctx.So(err, ShouldEqual, nil)
				ctx.So(n2, ShouldEqual, 2)
				ctx.So(received[:n2], ShouldResemble, sent[n:n+n2])

				s.lock.Lock()
				ctx.So(s.ack, ShouldEqual, 5)
				s.lock.Unlock()
				wg.Done()
			}()

			wg.Wait()

		})

		Convey("Test reading multiple chunks", func(ctx C) {
			var sent, received, buffer []byte
			var wg sync.WaitGroup
			sent = []byte{'H', 'E', 'L', 'L', 'O'}
			received = make([]byte, 3, 3)

			var n int
			var err error

			wg.Add(2)
			go func() {
				c.Write(sent)
				c.Write(sent)
				wg.Done()
			}()

			go func() {
				n, err = s.Read(received)
				buffer = append(buffer, received[:n]...)
				ctx.So(err, ShouldEqual, nil)
				ctx.So(n, ShouldEqual, 3)
				ctx.So(buffer, ShouldResemble, sent[:3])

				n, err = s.Read(received)
				buffer = append(buffer, received[:n]...)
				ctx.So(err, ShouldEqual, nil)
				ctx.So(n, ShouldEqual, 2)
				ctx.So(buffer, ShouldResemble, sent)

				n, err = s.Read(received)
				buffer = append(buffer, received[:n]...)
				ctx.So(err, ShouldEqual, nil)
				ctx.So(n, ShouldEqual, 3)
				ctx.So(buffer, ShouldResemble, append(sent, sent[:3]...))

				n, err = s.Read(received)
				buffer = append(buffer, received[:n]...)
				ctx.So(err, ShouldEqual, nil)
				ctx.So(n, ShouldEqual, 2)
				ctx.So(buffer, ShouldResemble, append(sent, sent...))

				s.lock.Lock()
				ctx.So(s.ack, ShouldEqual, 2*5)
				s.lock.Unlock()
				wg.Done()
			}()
			wg.Wait()

		})
	})

}
