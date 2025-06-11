package main

import (
	"net"
	"time"
)

type CustomConn struct {
	u      net.Conn
	readn  int64
	writen int64
}

func (cr *CustomConn) Read(p []byte) (n int, err error) {
	n, err = cr.u.Read(p)
	cr.readn += int64(n)
	return n, err
}

func (cr *CustomConn) Write(p []byte) (n int, err error) {
	n, err = cr.u.Write(p)
	cr.writen += int64(n)
	return n, err
}

func (cr *CustomConn) Close() error {
	err := cr.u.Close()
	if err != nil {
		return err
	}
	// removed nil assignment to prevent nil pointer dereference (not too sure why it is happening)
	return nil
}

func (cr *CustomConn) LocalAddr() net.Addr {
	return cr.u.LocalAddr()
}
func (cr *CustomConn) RemoteAddr() net.Addr {
	return cr.u.RemoteAddr()
}
func (cr *CustomConn) SetDeadline(t time.Time) error {
	return cr.u.SetDeadline(t)
}
func (cr *CustomConn) SetReadDeadline(t time.Time) error {
	return cr.u.SetReadDeadline(t)
}
func (cr *CustomConn) SetWriteDeadline(t time.Time) error {
	return cr.u.SetWriteDeadline(t)
}

// func (cr *CustomConn) Underlying() net.Conn {
// 	return cr.u
// }

func (cr *CustomConn) Readn() int64 {
	return cr.readn
}
func (cr *CustomConn) Writen() int64 {
	return cr.writen
}
func (cr *CustomConn) BytesTransferred() int64 {
	return cr.readn + cr.writen
}
