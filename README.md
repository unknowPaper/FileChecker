# FileChecker

This tool can help you check your system files stay there original checksum.

And send notification when some files have been changed by another one.

## Features

* Generate md5 for any files.
* Support recursive scan all sub folders.
* Store in MySQL.
* Notifycation by E-mail.
* Print diff string in notification.

## Install

## Usage

#### Step 1 - Edit config

You must set mysql username and password.

All config example:

```yaml
scanDir: /bin, /sbin
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
  database: filesmd5
```


#### Step 2 - Run scan

You must scan all files at first time.

```
fileChecker -r scan
```

-r is recursive

If you don't give -d flag, fileChecker will read scanDir in config.

or you can add another directory in the command.

```
fileChecker -d /usr/sbin scan
```
It will scan /usr/sbin and scanDir in the config.


#### Step 3 - Run check

You can add this command into your cron job.

```
fileChecker -r check
```

```
fileChecker -r -d /usr/sbin check
```

#### Step 4 - Renew

If you update your system, files md5 absolutely changed.

So you must run renew to update their md5 hash.

```
fileChecker -r renew
```