package utils

import (
	"os"
)

func IsDirExist(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func Joins(str1, str2 string) string {
	if str1 == "" {
		return str2
	} else if str2 == "" {
		return str1
	}

	if str1[len(str1)-1:] != "/" {
		str1 += "/"
	}

	return str1 + str2
}

func IsFileInDir(fileName, dirName string) bool {
	if !IsDirExist(dirName) {
		return false
	}

	entries, err := os.ReadDir(dirName)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.Name() == fileName {
			return true
		}
	}
	return false
}

func CreateEmptyFile(fileName, dirName string) {
	_, err := os.OpenFile(Joins(dirName, fileName), os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
}
