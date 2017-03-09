package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"database/sql"
	"encoding/hex"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/unknowPaper/FileChecker/config"
	"github.com/unknowPaper/FileChecker/logger"
	"github.com/urfave/cli"
	"net/smtp"
	"regexp"
	"strings"
	"os/user"
)

var db *sql.DB

var conf *config.Engine

var l *logger.Logger

var scanDir []string
var diffFileExtension []string
var excludeDir []string
var excludeFile []string

var isCheck = false
var isRenew = false

var findFileStmt *sql.Stmt
var insertFileStmt *sql.Stmt
var updateFileStmt *sql.Stmt

var DEBUG = false

func main() {
	app := cli.NewApp()

	app.Name = "File MD5 Record"
	app.Version = "0.1"
	app.Usage = ""
	app.UsageText = "FileChecker -d DIR_NAME scan\n   FileChecker -cfg config.yaml scan"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "dirictory, d",
			Value: "",
			Usage: "Scan directory location",
		},
		cli.BoolFlag{
			Name:  "recursive, r",
			Usage: "Scan recursively",
		},
		cli.StringFlag{
			Name:  "config, cfg",
			Value: "config.yaml",
			Usage: "Config file location",
		},
		cli.StringFlag{
			Name:  "username, u",
			Value: "",
			Usage: "MySQL username",
		},
		cli.StringFlag{
			Name:  "password, p",
			Value: "",
			Usage: "MySQL user password",
		},
		cli.StringFlag{
			Name:  "database, db",
			Value: "",
			Usage: "MySQL database name",
		},
		cli.StringFlag{
			Name: "log",
			Value: "",
			Usage: "Set log file location.",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug mode",
			Destination: &DEBUG,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:            "scan",
			Aliases:         []string{"s"},
			Usage:           "Scan all files in the directory for the init time.",
			SkipFlagParsing: true,
			Action: func(c *cli.Context) error {
				return commandAction(c)
			},
		},
		{
			Name:    "check",
			Aliases: []string{"c"},
			Usage:   "Check files change",
			Action: func(c *cli.Context) error {
				isCheck = true

				return commandAction(c)
			},
		},
		{
			Name:    "renew",
			Aliases: []string{"re"},
			Usage:   "Renew all files MD5 and content",
			Action: func(c *cli.Context) error {
				isRenew = true

				return commandAction(c)
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Command.Name != "" {
			// read config
			cfg := c.GlobalString("cfg")
			if err := readConfig(cfg); err != nil {
				fmt.Println("Warning: Read config failed! you can use -cfg flag to set config location")
			}

			// create log
			log := c.GlobalString("log")
			createLogFile(log)

			// connect to db
			guser := c.GlobalString("u")
			gpass := c.GlobalString("p")
			gdbname := c.GlobalString("db")

			connectDb(guser, gpass, gdbname)

			// check dir is not empty
			dir := c.GlobalString("d")

			if dir == "" && len(scanDir) == 0 {
				return cli.NewExitError("", 0)
			}
		}


		return nil
	}

	defer func() {
		if db != nil {
			db.Close()
		}

		if findFileStmt != nil {
			findFileStmt.Close()
		}

		if insertFileStmt != nil {
			insertFileStmt.Close()
		}

		if updateFileStmt != nil {
			updateFileStmt.Close()
		}
	}()

	app.Run(os.Args)

}

func createLogFile (logPath string) {
	var homeDir string
	usr, err := user.Current()
	if err != nil {
		homeDir = os.TempDir()
	} else {
		homeDir = usr.HomeDir
	}


	if logPath == "" && conf.GetString("logPath") == "" && DEBUG {
		os.Mkdir(homeDir + "/FileChecker", 0755)
		l = logger.New(homeDir + "/FileChecker/debug.log")
	} else {
		if conf.GetString("logPath") != "" {
			l = logger.New(conf.GetString("logPath"))
		} else if logPath != "" {
			l = logger.New(logPath)
		} else {
			os.Mkdir(homeDir + "/FileChecker", 0755)

			l = logger.New(homeDir + "/FileChecker/FileChecker.log")
		}
	}
}

func readConfig(configPath string) error {
	conf = &config.Engine{}
	err := conf.Load(configPath)
	if err != nil {
		return err
	}

	scanDir = strings.Split(conf.GetString("scanDir"), ",")
	diffFileExtension = strings.Split(conf.GetString("diffExtension"), ",")
	excludeDir = strings.Split(conf.GetString("excludeDir"), ",")
	excludeFile = strings.Split(conf.GetString("excludeFile"), ",")

	return nil
}

func connectDb(globalUser, globalPass, globalDbname string) {
	var conErr error

	driver := conf.GetString("storeDriver")
	if driver == "" {
		driver = "mysql"
	}

	user := conf.GetString(driver + ".username")
	if user == "" {
		user = globalUser
	}

	pass := conf.GetString(driver + ".password")
	if pass == "" {
		pass = globalPass
	}

	dbname := conf.GetString(driver + ".database")
	if dbname == "" {
		dbname = globalDbname
	}

	if DEBUG {
		fmt.Printf("Ready to connect MySQL, username: %s, password: %s, dbname: %s", user, pass, dbname)
		l.Debug(fmt.Sprintf("Ready to connect MySQL, username: %s, password: %s, dbname: %s", user, pass, dbname))
	}

	db, conErr = sql.Open(driver, user+":"+pass+"@/"+dbname)

	if conErr != nil {

		panic(conErr.Error())
	}
}

func commandAction(c *cli.Context) error {
	dirFlag := c.GlobalString("d")
	if dirFlag != "" {
		dirFlags := strings.Split(dirFlag, ",")
		scanDir = append(scanDir, dirFlags...)
	}

	recursive := c.GlobalBool("r")

	if DEBUG {
		l.Debug(fmt.Sprintf("scanDir: %v, recursive: %v", scanDir, recursive))
	}

	for _, dir := range scanDir {
		scanFiles(dir, recursive)
	}

	return nil
}

func scanFiles(path string, recursive bool) {

	path, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		l.Error("Convert file absolute path error: " + path)

		return
	}
	path = path + "/"

	files, err := ioutil.ReadDir(path)
	if err != nil {
		l.Error(fmt.Sprintf("Can not read dir, path: %s, error: %v", path, err.Error()))

		return
	}

	if DEBUG {
		l.Debug(fmt.Sprintf("scanFiles in %s, have %d files, is recursive? %v", path, len(files), recursive))
	}

	for _, file := range files {
		if file.IsDir() {
			if !recursive || inSlice(file.Name(), excludeDir) {
				continue
			}

			scanFiles(path+file.Name(), recursive)

			continue
		}

		// skip exclude file
		if inSlice(file.Name(), excludeFile) {
			continue
		}

		fileMd5, content := getContentWithMD5(path + file.Name())
		// skip when get md5 failed
		if fileMd5 == "" && content == "" {
			continue
		}

		file_in_db := findFile(path + file.Name())
		// skip when db error
		if file_in_db == nil {
			continue
		}

		if DEBUG {
			l.Debug(fmt.Sprintf("Current file: %s, MD5: %s, content: %s, DB data: %v", path+file.Name(), fileMd5, content, file_in_db))
		}

		if file_in_db["md5"] == "NULL" { // new file
			if isCheck {
				// 寄信通知有新增檔案
				body := fmt.Sprintf("New file found in %s, file name is %s, MD5: %s", path, file.Name(), fileMd5)
				if content != "" {
					body += "\ncontent: \n" + content
				}

				l.Danger(body)
				sendEmail(body)
			}

			// insert to db
			if insertFileStmt == nil {
				insert_sql := "INSERT INTO files (path, md5, content, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())"
				insertFileStmt, err = db.Prepare(insert_sql)
				if err != nil {
					//panic(err.Error())
					l.Error(fmt.Sprintf("Prepare insert statement error! error: %v", err.Error()))

					continue
				}
			}

			_, err = insertFileStmt.Exec(path+file.Name(), fileMd5, content)
			if err != nil {
				//panic(err.Error())
				l.Error(fmt.Sprintf("Insert error! path: %s, \nmd5: %s, \ncontent: %s, \nerror: %s", path+file.Name(), fileMd5, content, err))
			}
		} else {
			if file_in_db["md5"] == fileMd5 {
				continue
			}

			if isRenew {
				// update md5
				if updateFileStmt == nil {
					var err error
					update_sql := "UPDATE files SET md5 = ? where path = ?"
					updateFileStmt, err = db.Prepare(update_sql)
					if err != nil {
						//panic(err.Error())
						l.Error(fmt.Sprintf("Prepare update statement error! %v", err.Error()))

						continue
					}
				}

				_, err = updateFileStmt.Exec(fileMd5, file_in_db["path"])
				if err != nil {
					//panic(err.Error())
					l.Error(fmt.Sprintf("Update error! path: %s, md5: %s, error: %s", file_in_db["path"], fileMd5, err))
				}

				continue
			}

			if isCheck {
				// alert
				body := fmt.Sprintf("Alert! path: %s, old md5: %s, new md5: %s\n", path+file.Name(), file_in_db["md5"], fileMd5)

				if file_in_db["content"] != "" {
					l.Danger(body + fmt.Sprintf("\ndiff: \n", checkDiffText(file_in_db["content"], content)))

					body += fmt.Sprintf("\ndiff: \n<br>", checkDiffHTML(file_in_db["content"], content))
				}

				sendEmail(body)
			}

		}
	}
}

func inSlice(search string, slice []string) bool {
	for _, value := range slice {
		if strings.TrimSpace(value) == search {
			return true
		}
	}

	return false
}

func getContentWithMD5(path string) (md5, content string) {
	//path, err := filepath.Abs(path)
	//if err != nil {
	//	panic("Convert file absolute path error: " + path)
	//}

	f, err := os.Open(path)
	if err != nil {
		//log.Fatal(err)
		l.Error(fmt.Sprintf("Open file error! Path: %s, Error: %s", path, err.Error()))

		return "", ""
	}
	defer f.Close()

	var re = regexp.MustCompile("^.*(" + strings.Join(diffFileExtension, "|") + ")$")

	if re.MatchString(path) {
		content = getContent(f)
	} else {
		content = ""
	}

	f.Seek(0, 0)
	md5 = genMd5(f)

	return md5, content
}

func genMd5(file *os.File) string {
	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		//log.Fatal(err)
		l.Error(fmt.Sprintf("%s genMD5 error! error: %s", file.Name(), err.Error()))

		return ""
	}

	return hex.EncodeToString(h.Sum(nil))
}

func getContent(fi *os.File) string {
	contentB, err := ioutil.ReadAll(fi)
	if err != nil {
		//panic(err.Error())
		l.Error(fmt.Sprintf("%s getContent Error! %s", fi.Name(), err.Error()))

		return ""
	}

	return string(contentB)
}

func findFile(path string) map[string]string {
	// select md5 from db
	find_file_sql := "SELECT * FROM files WHERE path = ?"
	if findFileStmt == nil {
		var err error
		findFileStmt, err = db.Prepare(find_file_sql)
		if err != nil {
			//panic(err.Error()) // proper error handling instead of panic in your app
			l.Error(fmt.Sprintf("Prepare findFile sql Error! %s", err.Error()))

			return nil
		}
	}

	rows, err := findFileStmt.Query(path)
	defer rows.Close()
	if err != nil {
		//panic(err)
		l.Error(fmt.Sprintf("findFile Query error! %s", err.Error()))

		return nil
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		l.Error(fmt.Sprintf("findFile get Columns error! %s", err.Error()))

		return nil
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	rows.Next()
	if err = rows.Err(); err != nil {
		l.Error(fmt.Sprintf("findFile rows.Next() error! %s", err.Error()))

		return nil
	}

	// get RawBytes from data
	rows.Scan(scanArgs...)

	// Now do something with the data.
	// Here we just print each column as a string.
	var value string
	res := make(map[string]string)
	for i, col := range values {
		// Here we can check if the value is nil (NULL value)
		if col == nil {
			value = "NULL"
		} else {
			value = string(col)
		}
		res[columns[i]] = value
	}

	return res
}

func checkDiffText(text1, text2 string) string {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(text1, text2, false)

	return dmp.DiffPrettyText(diffs)
}

func checkDiffHTML(text1, text2 string) string {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(text1, text2, false)

	return dmp.DiffPrettyHtml(diffs)
}

func sendEmail(body string) error {
	host := conf.GetString("notification.smtp")
	port := conf.GetString("notification.port")
	account := conf.GetString("notification.account")
	pass := conf.GetString("notification.pass")
	from := conf.GetString("notification.from")
	to := conf.GetString("notification.to")

	title := "Alert! file changed found!"

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	msg := "Subject: " + title + "\n" +
		mime + "\n<html><body>" +
		body + "</body></html>"

	err := smtp.SendMail(host+":"+port,
		smtp.PlainAuth(account, from, pass, host),
		from, []string{to}, []byte(msg))

	if err != nil {
		l.Error(fmt.Sprintf("smtp error: %s", err))

		return err
	}

	if DEBUG {
		l.Debug(fmt.Sprintf("Email send successful."))
	}

	return nil
}
