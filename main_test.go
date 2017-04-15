package main

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/unknowPaper/FileChecker/config"
	"github.com/unknowPaper/FileChecker/logger"
	"io"
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

//func TestMain(m *testing.M) {
//	//mySetupFunction()
//
//
//	retCode := m.Run()
//	//myTeardownFunction()
//	os.Exit(retCode)
//}

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

//func TestSendMail(t *testing.T) {
//	testMailBody := `test mail body`
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mock_sender := mock_main.NewMockEmailSender(ctrl)
//
//	mock_sender.EXPECT().Send([]string{"to@gmail.com"}, NotificationTitle, testMailBody).Return(nil)
//
//	smtp.MOCK().SetController(ctrl)
//
//	err := sendEmail(testMailBody)
//	fmt.Printf("%v", err.Error())
//
//	//realSendEmail(mock_sender, "to@gmail.com", testMailBody)
//}

func TestGetMd5(t *testing.T) {
	createConfigFileForTest()

	l = logger.New("/dev/null")

	var returnMD5String string
	file, err := os.Open(testConfigFileName)
	if err != nil {
		print(err.Error())
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		print(err.Error())
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)

	testF, _ := os.Open(testConfigFileName)
	defer testF.Close()

	assert.Equal(t, returnMD5String, genMd5(testF))

	deleteTestConfigFile()
}

func TestGetContent(t *testing.T) {
	createConfigFileForTest()

	f, _ := os.Open(testConfigFileName)
	absPath, perr := getAbsPath(testConfigFileName)
	if perr != nil {
		print(perr.Error())
	}
	content := getContent(f, absPath)
	assert.Equal(t, "", content)

	diffFileExtension = append(diffFileExtension, " yaml")
	f.Seek(0, 0)
	content = getContent(f, absPath)
	assert.Equal(t, testConfigContent, content)

	deleteTestConfigFile()
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
