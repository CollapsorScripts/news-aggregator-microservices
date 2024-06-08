package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var mutex sync.Mutex
var fullPathToFile string

func writeToLog(str string, a ...any) error {
	mutex.Lock()
	// Открываем файл с логом
	file, err := os.OpenFile(fullPathToFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}

	defer func() {
		mutex.Unlock()
		file.Close()
	}()

	// Дозаписываем данные в файл
	data := []byte(fmt.Sprintf(str+"\n", a...))
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// New - инициализация журнала, создание папки и файла с логом
func New() error {
	//Текущая директория
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	//Файл с логом
	fileName := "journal.log"

	//Папка куда будет складывать лог
	folderName := "history"

	//Полный путь к папке
	dirPath := filepath.Join(pwd, folderName)

	//Полный путь к файлу с логом
	fullPathToFile = filepath.Join(dirPath, fileName)

	// Создаем папку с правами доступа для текущего пользователя
	err = os.Mkdir(dirPath, 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Создаем файл с логом
	file, err := os.OpenFile(fullPathToFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	mutex = sync.Mutex{}

	return nil
}

// Write - запись события в журнал
func Write(format string, a ...any) {
	str := fmt.Sprintf("%s", format)

	err := writeToLog(str, a...)
	if err != nil {
		fmt.Printf("Ошибка при записи в лог: %v\n", err)
		return
	}
}
