package network

import (
	"dq/log"
	"net"
	"sync"
	"time"
)

type ConnSet map[net.Conn]struct{}

type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool
	msgParser *MsgParser

	//needwritedata []byte
}

func newTCPConn(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn

	//	testcon, ok := conn.(*net.TCPConn)
	//	if !ok {
	//		log.Debug("conn.(*TCPConn) err ")
	//		//error handle
	//	}
	//	testcon.SetWriteBuffer(1024 * 1024 * 8)
	//	testcon.SetReadBuffer(1024 * 1024 * 8)

	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser

	//tcpConn.needwritedata = make([]byte, 0)

	//	go func() {
	//		for {
	//			if tcpConn.SendData() == false {
	//				break
	//			}
	//			time.Sleep(time.Millisecond * 5)
	//		}

	//		conn.Close()
	//		tcpConn.Lock()
	//		tcpConn.closeFlag = true
	//		tcpConn.Unlock()
	//	}()

	go func() {
		for b := range tcpConn.writeChan {
			if b == nil {
				break
			}

			_, err := conn.Write(b)
			if err != nil {
				break
			}
		}

		conn.Close()
		tcpConn.Lock()
		tcpConn.closeFlag = true
		tcpConn.Unlock()
	}()

	return tcpConn
}

//func (tcpConn *TCPConn) SendData() bool {
//	tcpConn.Lock()
//	defer tcpConn.Unlock()

//	if len(tcpConn.needwritedata) <= 0 {
//		return true
//	}
//	//tcpConn.needwritedata = append(tcpConn.needwritedata, b...)
//	_, err := tcpConn.conn.Write(tcpConn.needwritedata)
//	if err != nil {
//		return false
//	}
//	tcpConn.needwritedata = make([]byte, 0)

//	return true

//}

func (tcpConn *TCPConn) doDestroy() {
	tcpConn.conn.(*net.TCPConn).SetLinger(0)
	tcpConn.conn.Close()

	if !tcpConn.closeFlag {
		close(tcpConn.writeChan)
		tcpConn.closeFlag = true
	}
}

func (tcpConn *TCPConn) Destroy() {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	tcpConn.doDestroy()
}

func (tcpConn *TCPConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}

	tcpConn.doWrite(nil)
	tcpConn.closeFlag = true
}

func (tcpConn *TCPConn) doWrite(b []byte) {
	for len(tcpConn.writeChan) >= cap(tcpConn.writeChan) {
		log.Debug("conn: channel full %d  %d", len(tcpConn.writeChan), cap(tcpConn.writeChan))
		time.Sleep(time.Millisecond * 2)
		//tcpConn.doDestroy()
		//return
	}

	tcpConn.writeChan <- b

	//	_, err := tcpConn.conn.Write(b)
	//	if err != nil {
	//		tcpConn.conn.Close()
	//		tcpConn.closeFlag = true
	//	}
}

// b must not be modified by the others goroutines
func (tcpConn *TCPConn) Write(b []byte) {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag || b == nil {
		return
	}

	//tcpConn.needwritedata = append(tcpConn.needwritedata, b...)

	tcpConn.doWrite(b)
}

func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

func (tcpConn *TCPConn) ReadMsg() ([]byte, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

func (tcpConn *TCPConn) WriteMsg(args []byte) error {
	return tcpConn.msgParser.Write(tcpConn, args)
}
func (tcpConn *TCPConn) ReadSucc() {

}
