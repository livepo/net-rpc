package main


import (
    "net/rpc"
    "log"
    "fmt"
    "example/therpc"
    "time"
    "github.com/levigross/grequests"
)


func call(rpcname string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", "127.0.0.1:8973")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

type Method func(string, *grequests.RequestOptions) (*grequests.Response, error)

func requests(apis []therpc.Api, c int) *therpc.Result {
    respCode := make(map[int]int)
    start := time.Now()
    for i := 0; i < c; i ++ {
        for _, api := range apis {
            var method Method
            switch api.Method {
            case "get":
                method = grequests.Get
            case "post":
                method = grequests.Post
            case "put":
                method = grequests.Put
            case "delete":
                method = grequests.Delete
            }
            fmt.Println("make a request", api.Url.String())
            resp, err := method(api.Url.String(), api.RequestOptions)
            fmt.Println("resp", resp, err)
            respCode[resp.StatusCode] ++
            fmt.Println("respcode", respCode)
        }
    }
    duration := time.Now().Sub(start)
    fmt.Println(duration)
    return &therpc.Result{Duration: duration, StatusCode: respCode}
}


func consume(taskCh chan *therpc.Task) error {
    for task := range taskCh {
        fmt.Println("consume task", task)
        result := requests(task.Api, task.C)
        result.TaskId = task.Uid
        msg := &therpc.ResponseMsg{}
        call("Server.Report", result, &msg)
        fmt.Println(msg)
    }
    return nil
}


func main() {
    rsp := &therpc.Response{}
    req := &therpc.RequestName{WorkerName: therpc.ResolveHostIp()}
    call("Server.Register", req, rsp)
    fmt.Println(rsp)
    ch := make(chan *therpc.Task)
    go consume(ch)
    tick := time.Tick(time.Second)
    reqid := &therpc.RequestUid{WorkerUid: rsp.WorkerUid}
    for {
        select {
        case <-tick:
            rsp = &therpc.Response{}
            call("Server.Ping", reqid, rsp)
            fmt.Println(rsp)
            if rsp.HasTask {
                task := &therpc.Task{}
                call("Server.Fetch", reqid, task)
                fmt.Println("fetch a task", task)
                ch <- task
            }
        }
    }
}
