package types

import (
	"fmt"
	"net/http"

	"httes/core/proxy"
	"httes/core/util"
)

// Константы, используемые для определения значений полей Heart.
const (
	// Типы нагрузки (Load Types) определяют, как будут распределяться запросы.
	LoadTypeLinear      = "linear"      // Линейная нагрузка (равномерное распределение запросов).
	LoadTypeIncremental = "incremental" // Инкрементная нагрузка (увеличение интенсивности с течением времени).
	LoadTypeWaved       = "waved"       // Волнообразная нагрузка (периодическое увеличение и снижение интенсивности).

	// Значения по умолчанию для различных параметров Heart.
	DefaultIterCount  = 100            // Общее количество итераций по умолчанию.
	DefaultLoadType   = LoadTypeLinear // Тип нагрузки по умолчанию.
	DefaultDuration   = 10             // Продолжительность теста в секундах по умолчанию.
	DefaultTimeout    = 5              // Таймаут (в секундах) для каждого запроса по умолчанию.
	DefaultMethod     = http.MethodGet // HTTP-метод по умолчанию (GET).
	DefaultOutputType = "stdout"       // Формат вывода по умолчанию.
)

// Список всех поддерживаемых типов нагрузки. Используется для проверки корректности входных данных.
var loadTypes = [...]string{LoadTypeLinear, LoadTypeIncremental, LoadTypeWaved}

// TimeRunCount представляет структуру данных для ручной настройки нагрузки.
// Она описывает длительность (в секундах) и количество запросов за это время.
type TimeRunCount []struct {
	Duration int // Длительность в секундах.
	Count    int // Количество запросов.
}

// Heart — основной объект, описывающий метаданные нагрузки и параметры атаки.
// Используется для конфигурации и инициализации движка нагрузки.
type Heart struct {
	IterationCount    int                    // Общее количество итераций для выполнения.
	LoadType          string                 // Тип нагрузки, например, "linear", "incremental" или "waved".
	TestDuration      int                    // Общая продолжительность теста в секундах.
	TimeRunCountMap   TimeRunCount           // Карта, отображающая количество запросов за определённые промежутки времени.
	Scenario          Scenario               // Тестовый сценарий, содержащий шаги выполнения нагрузки.
	Proxy             proxy.Proxy            // Прокси-серверы, которые будут использоваться для выполнения запросов.
	ReportDestination string                 // Место назначения для записи данных о результатах теста.
	Others            map[string]interface{} // Динамическое поле для дополнительных параметров, которые могут быть добавлены пользователем.
	Debug             bool                   // Флаг для включения/выключения режима отладки.
}

// Validate проверяет корректность конфигурации Heart.
// Метод выполняет базовую валидацию всех ключевых полей и вызывает проверки зависимых служб.
func (h *Heart) Validate() error {
	// Проверка, что сценарий содержит хотя бы один шаг.
	if len(h.Scenario.Steps) == 0 {
		return fmt.Errorf("scenario or target is empty") // Ошибка, если сценарий или цель отсутствуют.
	} else if err := h.Scenario.validate(); err != nil {
		return err // Возврат ошибки, если валидация сценария не удалась.
	}

	// Проверка, что указанный тип нагрузки поддерживается.
	if h.LoadType != "" && !util.StringInSlice(h.LoadType, loadTypes[:]) {
		return fmt.Errorf("unsupported LoadType: %s", h.LoadType) // Ошибка, если LoadType некорректен.
	}

	// Проверка валидности значений в TimeRunCountMap.
	if len(h.TimeRunCountMap) > 0 {
		for _, t := range h.TimeRunCountMap {
			// Убедимся, что длительность больше 0.
			if t.Duration < 1 {
				return fmt.Errorf("duration in manual_load should be greater than 0")
			}
		}
	}

	// Если все проверки пройдены успешно, возвращаем nil (ошибок нет).
	return nil
}
