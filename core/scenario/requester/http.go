package requester

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"httes/core/scenario/scripting/extraction"
	"httes/core/scenario/scripting/injection"
	"httes/core/types"
	"httes/core/types/regex"

	"github.com/google/uuid"
	"golang.org/x/net/http2"
)

type HttpRequester struct {
	ctx             context.Context    // Контекст для управления запросами
	proxyAddr       *url.URL           // Адрес прокси-сервера
	packet          types.ScenarioStep // Шаг сценария с данными запроса
	client          *http.Client       // HTTP-клиент для отправки запросов
	requestSettings struct {           // Новая подструктура вместо *http.Request
		Header http.Header
		Host   string
		Close  bool
	}
	ei                   *injection.EnvironmentInjector // Инжектор переменных
	containsDynamicField map[string]bool                // Флаги наличия динамических переменных
	containsEnvVar       map[string]bool                // Флаги наличия окружных переменных
	debug                bool                           // Режим отладки
	dynamicRgx           *regexp.Regexp                 // Регулярка для динамических переменных
	envRgx               *regexp.Regexp                 // Регулярка для окружных переменных
}

// Init создаёт клиента для указанного шага сценария. HttpRequester использует один http.Client для всех запросов
func (h *HttpRequester) Init(ctx context.Context, s types.ScenarioStep, proxyAddr *url.URL, debug bool) (err error) {
	h.ctx = ctx
	h.packet = s
	h.proxyAddr = proxyAddr
	h.ei = &injection.EnvironmentInjector{}
	h.ei.Init()
	h.containsDynamicField = make(map[string]bool)
	h.containsEnvVar = make(map[string]bool)
	h.debug = debug
	h.dynamicRgx = regexp.MustCompile(regex.DynamicVariableRegex) // Инициализация регулярки для {{var}}
	h.envRgx = regexp.MustCompile(regex.EnvironmentVariableRegex) // Инициализация регулярки для ${var}

	// Настройка TLS
	tlsConfig := h.initTLSConfig()

	// Настройка транспорта
	tr := h.initTransport(tlsConfig)

	// Создание HTTP-клиента
	h.client = &http.Client{Transport: tr, Timeout: time.Duration(h.packet.Timeout) * time.Second}
	if val, ok := h.packet.Custom["disable-redirect"]; ok { // Проверка настройки отключения редиректов
		val := val.(bool)
		if val {
			h.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Отключение редиректов
			}
		}
	}

	// Инициализация экземпляра запроса
	err = h.initRequestInstance()
	if err != nil {
		return
	}

	// Обработка тела запроса
	if h.dynamicRgx.MatchString(h.packet.Payload) { // Проверка на динамические переменные в теле
		_, err = h.ei.InjectDynamic(h.packet.Payload)
		if err != nil {
			return
		}
		h.containsDynamicField["body"] = true
	}

	if h.envRgx.MatchString(h.packet.Payload) { // Проверка на окружные переменные в теле
		h.containsEnvVar["body"] = true
	}

	// Обработка URL
	if h.dynamicRgx.MatchString(h.packet.URL) { // Проверка на динамические переменные в URL
		_, err = h.ei.InjectDynamic(h.packet.URL)
		if err != nil {
			return
		}
		h.containsDynamicField["url"] = true
	}

	if h.envRgx.MatchString(h.packet.URL) { // Проверка на окружные переменные в URL
		h.containsEnvVar["url"] = true
	}

	// Обработка заголовков
	for k, values := range h.requestSettings.Header {
		for _, v := range values {
			if h.dynamicRgx.MatchString(k) || h.dynamicRgx.MatchString(v) { // Динамические переменные в заголовках
				_, err = h.ei.InjectDynamic(k)
				if err != nil {
					return
				}
				_, err = h.ei.InjectDynamic(v)
				if err != nil {
					return
				}
				h.containsDynamicField["header"] = true
			}
			if h.envRgx.MatchString(k) || h.envRgx.MatchString(v) { // Окружные переменные в заголовках
				h.containsEnvVar["header"] = true
			}
		}
	}

	// Обработка базовой авторизации
	if h.dynamicRgx.MatchString(h.packet.Auth.Username) || h.dynamicRgx.MatchString(h.packet.Auth.Password) { // Динамические переменные в логине/пароле
		_, err = h.ei.InjectDynamic(h.packet.Auth.Username)
		if err != nil {
			return
		}
		_, err = h.ei.InjectDynamic(h.packet.Auth.Password)
		if err != nil {
			return
		}
		h.containsDynamicField["basicauth"] = true
	}

	return
}

// Done закрывает неиспользуемые соединения клиента
func (h *HttpRequester) Done() {
	// Настройки MaxIdleConnsPerHost и MaxIdleConns в Transport позволяют повторно использовать соединения при включённом keep-alive (по умолчанию).
	// После завершения задачи закрываем неактивные соединения, чтобы избежать блокировки сокетов в состоянии TIME_WAIT.
	// Иначе следующая задача не сможет использовать эти сокеты для текущего хоста.
	h.client.CloseIdleConnections()
}

func (h *HttpRequester) Send(envs map[string]interface{}) (res *types.ScenarioStepResult) {
	var statusCode int                // Код ответа
	var contentLength int64           // Длина контента ответа
	var requestErr types.RequestError // Ошибка запроса
	var reqStartTime = time.Now()     // Время начала запроса

	// Для режима отладки
	var copiedReqBody bytes.Buffer                   // Копия тела запроса
	var respBody []byte                              // Тело ответа
	var respHeaders http.Header                      // Заголовки ответа
	var debugInfo map[string]interface{}             // Информация для отладки
	var bodyRead bool                                // Флаг чтения тела ответа
	var bodyReadErr error                            // Ошибка чтения тела
	var extractedVars = make(map[string]interface{}) // Извлечённые переменные
	var failedCaptures = make(map[string]string, 0)  // Неудачные извлечения

	var usableVars = make(map[string]interface{}, len(envs)) // Используемые переменные
	for k, v := range envs {
		usableVars[k] = v // Копируем переданные переменные окружения
	}

	durations := &duration{}                        // Структура для хранения длительностей
	trace := newTrace(durations, h.proxyAddr)       // Трассировка сетевых операций
	httpReq, err := h.prepareReq(usableVars, trace) // Подготовка запроса

	if err != nil { // Не удалось подготовить запрос
		requestErr.Type = types.ErrorInvalidRequest
		requestErr.Reason = fmt.Sprintf("Не удалось подготовить запрос, %s", err.Error())
		res = &types.ScenarioStepResult{
			StepID:    h.packet.ID,
			StepName:  h.packet.Name,
			RequestID: uuid.New(),
			Err:       requestErr,
		}
		return res
	}

	if h.debug { // В режиме отладки копируем тело запроса
		io.Copy(&copiedReqBody, httpReq.Body)
		httpReq.Body = io.NopCloser(bytes.NewReader(copiedReqBody.Bytes()))
	}

	durations.setReqStart() // Фиксация времени начала запроса

	// Выполнение запроса
	httpRes, err := h.client.Do(httpReq)
	if err != nil { // Ошибка выполнения запроса
		requestErr = fetchErrType(err)
		failedCaptures = h.captureEnvironmentVariables(nil, nil, extractedVars)
	}

	// Чтение тела ответа для повторного использования соединений
	if httpRes != nil {
		if len(h.packet.EnvsToCapture) > 0 { // Если нужно извлечь переменные
			respBody, bodyReadErr = io.ReadAll(httpRes.Body)
			bodyRead = true
			if bodyReadErr != nil {
				requestErr = fetchErrType(bodyReadErr)
			}
			failedCaptures = h.captureEnvironmentVariables(httpRes.Header, respBody, extractedVars)
		}

		if !bodyRead { // Если тело ещё не прочитано
			if h.debug { // В режиме отладки сохраняем тело
				respBody, bodyReadErr = io.ReadAll(httpRes.Body)
			} else { // Иначе просто читаем без сохранения
				_, bodyReadErr = io.Copy(io.Discard, httpRes.Body)
			}
			if bodyReadErr != nil {
				requestErr = fetchErrType(bodyReadErr)
			}
		}

		httpRes.Body.Close() // Закрытие тела ответа
		respHeaders = httpRes.Header
		contentLength = httpRes.ContentLength
		statusCode = httpRes.StatusCode
	}
	// Фиксация времени получения ответа после чтения тела
	durations.setResDur()

	var ddResTime time.Duration // Время ответа от сервера (если указано)
	if httpRes != nil && httpRes.Header.Get("x-server-response-time") != "" {
		resTime, _ := strconv.ParseFloat(httpRes.Header.Get("x-server-response-time"), 64)
		ddResTime = time.Duration(resTime*1000) * time.Millisecond
	}

	if h.debug { // Сбор отладочной информации
		debugInfo = map[string]interface{}{
			"url":             httpReq.URL.String(),
			"method":          httpReq.Method,
			"requestHeaders":  httpReq.Header,
			"requestBody":     copiedReqBody.Bytes(),
			"responseBody":    respBody,
			"responseHeaders": respHeaders,
		}
	}

	// Формирование результата
	res = &types.ScenarioStepResult{
		StepID:        h.packet.ID,
		StepName:      h.packet.Name,
		RequestID:     uuid.New(),
		StatusCode:    statusCode,
		RequestTime:   reqStartTime,
		Duration:      durations.totalDuration(), // Общая длительность
		ContentLength: contentLength,
		Err:           requestErr,
		DebugInfo:     debugInfo,
		Custom: map[string]interface{}{
			"dnsDuration":           durations.getDNSDur(),           // Время DNS
			"connDuration":          durations.getConnDur(),          // Время соединения
			"reqDuration":           durations.getReqDur(),           // Время отправки запроса
			"resDuration":           durations.getResDur(),           // Время получения ответа
			"serverProcessDuration": durations.getServerProcessDur(), // Время обработки сервером
		},
		ExtractedEnvs:  extractedVars,  // Извлечённые переменные
		UsableEnvs:     usableVars,     // Используемые переменные
		FailedCaptures: failedCaptures, // Неудачные извлечения
	}

	if strings.EqualFold(httpReq.URL.Scheme, types.ProtocolHTTPS) { // Если HTTPS, добавляем время TLS
		res.Custom["tlsDuration"] = durations.getTLSDur()
	}

	if ddResTime != 0 { // Добавляем время ответа от сервера, если есть
		res.Custom["ddResponseTime"] = ddResTime
	}

	return
}

func (h *HttpRequester) prepareReq(envs map[string]interface{}, trace *httptrace.ClientTrace) (*http.Request, error) {
	re := regexp.MustCompile(regex.DynamicVariableRegex)

	// Обработка тела запроса
	body := h.packet.Payload
	var err error
	if h.containsDynamicField["body"] {
		body, _ = h.ei.InjectDynamic(body)
	}
	if h.containsEnvVar["body"] {
		body, err = h.ei.InjectEnv(body, envs)
		if err != nil {
			return nil, err
		}
	}

	// Обработка URL
	hostURL := h.packet.URL
	var errURL error
	if h.containsDynamicField["url"] {
		hostURL, _ = h.ei.InjectDynamic(hostURL)
	}
	if h.containsEnvVar["url"] {
		hostURL, errURL = h.ei.InjectEnv(hostURL, envs)
		if errURL != nil {
			return nil, errURL
		}
	}

	// Создание нового запроса
	httpReq, err := http.NewRequest(h.packet.Method, hostURL, io.NopCloser(bytes.NewBufferString(body)))
	if err != nil {
		return nil, err
	}
	httpReq.ContentLength = int64(len(body))

	// Установка заголовков из h.requestSettings.Header
	httpReq.Header = h.requestSettings.Header.Clone()

	// Установка хоста, если он был задан
	if h.requestSettings.Host != "" {
		httpReq.Host = h.requestSettings.Host
	}

	// Обработка заголовков с динамическими переменными
	if h.containsDynamicField["header"] {
		for k, values := range httpReq.Header {
			for _, v := range values {
				kk := k
				vv := v
				if re.MatchString(v) {
					vv, _ = h.ei.InjectDynamic(v)
				}
				if re.MatchString(k) {
					kk, _ = h.ei.InjectDynamic(k)
					httpReq.Header.Del(k)
				}
				httpReq.Header.Set(kk, vv)
			}
		}
	}

	// Обработка заголовков с окружными переменными
	if h.containsEnvVar["header"] {
		for k, v := range httpReq.Header {
			for i, vv := range v {
				if h.envRgx.MatchString(vv) {
					vvv, err := h.ei.InjectEnv(vv, envs)
					if err != nil {
						return nil, err
					}
					v[i] = vvv
				}
			}
			httpReq.Header.Set(k, strings.Join(v, ","))

			if h.envRgx.MatchString(k) {
				kk, err := h.ei.InjectEnv(k, envs)
				if err != nil {
					return nil, err
				}
				httpReq.Header.Del(k)
				httpReq.Header.Set(kk, strings.Join(v, ","))
			}
		}
	}

	// Обработка базовой авторизации
	if h.packet.Auth != (types.Auth{}) {
		username := h.packet.Auth.Username
		password := h.packet.Auth.Password
		if h.containsDynamicField["basicauth"] {
			username, _ = h.ei.InjectDynamic(username)
			password, _ = h.ei.InjectDynamic(password)
		}
		httpReq.SetBasicAuth(username, password)
	}

	// Установка настройки keep-alive
	httpReq.Close = h.requestSettings.Close

	// Добавление трассировки
	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))
	return httpReq, nil
}

// На данный момент точный тип ошибки определить нельзя, нужен более элегантный способ
func fetchErrType(err error) types.RequestError {
	var requestErr types.RequestError = types.RequestError{
		Type:   types.ErrorUnkown, // Неизвестная ошибка по умолчанию
		Reason: err.Error()}

	ue, ok := err.(*url.Error) // Проверка, является ли ошибка URL-ошибкой
	if ok {
		errString := ue.Error()
		if strings.Contains(errString, "proxyconnect") { // Ошибки прокси
			if strings.Contains(errString, "connection refused") {
				requestErr = types.RequestError{Type: types.ErrorProxy, Reason: types.ReasonProxyFailed}
			} else if strings.Contains(errString, "Client.Timeout") {
				requestErr = types.RequestError{Type: types.ErrorProxy, Reason: types.ReasonProxyTimeout}
			} else {
				requestErr = types.RequestError{Type: types.ErrorProxy, Reason: errString}
			}
		} else if strings.Contains(errString, context.DeadlineExceeded.Error()) { // Таймаут контекста
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: types.ReasonConnTimeout}
		} else if strings.Contains(errString, "i/o timeout") { // Таймаут ввода-вывода
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: types.ReasonReadTimeout}
		} else if strings.Contains(errString, "connection refused") { // Отказ в соединении
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: types.ReasonConnRefused}
		} else if strings.Contains(errString, context.Canceled.Error()) { // Отмена контекста
			requestErr = types.RequestError{Type: types.ErrorIntented, Reason: types.ReasonCtxCanceled}
		} else if strings.Contains(errString, "connection reset by peer") { // Сброс соединения
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: "connection reset by peer"}
		} else { // Прочие ошибки соединения
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: errString}
		}
	}

	return requestErr
}

func (h *HttpRequester) initTransport(tlsConfig *tls.Config) *http.Transport {
	tr := &http.Transport{
		TLSClientConfig:     tlsConfig,
		Proxy:               http.ProxyURL(h.proxyAddr),
		MaxIdleConnsPerHost: 60000,
		MaxIdleConns:        0,
	}

	tr.DisableKeepAlives = false
	if val, ok := h.packet.Custom["keep-alive"]; ok {
		tr.DisableKeepAlives = !val.(bool)
	}
	if val, ok := h.packet.Custom["disable-compression"]; ok {
		tr.DisableCompression = val.(bool)
	}
	if val, ok := h.packet.Custom["h2"]; ok {
		val := val.(bool)
		if val {
			http2.ConfigureTransport(tr)
		}
	}
	return tr
}

func (h *HttpRequester) initTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	if h.packet.CertPool != nil && h.packet.Cert.Certificate != nil {
		tlsConfig.RootCAs = h.packet.CertPool
		tlsConfig.Certificates = []tls.Certificate{h.packet.Cert}
	}

	if val, ok := h.packet.Custom["hostname"]; ok {
		tlsConfig.ServerName = val.(string)
	}
	return tlsConfig
}

func (h *HttpRequester) initRequestInstance() (err error) {
	// Заголовки
	header := make(http.Header)
	for k, v := range h.packet.Headers {
		if strings.EqualFold(k, "Host") {
			h.requestSettings.Host = v
		} else {
			header.Set(k, v)
		}
	}
	h.requestSettings.Header = header

	// Настройка keep-alive
	h.requestSettings.Close = false
	if val, ok := h.packet.Custom["keep-alive"]; ok {
		h.requestSettings.Close = !val.(bool)
	}
	return
}

func newTrace(duration *duration, proxyAddr *url.URL) *httptrace.ClientTrace {
	var dnsStart, connStart, tlsStart, reqStart, serverProcessStart time.Time // Временные метки этапов

	// Некоторые хуки могут срабатывать многократно (повторные соединения, "Happy Eyeballs" и т.д.).
	// Также некоторые хуки могут срабатывать после TCP-раунда, если запрос не завершён успешно.
	// Для фиксации времени только при первом срабатывании и предотвращения гонки данных используем мьютекс.
	var m sync.Mutex

	return &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) { // Начало DNS-запроса
			m.Lock()
			if dnsStart.IsZero() { // Фиксируем время только при первом вызове
				dnsStart = time.Now()
			}
			m.Unlock()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) { // Завершение DNS
			m.Lock()
			if dnsInfo.Err == nil { // Если нет ошибки, считаем длительность
				duration.setDNSDur(time.Since(dnsStart))
			}
			m.Unlock()
		},
		ConnectStart: func(network, addr string) { // Начало соединения
			m.Lock()
			if connStart.IsZero() { // Фиксируем время только при первом вызове
				connStart = time.Now()
			}
			m.Unlock()
		},
		ConnectDone: func(network, addr string, err error) { // Завершение соединения
			m.Lock()
			if err == nil { // Если нет ошибки, считаем длительность
				duration.setConnDur(time.Since(connStart))
			}
			m.Unlock()
		},
		TLSHandshakeStart: func() { // Начало TLS-рукопожатия
			m.Lock()
			// Хук может срабатывать дважды (прокси и цель HTTPS), фиксируем последнее время
			tlsStart = time.Now()
			m.Unlock()
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, e error) { // Завершение TLS
			m.Lock()
			if e == nil { // Если нет ошибки
				if proxyAddr == nil || proxyAddr.Hostname() != cs.ServerName { // Считаем только для цели, не прокси
					duration.setTLSDur(time.Since(tlsStart))
				}
			}
			m.Unlock()
		},
		GotConn: func(connInfo httptrace.GotConnInfo) { // Получение соединения
			m.Lock()
			if reqStart.IsZero() { // Фиксируем время только при первом вызове
				reqStart = time.Now()
			}
			m.Unlock()
		},
		WroteRequest: func(w httptrace.WroteRequestInfo) { // Запрос отправлен
			m.Lock()
			if w.Err == nil { // Если нет ошибки
				duration.setReqDur(time.Since(reqStart))
				serverProcessStart = time.Now() // Начало обработки сервером
			}
			m.Unlock()
		},
		GotFirstResponseByte: func() { // Получен первый байт ответа
			m.Lock()
			duration.setServerProcessDur(time.Since(serverProcessStart)) // Время обработки сервером
			duration.setResStartTime(time.Now())                         // Время начала ответа
			m.Unlock()
		},
	}
}

func (h *HttpRequester) captureEnvironmentVariables(header http.Header, respBody []byte,
	extractedVars map[string]interface{}) map[string]string {
	var err error
	failedCaptures := make(map[string]string, 0) // Карта для ошибок извлечения
	var captureError extraction.ExtractionError

	// Если запрос провалился, устанавливаем значения по умолчанию
	if header == nil && respBody == nil {
		for _, ce := range h.packet.EnvsToCapture {
			extractedVars[ce.Name] = ""                // Пустое значение по умолчанию
			failedCaptures[ce.Name] = "request failed" // Причина ошибки
		}
		return failedCaptures
	}

	// Извлечение переменных из ответа
	for _, ce := range h.packet.EnvsToCapture {
		var val interface{}
		switch ce.From {
		case types.Header: // Извлечение из заголовков
			val, err = extraction.Extract(header, ce)
		case types.Body: // Извлечение из тела ответа
			val, err = extraction.Extract(respBody, ce)
		}
		if err != nil && errors.As(err, &captureError) { // Если ошибка извлечения
			extractedVars[ce.Name] = ""                    // Устанавливаем пустое значение
			failedCaptures[ce.Name] = captureError.Error() // Записываем ошибку
			continue                                       // Продолжаем для остальных переменных
		}
		extractedVars[ce.Name] = val // Сохраняем извлечённое значение
	}

	return failedCaptures
}
