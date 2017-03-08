package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestNewFunc(t *testing.T) {
	c := New()
	err := c.Load("tests/test.yaml")
	assert.NoError(t, err)

	err = c.Load("tests/test.yml")
	assert.NoError(t, err)

	err = c.Load("tests/error.yml")
	assert.Error(t, err)
	assert.Equal(t, "Can not parse tests/error.yml config", err.Error())
}

func TestMissingConfig(t *testing.T) {
	c := New()
	err := c.Load("tests/test3.yaml")

	assert.Error(t, err)

	err = c.Load("tests/test3.txt")

	assert.Error(t, err)
	assert.Equal(t, "Can not load tests/test3.txt config", err.Error())
}

func TestNewWithDefaultConfig(t *testing.T) {
	c := New(map[interface{}]interface{}{"mode": "ONLINE", "DEFAULT": "DD"})
	err := c.Load("tests/test.yaml")

	assert.NoError(t, err)

	mode := c.Get("mode")
	defaultV := c.Get("DEFAULT")

	assert.Equal(t, "TEST", mode)
	assert.Equal(t, "DD", defaultV)
}

type ConfigTestSuite struct {
	suite.Suite
	C *Engine
}

func (suite *ConfigTestSuite) SetupTest() {
	suite.C = &Engine{}

	if err := suite.C.Load("tests/test.yaml"); err != nil {
		panic("failed to load config.yml")
	}
}

func (suite *ConfigTestSuite) TestGet() {
	isExist := suite.C.Get("db.appleboy")
	assert.Nil(suite.T(), isExist)

	// get integer
	port := suite.C.Get("port")
	assert.Equal(suite.T(), 8080, port)

	// get boolean
	enable := suite.C.Get("enable")
	assert.Equal(suite.T(), true, enable)

	// get string
	mode := suite.C.Get("mode")
	assert.Equal(suite.T(), "TEST", mode)

	// missing key
	isNil := suite.C.Get("not_exists")
	assert.Nil(suite.T(), isNil)

	// test subtree key
	// get string
	dbDriver := suite.C.Get("db.Sqldriver")
	assert.Equal(suite.T(), "MySQL", dbDriver)

	// get boolean
	dbEnable := suite.C.Get("db.enable")
	assert.Equal(suite.T(), true, dbEnable)

	// get int
	dbPort := suite.C.Get("db.port")
	assert.Equal(suite.T(), 3306, dbPort)

	// missing key
	test := suite.C.Get("db.test")
	assert.Nil(suite.T(), test)

	gpa := suite.C.Get("db.gpa")
	assert.Equal(suite.T(), "3.7", gpa)

	average := suite.C.Get("db.average")
	assert.Equal(suite.T(), 2.0, average)

	floatzero := suite.C.Get("db.floatzero")
	assert.Equal(suite.T(), 0.0, floatzero)
}

func (suite *ConfigTestSuite) TestGetString() {
	// get string
	dbDriver := suite.C.GetString("db.Sqldriver")
	assert.Equal(suite.T(), "MySQL", dbDriver)

	// get boolean
	dbEnable := suite.C.GetString("db.enable")
	assert.Equal(suite.T(), "true", dbEnable)

	// get boolean
	dbDisable := suite.C.GetString("db.disable")
	assert.Equal(suite.T(), "false", dbDisable)

	// get int
	dbPort := suite.C.GetString("db.port")
	assert.Equal(suite.T(), "3306", dbPort)

	// missing key
	test := suite.C.GetString("db.test")
	assert.Equal(suite.T(), "", test)

	gpa := suite.C.GetString("db.gpa")
	assert.Equal(suite.T(), "3.7", gpa)

	average := suite.C.GetString("db.average")
	assert.Equal(suite.T(), "2", average)

	floatzero := suite.C.GetString("db.floatzero")
	assert.Equal(suite.T(), "0", floatzero)

	sint := suite.C.GetString("db.sint")
	assert.Equal(suite.T(), "2000", sint)

	coverage := suite.C.GetString("db.coverage")
	assert.Equal(suite.T(), "2.1", coverage)
}

func (suite *ConfigTestSuite) TestGetInt() {
	// get string
	dbDriver := suite.C.GetInt("db.Sqldriver")
	assert.Equal(suite.T(), 0, dbDriver)

	// get boolean
	dbEnable := suite.C.GetInt("db.enable")
	assert.Equal(suite.T(), 1, dbEnable)

	// get boolean
	dbDisable := suite.C.GetInt("db.disable")
	assert.Equal(suite.T(), 0, dbDisable)

	// get int
	dbPort := suite.C.GetInt("db.port")
	assert.Equal(suite.T(), 3306, dbPort)

	// missing key
	test := suite.C.GetInt("db.test")
	assert.Equal(suite.T(), 0, test)

	gpa := suite.C.GetInt("db.gpa")
	assert.Equal(suite.T(), 0, gpa)

	average := suite.C.GetInt("db.average")
	assert.Equal(suite.T(), 2, average)

	floatzero := suite.C.GetInt("db.floatzero")
	assert.Equal(suite.T(), 0, floatzero)

	sint := suite.C.GetInt("db.sint")
	assert.Equal(suite.T(), 2000, sint)

	coverage := suite.C.GetInt("db.coverage")
	assert.Equal(suite.T(), 2, coverage)
}

func (suite *ConfigTestSuite) TestGetBool() {
	// get string
	dbDriver := suite.C.GetBool("db.Sqldriver")
	assert.Equal(suite.T(), false, dbDriver)

	isTest := suite.C.GetBool("db.isTest")
	assert.Equal(suite.T(), true, isTest)

	isTrue := suite.C.GetBool("db.isTrue")
	assert.Equal(suite.T(), false, isTrue)

	// get boolean
	dbEnable := suite.C.GetBool("db.enable")
	assert.Equal(suite.T(), true, dbEnable)

	// get boolean
	dbDisable := suite.C.GetBool("db.disable")
	assert.Equal(suite.T(), false, dbDisable)

	// get int
	dbPort := suite.C.GetBool("db.port")
	assert.Equal(suite.T(), true, dbPort)

	isFalse := suite.C.GetBool("db.isFalse")
	assert.Equal(suite.T(), true, isFalse)

	dbTimeout := suite.C.GetBool("db.timeout")
	assert.Equal(suite.T(), false, dbTimeout)

	// missing key
	test := suite.C.GetBool("db.test")
	assert.Equal(suite.T(), false, test)

	average := suite.C.GetBool("db.average")
	assert.Equal(suite.T(), true, average)

	floatzero := suite.C.GetBool("db.floatzero")
	assert.Equal(suite.T(), false, floatzero)
}

func (suite *ConfigTestSuite) TestGetFloat64() {
	// get string
	dbDriver := suite.C.GetFloat64("db.Sqldriver")
	assert.Equal(suite.T(), float64(0), dbDriver)

	// get boolean
	dbEnable := suite.C.GetFloat64("db.enable")
	assert.Equal(suite.T(), float64(1), dbEnable)

	// get boolean
	dbDisable := suite.C.GetFloat64("db.disable")
	assert.Equal(suite.T(), float64(0), dbDisable)

	// get int
	dbPort := suite.C.GetFloat64("db.port")
	assert.Equal(suite.T(), float64(3306), dbPort)

	// missing key
	test := suite.C.GetFloat64("db.test")
	assert.Equal(suite.T(), float64(0), test)

	average := suite.C.GetFloat64("db.average")
	assert.Equal(suite.T(), 2.0, average)

	gpa := suite.C.GetFloat64("db.gpa")
	assert.Equal(suite.T(), 3.7, gpa)

	timeout := suite.C.GetFloat64("db.timeout")
	assert.Equal(suite.T(), 0.0, timeout)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func TestConfig_OverLoad(t *testing.T) {
	c := &Engine{}
	err := c.Load("tests/test.yaml")

	mode := c.Get("mode")
	assert.NoError(t, err)
	assert.Equal(t, "TEST", mode)

	err = c.Load("tests/test2.yaml")
	mode = c.Get("mode")

	assert.NoError(t, err)
	assert.Equal(t, "ONLINE", mode)
}

type testStruct struct {
	Sqldriver string
	Protocol  string
	Hostname  string
	Username  string
	Password  string
	Dbname    string
	Charset   string
}

func TestGetStruct(t *testing.T) {
	c := &Engine{}
	err := c.Load("tests/test.yaml")

	assert.NoError(t, err)

	ts := &testStruct{}
	c.GetStruct("db", ts)

	assert.Equal(t, "MySQL", ts.Sqldriver)
}

type testMultiStruct struct {
	Second struct {
		Third struct {
			Value string
		}
	}
}

func TestGetMultiStruct(t *testing.T) {
	c := &Engine{}
	err := c.Load("tests/test.yaml")

	assert.NoError(t, err)

	tms := &testMultiStruct{}
	c.GetStruct("First", tms)

	assert.Equal(t, "Test", tms.Second.Third.Value)
}
