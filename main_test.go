package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/unknowPaper/FileChecker/config"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"strings"
	"testing"
)

var testConfigFileName = "test_config.yaml"
var testConfigContent = `scanDir: /bin, /sbin
excludeDir: .git
excludeFile: .gitignore
diffExtension: go, php
storeDriver: mysql
logPath: mylog.log
notification:
  smtp: smtp.gmail.com
  port: 587
  account: account
  pass: password
  from: account@gmail.com
  to: to@gmail.com
mysql:
  Protocol: tcp
  host: localhost
  username: root
  password:
  database: filesmd5`

func createConfigFileForTest() {

	// write the whole body at once
	err := ioutil.WriteFile(testConfigFileName, []byte(testConfigContent), 0644)
	if err != nil {
		panic(err)
	}
}

func deleteTestConfigFile() {
	os.Remove(testConfigFileName)
}

func TestReadConfig(t *testing.T) {
	createConfigFileForTest()
	defer deleteTestConfigFile()

	readConfig(testConfigFileName)
	assert.Equal(t, "/bin, /sbin", conf.GetString("scanDir"))

	assert.Equal(t, strings.Split(conf.GetString("scanDir"), ","), scanDir)
}

func TestConnectDbFailed(t *testing.T) {

	err := connectDb("user", "pass", "dbname")

	assert.Error(t, err)
}

func TestScanDirEmpty(t *testing.T) {
	// read from config
	err := mixScanDir("")
	assert.NoError(t, err)

	// read from config and flag
	expectScanDir := append(scanDir, "/tmp")
	err = mixScanDir("/tmp")
	assert.NoError(t, err)
	assert.Equal(t, expectScanDir, scanDir)

	// no any scan dir
	scanDir = []string{}
	err = mixScanDir("")
	assert.Error(t, err)

	// read from flag only
	err = mixScanDir("/bin")
	assert.NoError(t, err)
}

func TestInSlice(t *testing.T) {
	testSlice := []string{"abc", "defg", " hij "}

	// test find
	assert.Equal(t, true, inSlice("abc", testSlice))

	// test find with space
	assert.Equal(t, true, inSlice("hij", testSlice))

	// test can not search
	assert.Equal(t, false, inSlice("abcd", testSlice))
}

func TestCreateLogFileWithDefault(t *testing.T) {
	DEBUG = false
	conf = config.New()

	createLogFile("")

	assert.NotNil(t, l)

	usr, err := user.Current()
	assert.NoError(t, err)

	d, err := os.Open(usr.HomeDir + "/FileChecker")
	defer d.Close()
	assert.NoError(t, err)

	l.Info("test")

	f, err := os.Open(usr.HomeDir + "/FileChecker/FileChecker.log")
	defer f.Close()
	assert.NoError(t, err)

	content, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	//assert.Equal(t, "test", string(content))
	assert.Regexp(t, regexp.MustCompile(`\[INFO\].*test`), string((content)))

	os.Remove(usr.HomeDir + "/FileChecker/FileChecker.log")
	os.Remove(usr.HomeDir + "/FileChecker")

	l = nil
}

func TestCreateLogFileWithDebugDefault(t *testing.T) {
	DEBUG = true
	conf = config.New()

	createLogFile("")
	assert.NotNil(t, l)

	usr, err := user.Current()
	assert.NoError(t, err)

	d, err := os.Open(usr.HomeDir + "/FileChecker")
	defer d.Close()
	assert.NoError(t, err)

	l.Debug("test")

	f, err := os.Open(usr.HomeDir + "/FileChecker/debug.log")
	defer f.Close()
	assert.NoError(t, err)

	content, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	//assert.Equal(t, "test", string(content))
	assert.Regexp(t, regexp.MustCompile(`\[DEBUG\].*test`), string((content)))

	os.Remove(usr.HomeDir + "/FileChecker/debug.log")
	os.Remove(usr.HomeDir + "/FileChecker")

	l = nil
}

func TestCreateLogFileWithConfig(t *testing.T) {
	DEBUG = false

	usr, err := user.Current()
	assert.NoError(t, err)

	c := map[interface{}]interface{}{
		"logPath": usr.HomeDir + "/testlog.log",
	}
	conf = config.New(c)

	createLogFile("")
	assert.NotNil(t, l)

	l.Info("test")

	f, err := os.Open(usr.HomeDir + "/testlog.log")
	defer f.Close()
	assert.NoError(t, err)

	content, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	//assert.Equal(t, "test", string(content))
	assert.Regexp(t, regexp.MustCompile(`\[INFO\].*test`), string((content)))

	os.Remove(usr.HomeDir + "/testlog.log")

	l = nil
}

func TestCreateLogFileWithFlag(t *testing.T) {
	DEBUG = true
	conf = config.New()

	usr, err := user.Current()
	assert.NoError(t, err)

	filePath := usr.HomeDir + "/flaglog.log"

	createLogFile(filePath)
	assert.NotNil(t, l)

	l.Info("test")

	f, err := os.Open(filePath)
	defer f.Close()
	assert.NoError(t, err)

	content, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	//assert.Equal(t, "test", string(content))
	assert.Regexp(t, regexp.MustCompile(`\[INFO\].*test`), string((content)))

	os.Remove(filePath)

	l = nil
}
