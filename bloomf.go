package main 

import (
	"fmt"
    "os"
    "log"
    "bufio"
    "strings"
    "strconv"
    "io"
    "./bloom"
    "net/http"
    "time"
    "runtime"
	"io/ioutil"
	"os/exec"
)


var myBloom *bloom.BloomFilter
var appendFileName = "append.txt"
var dbFileName = "data.db"
var tmpdbFileName = "data_tmp.db"
var checkAppendTime = 5 * time.Second //seconds

var Usage = func() {
    fmt.Println("Usage: \n ./bloomf append file1\n./bloomf search key1\n./bloomf web 9090")
}

func checkAppend() {

    for {
        runtime.Gosched()
        log.Println("check append...")
        if _, err := os.Stat(appendFileName); err == nil {
            appendData(appendFileName)
            os.Rename(appendFileName,"./datas/"+appendFileName+"_"+ strconv.FormatInt(time.Now().Unix(),10))
        }
        time.Sleep(checkAppendTime)
    }
}

func search(w http.ResponseWriter, r *http.Request) {
    key := r.FormValue("key")
    if myBloom.TestString(key) == true {
        io.WriteString(w, "1")
    } else {
        io.WriteString(w, "0")
    }
}

var proNum int64

func appendPerFile(fileName string, ch chan int) {
    f, err := os.Open(fileName)
    if err != nil {
        log.Printf("error: %s\n",err)
        return
    }
    defer f.Close()

    buf := bufio.NewReader(f)
    for {
        line, err := buf.ReadString('\n')
        line = strings.TrimSpace(line)
        proNum++
        //fmt.Printf("\r%3d", proNum)

        if line != "" {
            myBloom.AddString(line)
            //log.Println("AddString: "+ line)
        }
        if err != nil {
            if err == io.EOF {
                break
            }
        }
    }
    os.Remove("./"+fileName)
    ch <- 1
}

func splitFile(inputFile string) {
    pid := strconv.Itoa(os.Getpid())
    fileArr := make([]string,100)
    cmd := exec.Command("/usr/bin/split","-b 512m","-d", inputFile, pid+"_sub_")
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }

	files, _ := ioutil.ReadDir("./")
	for i, file := range files {
		if file.IsDir() {
			continue
		} else {
            if strings.Contains(file.Name(),pid+"_sub_") == true {
                fileArr[i] = file.Name()
            }
        }
    }
	chs := make([]chan int, 100)

   //fmt.Println("append rows:")
    for i,fn := range(fileArr) {
        if fn == "" {
            continue
        }
		chs[i] = make(chan int)
		go appendPerFile(fn, chs[i])
    }

    for i,fn := range(fileArr) {
        if fn == "" {
            continue
        }
		<-chs[i]
    }
}

func appendData(fileName string) {
        stream_tmp, err := os.OpenFile("./"+tmpdbFileName, os.O_RDWR|os.O_CREATE,0666)
        if err != nil {
            log.Printf("error: %s\n",err)
        }
        defer stream_tmp.Close()

        log.Printf("begining append file: %s\n",fileName)
        splitFile(fileName)


        myBloom.WriteTo(stream_tmp) // Write to db file
        os.Rename("./"+tmpdbFileName,"./"+dbFileName)
        log.Println("append ok.")
}

func main() {
    runtime.GOMAXPROCS(20)
    args := os.Args
    if args == nil || len(args) < 3 {
        Usage()
        return
    }
    myBloom = bloom.NewWithEstimates(400000000, 0.0001)
    stream, err := os.OpenFile("./"+dbFileName,os.O_RDWR|os.O_CREATE,0666)
    if err != nil {
        log.Println("Error:",err)
        return
    }
    defer stream.Close()

    log.Println("begin loading from db...")
    myBloom.ReadFrom(stream)  // Loading data from db file
    log.Println("loading db end.")

    switch args[1] {
    case "append":
        appendData(args[2])
    case "search":
        log.Printf("search %s, %v\n", args[2],myBloom.TestString(args[2]))
    case "web":
        go checkAppend()
        http.HandleFunc("/search",search)
        err := http.ListenAndServe(":"+args[2],nil)
        if err != nil {
            log.Fatal("ListenAndServer faild:",err)
        }
    default:
            Usage()
    }
}

