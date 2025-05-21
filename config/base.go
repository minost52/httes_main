package config

import (
	"fmt"
	"reflect"

	"httes/core/types"
)

// AvailableConfigReader хранит доступные реализации интерфейса ConfigReader.
// Ключ - строка, представляющая тип конфигурации (например, "json").
// Значение - объект, который реализует интерфейс ConfigReader.
var AvailableConfigReader = make(map[string]ConfigReader)

// ConfigReader - интерфейс для абстрагирования различных реализаций чтения конфигурации.
// Он предоставляет методы для инициализации конфигурации и создания объекта Hammer.
type ConfigReader interface {
	// Init принимает срез байтов с данными конфигурации и инициализирует объект.
	Init([]byte) error

	// CreateHammer создает объект Hammer на основе данных конфигурации.
	CreateHammer() (types.Heart, error)
}

// NewConfigReader - фабричный метод для создания объекта ConfigReader.
// Принимает:
// - config: данные конфигурации в виде среза байтов.
// - configType: строка, представляющая тип конфигурации (например, "json").
// Возвращает:
// - reader: объект, реализующий интерфейс ConfigReader.
// - err: ошибка, если тип конфигурации не поддерживается или произошла ошибка инициализации.
func NewConfigReader(config []byte, configType string) (reader ConfigReader, err error) {
	// Проверяем, поддерживается ли указанный тип конфигурации.
	if val, ok := AvailableConfigReader[configType]; ok {
		// Создаем новый объект указанного типа с использованием рефлексии.
		reader = reflect.New(reflect.TypeOf(val).Elem()).Interface().(ConfigReader)

		// Инициализируем объект переданными данными конфигурации.
		err = reader.Init(config)
	} else {
		// Возвращаем ошибку, если тип конфигурации не найден в AvailableConfigReader.
		err = fmt.Errorf("unsupported config reader type: %s", configType)
	}
	return
}
