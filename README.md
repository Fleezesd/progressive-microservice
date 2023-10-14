# 1.base-server实现
### 1.基础server
```go
// main.go
import (  
    "fmt"  
    "io"
	"log"
    "net/http"
)  
  
func main() {  
    http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {  
       log.Println("hello world")  
       d, err := io.ReadAll(r.Body)  
       if err != nil {  
          http.Error(rw, err.Error(), http.StatusBadRequest)  
          return  
       }  
       fmt.Fprintf(rw, "Hello %s", d)  
    })  
    http.HandleFunc("/goodbye", func(rw http.ResponseWriter, r *http.Request) {  
       log.Println("Goodbye World")  
    })  
    http.ListenAndServe(":9090", nil)  // nil为DefaultServeMux
}
```
- 仅通过官方http库实现server
  ----> 让启动main文件瘦身, 拆handler逻辑
## 2.handler 逻辑拆分
- 扩展前提 Reference https://pkg.go.dev/net/http#Handler
```go
// Handler responds to an HTTP request.
// ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.   
// ServeHTTP 将请求分派到其模式与请求 URL 最匹配的处理程序。

type Handler interface {  
    ServeHTTP(ResponseWriter, *Request)  
}
```
- 即我们扩展的handler 需完成ServeHTTP方法
```go
// handler/hello.go

type Hello struct {  
    l *log.Logger  
}  
  
func NewHello(l *log.Logger) *Hello {  
    return &Hello{  
       l: l,  
    }}  
  
func (h *Hello) ServeHTTP(rw http.ResponseWriter, r *http.Request) {  
    h.l.Println("hello world")  
    d, err := io.ReadAll(r.Body)  
    if err != nil {  
       http.Error(rw, err.Error(), http.StatusBadRequest)  
       return  
    }  
    fmt.Fprintf(rw, "Hello %s", d)  
}
```
- logger 变成可配置，后续可以通过依赖注入来变化配置，使其变得灵活
- ServeMux: ServeMux 是一个 HTTP 请求多路复用器。它将每个传入请求的 URL 与注册模式列表进行匹配，并调用与 URL 最匹配的模式的处理程序。
  reference: https://pkg.go.dev/log https://pkg.go.dev/net/http#ServeMux
```go
// main.go
func main() {  
    l := log.New(os.Stdout, "product-api ", log.Lshortfile|log.LstdFlags)  
    hh := handler.NewHello(l)  
    sm := http.NewServeMux()  
    sm.Handle("/", hh)  
    http.ListenAndServe(":9090", sm)  
}
```

## 3.配置Server
- 对server进行配置，并将handler封装进去
- Reference: https://pkg.go.dev/net/http#Server
```go
// main.go
func main() {  
	....
	s := http.Server{  
	    Addr:        ":9090",  
	    Handler:     sm,  
	    IdleTimeout: 1 * time.Second,  
	    ReadTimeout: 1 * time.Second,  
	}  
	s.ListenAndServe()
}
```
## 4.优雅关闭
- 接收关闭信号量
- 优雅关闭server，设置超时时间
- Reference: https://pkg.go.dev/os/signal
  https://pkg.go.dev/net/http#Server.Shutdown
```go
// main.go
func main() {  
    l := log.New(os.Stdout, "product-api ", log.Lshortfile|log.LstdFlags)  
    hh := handler.NewHello(l)  
    sm := http.NewServeMux()  
    sm.Handle("/", hh)  
  
    s := http.Server{  
       Addr:        ":9090",  
       Handler:     sm,  
       IdleTimeout: 1 * time.Second,  
       ReadTimeout: 1 * time.Second,  
    }
    // 防止阻塞	
    go func() {  
       if err := s.ListenAndServe(); err != nil {  
          l.Fatal(err)  
       }    
    }()
    
    // 识别signal  
    sigChan := make(chan os.Signal)  
    signal.Notify(sigChan, os.Kill, os.Interrupt, syscall.SIGTERM)  
    sig := <-sigChan    l.Println("Received terminate, graceful shutdown", sig)  
  
    // 优雅关闭  
    tc, cancel := context.WithTimeout(context.Background(), 10*time.Second)  
    defer cancel()  
    s.Shutdown(tc)  
}
```