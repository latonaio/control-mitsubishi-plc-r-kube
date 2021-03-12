package cmd

import "net"

type PlcClient struct {
	conn *net.TCPConn
}

type PlcReader interface {
	NewClient(targetAddress, targetPort string) (PlcReader, error)
	Close() error
	Read(data []byte) (int, error)
	Conn() *net.TCPConn
}

func (pc *PlcClient) NewClient(targetAddress, targetPort string) (PlcReader, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", targetAddress+":"+targetPort)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", tcpAddr, nil)
	if err != nil {
		return nil, err
	}

	return &PlcClient{
		conn: conn,
	}, nil
}

func (pc *PlcClient) Close() error {
	if pc.conn != nil {
		if err := pc.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (pc *PlcClient) Read(data []byte) (int, error) {
	return pc.conn.Read(data)
}

func (pc *PlcClient) Conn() *net.TCPConn {
	return pc.conn
}
