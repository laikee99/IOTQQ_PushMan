package main

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
    "fmt"
    "log"
    "os"
    "runtime"
    "strconv"
    "strings"
    "time"

    "github.com/graarh/golang-socketio"
    "github.com/graarh/golang-socketio/transport"
    "github.com/mcoo/iotqq-plugins-demo/Go/model"
)
var url1, qq, token string
var conf iotqq.Conf
var zanok, qd []int64
var msgs data
type msg struct {
    Title string `json:"title"`
    Content string `json:"content"`
    From string `json:"from"`
    Qq string `json:"qq"`
    At string `json:"at"`
    Date int64 `json:"date"`
}
type data struct {
    Status int `json:"status"`
    Msg string `json:"msg""`
    Data []msg `json:"data"`
    Tpl string `json:"qtpl"`
}
func get(url string) data {

    client := &http.Client{}
    request, _ := http.NewRequest("GET", url, nil)
    // request.Header.Set("Connection", "keep-alive")
    response, _ := client.Do(request)
    var d = data{}
    d.Status = 0
    d.Msg = "api error"
    if response.StatusCode == 200 {
        body, _ := ioutil.ReadAll(response.Body)
        json.Unmarshal(body, &d) // fmt.Println(string(body))
        fmt.Println(d.Msg)
    }
    defer response.Body.Close()
    return d
}
func init() {
file, err := os.Open("main.conf")
conf = iotqq.Conf{true, make(map[string]int)}
//log.Println(file)
if err != nil {
    log.Println(err)
    os.Create("main.conf")
    f, _ := os.OpenFile("main.conf", os.O_APPEND, 0644)
    defer f.Close()
    enc := json.NewEncoder(f)
    conf.Enable = true
    conf.GData = make(map[string]int)
    enc.Encode(conf)
}
defer file.Close()
tmp := json.NewDecoder(file)
//log.Println(tmp)
for tmp.More() {
    err := tmp.Decode(&conf)
    if err != nil {
        fmt.Println("Error:", err)
    }
    //fmt.Println(conf)
}
}
func periodlycall(d time.Duration, f func()) {
for x := range time.Tick(d) {
    f()
    log.Println(x)
}
}
func s(d data,m msg,q int){
    var c string
    timeLayout := "2006-01-02 15:04:05"  //转化所需模板
    c = strings.Replace(d.Tpl ,"{title}", m.Title, -1)
    c = strings.Replace(c ,"{content}", m.Content, -1)
    c = strings.Replace(c ,"{time}", time.Unix(m.Date, 0).Format(timeLayout), -1)
    iotqq.Send(q, 2, c)
    fmt.Println("\n发送消息一次")
}
func getmsg() {
    url := "https://mp.application.pub/"+ token +".qmsg"
    fmt.Println(url)
    d := get(url)
    var q int
    if d.Status == 1{
        for i := 0; i< len(d.Data); i++{
            m := d.Data[i]
            iotqq.Set(url1, m.From)
            if strings.Index(m.Qq, ",") == -1{
                // 不存在
                q,_ = strconv.Atoi(m.Qq)
                s(d,m,q)
            }else{
                arr := strings.Split(m.Qq, ",")
                for j := 0; j< len(arr); j++{
                    q,_ = strconv.Atoi(arr[j])
                    s(d,m,q)
                }
            }
        }
    }else{

    }
}
func SendJoin(c *gosocketio.Client) {
  log.Println("Get QQ connection...")
  result, err := c.Ack("GetWebConn", qq, time.Second*5)
  if err != nil {
      log.Fatal(err)
  } else {
      log.Println("emit", result)
  }
}
func main() {
  var site string
  var port int
  port = 8888
  fmt.Println("Push Man - Based on SocketIO V0.0.1")
  fmt.Println("\n请输入Iotqq的Web地址(无需http://和端口): ")
  fmt.Scan(&site)
  fmt.Println("\n请输入Iotqq的端口号: ")
  fmt.Scan(&port)
  fmt.Println("\n请输入QQ机器人账号（兼容多账号）: ")
  fmt.Scan(&qq)
  fmt.Println("\n请输入Push Man的APP token: ")
  fmt.Scan(&token)
  runtime.GOMAXPROCS(runtime.NumCPU())
  url1 = site + ":" + strconv.Itoa(port)
      iotqq.Set(url1, qq)
  c, err := gosocketio.Dial(
      gosocketio.GetUrl(site, port, false),
      transport.GetDefaultWebsocketTransport())
  if err != nil {
      log.Fatal(err)
  }
  err = c.On("OnGroupiotqqs", func(h *gosocketio.Channel, args iotqq.Message) {
      var mess iotqq.Data = args.CurrentPacket.Data
      /*
          mess.Content 消息内容 string
          mess.FromGroupID 来源QQ群 int
          mess.FromUserID 来源QQ int64
          mess.iotqqType 消息类型 string
      */
      log.Println("群聊消息: ", mess.FromNickName+"<"+strconv.FormatInt(mess.FromUserID, 10)+">: "+mess.Content)
      // cm := strings.Split(mess.Content, " ")

  })
  if err != nil {
      log.Fatal(err)
  }
  /*
  err = c.On("OnFriendiotqqs", func(h *gosocketio.Channel, args iotqq.Message) {
      log.Println("私聊消息: ", args.CurrentPacket.Data.Content)
  })
  if err != nil {
      log.Fatal(err)
  }
  */
  err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
      log.Fatal("Disconnected")
  })
  if err != nil {
      log.Fatal(err)
  }
  err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
      log.Println("Connect success")
  })
  if err != nil {
      log.Fatal(err)
  }

  time.Sleep(1 * time.Second)
  go SendJoin(c)
  periodlycall(10*time.Second, getmsg)
  home:
  time.Sleep(600 * time.Second)
  SendJoin(c)
  goto home
  log.Println(" [x] Complete")
}
