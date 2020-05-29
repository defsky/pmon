package main

import (
    //"log"
    //"net/http"
    "os"
    "bytes"
    "os/exec"
    "errors"
    "strconv"
    "regexp"
    "time"
    "sync"
    "path/filepath"
    "github.com/gomodule/redigo/redis"
    "github.com/jander/golog/logger"
    "github.com/kardianos/service"
    "github.com/Unknwon/goconfig"
)

type program struct {
    IsRunning bool
    Pname string
    RedisHost string
    RedisKey string
    RedisConn redis.Conn
}

func (p *program) Start(s service.Service) error {
    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Fatal(err)
	}
    cfg, err := goconfig.LoadConfigFile(dir + "\\config.ini")
	if err != nil{
        logger.Fatal(err)
	}

	pname, err := cfg.GetValue("process", "name")
	if err != nil {
        logger.Fatal(err)
	}
    p.Pname = pname

	redisHost, err := cfg.GetValue("redis", "host")
	if err != nil {
        logger.Fatal(err)
	}
    p.RedisHost = redisHost

	redisKey, err := cfg.GetValue("redis", "key")
	if err != nil {
        logger.Fatal(err)
    }
    p.RedisKey = redisKey

    p.IsRunning = true
    go p.run()
    return nil
}

func (p *program) run() {
    var wg sync.WaitGroup

    statusCh := make(chan int)
    
    wg.Add(2)

    go func() {
        redisNeedReconnect := make(chan int)

        touchRedis:
            conn, err := redis.Dial("tcp",p.RedisHost)
            if err != nil {
                time.Sleep(3 * time.Second)
                goto touchRedis
            }
        
        p.RedisConn = conn

        defer func() {
            wg.Done()
        }()

        for ;p.IsRunning; {
            select {
            case <-redisNeedReconnect:
                conn, err = redis.Dial("tcp", p.RedisHost)
                if err != nil {
                    time.Sleep(3 * time.Second)
                    redisNeedReconnect <- 1
                }
                p.RedisConn = conn
            case v := <-statusCh:
                _, err := conn.Do("SET",p.RedisKey, v)
            
                if err != nil {
                    redisNeedReconnect <- 1
                }
            default:
                time.Sleep(1 * time.Second)
            }
        }
    }()

    go func() {
        defer wg.Done()

        for ;p.IsRunning; {
            pid, err := findProcessID(p.Pname)
            if err == nil {
                process, err := os.FindProcess(pid)
                if err == nil {
                    statusCh <- 1
                    process.Wait()
                }
            }
            
            statusCh <- 0
            
            time.Sleep(1 * time.Second)
        }
    }()

    wg.Wait()
}

func (p *program) Stop(s service.Service) error {
    p.RedisConn.Do("SET", p.RedisKey, 0)
    p.RedisConn.Close()

    p.IsRunning = false

    return nil
}

/**
* MAIN函数，程序入口
*/

func main() {
    svcConfig := &service.Config{
        Name:        "pmonsvc", //服务名称
        DisplayName: "Process Monitor", //服务显示名称
        Description: "此服务用作监控指定进程是否存活", //服务描述
    }

    prg := &program{}
    s, err := service.New(prg, svcConfig)
    if err != nil {
        logger.Fatal(err)
    }

    if err != nil {
        logger.Fatal(err)
    }

    if len(os.Args) > 1 {
        if os.Args[1] == "install" {
            s.Install()
            logger.Println("服务安装成功")
            return
        }

        if os.Args[1] == "remove" {
            s.Uninstall()
            logger.Println("服务卸载成功")
            return
        }
    }

    err = s.Run()
    if err != nil {
        logger.Error(err)
    }
}

func findProcessID(processName string)(int, error) {
    buf := bytes.Buffer{}
    cmd := exec.Command("wmic","process","get","processid,name")
    cmd.Stdout = &buf
    cmd.Run()
    
    cmd2 := exec.Command("findstr", processName)
    cmd2.Stdin = &buf
    data, _ := cmd2.CombinedOutput()
    
    if len(data) == 0 {
        return -1, errors.New("not found")
    }
    
    info := string(data)
    reg := regexp.MustCompile(`[0-9]+`)
    pid := reg.FindString(info)
    
    return strconv.Atoi(pid)
}