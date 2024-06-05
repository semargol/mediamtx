package api

import "fmt"

type ErrorDescription struct {
	English string
	Russian string
}

func getErrorDescription(errorCode int, str string, opts ...interface{}) string {
	errors := map[int]ErrorDescription{
		100: {
			English: "Cannot extract ID",
			Russian: "Невозможно получить ID",
		},
		101: {
			English: "Сan't find pipe with specified ID",
			Russian: "Невозможно найти pipe с задданым ID",
		},
		102: {
			English: "Pipe with specified ID already exists",
			Russian: "Pipe с заданным ID уже существует",
		},
		103: {
			English: "Can't delete pipe with specified ID",
			Russian: "Невозможно удалить pipe с заданным ID",
		},
		104: {
			English: fmt.Sprintf("Field '%s' not found in pipe config", str),
			Russian: fmt.Sprintf("Поле '%s' не найдено в конфигурации pipe", str),
		},
		105: {
			English: fmt.Sprintf("Unsupported type for field '%s' ", str),
			Russian: fmt.Sprintf("Неподдерживаемый тип для поля '%s' ", str),
		},
	}

	description, exists := errors[errorCode]
	if !exists {
		description = ErrorDescription{
			English: "Unknown Error",
			Russian: "Неизвестная ошибка",
		}
	}

	language := "English"
	print := false

	for _, opt := range opts {
		switch v := opt.(type) {
		case string:
			language = v
		case bool:
			print = v
		}
	}

	var message string
	switch language {
	case "Russian":
		message = description.Russian
	default:
		message = description.English
	}

	if print {
		fmt.Printf("Error Code: %d, Description: %s\n", errorCode, message)
	}
	return message
}
