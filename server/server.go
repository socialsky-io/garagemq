package server

import (
	"net"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/valinurovam/garagemq/auth"
)

type Server struct {
	host         string
	port         string
	protoVersion string
	listener     net.Listener
	connSeq      int64
	connections  map[int64]*Connection
	config       *ServerConfig
	users        map[string]string
}

func NewServer(host string, port string, protoVersion string, config *ServerConfig) (server *Server) {
	server = &Server{
		host:         host,
		port:         port,
		connections:  make(map[int64]*Connection),
		protoVersion: protoVersion,
		config:       config,
		users:        make(map[string]string),
	}

	server.initUsers()
	return
}

func (srv *Server) Start() (err error) {
	address := srv.host + ":" + srv.port
	srv.listener, err = net.Listen("tcp", address)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"network": "tcp",
			"address": address,
		}).Error("Error on listener start")
		os.Exit(1)
	}

	log.WithFields(log.Fields{
		"network": "tcp",
		"address": address,
	}).Info("Server start")

	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			log.WithError(err).Error("Error accept connection")
			os.Exit(1)
		}
		log.WithFields(log.Fields{
			"address": conn.RemoteAddr().String(),
			"network": conn.RemoteAddr().Network(),
		}).Info("New connection")
		go srv.acceptConnection(conn)
	}

	return
}

func (srv *Server) Stop() {
	srv.listener.Close()
}

func (srv *Server) acceptConnection(conn net.Conn) {
	connection := NewConnection(srv, conn)
	srv.connections[connection.id] = connection
	srv.connections[connection.id].handleConnection()
}

func (srv *Server) checkAuth(saslData auth.SaslData) bool {
	for userName, passwordHash := range srv.users {
		if userName != saslData.Username {
			continue
		}

		return auth.CheckPasswordHash(saslData.Password, passwordHash);
	}
	return false
}

func (srv *Server) initUsers() {
	for _, user := range srv.config.Users {
		srv.users[user.Username] = user.Password
	}
}