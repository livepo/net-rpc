package therpc

import (
    "fmt"
    "time"
    "github.com/jakehl/goid"
    "github.com/levigross/grequests"
    "sync"
    "errors"
    "net/url"
)


type Server struct {
    Mutex sync.Mutex
    HeartBeat map[string]time.Time
    TaskQueue []Task
}


type RequestName struct {
    WorkerName string
}


type RequestUid struct {
    WorkerUid string
}


type Response struct {
    WorkerUid string
    HasTask bool
}


type Api struct {
    Method string
    Url *url.URL
    RequestOptions *grequests.RequestOptions
}


type Task struct {
    Api []Api
    C int
    Uid string
}


type Result struct {
    Duration time.Duration
    StatusCode map[int]int
    TaskId string
}


type ResponseMsg struct {
    Msg string
}


func GenerateUUID() string {
    v4uuid := goid.NewV4UUID()
    return v4uuid.String()
}


func ResolveHostIp() string {
    netInterfaceAddresses, err := net.InterfaceAddrs()
    if err != nil { return "" }
    for _, netInterfaceAddress := range netInterfaceAddresses {
        networkIp, ok := netInterfaceAddress.(*net.IPNet)
        if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil {
            ip := networkIp.IP.String()
            fmt.Println("Resolved Host IP: " + ip)
            return ip
        }
    }
    return ""
}

func (s *Server) Register(req *RequestName, rsp *Response) error {
    rsp.WorkerUid = req.WorkerName + "-" + GenerateUUID()
    return nil
}


func (s *Server) Ping(req *RequestUid, rsp *Response) error {
    s.HeartBeat[req.WorkerUid] = time.Now()
    rsp.WorkerUid = req.WorkerUid
    s.Mutex.Lock()
    if len(s.TaskQueue) > 0 {
        rsp.HasTask = true
    } else {
        rsp.HasTask = false
    }
    s.Mutex.Unlock()
    for k, v := range s.HeartBeat {
        if time.Now().Sub(v) > time.Second * 10 {
            fmt.Println("delete disconnect worker", k)
            delete(s.HeartBeat, k)
        }
    }
    return nil
}


func (s *Server) Fetch(req *RequestUid, rsp *Task) error {
    defer func() {
        recover()
    }()

    if _, ok := s.HeartBeat[req.WorkerUid]; !ok {
        return errors.New("worker need register")
    }
    s.Mutex.Lock()
    fmt.Println("sasaaaaaaaaaaaaa", s.TaskQueue)
    task := s.TaskQueue[0]
    s.TaskQueue = s.TaskQueue[1:]

    s.Mutex.Unlock()
    rsp.Api = task.Api
    rsp.C = task.C
    rsp.Uid = task.Uid
    return nil
}


func (s *Server) Report(req *Result, rsp *ResponseMsg) error {
    fmt.Println("a worker finished a task", req, rsp)
    return nil
}


func (s *Server) Requests(req *Task) error {
    if len(s.HeartBeat) == 0 {
        return errors.New("register worker first")
    }
    cnt := req.C / len(s.HeartBeat)
    requestTimes := cnt

    defer s.Mutex.Unlock()
    s.Mutex.Lock()

    for i := 1; i <= len(s.HeartBeat); i ++ {
        if i == len(s.HeartBeat) {
            requestTimes = req.C - (i-1)*cnt
        }
        s.TaskQueue = append(s.TaskQueue, Task{Api: req.Api, C: requestTimes, Uid: GenerateUUID()})
    }
    return nil
}
