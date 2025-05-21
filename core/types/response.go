package types

import (
	"net/url"
	"time"

	"github.com/google/uuid"
)

// ScenarioResult соответствует сценарию. Каждый сценарий имеет ScenarioResult после его выполнения.
type ScenarioResult struct {
	// Время начала первого запроса для сценария
	StartTime time.Time

	ProxyAddr   *url.URL
	StepResults []*ScenarioStepResult

	// Динамическое поле для дополнительных данных, необходимых потребителям объекта ответа.
	Others map[string]interface{}
}

// ScenarioStepResult соответствует шагу сценария.
type ScenarioStepResult struct {
	// ID шага сценария
	StepID uint16

	// Имя шага сценария
	StepName string

	// Каждый запрос имеет уникальный ID.
	RequestID uuid.UUID

	// Возвращенный статус-код. Имеет разное значение для разных протоколов.
	StatusCode int

	// Время выполнения запроса.
	RequestTime time.Time

	// Общая продолжительность. От отправки запроса до полного получения ответа.
	Duration time.Duration

	// Длина содержимого ответа
	ContentLength int64

	// Ошибка, возникшая во время выполнения запроса.
	Err RequestError

	// Подробная информация для отладки
	DebugInfo map[string]interface{}

	// Метрики, специфичные для протокола. Например: DNSLookupDuration: 1s для HTTP
	Custom map[string]interface{}

	// Используемые переменные окружения на этом шаге
	UsableEnvs map[string]interface{}

	// Захваченные переменные окружения на этом шаге
	ExtractedEnvs map[string]interface{}

	// Неудачные захваты и их причины
	FailedCaptures map[string]string
}
