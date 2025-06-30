package logger

import (
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type Log struct {
	packageName string
	logger      *log.Logger
	writerOut   io.Writer
	writerErr   io.Writer
	fileOutName string
	fileErrName string
	fileOut     io.Writer
	fileErr     io.Writer
	fileOutNum  int
	fileErrNum  int
	maxSize     int64
}

var logSingleton *Log = nil

// NewLog Create a new Log:
// - packageName should be the name of the go package using this log.
// - fileOut will be the log file used to write the log data, if empty it will not be used.
// - fileErr will be the log file used to write the log data of FATAL and ERROR log, if empty it will not be used.
// - maxSize will be the max bytes size used in fileOut and fileErr. Once the maxSize is reached a new file will be
// created and used.
// os.Stdout and os.Stderr are used no matter what.
func NewLog() *Log {
	if logSingleton != nil {
		return logSingleton
	}
	logSingleton = &Log{
		packageName: "",
		logger:      log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
		writerOut:   os.Stdout,
		writerErr:   os.Stderr,
		fileOutName: "",
		fileErrName: "",
		fileOut:     nil,
		fileErr:     nil,
		fileOutNum:  0,
		fileErrNum:  0,
		maxSize:     0,
	}
	return logSingleton
}

func (log *Log) SetFileOutName(fileOutName string) {
	var fileOut io.Writer

	if fileOutName != "" {
		log.fileOutNum = getFileNum(fileOutName)
		if fileOutName[len(fileOutName)-4:] != ".log" {
			fileOutName = fileOutName + "." + strconv.Itoa(log.fileOutNum) + ".log"
		} else {
			fileOutName = fileOutName[:len(fileOutName)-4] + "." + strconv.Itoa(log.fileOutNum) + ".log"
		}
		fOut, err := os.OpenFile(fileOutName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		fileOut = fOut
	}
	log.fileOutName = fileOutName
	log.fileOut = fileOut
}

func (log *Log) SetFileErrName(fileErrName string) {
	var fileErr io.Writer

	if fileErrName != "" {
		log.fileErrNum = getFileNum(fileErrName)
		if fileErrName[len(fileErrName)-4:] != ".log" {
			fileErrName = fileErrName + "." + strconv.Itoa(log.fileErrNum) + ".log"
		} else {
			fileErrName = fileErrName[:len(fileErrName)-4] + "." + strconv.Itoa(log.fileErrNum) + ".log"
		}
		fErr, err := os.OpenFile(fileErrName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		fileErr = fErr
	}

	log.fileErrName = fileErrName
	log.fileErr = fileErr
}

func (log *Log) SetMaxSize(maxSize int64) {
	log.maxSize = maxSize
}

func (log *Log) Info(data string, v ...any) {
	log.logger.SetOutput(log.writerOut)
	log.Print(data, "INFO", false, v...)

	if log.fileOut != nil {
		log.logger.SetOutput(log.fileOut)
		log.Print(data, "INFO", false, v...)
	}
}

func (log *Log) Debug(data string, v ...any) {
	log.logger.SetOutput(log.writerOut)
	log.Print(data, "DEBUG", false, v...)
	if log.fileOut != nil {
		log.logger.SetOutput(log.fileOut)
		log.Print(data, "DEBUG", false, v...)
	}
}

func (log *Log) Warn(data string, v ...any) {
	log.logger.SetOutput(log.writerOut)
	log.Print(data, "WARN", false, v...)
	if log.fileOut != nil {
		log.logger.SetOutput(log.fileOut)
		log.Print(data, "WARN", false, v...)
	}
}

func (log *Log) Error(data string, v ...any) {
	if log.fileOut != nil {
		log.logger.SetOutput(log.fileOut)
		log.Print(data, "ERROR", false, v...)
	}
	if log.fileErr != nil {
		log.logger.SetOutput(log.fileErr)
		log.Print(data, "ERROR", false, v...)
	}
	log.logger.SetOutput(log.writerErr)
	log.Print(data, "ERROR", true, v...)
}

func (log *Log) Fatal(data string, v ...any) {
	if log.fileOut != nil {
		log.logger.SetOutput(log.fileOut)
		log.Print(data, "FATAL", false, v...)
	}
	if log.fileErr != nil {
		log.logger.SetOutput(log.fileErr)
		log.Print(data, "FATAL", false, v...)
	}
	log.logger.SetOutput(log.writerErr)
	log.Print(data, "FATAL", true, v...)
}

func (log *Log) Print(data string, level string, exit bool, v ...any) {
	pc, file, line, ok := runtime.Caller(2)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	funcName := parts[pl-1]

	if !ok {
		file = "???"
		line = 0
	}

	short := file

	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
		file = short
	}

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		log.packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		log.packageName = strings.Join(parts[0:pl-1], ".")
	}

	log.logger.SetPrefix("[" + log.packageName + ":" + funcName + "] | " + level + " : " +
		short + ":" + strconv.Itoa(line) + ": ")

	log.checkLogFileSize()

	if level == "FATAL" && exit {
		log.logger.Fatalf(data, v...)
	} else if level == "ERROR" && exit {
		log.logger.Panicf(data, v...)
	} else {
		log.logger.Printf(data, v...)
	}
}

func (log *Log) checkLogFileSize() {
	fOutInfo, err := os.Stat(log.fileOutName)
	if err != nil {
		panic(err)
	}

	if fOutInfo.Size() >= log.maxSize {
		var fOut io.Writer

		log.fileOutName = log.fileOutName[:len(log.fileOutName)-getIndexLogFile(log.fileOutName)] +
			strconv.Itoa(getNumLogFile(log.fileOutName)+1) + ".log"
		fOut, err = os.Create(log.fileOutName)
		if err != nil {
			panic(err)
		}
		log.fileOut = fOut
	}

	fErrInfo, err := os.Stat(log.fileErrName)
	if err != nil {
		panic(err)
	}

	if fErrInfo.Size() >= log.maxSize {
		var fErr io.Writer

		log.fileErrName = log.fileErrName[:len(log.fileErrName)-getIndexLogFile(log.fileErrName)] +
			strconv.Itoa(getNumLogFile(log.fileErrName)+1) + ".log"
		fErr, err = os.Create(log.fileErrName)
		if err != nil {
			panic(err)
		}
		log.fileErr = fErr
	}
}

func getIndexLogFile(fileName string) int {
	counter := 0

	for i := len(fileName) - 1; i > 0; i-- {
		if fileName[i] == '.' {
			counter++
		}

		if counter == 2 {
			return len(fileName) - i - 1
		}
	}
	return 0
}

func getFileNum(filepath string) int {
	var dir string
	var filename string
	num := 0

	if strings.Contains(filepath, "/") {
		dir = filepath[0:strings.LastIndex(filepath, "/")]
		filename = filepath[strings.LastIndex(filepath, "/")+1:]
		if filename[len(filename)-4:] == ".log" {
			filename = filename[:len(filename)-4]
		}
	} else {
		dir = "."
		filename = filepath
		if filename[len(filename)-4:] == ".log" {
			filename = filename[:len(filename)-4]
		}
	}

	files, err := os.ReadDir(dir)

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		truncFileName := file.Name()[:getIndexLogFile(file.Name())-1]
		if !file.IsDir() && truncFileName == filename {
			if num < getNumLogFile(file.Name()) {
				num = getNumLogFile(file.Name())
			}
		}
	}

	return num
}

func getNumLogFile(fileName string) int {
	var counter int
	var num string
	checkToken := false

	for i := len(fileName) - 1; i > 0; i-- {
		if fileName[i] == '.' {
			counter++
		}

		if counter == 2 {
			checkToken = true
			counter++
			i++
		}

		if checkToken {
			for j := i; j < len(fileName)-1; j++ {
				if fileName[j] == '.' {
					ret, _ := strconv.Atoi(num)
					return ret
				}
				num += string(fileName[j])
			}
		}
	}
	return 0
}
