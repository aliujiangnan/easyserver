package gate

import (
    "log"
    "net"
    "flag"
    "net/http"

    "easyserver/lib/rpc"
    "easyserver/lib/util"
    proto "easyserver/lib/rpc"
    
    "google.golang.org/grpc"
)

type Gate struct{}

var gInstance = &Gate{}

func (this *Gate) RpcStream(stream proto.Rpc_RpcStreamServer) error {
    return rpc.NewRpc(stream, func(id int, msg string){
        GetSession(id).SendMsg(msg)
    })
}

func onSessionReq(server string, id int, msg []byte){
    worker := rpc.GetRpc(server)
    if worker == nil {
        log.Printf("could not find worker %d", id)
        return
    }
    data := make(map[string]interface{})
    data["type"] = 1
    data["id"] = id
    data["msg"] = string(msg)

    worker.SendMsg(util.JsonToString(data))
}

func onSessionOffline(server string, id int){
    worker := rpc.GetRpc(server)
    if worker == nil {
        log.Printf("could not find worker %d", id)
        return
    }
    data := make(map[string]interface{})
    data["type"] = 0
    data["id"] = id

    worker.SendMsg(util.JsonToString(data))
}

func (this *Gate)handleMsg(w http.ResponseWriter, r *http.Request) {
    NewSession(w, r, onSessionReq, onSessionOffline)
}

func Open(host string) {
    go func(){
        server := grpc.NewServer()
        proto.RegisterRpcServer(server, gInstance)

        address, _ := net.Listen("tcp", ":43210")
        log.Fatal(server.Serve(address))
    }()
    
    go func(){
        var addr = flag.String("addr", host, "http service address")
        
        http.HandleFunc("/socket", gInstance.handleMsg)

        log.Fatal(http.ListenAndServe(*addr, nil))
    }()

}