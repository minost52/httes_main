package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"httes/core/proxy"
	"httes/core/types"
)

// Константа для обозначения типа конфигурации "jsonReader".
const ConfigTypeJson = "jsonReader"

// Функция init выполняется при инициализации пакета.
// Регистрирует реализацию JsonReader в карте AvailableConfigReader.
func init() {
	AvailableConfigReader[ConfigTypeJson] = &JsonReader{}
}

// Структура timeRunCount описывает нагрузку вручную, с указанием длительности (в мс) и количества запросов за этот промежуток времени.
type timeRunCount []struct {
	Duration int `json:"duration"`
	Count    int `json:"count"`
}

// Структура auth описывает параметры аутентификации, такие как тип, имя пользователя и пароль.
type auth struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Структура multipartFormData описывает данные для multipart-запросов.
// Поля:
// - Name: имя поля.
// - Value: значение поля (если это текст).
// - Type: тип данных (например, текст или файл).
// - Src: путь к файлу, если это файл.
type multipartFormData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
	Src   string `json:"src"`
}

// Структура RegexCaptureConf описывает конфигурацию захвата данных с использованием регулярных выражений.
// Поля:
// - Exp: строка регулярного выражения.
// - No: номер совпадения, которое нужно извлечь.
type RegexCaptureConf struct {
	Exp *string `json:"exp"`
	No  int     `json:"matchNo"`
}

// Структура capturePath описывает, как захватывать данные из ответа.
// Поля:
// - JsonPath, XPath, RegExp: способы извлечения данных (JSONPath, XPath или регулярное выражение).
// - From: источник данных (например, тело ответа или заголовок).
// - HeaderKey: ключ заголовка, если данные берутся из заголовков.
type capturePath struct {
	JsonPath  *string           `json:"jsonPath"`
	XPath     *string           `json:"xPath"`
	RegExp    *RegexCaptureConf `json:"regExp"`
	From      string            `json:"from"`
	HeaderKey *string           `json:"headerKey"`
}

// Структура step описывает один шаг сценария.
// Поля включают URL, метод запроса, заголовки, тело, а также параметры для аутентификации, времени ожидания и другие.
type step struct {
	Id               uint16                 `json:"id"`
	Name             string                 `json:"name"`
	Url              string                 `json:"url"`
	Auth             auth                   `json:"auth"`
	Method           string                 `json:"method"`
	Headers          map[string]string      `json:"headers"`
	Payload          string                 `json:"payload"`
	PayloadFile      string                 `json:"payload_file"`
	PayloadMultipart []multipartFormData    `json:"payload_multipart"`
	Timeout          int                    `json:"timeout"`
	Sleep            string                 `json:"sleep"`
	Others           map[string]interface{} `json:"others"`
	CertPath         string                 `json:"cert_path"`
	CertKeyPath      string                 `json:"cert_key_path"`
	CaptureEnv       map[string]capturePath `json:"captureEnv"`
}

// Метод UnmarshalJSON для структуры step.
// Используется для настройки значений по умолчанию при десериализации JSON.
func (s *step) UnmarshalJSON(data []byte) error {
	type stepAlias step // Создаем псевдоним для step, чтобы избежать рекурсии.
	defaultFields := &stepAlias{
		Method:  types.DefaultMethod,  // Метод по умолчанию.
		Timeout: types.DefaultTimeout, // Тайм-аут по умолчанию.
	}

	// Парсим данные JSON в структуру с полями по умолчанию.
	err := json.Unmarshal(data, defaultFields)
	if err != nil {
		return err
	}

	// Присваиваем значения оригинальной структуре.
	*s = step(*defaultFields)
	return nil
}

// Структура JsonReader описывает читатель конфигураций в формате JSON.
// Включает поля для параметров нагрузки, шагов сценария, прокси, окружения и других.
type JsonReader struct {
	ReqCount     *int                   `json:"request_count"`
	IterCount    *int                   `json:"iteration_count"`
	LoadType     string                 `json:"load_type"`
	Duration     int                    `json:"duration"`
	TimeRunCount timeRunCount           `json:"manual_load"`
	Steps        []step                 `json:"steps"`
	Output       string                 `json:"output"`
	Proxy        string                 `json:"proxy"`
	Envs         map[string]interface{} `json:"env"`
	Debug        bool                   `json:"debug"`
}

// Метод UnmarshalJSON для JsonReader.
// Настраивает значения по умолчанию для LoadType, Duration и Output.
func (j *JsonReader) UnmarshalJSON(data []byte) error {
	type jsonReaderAlias JsonReader
	defaultFields := &jsonReaderAlias{
		LoadType: types.DefaultLoadType,   // Тип нагрузки по умолчанию.
		Duration: types.DefaultDuration,   // Длительность по умолчанию.
		Output:   types.DefaultOutputType, // Тип вывода по умолчанию.
	}

	// Десериализуем JSON с настройкой полей по умолчанию.
	err := json.Unmarshal(data, defaultFields)
	if err != nil {
		return err
	}

	// Присваиваем значения оригинальной структуре.
	*j = JsonReader(*defaultFields)
	return nil
}

// Метод Init для JsonReader.
// Проверяет, что переданный JSON валиден, и десериализует его в структуру.
func (j *JsonReader) Init(jsonByte []byte) (err error) {
	// Проверяем валидность JSON.
	if !json.Valid(jsonByte) {
		err = fmt.Errorf("provided json is invalid")
		return
	}

	// Десериализуем JSON в объект JsonReader.
	err = json.Unmarshal(jsonByte, &j)
	return
}

func (j *JsonReader) CreateHammer() (h types.Heart, err error) {
	// Создание сценария на основе шагов и переменных окружения.
	s := types.Scenario{
		Envs: j.Envs, // Переменные окружения для сценария.
	}
	var si types.ScenarioStep
	for _, step := range j.Steps {
		// Преобразование каждого шага в тип ScenarioStep.
		si, err = stepToScenarioStep(step)
		if err != nil {
			return
		}
		// Добавление шага в сценарий.
		s.Steps = append(s.Steps, si)
	}

	// Создание конфигурации прокси, если она указана.
	var proxyURL *url.URL
	if j.Proxy != "" {
		proxyURL, err = url.Parse(j.Proxy)
		if err != nil {
			return
		}
	}
	p := proxy.Proxy{
		Strategy: proxy.ProxyTypeSingle, // Используется одна прокси-стратегия.
		Addr:     proxyURL,
	}

	// Обратная совместимость: установка количества итераций.
	var iterationCount int
	if j.IterCount != nil {
		iterationCount = *j.IterCount
	} else if j.ReqCount != nil {
		iterationCount = *j.ReqCount
	} else {
		iterationCount = types.DefaultIterCount // Значение по умолчанию.
	}
	j.IterCount = &iterationCount

	// Обновление параметров на основе TimeRunCount, если он задан.
	if len(j.TimeRunCount) > 0 {
		*j.IterCount, j.Duration = 0, 0
		for _, t := range j.TimeRunCount {
			*j.IterCount += t.Count
			j.Duration += t.Duration
		}
	}

	// Создание объекта Hammer, который содержит конфигурацию нагрузки.
	h = types.Heart{
		IterationCount:    *j.IterCount,
		LoadType:          strings.ToLower(j.LoadType),
		TestDuration:      j.Duration,
		TimeRunCountMap:   types.TimeRunCount(j.TimeRunCount),
		Scenario:          s,
		Proxy:             p,
		ReportDestination: j.Output,
		Debug:             j.Debug,
	}
	return
}

func stepToScenarioStep(s step) (types.ScenarioStep, error) {
	var payload string
	var err error

	// Подготовка payload для multipart-запросов.
	if len(s.PayloadMultipart) > 0 {
		if s.Headers == nil {
			s.Headers = make(map[string]string)
		}

		payload, s.Headers["Content-Type"], err = prepareMultipartPayload(s.PayloadMultipart)
		if err != nil {
			return types.ScenarioStep{}, err
		}
	} else if s.PayloadFile != "" { // Если указан файл для payload.
		buf, err := ioutil.ReadFile(s.PayloadFile)
		if err != nil {
			return types.ScenarioStep{}, err
		}
		payload = string(buf)
	} else { // Если указан обычный payload.
		payload = s.Payload
	}

	// Установка типа аутентификации по умолчанию.
	if s.Auth != (auth{}) && s.Auth.Type == "" {
		s.Auth.Type = types.AuthHttpBasic
	}

	// Проверка валидности URL.
	err = types.IsTargetValid(s.Url)
	if err != nil {
		return types.ScenarioStep{}, err
	}

	// Настройка захвата данных из ответа.
	var capturedEnvs []types.EnvCaptureConf
	for name, path := range s.CaptureEnv {
		capConf := types.EnvCaptureConf{
			JsonPath: path.JsonPath,
			Xpath:    path.XPath,
			Name:     name,
			From:     types.SourceType(path.From),
			Key:      path.HeaderKey,
		}

		if path.RegExp != nil {
			capConf.RegExp = &types.RegexCaptureConf{
				Exp: path.RegExp.Exp,
				No:  path.RegExp.No,
			}
		}

		capturedEnvs = append(capturedEnvs, capConf)
	}

	// Создание объекта ScenarioStep.
	item := types.ScenarioStep{
		ID:            s.Id,
		Name:          s.Name,
		URL:           s.Url,
		Auth:          types.Auth(s.Auth),
		Method:        strings.ToUpper(s.Method),
		Headers:       s.Headers,
		Payload:       payload,
		Timeout:       s.Timeout,
		Sleep:         strings.ReplaceAll(s.Sleep, " ", ""),
		Custom:        s.Others,
		EnvsToCapture: capturedEnvs,
	}

	// Настройка TLS-сертификатов.
	if s.CertPath != "" && s.CertKeyPath != "" {
		cert, pool, err := types.ParseTLS(s.CertPath, s.CertKeyPath)
		if err != nil {
			return item, err
		}

		item.Cert = cert
		item.CertPool = pool
	}

	return item, nil
}

func prepareMultipartPayload(parts []multipartFormData) (body string, contentType string, err error) {
	byteBody := &bytes.Buffer{}
	writer := multipart.NewWriter(byteBody)

	// Обработка multipart-данных.
	for _, part := range parts {
		if strings.EqualFold(part.Type, "file") { // Если часть является файлом.
			if strings.EqualFold(part.Src, "remote") { // Если файл загружается по URL.
				response, err := http.Get(part.Value)
				if err != nil {
					return "", "", err
				}
				defer response.Body.Close()

				u, _ := url.Parse(part.Value)
				formPart, err := writer.CreateFormFile(part.Name, path.Base(u.Path))
				if err != nil {
					return "", "", err
				}

				_, err = io.Copy(formPart, response.Body)
				if err != nil {
					return "", "", err
				}
			} else { // Если файл находится локально.
				file, err := os.Open(part.Value)
				defer file.Close()
				if err != nil {
					return "", "", err
				}

				formPart, err := writer.CreateFormFile(part.Name, filepath.Base(file.Name()))
				if err != nil {
					return "", "", err
				}

				_, err = io.Copy(formPart, file)
				if err != nil {
					return "", "", err
				}
			}
		} else { // Если часть является текстовым полем.
			err = writer.WriteField(part.Name, part.Value)
			if err != nil {
				return "", "", err
			}
		}
	}

	writer.Close()
	return byteBody.String(), writer.FormDataContentType(), err
}
