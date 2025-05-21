package extraction

import (
	"errors"
	"fmt"
	"net/http"

	"httes/core/types"
)

// Функция Extract извлекает данные из источника (например, HTTP-ответа)
// согласно параметрам конфигурации захвата окружения (EnvCaptureConf).
func Extract(source interface{}, ce types.EnvCaptureConf) (val interface{}, err error) {
	defer func() {
		// Обработка паники, если что-то пошло не так
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			val = nil
		}
	}()

	// Проверка на nil-источник
	if source == nil {
		return "", ExtractionError{
			msg: "source is nil",
		}
	}

	// Извлечение в зависимости от источника данных
	switch ce.From {
	case types.Header: // Если источник — HTTP-заголовки
		header := source.(http.Header)
		if ce.Key != nil {
			val = header.Get(*ce.Key)
			if val == "" {
				err = fmt.Errorf("http header %s not found", *ce.Key)
			} else if ce.RegExp != nil {
				// Применяем регулярное выражение к значению заголовка
				val, err = extractWithRegex(val, *ce.RegExp)
			}
		} else {
			err = fmt.Errorf("http header key not specified")
		}
	case types.Body: // Если источник — тело ответа
		if ce.JsonPath != nil {
			val, err = extractFromJson(source, *ce.JsonPath)
		} else if ce.RegExp != nil {
			val, err = extractWithRegex(source, *ce.RegExp)
		} else if ce.Xpath != nil {
			val, err = extractFromXml(source, *ce.Xpath)
		}
	}

	// Обработка ошибки извлечения
	if err != nil {
		return "", ExtractionError{
			msg:        fmt.Sprintf("%v", err),
			wrappedErr: err,
		}
	}
	return val, nil
}

// Извлечение с помощью регулярного выражения
func extractWithRegex(source interface{}, regexConf types.RegexCaptureConf) (val interface{}, err error) {
	re := regexExtractor{}
	re.Init(*regexConf.Exp)
	switch s := source.(type) {
	case []byte:
		return re.extractFromByteSlice(s, regexConf.No)
	case string:
		return re.extractFromString(s, regexConf.No)
	default:
		return "", fmt.Errorf("Unsupported type for extraction source")
	}
}

// Извлечение из JSON по пути (JsonPath)
func extractFromJson(source interface{}, jsonPath string) (interface{}, error) {
	je := jsonExtractor{}
	switch s := source.(type) {
	case []byte:
		return je.extractFromByteSlice(s, jsonPath)
	case string:
		return je.extractFromString(s, jsonPath)
	default:
		return "", fmt.Errorf("Unsupported type for extraction source")
	}
}

// Извлечение из XML по XPath
func extractFromXml(source interface{}, xPath string) (interface{}, error) {
	xe := xmlExtractor{}
	switch s := source.(type) {
	case []byte:
		return xe.extractFromByteSlice(s, xPath)
	default:
		return "", fmt.Errorf("Unsupported type for extraction source")
	}
}

// Кастомная ошибка для случаев неудачного извлечения
type ExtractionError struct {
	msg        string
	wrappedErr error
}

func (sc ExtractionError) Error() string {
	return sc.msg
}

func (sc ExtractionError) Unwrap() error {
	return sc.wrappedErr
}
