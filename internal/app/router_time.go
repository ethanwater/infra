package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"vivian.infra/internal/pkg/socket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var calls atomic.Int32

func HandleWebSocketTimestamp(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		VivianServerLogger.SetProtocol(1)
		defer VivianServerLogger.DefaultProtocol()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			VivianServerLogger.LogError("vivian: socket: [error] handshake failure", err)
		} else {
			VivianServerLogger.LogSuccess(fmt.Sprintf("handshake success: remote:%v local:%v", conn.RemoteAddr(), conn.LocalAddr()))
		}
		defer conn.Close()

		disconnectChannel := make(chan int)
		defer close(disconnectChannel)

		go func() {
			for {
				select {
				case <-disconnectChannel:
					VivianServerLogger.LogDebug("handshake disconnected")
					return
				case <-ctx.Done():
					VivianServerLogger.LogWarning("lost context")
					return
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				VivianServerLogger.LogWarning("lost context")
				return
			default:
				timestamp := socket.TimeRFC3339Local()
				err = conn.WriteMessage(websocket.TextMessage, timestamp)
				if err != nil {
					VivianServerLogger.LogWarning(fmt.Sprintf("%v", err))
					disconnectChannel <- 1
					return
				}
				time.Sleep(time.Second)
			}
		}
	})
}

var liveConn *websocket.Conn
var socketSync sync.Mutex

func SocketCalls(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		VivianServerLogger.SetProtocol(1)
		defer VivianServerLogger.DefaultProtocol()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			VivianServerLogger.LogError("handshake failure", websocket.ErrBadHandshake)
			return
		}
		defer conn.Close()

		VivianServerLogger.LogSuccess(fmt.Sprintf("handshake success: remote:%v local:%v", conn.RemoteAddr(), conn.LocalAddr()))

		socketSync.Lock()
		if liveConn != nil {
			liveConn.Close()
		}
		socketSync.Unlock()
		liveConn = conn

		reconnectChannel := make(chan int)
		defer close(reconnectChannel)

		//TODO: does this even work?
		//should probably write a testcase for this shit im ngl
		go func() {
			for {
				select {
				case <-reconnectChannel:
					VivianServerLogger.LogDebug("reconnected")
					return
				case <-ctx.Done():
					VivianServerLogger.LogWarning("lost context")
					return
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				VivianServerLogger.LogWarning("lost context")
				return
			default:
				marshal_data, _ := json.Marshal(uint32(calls.Load()))
				err = liveConn.WriteMessage(websocket.TextMessage, marshal_data)
				if err != nil {
					VivianServerLogger.LogError("disconnected <- broken pipe?", err)
					reconnectChannel <- 1
					return
				}
				time.Sleep(time.Second)
			}
		}
	})
}
