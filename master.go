package main


import (
    "net"
    "net/http"
    "net/rpc"
    "net/url"
    "example/therpc"
    "time"
    "sync"
)


func TestTask(s *therpc.Server) {
    url1 := &url.URL{
        Scheme:      "https",
        Host:        "baidu.com",
    }
    url2 := &url.URL{
        Scheme:      "https",
        Host:        "sina.com",
    }

    api1 := therpc.Api{
        Method: "get",
        Url: url1,
    }
    api2 := therpc.Api{
        Method: "get",
        Url: url2,
    }
    task := therpc.Task{
        Api: []therpc.Api{api1, api2},
        C: 100,
        Uid: therpc.GenerateUUID(),
    }

    tick := time.Tick(time.Second * 20)
    for {
        select {
        case <-tick:
            s.Requests(&task)
        }
    }
}


func main() {
    s := &therpc.Server{
        HeartBeat: make(map[string]time.Time),
        TaskQueue: make([]therpc.Task, 0),
        Mutex: sync.Mutex{},
    }

    go TestTask(s)
    rpc.Register(s)
    rpc.HandleHTTP()
    lis, _ := net.Listen("tcp", ":8973")
    http.Serve(lis, nil)
}
