// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package network tcp服务器
package network

import (
	"crypto/tls"
	"github.com/liangdas/mqant/log"
	"net"
	"sync"
	"time"
)

// TCPServer tcp服务器
type TCPServer struct {
	Addr       string
	TLS        bool //是否支持tls
	CertFile   string
	KeyFile    string
	MaxConnNum int
	NewAgent   func(*TCPConn) Agent
	ln         net.Listener
	mutexConns sync.Mutex
	wgLn       sync.WaitGroup
	wgConns    sync.WaitGroup
}

// Start 开始tcp监听
func (server *TCPServer) Start() {
	server.init()
	log.Info("TCP Listen :%s", server.Addr)
	go server.run()
}

func (server *TCPServer) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Warning("%v", err)
	}

	if server.NewAgent == nil {
		log.Warning("NewAgent must not be nil")
	}
	if server.TLS {
		tlsConf := new(tls.Config)
		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err == nil {
			ln = tls.NewListener(ln, tlsConf)
			log.Info("TCP Listen TLS load success")
		} else {
			log.Warning("tcp_server tls :%v", err)
		}
	}

	server.ln = ln
}
func (server *TCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Info("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0
		tcpConn := newTCPConn(conn)
		agent := server.NewAgent(tcpConn)
		go func() {
			server.wgConns.Add(1)
			agent.Run()

			// cleanup
			tcpConn.Close()
			agent.OnClose()

			server.wgConns.Done()
		}()
	}
}

// Close 关闭TCP监听
func (server *TCPServer) Close() {
	server.ln.Close()
	server.wgLn.Wait()
	server.wgConns.Wait()
}
