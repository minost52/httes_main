package requester

import (
	"context"
	"net/url"

	"httes/core/types"
)

// Отправитель запроса - это интерфейс, который абстрагирует реализации отправки запросов по различным протоколам.
// // Поле протокола в типах.Шаг сценария определяет, какую реализацию отправителя запроса использовать.
type Requester interface {
	Init(ctx context.Context, ss types.ScenarioStep, url *url.URL, debug bool) error
	Send(envs map[string]interface{}) *types.ScenarioStepResult
	Done()
}

// // Новый запросчик - это заводской метод отправителя запроса.
func NewRequester(s types.ScenarioStep) (requester Requester, err error) {
	requester = &HttpRequester{} // на данный момент у нас есть только тип HttpRequester.
	return
}
