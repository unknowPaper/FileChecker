package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"os"
	"github.com/unknowPaper/FileChecker/config"
	//"io/ioutil"
	"os/user"
	"io/ioutil"
	"regexp"
)

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