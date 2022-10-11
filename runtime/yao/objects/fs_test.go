package objects

import (
	"fmt"
	iofs "io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/gou/fs"
	"github.com/yaoapp/gou/runtime/yao/bridge"
	"rogchap.com/v8go"
)

func TestFSObjectReadFile(t *testing.T) {
	testFsClear(t)
	f := testFsFiles(t)
	data := testFsMakeF1(t)

	initTestEngine()
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	fs := &FSOBJ{}
	global := v8go.NewObjectTemplate(iso)
	global.Set("FS", fs.ExportFunction(iso))

	ctx := v8go.NewContext(iso, global)
	defer ctx.Close()

	// ReadFile
	v, err := ctx.RunScript(fmt.Sprintf(`
	function ReadFile() {
		var fs = new FS("system")
		var data = fs.ReadFile("%s");
		return data
	}
	ReadFile()
	`, f["F1"]), "")

	if err != nil {
		t.Fatal(err)
	}

	res, err := bridge.ToInterface(v)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(data), res)

	// ReadFileBuffer
	v, err = ctx.RunScript(fmt.Sprintf(`
		function ReadFileBuffer() {
			var fs = new FS("system")
			var data = fs.ReadFileBuffer("%s");
			return data
		}
		ReadFileBuffer()
		`, f["F1"]), "")

	assert.True(t, v.IsUint8Array())

	res, err = bridge.ToInterface(v)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, data, res)

	// ReadFileBuffer
	v, err = ctx.RunScript(fmt.Sprintf(`
	function ReadFileBuffer() {
		var fs = new FS("system")
		var data = fs.ReadFileBuffer("%s");
		return data
	}
	ReadFileBuffer()
	`, f["F1"]), "")

	if err != nil {
		t.Fatal(err)
	}

	res, err = bridge.ToInterface(v)
	assert.True(t, v.IsUint8Array())
	assert.Equal(t, data, res)
}

func TestFSObjectWriteFile(t *testing.T) {
	testFsClear(t)
	f := testFsFiles(t)
	data := testFsMakeF1(t)

	initTestEngine()
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	fs := &FSOBJ{}
	global := v8go.NewObjectTemplate(iso)
	global.Set("FS", fs.ExportFunction(iso))

	ctx := v8go.NewContext(iso, global)
	defer ctx.Close()

	// WriteFile
	v, err := ctx.RunScript(fmt.Sprintf(`
	function WriteFile() {
		var fs = new FS("system")
		return fs.WriteFile("%s", "%s", 0644);
	}
	WriteFile()
	`, f["F1"], string(data)), "")

	if err != nil {
		t.Fatal(err)
	}

	res, err := bridge.ToInterface(v)
	if err != nil {
		t.Fatal(err)
	}

	// WriteFileBuffer
	v, err = ctx.RunScript(fmt.Sprintf(`
		function WriteFileBuffer() {
			var fs = new FS("system")
			var data = fs.ReadFileBuffer("%s")
			return fs.WriteFileBuffer("%s", data, 0644);
		}
		WriteFileBuffer()
		`, f["F1"], f["F2"]), "")

	if err != nil {
		t.Fatal(err)
	}

	res, err = bridge.ToInterface(v)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(data), res)

}

func TestFSObjectExistRemove(t *testing.T) {

	testFsClear(t)
	testFsMakeF1(t)
	testFsMakeD1D2F1(t)
	f := testFsFiles(t)

	initTestEngine()
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	fs := &FSOBJ{}
	global := v8go.NewObjectTemplate(iso)
	global.Set("FS", fs.ExportFunction(iso))

	ctx := v8go.NewContext(iso, global)
	defer ctx.Close()

	// WriteFile
	v, err := ctx.RunScript(fmt.Sprintf(`
	function ExistRemove() {
		var res = {}
		var fs = new FS("system")
		res["ExistsTrue"] = fs.Exists("%s");
		res["ExistsFalse"] = fs.Exists("%s");
		res["IsDirTrue"] = fs.IsDir("%s");
		res["IsDirFalse"] = fs.IsDir("%s");
		res["IsFileTrue"] = fs.IsFile("%s");
		res["IsFileFalse"] = fs.IsFile("%s");
		res["Remove"] = fs.Remove("%s");
		res["RemoveNotExists"] = fs.Remove("%s");
		try {
			fs.Remove("%s");
		} catch( err ) {
			res["RemoveError"] = err.message
		}
		res["RemoveAll"] = fs.RemoveAll("%s");
		res["RemoveAllNotExists"] = fs.RemoveAll("%s");
		return res
	}
	ExistRemove()
	`,
		f["F1"], f["F2"], f["D1"], f["F1"], f["F1"], f["D1"], f["F1"], f["F2"], f["D1"], f["D1"], f["D1_D2"],
	), "")

	if err != nil {
		t.Fatal(err)
	}

	res, err := bridge.ToInterface(v)
	if err != nil {
		t.Fatal(err)
	}

	retval := res.(map[string]interface{})
	assert.Equal(t, true, retval["ExistsTrue"])
	assert.Equal(t, false, retval["ExistsFalse"])
	assert.Equal(t, true, retval["IsDirTrue"])
	assert.Equal(t, false, retval["IsDirFalse"])
	assert.Equal(t, true, retval["IsFileTrue"])
	assert.Equal(t, false, retval["IsFileFalse"])
	assert.Nil(t, retval["Remove"])
	assert.Nil(t, retval["RemoveNotExists"])
	assert.Contains(t, retval["RemoveError"], "directory not empty")
	assert.Nil(t, retval["RemoveAll"])
	assert.Nil(t, retval["RemoveAllNotExists"])
}

func TestFSObjectFileInfo(t *testing.T) {

	testFsClear(t)
	data := testFsMakeF1(t)
	testFsMakeD1D2F1(t)
	f := testFsFiles(t)

	initTestEngine()
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	fs := &FSOBJ{}
	global := v8go.NewObjectTemplate(iso)
	global.Set("FS", fs.ExportFunction(iso))

	ctx := v8go.NewContext(iso, global)
	defer ctx.Close()

	// WriteFile
	v, err := ctx.RunScript(fmt.Sprintf(`
	function FileInfo() {
		var res = {}
		var fs = new FS("system")
		res["BaseName"] = fs.BaseName("%s");
		res["DirName"] = fs.DirName("%s");
		res["ExtName"] = fs.ExtName("%s");
		res["MimeType"] = fs.MimeType("%s");
		res["Size"] = fs.Size("%s");
		res["ModTime"] = fs.ModTime("%s");
		res["Mode"] = fs.Mode("%s");
		res["Chmod"] = fs.Chmod("%s", 0755);
		res["ModeAfter"] = fs.Mode("%s");
		return res
	}
	FileInfo()
	`,
		f["F1"], f["F1"], f["F1"], f["F1"], f["F1"], f["F1"], f["F1"], f["F1"], f["F1"],
	), "")

	if err != nil {
		t.Fatal(err)
	}

	res, err := bridge.ToInterface(v)
	if err != nil {
		t.Fatal(err)
	}

	ret := res.(map[string]interface{})
	assert.Equal(t, "f1.file", ret["BaseName"])
	assert.Equal(t, f["root"], ret["DirName"])
	assert.Equal(t, "file", ret["ExtName"])
	assert.Equal(t, "text/plain; charset=utf-8", ret["MimeType"])
	assert.Equal(t, len(data), int(ret["Size"].(float64)))
	assert.Equal(t, iofs.FileMode(0644), iofs.FileMode(int(ret["Mode"].(float64))))
	assert.Equal(t, iofs.FileMode(0755), iofs.FileMode(int(ret["ModeAfter"].(float64))))
	assert.Equal(t, true, int(time.Now().Unix()) >= int(ret["ModTime"].(float64)))
	assert.Nil(t, ret["Chmod"])

}

func TestFSObjectDir(t *testing.T) {
	testFsClear(t)
	f := testFsFiles(t)
	initTestEngine()
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	fs := &FSOBJ{}
	global := v8go.NewObjectTemplate(iso)
	global.Set("FS", fs.ExportFunction(iso))

	ctx := v8go.NewContext(iso, global)
	defer ctx.Close()

	// DirTest
	v, err := ctx.RunScript(fmt.Sprintf(`
	function DirTest() {
		var fs = new FS("system");
		fs.Mkdir("%s");
		fs.MkdirAll("%s");
		fs.MkdirTemp()
		fs.MkdirTemp("%s")
		fs.MkdirTemp("%s", "*-logs")
		return fs.ReadDir("%s", true)
	}
	DirTest()
	`, f["D2"], f["D1_D2"], f["D1"], f["D1"], f["root"]), "")

	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, v.IsArray())
	res, err := bridge.ToInterface(v)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 5, len(res.([]interface{})))
}

func testFsMakeF1(t *testing.T) []byte {
	data := testFsData(t)
	f := testFsFiles(t)

	// Write
	_, err := fs.WriteFile(fs.FileSystems["system"], f["F1"], data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func testFsData(t *testing.T) []byte {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(10) + 1
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return []byte(fmt.Sprintf("HELLO WORLD %s", string(b)))
}

func testFsFiles(t *testing.T) map[string]string {

	root := filepath.Join(os.Getenv("GOU_TEST_APP_ROOT"), "data")
	return map[string]string{
		"root":     root,
		"F1":       filepath.Join(root, "f1.file"),
		"F2":       filepath.Join(root, "f2.file"),
		"F3":       filepath.Join(root, "f3.js"),
		"D1_F1":    filepath.Join(root, "d1", "f1.file"),
		"D1_F2":    filepath.Join(root, "d1", "f2.file"),
		"D2_F1":    filepath.Join(root, "d2", "f1.file"),
		"D2_F2":    filepath.Join(root, "d2", "f2.file"),
		"D1_D2_F1": filepath.Join(root, "d1", "d2", "f1.file"),
		"D1_D2_F2": filepath.Join(root, "d1", "d2", "f2.file"),
		"D1":       filepath.Join(root, "d1"),
		"D2":       filepath.Join(root, "d2"),
		"D1_D2":    filepath.Join(root, "d1", "d2"),
	}

}

func testFsClear(t *testing.T) {
	stor := fs.FileSystems["system"]
	root := filepath.Join(os.Getenv("GOU_TEST_APP_ROOT"), "data")
	err := os.RemoveAll(root)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	err = fs.MkdirAll(stor, root, int(os.ModePerm))
	if err != nil {
		t.Fatal(err)
	}
}

func testFsMakeD1D2F1(t *testing.T) []byte {
	data := testFsData(t)
	f := testFsFiles(t)
	_, err := fs.WriteFile(fs.FileSystems["system"], f["D1_D2_F1"], data, 0644)
	if err != nil {
		t.Fatal(err)
	}
	return data
}