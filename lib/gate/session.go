package gate

import (
    "log"
    "net/http"

    "easyserver/lib/util"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Session struct{
    conn        *websocket.Conn
    id          int
    server      string
}

var gid = 0
var gSessions = make(map[int] *Session)

func genSessionId() int{
    for i := 0; i < gid; i++{
        if gSessions[i] == nil {
            return i
        }
    }
    gid++
    return gid - 1
}

func NewSession(w http.ResponseWriter, r *http.Request, onReq func(server string, id int, msg []byte), onOffline func(server string, id int)) error{
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return err
    }

    session := &Session{conn:c, id:genSessionId(), server:"0"}
    if gSessions[session.id] == nil {
        gSessions[session.id] = session
    }

    log.Printf("new session: %d", session.id)
    defer c.Close()
    for {
        mt, message, err := c.ReadMessage()
        if err != nil {
            log.Println("session err")
            DelSession(session.id)
            goto ERR
        }

        if mt == websocket.CloseMessage {
            log.Println("session close")
            onOffline(session.server, session.id)
            DelSession(session.id)
            goto ERR

        } else {
            onReq(session.server, session.id, message)
            log.Printf("session recv: %s", message)
        }
    }

    ERR:
        c.Close()

    return err
}

func DelSession(id int){
    delete(gSessions, id)
}

func GetSession(id int) *Session{
    return gSessions[id]
}

func (this *Session)SendMsg (msg string) error{
    err := this.conn.WriteMessage(websocket.TextMessage, util.StringToBytes(msg))
    if err != nil {
        log.Printf("SendMsg failed: [%v]\n", err)
    }
    
    return err
}