package rpc

import (
    "io"
    "log"
    "net/http"
    "encoding/json"
    
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Rpc struct{
    stream      *Rpc_RpcStreamServer
    id          string
}

var gRpcs = make(map[string] *Rpc)

func NewRpc(stream Rpc_RpcStreamServer, sendMsg func(id int, msg string)) error{
    ctx := stream.Context()
    var err error
    var rpc *Rpc
    for {
        select {
        case <-ctx.Done():
            log.Println("rpc down...")
            DelRpc(rpc.id)

            err = ctx.Err()
            return err

        default:
            input, err := stream.Recv()
            if err == io.EOF {
                log.Println("stream over...")
                DelRpc(rpc.id)
                goto ERR

            }
            if err != nil {
                log.Printf("rpc error: [%v]\n", err)
                DelRpc(rpc.id)
                goto ERR
            }
            
            switch input.Type {
            case 0:
                log.Println("close stream...")
                DelRpc(rpc.id)
                goto ERR

            case 1:
                log.Println("register worker...")
                rpc = &Rpc{stream:&stream, id:input.Msg}
                if gRpcs[input.Msg] == nil {
                    gRpcs[input.Msg] = rpc
                }
            case 2:
                var data map[string]interface{}
                if err := json.Unmarshal([]byte(input.Msg), &data); err == nil {
                    sendMsg(int(data["id"].(float64)), data["msg"].(string))                    
                } else {
                    log.Println(err)
                }
            default:
                log.Printf("rpc recv: %s", input.Msg)
            }
        }
    }

    ERR:
        ctx.Done()

    return err
}

func DelRpc(id string){
    log.Printf("delete rpc: %s", id)
    delete(gRpcs,id)
}

func GetRpc(id string) *Rpc{
    return gRpcs[id]
}

func (this *Rpc)SendMsg (msg string) error{
    stream := *this.stream
    err := stream.Send(&B2WMSG{Msg:msg})
    log.Println("rpc send: %s", msg)
    if err != nil {
        log.Fatal(err)
    }
    
    return err
}