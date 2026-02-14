package wrapsocket

type Hooks struct {
	OnConnect    func(conn *Conn)
	OnDisconnect func(conn *Conn)
	OnMessage    func(conn *Conn, msg *Message)
	OnError      func(conn *Conn, err error)
}
