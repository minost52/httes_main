package report

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"httes/core/types"
)

// verboseRequest представляет подробную информацию об HTTP-запросе.
type verboseRequest struct {
	Url     string            `json:"url"`     // URL запроса
	Method  string            `json:"method"`  // HTTP-метод (GET, POST и т.д.)
	Headers map[string]string `json:"headers"` // Заголовки запроса
	Body    interface{}       `json:"body"`    // Тело запроса
}

// verboseResponse представляет подробную информацию об HTTP-ответе.
type verboseResponse struct {
	StatusCode int               `json:"statusCode"` // Код статуса ответа (200, 404 и т.д.)
	Headers    map[string]string `json:"headers"`    // Заголовки ответа
	Body       interface{}       `json:"body"`       // Тело ответа
}

// verboseHttpRequestInfo объединяет информацию о запросе, ответе и состоянии окружения.
type verboseHttpRequestInfo struct {
	StepId         uint16                 `json:"stepId"`         // Идентификатор шага сценария
	StepName       string                 `json:"stepName"`       // Имя шага сценария
	Request        verboseRequest         `json:"request"`        // Информация о запросе
	Response       verboseResponse        `json:"response"`       // Информация об ответе
	Envs           map[string]interface{} `json:"envs"`           // Используемые переменные окружения
	FailedCaptures map[string]string      `json:"failedCaptures"` // Переменные окружения, которые не удалось захватить
	Error          string                 `json:"error"`          // Ошибка, если шаг не выполнен
}

// Преобразует результат шага сценария (ScenarioStepResult) в структуру verboseHttpRequestInfo.
// Используется для получения подробной информации о запросах, ответах и ошибках.
func ScenarioStepResultToVerboseHttpRequestInfo(sr *types.ScenarioStepResult) verboseHttpRequestInfo {
	var verboseInfo verboseHttpRequestInfo

	verboseInfo.StepId = sr.StepID     // Устанавливаем ID шага
	verboseInfo.StepName = sr.StepName // Устанавливаем имя шага

	if sr.Err.Type == types.ErrorInvalidRequest {
		// Если запрос не удалось подготовить, записываем ошибку
		verboseInfo.Error = sr.Err.Error()
		return verboseInfo
	}

	// Декодируем заголовки и тело запроса
	requestHeaders, requestBody, _ := decode(sr.DebugInfo["requestHeaders"].(http.Header),
		sr.DebugInfo["requestBody"].([]byte))
	verboseInfo.Request = verboseRequest{
		Url:     sr.DebugInfo["url"].(string),
		Method:  sr.DebugInfo["method"].(string),
		Headers: requestHeaders,
		Body:    requestBody,
	}

	if sr.Err.Type != "" {
		// Если произошла ошибка, записываем её
		verboseInfo.Error = sr.Err.Error()
	} else {
		// Декодируем заголовки и тело ответа
		responseHeaders, responseBody, _ := decode(sr.DebugInfo["responseHeaders"].(http.Header),
			sr.DebugInfo["responseBody"].([]byte))
		verboseInfo.Response = verboseResponse{
			StatusCode: sr.StatusCode,
			Headers:    responseHeaders,
			Body:       responseBody,
		}
	}

	// Сохраняем переменные окружения и неудачные захваты
	verboseInfo.Envs = sr.UsableEnvs
	verboseInfo.FailedCaptures = sr.FailedCaptures

	return verboseInfo
}

// Декодирует HTTP-заголовки и тело запроса/ответа в человекочитаемый формат.
func decode(headers http.Header, byteBody []byte) (map[string]string, interface{}, error) {
	contentType := headers.Get("Content-Type") // Получаем Content-Type
	var reqBody interface{}

	hs := make(map[string]string, 0)
	for k, v := range headers {
		values := strings.Join(v, ",") // Преобразуем список значений в строку
		hs[k] = values
	}

	if strings.Contains(contentType, "text/html") {
		// Если это HTML, преобразуем в текст с экранированными символами
		unescapedHtml := html.UnescapeString(string(byteBody))
		reqBody = unescapedHtml
	} else if strings.Contains(contentType, "application/json") {
		// Если это JSON, десериализуем его
		err := json.Unmarshal(byteBody, &reqBody)
		if err != nil {
			reqBody = string(byteBody) // Если ошибка, возвращаем сырой текст
		}
	} else {
		// Для остальных типов возвращаем сырой текст
		reqBody = string(byteBody)
	}

	return hs, reqBody, nil
}

// Проверяет, является ли структура verboseRequest пустой.
func isVerboseInfoRequestEmpty(req verboseRequest) bool {
	if req.Url == "" && req.Method == "" && req.Headers == nil && req.Body == nil {
		return true
	}
	return false
}
