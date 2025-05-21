package types

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"httes/core/util"

	validator "github.com/asaskevich/govalidator"
)

// Константы для типов протоколов
const (
	// Протокол HTTP
	ProtocolHTTP = "HTTP"
	// Протокол HTTPS
	ProtocolHTTPS = "HTTPS"

	// Тип аутентификации HTTP Basic
	AuthHttpBasic = "basic"

	// Максимальная задержка (90 секунд) в миллисекундах
	maxSleep = 90000

	// Регулярное выражение для проверки переменных окружения
	// Оно должно соответствовать формату {{varName}}, но игнорировать переменные, начинающиеся с "_"
	EnvironmentVariableRegexStr = `\{{[^_]\w+\}}`
)

// Поддерживаемые протоколы, которые нужно обновлять при добавлении нового интерфейса requester.Requester
var SupportedProtocols = [...]string{ProtocolHTTP, ProtocolHTTPS}

// Методы HTTP, поддерживаемые приложением
var supportedProtocolMethods = []string{
	http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
	http.MethodPatch, http.MethodHead, http.MethodOptions,
}

// Поддерживаемые типы аутентификации
var supportedAuthentications = []string{
	AuthHttpBasic,
}

// Регулярное выражение для проверки переменных окружения, компилируемое при инициализации
var envVarRegexp *regexp.Regexp

// Инициализация регулярного выражения для переменных окружения
func init() {
	envVarRegexp = regexp.MustCompile(EnvironmentVariableRegexStr)
}

// Scenario описывает сценарий, состоящий из шагов и окружения
type Scenario struct {
	// Шаги сценария
	Steps []ScenarioStep
	// Глобальные переменные окружения, доступные для всех шагов
	Envs map[string]interface{}
}

// validate проверяет уникальность ID шагов и валидность использования переменных окружения
func (s *Scenario) validate() error {
	// Хранение уникальных ID шагов
	stepIds := make(map[uint16]struct{}, len(s.Steps))
	// Переменные окружения, определённые глобально или захваченные на предыдущих шагах
	definedEnvs := map[string]struct{}{}

	// Добавляем глобальные переменные окружения
	for key := range s.Envs {
		definedEnvs[key] = struct{}{}
	}

	// Проходим по всем шагам сценария
	for _, st := range s.Steps {
		// Валидация шага
		if err := st.validate(definedEnvs); err != nil {
			return err
		}

		// Добавляем переменные, захваченные из текущего шага
		for _, ce := range st.EnvsToCapture {
			definedEnvs[ce.Name] = struct{}{}
		}
		// Проверяем уникальность ID шага
		if _, ok := stepIds[st.ID]; ok {
			return fmt.Errorf("duplicate step id: %d", st.ID)
		}
		stepIds[st.ID] = struct{}{}
	}
	return nil
}

// checkEnvsValidInStep проверяет использование переменных окружения в шаге
func checkEnvsValidInStep(st *ScenarioStep, definedEnvs map[string]struct{}) error {
	var err error

	// Вспомогательная функция для проверки наличия переменных в окружении
	matchInEnvs := func(matches []string) error {
		for _, v := range matches {
			// Проверяем, существует ли переменная в окружении
			if _, ok := definedEnvs[v[2:len(v)-2]]; !ok { // {{...}}
				return EnvironmentNotDefinedError{
					msg: fmt.Sprintf("%s is not defined to use by global and captured environments", v),
				}
			}
		}
		return nil
	}

	// Функция для поиска и проверки переменных окружения в заданной строке
	f := func(source string) error {
		matches := envVarRegexp.FindAllString(source, -1)
		return matchInEnvs(matches)
	}

	// Проверка переменных окружения в URL
	err = f(st.URL)
	if err != nil {
		return err
	}

	// Проверка переменных окружения в заголовках
	for k, v := range st.Headers {
		err = f(k)
		if err != nil {
			return err
		}

		err = f(v)
		if err != nil {
			return err
		}
	}

	// Проверка переменных окружения в полезной нагрузке
	err = f(st.Payload)
	return err
}

// ScenarioStep представляет один шаг сценария.
// Эта структура должна включать все необходимые данные в сетевом пакете для поддерживаемых протоколов.
type ScenarioStep struct {
	// ID элемента. Должен быть предоставлен клиентом.
	ID uint16

	// Имя элемента.
	Name string

	// Метод запроса.
	Method string

	// Аутентификация.
	Auth Auth

	// TLS-сертификат.
	Cert tls.Certificate

	// Пул TLS-сертификатов.
	CertPool *x509.CertPool

	// Заголовки запроса.
	Headers map[string]string

	// Тело запроса.
	Payload string

	// Целевой URL.
	URL string

	// Длительность таймаута соединения для запроса в секундах.
	Timeout int

	// Длительность ожидания после выполнения шага. Может быть указана как диапазон, например, "300-500", или точное значение, например, "350" в мс.
	Sleep string

	// Параметры запроса, специфичные для протокола. Например: DisableRedirects:true для HTTP-запросов.
	Custom map[string]interface{}

	// Переменные окружения, которые нужно извлечь из ответа на этот шаг.
	EnvsToCapture []EnvCaptureConf
}

type SourceType string

const (
	Header SourceType = "header"
	Body   SourceType = "body"
)

type RegexCaptureConf struct {
	Exp *string `json:"exp"`
	No  int     `json:"matchNo"`
}

type EnvCaptureConf struct {
	JsonPath *string           `json:"jsonPath"`
	Xpath    *string           `json:"xpath"`
	RegExp   *RegexCaptureConf `json:"regExp"`
	Name     string            `json:"as"`
	From     SourceType        `json:"from"`
	Key      *string           `json:"headerKey"` // Ключ заголовка
}

// Auth должна включать все необходимые данные для аутентификации для поддерживаемых типов аутентификации.
type Auth struct {
	Type     string
	Username string
	Password string
}

func (si *ScenarioStep) validate(definedEnvs map[string]struct{}) error {
	if !util.StringInSlice(si.Method, supportedProtocolMethods) {
		return fmt.Errorf("неподдерживаемый метод запроса: %s", si.Method)
	}
	if si.Auth != (Auth{}) && !util.StringInSlice(si.Auth.Type, supportedAuthentications) {
		return fmt.Errorf("неподдерживаемый метод аутентификации (%s)", si.Auth.Type)
	}
	if si.ID == 0 {
		return fmt.Errorf("ID шага должен быть больше нуля")
	}
	if !envVarRegexp.MatchString(si.URL) && !validator.IsURL(strings.ReplaceAll(si.URL, " ", "_")) {
		return fmt.Errorf("цель недействительна: %s", si.URL)
	}
	if si.Sleep != "" {
		sleep := strings.Split(si.Sleep, "-")

		// Избегайте некорректного синтаксиса, например, "-300-500"
		if len(sleep) > 2 {
			return fmt.Errorf("выражение ожидания недействительно: %s", si.Sleep)
		}

		// Проверка преобразования строки в число
		for _, s := range sleep {
			dur, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("время ожидания недействительно: %s", si.Sleep)
			}

			if dur > maxSleep {
				return fmt.Errorf("превышен максимальный предел ожидания. указано: %d мс, максимум: %d мс", dur, maxSleep)
			}
		}
	}

	for _, conf := range si.EnvsToCapture {
		err := validateCaptureConf(conf)
		if err != nil {
			return wrapAsScenarioValidationError(err)
		}
	}

	// Проверьте, были ли уже определены переменные окружения, на которые ссылается текущий шаг
	if err := checkEnvsValidInStep(si, definedEnvs); err != nil {
		return wrapAsScenarioValidationError(err)
	}

	return nil
}

func wrapAsScenarioValidationError(err error) ScenarioValidationError {
	return ScenarioValidationError{
		msg:        fmt.Sprintf("Ошибка проверки сценария: %v", err),
		wrappedErr: err,
	}
}

func validateCaptureConf(conf EnvCaptureConf) error {
	if !(conf.From == Header || conf.From == Body) {
		return CaptureConfigError{
			msg: fmt.Sprintf("некорректный тип \"from\" в настройках извлечения: %s", conf.From),
		}
	}

	if conf.From == Header && conf.Key == nil {
		return CaptureConfigError{
			msg: fmt.Sprintf("%s, необходимо указать ключ заголовка", conf.Name),
		}
	}

	if conf.From == Body && conf.JsonPath == nil && conf.RegExp == nil && conf.Xpath == nil {
		return CaptureConfigError{
			msg: fmt.Sprintf("%s, необходимо указать один из jsonPath, regExp или xPath для извлечения из тела", conf.Name),
		}
	}

	return nil
}

func ParseTLS(certFile, keyFile string) (tls.Certificate, *x509.CertPool, error) {
	if certFile == "" || keyFile == "" {
		return tls.Certificate{}, nil, nil
	}

	// Чтение пары ключей для создания сертификата
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	// Создание пула сертификатов ЦС и добавление cert.pem в него
	caCert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	return cert, pool, nil
}

func IsTargetValid(url string) error {
	if !envVarRegexp.MatchString(url) && !validator.IsURL(strings.ReplaceAll(url, " ", "_")) {
		return fmt.Errorf("цель недействительна: %s", url)
	}
	return nil
}
