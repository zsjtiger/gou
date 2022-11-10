package gou

import (
	"fmt"
	"io/ioutil"
	"net"
	nethttp "net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/gou/http"
	"github.com/yaoapp/kun/any"
)

func TestHTTPGet(t *testing.T) {

	shutdown, ready, host := setup()
	go start(t, &host, shutdown, ready)
	defer stop(shutdown, ready)
	<-ready

	v := NewProcess("http.Get", fmt.Sprintf("%s/get?foo=bar", host),
		map[string]string{"hello": "world"},
		map[string]string{"Auth": "Test"},
	).Run()

	resp, ok := v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	assert.Equal(t, 200, resp.Status)
	res := any.Of(resp.Data).MapStr().Dot()
	assert.Equal(t, "bar", res.Get("query.foo[0]"))
	assert.Equal(t, "world", res.Get("query.hello[0]"))
	assert.Equal(t, "Test", res.Get("headers.Auth[0]"))

	v = NewProcess("http.Get", fmt.Sprintf("%s/get?foo=bar", host),
		map[int]int{1: 2},
		map[string]string{"Auth": "Test"},
	).Run()
	resp, ok = v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	assert.Equal(t, 400, resp.Status)
	assert.NotNil(t, resp.Message)
}

func TestHTTPHead(t *testing.T) {

	shutdown, ready, host := setup()
	go start(t, &host, shutdown, ready)
	defer stop(shutdown, ready)
	<-ready

	v := NewProcess("http.Head", fmt.Sprintf("%s/head?foo=bar", host),
		map[string]string{"name": "Lucy"},
		map[string]string{"hello": "world"},
		map[string]string{"Auth": "Test"},
	).Run()

	resp, ok := v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	assert.Equal(t, 200, resp.Status)

	v = NewProcess("http.Head", fmt.Sprintf("%s/head?foo=bar", host),
		map[string]string{"name": "Lucy"},
		map[int]int{1: 2},
		map[string]string{"Auth": "Test"},
	).Run()

	resp, ok = v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	assert.Equal(t, 400, resp.Status)
	assert.NotNil(t, resp.Message)
}

func TestHTTPPost(t *testing.T) {

	shutdown, ready, host := setup()
	go start(t, &host, shutdown, ready)
	defer stop(shutdown, ready)
	<-ready
	v := NewProcess("http.Post", fmt.Sprintf("%s/path?foo=bar", host),
		map[string]string{"name": "Lucy"},
		nil,
		map[string]string{"hello": "world"},
		map[string]string{"Auth": "Test"},
	).Run()

	resp, ok := v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	assert.Equal(t, 200, resp.Status)
	res := any.Of(resp.Data).MapStr().Dot()
	assert.Equal(t, "bar", res.Get("query.foo[0]"))
	assert.Equal(t, "world", res.Get("query.hello[0]"))
	assert.Equal(t, "Test", res.Get("headers.Auth[0]"))
	assert.Equal(t, `{"name":"Lucy"}`, res.Get("payload"))

	// Post File via payload
	root, file := tmpfile(t, "Hello World via payload")
	err := SetFileRoot(root)
	if err != nil {
		t.Fatal(err)
	}

	v = NewProcess("http.Post", fmt.Sprintf("%s/path?foo=bar", host),
		file,
		nil,
		map[string]string{"hello": "world"},
		map[string]string{"Auth": "Test", "Content-Type": "multipart/form-data"},
	).Run()

	resp, ok = v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}

	assert.Equal(t, 200, resp.Status)
	res = any.Of(resp.Data).MapStr().Dot()
	assert.Equal(t, "bar", res.Get("query.foo[0]"))
	assert.Equal(t, "world", res.Get("query.hello[0]"))
	assert.Equal(t, "Test", res.Get("headers.Auth[0]"))
	assert.Contains(t, res.Get("payload"), "Hello World via payload")

	// Post File via files
	root, file = tmpfile(t, "Hello World via files")
	err = SetFileRoot(root)
	if err != nil {
		t.Fatal(err)
	}

	v = NewProcess("http.Post", fmt.Sprintf("%s/path?foo=bar", host),
		map[string]string{"name": "Lucy"},
		map[string]interface{}{"file": file},
		map[string]string{"hello": "world"},
		map[string]string{"Auth": "Test", "Content-Type": "multipart/form-data"},
	).Run()

	resp, ok = v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}

	assert.Equal(t, 200, resp.Status)
	res = any.Of(resp.Data).MapStr().Dot()
	assert.Equal(t, "bar", res.Get("query.foo[0]"))
	assert.Equal(t, "world", res.Get("query.hello[0]"))
	assert.Equal(t, "Test", res.Get("headers.Auth[0]"))
	assert.Contains(t, res.Get("payload"), "Hello World via files")
	assert.Contains(t, res.Get("payload"), "Lucy")

	// Post Error
	v = NewProcess("http.Post", fmt.Sprintf("%s/path?foo=bar", host),
		map[string]string{"name": "Lucy"},
		nil,
		map[int]int{1: 2},
		map[string]string{"Auth": "Test"},
	).Run()

	resp, ok = v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	assert.Equal(t, 400, resp.Status)
	assert.NotNil(t, resp.Message)
}

func TestHTTPOthers(t *testing.T) {

	shutdown, ready, host := setup()
	go start(t, &host, shutdown, ready)
	defer stop(shutdown, ready)
	<-ready

	methods := []string{"http.Put", "http.Patch", "http.Delete"}
	for _, method := range methods {
		v := NewProcess(method, fmt.Sprintf("%s/path?foo=bar", host),
			map[string]string{"name": "Lucy"},
			map[string]string{"hello": "world"},
			map[string]string{"Auth": "Test"},
		).Run()

		resp, ok := v.(*http.Response)
		if !ok {
			t.Fatal(fmt.Errorf("response error %#v", v))
		}
		assert.Equal(t, 200, resp.Status)
		res := any.Of(resp.Data).MapStr().Dot()
		assert.Equal(t, "bar", res.Get("query.foo[0]"))
		assert.Equal(t, "world", res.Get("query.hello[0]"))
		assert.Equal(t, "Test", res.Get("headers.Auth[0]"))
		assert.Equal(t, `{"name":"Lucy"}`, res.Get("payload"))

		v = NewProcess(method, fmt.Sprintf("%s/path?foo=bar", host),
			map[string]string{"name": "Lucy"},
			map[int]int{1: 2},
			map[string]string{"Auth": "Test"},
		).Run()

		resp, ok = v.(*http.Response)
		if !ok {
			t.Fatal(fmt.Errorf("response error %#v", v))
		}
		assert.Equal(t, 400, resp.Status)
		assert.NotNil(t, resp.Message)
	}
}

func TestHTTPSend(t *testing.T) {

	shutdown, ready, host := setup()
	go start(t, &host, shutdown, ready)
	defer stop(shutdown, ready)
	<-ready

	v := NewProcess("http.Send", "POST", fmt.Sprintf("%s/path?foo=bar", host),
		map[string]string{"name": "Lucy"},
		map[string]string{"hello": "world"},
		map[string]string{"Auth": "Test"},
	).Run()

	resp, ok := v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	if resp.Status == 0 {
		fmt.Println(resp.Message)
	}

	assert.Equal(t, 200, resp.Status)
	res := any.Of(resp.Data).MapStr().Dot()
	assert.Equal(t, "bar", res.Get("query.foo[0]"))
	assert.Equal(t, "world", res.Get("query.hello[0]"))
	assert.Equal(t, "Test", res.Get("headers.Auth[0]"))
	assert.Equal(t, `{"name":"Lucy"}`, res.Get("payload"))

	v = NewProcess("http.Send", "POST", fmt.Sprintf("%s/path?foo=bar", host),
		map[string]string{"name": "Lucy"},
		map[int]int{1: 2},
		map[string]string{"Auth": "Test"},
	).Run()

	resp, ok = v.(*http.Response)
	if !ok {
		t.Fatal(fmt.Errorf("response error %#v", v))
	}
	assert.Equal(t, 400, resp.Status)
	assert.NotNil(t, resp.Message)
}

func setup() (chan bool, chan bool, string) {
	return make(chan bool, 1), make(chan bool, 1), ""
}

func start(t *testing.T, host *string, shutdown, ready chan bool) {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	errCh := make(chan error, 1)

	// Set router
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	router := gin.New()

	router.GET("/get", testHanlder)
	router.HEAD("/head", testHanlder)
	router.POST("/path", testHanlder)
	router.PUT("/path", testHanlder)
	router.PATCH("/path", testHanlder)
	router.DELETE("/path", testHanlder)

	// Listen
	l, err := net.Listen("tcp4", ":0")
	if err != nil {
		errCh <- fmt.Errorf("Error: can't get port")
	}

	srv := &nethttp.Server{Addr: ":0", Handler: router}
	defer func() {
		srv.Close()
		l.Close()
	}()

	// start serve
	go func() {
		fmt.Println("[TestServer] Starting")
		if err := srv.Serve(l); err != nil && err != nethttp.ErrServerClosed {
			fmt.Println("[TestServer] Error:", err)
			errCh <- err
		}
	}()

	addr := strings.Split(l.Addr().String(), ":")
	if len(addr) != 2 {
		errCh <- fmt.Errorf("Error: can't get port")
	}

	*host = fmt.Sprintf("http://127.0.0.1:%s", addr[1])
	time.Sleep(50 * time.Millisecond)
	ready <- true
	fmt.Printf("[TestServer] %s", *host)

	select {

	case <-shutdown:
		fmt.Println("[TestServer] Stop")
		break

	case <-interrupt:
		fmt.Println("[TestServer] Interrupt")
		break

	case err := <-errCh:
		fmt.Println("[TestServer] Error:", err.Error())
		break
	}
}

func stop(shutdown, ready chan bool) {
	ready <- false
	shutdown <- true
	time.Sleep(50 * time.Millisecond)
}

func tmpfile(t *testing.T, content string) (string, string) {
	file, err := os.CreateTemp("", "-data")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(file.Name(), []byte(content), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(file.Name()), filepath.Base(file.Name())
}

func testHanlder(c *gin.Context) {
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error(), "code": 400})
		return
	}
	c.JSON(200, gin.H{
		"payload": string(payload),
		"query":   c.Request.URL.Query(),
		"headers": c.Request.Header,
	})
}