package worker

import (
    "context"
    "io"
    "log"
    "encoding/json"

    "easyserver/lib/util"
    proto "easyserver/lib/rpc"

    "google.golang.org/grpc"
)

type Worker struct {
    stream *proto.Rpc_RpcStreamClient
}

var gInstance *Worker

func SendMsg(id int, msg string) error{
    stream := *gInstance.stream

    data := make(map[string]interface{})
    data["id"] = id
    data["msg"] = string(msg)

    err := stream.Send(&proto.W2BMSG{Type:2, Msg:util.JsonToString(data)})

    return err
}

func Open(host string, onReq func(id int, msg string), onOffline func(id int)) error{
    conn, err := grpc.Dial(host, grpc.WithInsecure())
    if err != nil {
        log.Printf("connect failed: [%v]\n", err)
        return err
    }
    defer conn.Close()
    
    client := proto.NewRpcClient(conn)
    ctx := context.Background()

    stream, err := client.RpcStream(ctx)
    if err != nil {
        log.Printf("create stream failed: [%v]\n", err)
    }

    gInstance = &Worker{stream:&stream}

    err = stream.Send(&proto.W2BMSG{Type:1, Msg:"0"})

    if err != nil {
        log.Printf("register failed: [%v]\n", err)
        return err
    }

    for {
        input, err := stream.Recv()
        if err == io.EOF {
            log.Println("stream over...")
            break
        }

        if err != nil {
            log.Println("worker error...")
            // log.Fatal(err)
        }
        
        log.Printf("worker recv: %s", input.Msg)

        var data map[string]interface{}
        if err := json.Unmarshal([]byte(input.Msg), &data); err == nil {
            if int(data["type"].(float64)) == 0 {
                onOffline(int(data["id"].(float64)))
                continue
            }

            onReq(int(data["id"].(float64)), data["msg"].(string))
        } else {
            log.Fatal(err)
        }
    }

    return nil
}