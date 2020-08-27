package utils

import (
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestSyncMapConn(t *testing.T) {
	syncMapConn := NewSyncMapConn()
	if reflect.TypeOf(syncMapConn).String() != "*utils.SyncMapConn" {
		t.Fatalf("is not *utils.SyncMapConn")
	}
	addr := "127.0.0.1:10000"
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var buf []byte
		for {
			conn, err := listen.Accept()
			if err != nil {
				t.Fatal(err)
			}
			tmp := make([]byte, 256)
			n, err := conn.Read(tmp)
			if err != nil {
				t.Fatal(err)
			}
			buf = append(buf, tmp[:n]...)
			if string(buf) == "hello" {
				t.Logf("The result is correct")
				conn.Close()
				break
			}
		}
		wg.Done()
	}()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	syncMapConn.Set("test", addr, &conn)
	time.Sleep(time.Second)
	result := syncMapConn.Get("test", addr)
	(*result).Write([]byte("hello"))
	wg.Wait()
	listen.Close()
	(*result).Close()
	syncMapConn.Clear("test")
	empty := syncMapConn.Get("test", addr)
	if empty != nil {
		t.Fatal("conn should be cleared")
	}
}
