package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/determined-ai/determined/master/pkg/actor"
	"github.com/determined-ai/determined/master/pkg/syncx/queue"
)

// TODO: Add a write size limit.

const (
	// pingWaitDuration is the duration to wait for a pong response to a ping.
	pingWaitDuration = 1 * time.Minute
	// pingInterval is the duration to wait for between pinging connections.
	pingInterval = 1 * time.Minute
)

const (
	// MaxWebsocketMessageSize is the maximum size of a websocket message that we send in bytes.
	// This is copied from MAX_WEBSOCKET_MSG_SIZE in determined/constants.py.
	MaxWebsocketMessageSize = 128 * 1024 * 1024
)

// WebSocketRequest notifies the actor that a websocket is attempting to connect.
type WebSocketRequest struct {
	Ctx echo.Context
}

// TODO: consider specializing this to agent/agents to simplify code
// Accept wraps the connecting websocket connection in an actor.
func AcceptWebSocketRequest[T any](
	w WebSocketRequest,
	destActor *actor.Ref,
	usePing bool,
) (*WebSocketManager[T], error) {
	// TODO: error callback

	if reflect.TypeOf(*new(T)).Kind() == reflect.Pointer {
		return nil, errors.New("WebSocket message types must not be a pointer")
	}

	httpReq := w.Ctx.Request()

	conn, err := upgrader.Upgrade(w.Ctx.Response(), httpReq, nil)
	if err != nil {
		return nil, errors.Wrap(err, "websocket connection error")
	}

	a := WrapSocket[T](httpReq.Context(), conn, destActor, usePing)
	return a, nil
}

// IsReconnect checks if agent is reconnecting after a network failure.
func (w *WebSocketRequest) IsReconnect() (bool, error) {
	return strconv.ParseBool(w.Ctx.QueryParam("reconnect"))
}

// WriteMessage is a message to a websocketActor asking it to write out the
// given message, encoding it to JSON.
type WriteMessage struct {
	actor.Message
}

// WriteRawMessage is a message to a websocketActor asking it to write out the
// given message without encoding to JSON.
type WriteRawMessage struct {
	actor.Message
}

// WriteResponse is the response to a successful WriteMessage.
type WriteResponse struct{}

// WriteSocketJSON writes a JSON-serializable object to a websocket actor.
func WriteSocketJSON(ctx *actor.Context, socket *actor.Ref, msg interface{}) error {
	resp := ctx.Ask(socket, WriteMessage{
		Message: msg,
	}).Get()

	switch resp := resp.(type) {
	case error:
		return errors.WithStack(resp)
	case WriteResponse:
		return nil
	default:
		return errors.Errorf("unknown response %T: %s", resp, resp)
	}
}

// WriteSocketRaw writes a string to a websocket actor.
func WriteSocketRaw(ctx *actor.Context, socket *actor.Ref, msg interface{}) error {
	resp := ctx.Ask(socket, WriteRawMessage{
		Message: msg,
	}).Get()

	switch resp := resp.(type) {
	case error:
		return errors.WithStack(resp)
	case WriteResponse:
		return nil
	default:
		return errors.Errorf("unknown response %T: %s", resp, resp)
	}
}

// WrapSocket wraps a websocket connection as an actor.
func WrapSocket[T any](ctx context.Context, conn *websocket.Conn, destActor *actor.Ref, usePing bool,
) *WebSocketManager[T] {
	m := &WebSocketManager[T]{
		conn: conn,
		// msgType:      reflect.TypeOf(msgType),
		destActor:    destActor,
		usePing:      usePing,
		pendingPings: make(map[string]time.Time),
	}

	if m.usePing {
		m.setupPingLoop(ctx)
	}

	go m.runReadLoop(ctx)

	return m
}

type WebSocketManager[T any] struct {
	conn      *websocket.Conn
	msgType   reflect.Type
	destActor *actor.Ref

	msgQueue    *queue.Queue[T]
	msgCallback func(msg T)
	errCallback func(err error)

	usePing      bool
	pingLock     sync.Mutex
	pendingPings map[string]time.Time
}

func (s *WebSocketManager[T]) Queue() *queue.Queue[T] {
	return s.msgQueue
}

// TODO(maybe): helper function that consumes queue and does recipient.Tell(msg) in a goroutine

// Receive implements the actor.Actor interface.
func (s *WebSocketManager[T]) Receive(ctx *actor.Context) error {
	switch msg := ctx.Message().(type) {
	case actor.PreStart:
		return nil
	case actor.PostStop:
		return s.conn.Close()
	case error: // Socket read errors.
		return msg
	case []byte: // Incoming messages on the socket.
		// TODO: replace this with a goroutine that reads from the queue and sends parsed messages
		parsed, err := s.parseMsg(msg, s.msgType)
		if err != nil {
			return err
		}
		// Notify the socket's parent actor of the incoming message.
		s.destActor.System().Tell(s.destActor, parsed)
		return nil
	case WriteMessage:
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(msg.Message); err != nil {
			return err
		}
		return s.processWriteMessage(ctx, buf)
	case WriteRawMessage:
		var buf bytes.Buffer
		if _, err := buf.WriteString(msg.Message.(string)); err != nil {
			return err
		}
		return s.processWriteMessage(ctx, buf)
	default:
		return actor.ErrUnexpectedMessage(ctx)
	}
}

func (s *WebSocketManager[T]) processWriteMessage(
	ctx *actor.Context,
	buf bytes.Buffer,
) error {
	if cur, max := buf.Len(), MaxWebsocketMessageSize; cur > max {
		ctx.Respond(errors.Errorf("message size %d exceeds maximum size %d", cur, max))
		return nil
	}

	ctx.Respond(WriteResponse{})

	return s.conn.WriteMessage(websocket.TextMessage, buf.Bytes())
}

func isClosingError(err error) bool {
	return err == websocket.ErrCloseSent || websocket.IsCloseError(err, websocket.CloseNormalClosure)
}

func (s *WebSocketManager[T]) setupPingLoop(ctx *actor.Context) {
	s.conn.SetPongHandler(func(data string) error {
		return s.handlePong(ctx, data)
	})
	go s.runPingLoop(ctx)
}

func (s *WebSocketManager[T]) handlePong(ctx *actor.Context, id string) error {
	now := time.Now()

	s.pingLock.Lock()
	defer s.pingLock.Unlock()

	if deadline, ok := s.pendingPings[id]; !ok {
		ctx.Log().Errorf("unknown ping %s", id)
		return nil
	} else if deadline.Before(now) {
		// This is a ping timeout, but let checkPendingPings report the error.
		return nil
	}
	delete(s.pendingPings, id)
	return nil
}

func (s *WebSocketManager[T]) checkPendingPings() error {
	now := time.Now()

	s.pingLock.Lock()
	defer s.pingLock.Unlock()

	var errs []error
	for id, deadline := range s.pendingPings {
		if deadline.Before(now) {
			errs = append(errs, errors.Errorf("ping %s did not receive pong response by %s", id, deadline))
			delete(s.pendingPings, id)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (s *WebSocketManager[T]) ping() error {
	s.pingLock.Lock()
	defer s.pingLock.Unlock()

	// According to the websocket specification [1], endpoints only have to acknowledge the most
	// recent ping message, so avoid having more than one outstanding ping request.
	//
	// [1] https://tools.ietf.org/html/rfc6455#section-5.5.3
	if len(s.pendingPings) > 0 {
		return nil
	}

	id := uuid.New().String()

	deadline := time.Now().Add(pingWaitDuration)
	err := s.conn.WriteControl(websocket.PingMessage, []byte(id), deadline)
	if e, ok := err.(net.Error); ok && e.Temporary() { //nolint: staticcheck
		// Temporary is deprecated but not sure the better alternative.
		return nil
	} else if err != nil {
		return err
	}

	s.pendingPings[id] = deadline

	return nil
}

func (s *WebSocketManager[T]) runPingLoop(ctx context.Context) {
	pingAndWait := func() error {
		if err := s.ping(); err != nil {
			return err
		}

		t := time.NewTimer(pingInterval)
		defer t.Stop()
		<-t.C
		return nil
	}

	// TODO: shut down using context.Context
	// Stop stops now and doesn't let the actor consume any more messages from the queue
	defer ctx.Self().Stop()

	for {
		if err := s.checkPendingPings(); err != nil {
			ctx.Tell(ctx.Self(), err)
			return
		}

		if err := pingAndWait(); isClosingError(err) {
			return
		} else if err != nil {
			//
			ctx.Tell(ctx.Self(), err)
			return
		}
	}
}

func (s *WebSocketManager[T]) runReadLoop(ctx context.Context) {
	read := func() ([]byte, error) {
		msgType, msg, err := s.conn.ReadMessage()
		if err != nil {
			return nil, err
		}
		if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
			return nil, errors.Errorf("unexpected message type: %d", msgType)
		}
		return msg, nil
	}

	// TODO: stop (the loops?)
	defer ctx.Self().Stop()

	for {
		msg, err := read()
		if isClosingError(err) {
			return
		} else if err != nil {
			// Socket read errors are sent to the socket actor rather than the parent. Exceptions
			// will bubble up the parent through the actor system.
			// ctx.Tell(ctx.Self(), err)

			// TODO: verify this is okay
			s.syslog.WithError(err).Error("error reading from WebSocket")
			s.Stop()
			return
		}

		// TODO: enqueue message
		// ctx.Tell(ctx.Self(), msg)

		parsedMsg, err := s.parseMsg(msg)
		if err != nil {
			// TODO: handle err
			continue
		}

		s.msgQueue.Put(*parsedMsg)
	}
}

func (s *WebSocketManager[T]) parseMsg(raw []byte) (*T, error) {
	val := new(T)

	// var parsed interface{}
	// if reflect.TypeOf(val).Kind() == reflect.Ptr {
	// 	parsed = reflect.New().Interface()
	// } else {
	// 	parsed = reflect.New(s.msgType).Interface()
	// }

	if err := json.Unmarshal(raw, val); err != nil {
		return nil, err
	}
	return val, nil
	// if s.msgType.Kind() == reflect.Ptr {
	// 	return parsed.(T), nil
	// }

	// val := reflect.ValueOf(parsed).Elem().Interface()
	// return val.(T), nil
}
