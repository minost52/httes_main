package scenario

import (
	"context"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"httes/core/scenario/requester"
	"httes/core/scenario/scripting/injection"
	"httes/core/types"
	"httes/core/types/regex"
)

// ScenarioService инкапсулирует информацию о прокси, сценариях и отправителях запросов.
type ScenarioService struct {
	clients     map[*url.URL][]scenarioItemRequester // Клиенты (по прокси): для каждого прокси — массив запросчиков
	scenario    types.Scenario                       // Сценарий выполнения
	ctx         context.Context                      // Контекст управления жизненным циклом
	clientMutex sync.Mutex                           // Мьютекс для конкурентного доступа к clients
	debug       bool                                 // Режим отладки
}

// NewScenarioService создает новый экземпляр ScenarioService.
func NewScenarioService() *ScenarioService {
	return &ScenarioService{}
}

// Init инициализирует ScenarioService.clients с использованием переданного types.Scenario и списка прокси.
// Передает переданный ctx в подчиненный запросчик, чтобы можно было управлять жизненным циклом каждого запроса.
func (s *ScenarioService) Init(ctx context.Context, scenario types.Scenario, proxies []*url.URL, debug bool) (err error) {
	s.scenario = scenario
	s.ctx = ctx
	s.debug = debug
	s.clients = make(map[*url.URL][]scenarioItemRequester, len(proxies))
	for _, p := range proxies {
		err = s.createRequesters(p)
		if err != nil {
			return
		}
	}
	return
}

// Do выполняет сценарий для указанного прокси.
// Возвращает "types.Response", заполненный запросчиком для данного прокси, и добавляет startTime в ответ.
// Возвращает ошибку только если types.Response.Err.Type равен types.ErrorProxy или types.ErrorIntented.
func (s *ScenarioService) Do(proxy *url.URL, startTime time.Time) (
	response *types.ScenarioResult, err *types.RequestError) {
	response = &types.ScenarioResult{StepResults: []*types.ScenarioStepResult{}}
	response.StartTime = startTime
	response.ProxyAddr = proxy

	requesters, e := s.getOrCreateRequesters(proxy)
	if e != nil {
		return nil, &types.RequestError{Type: types.ErrorUnkown, Reason: e.Error()}
	}

	// Запускаем окружения отдельно для каждой итерации
	envs := make(map[string]interface{}, len(s.scenario.Envs))
	for k, v := range s.scenario.Envs {
		envs[k] = v
	}
	// Внедряем динамические переменные заранее для каждой итерации
	injectDynamicVars(envs)

	for _, sr := range requesters {
		res := sr.requester.Send(envs)

		if res.Err.Type == types.ErrorProxy || res.Err.Type == types.ErrorIntented {
			err = &res.Err
			if res.Err.Type == types.ErrorIntented {
				return
			}
		}
		response.StepResults = append(response.StepResults, res)

		// Пауза перед выполнением следующего шага
		if sr.sleeper != nil && len(s.scenario.Steps) > 1 {
			sr.sleeper.sleep()
		}

		enrichEnvFromPrevStep(envs, res.ExtractedEnvs)
	}
	return
}

// enrichEnvFromPrevStep добавляет переменные из предыдущего шага в текущее окружение.
func enrichEnvFromPrevStep(m1 map[string]interface{}, m2 map[string]interface{}) {
	for k, v := range m2 {
		m1[k] = v
	}
}

// Done завершает работу всех запросчиков.
func (s *ScenarioService) Done() {
	for _, v := range s.clients {
		for _, r := range v {
			r.requester.Done()
		}
	}
}

// getOrCreateRequesters возвращает список запросчиков для указанного прокси или создает их, если они отсутствуют.
func (s *ScenarioService) getOrCreateRequesters(proxy *url.URL) (requesters []scenarioItemRequester, err error) {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	requesters, ok := s.clients[proxy]
	if !ok {
		err = s.createRequesters(proxy)
		if err != nil {
			return
		}
	}
	return s.clients[proxy], err
}

// createRequesters создает запросчиков для указанного прокси.
func (s *ScenarioService) createRequesters(proxy *url.URL) (err error) {
	s.clients[proxy] = []scenarioItemRequester{}
	for _, si := range s.scenario.Steps {
		var r requester.Requester
		r, err = requester.NewRequester(si)
		if err != nil {
			return
		}
		s.clients[proxy] = append(
			s.clients[proxy],
			scenarioItemRequester{
				scenarioItemID: si.ID,
				sleeper:        newSleeper(si.Sleep),
				requester:      r,
			},
		)

		err = r.Init(s.ctx, si, proxy, s.debug)
		if err != nil {
			return
		}
	}
	return err
}

// injectDynamicVars внедряет динамические переменные в окружение.
func injectDynamicVars(envs map[string]interface{}) {
	dynamicRgx := regexp.MustCompile(regex.DynamicVariableRegex)
	vi := &injection.EnvironmentInjector{}
	vi.Init()
	for k, v := range envs {
		vStr := v.(string)
		if dynamicRgx.MatchString(vStr) {
			injected, err := vi.InjectDynamic(vStr)
			if err != nil {
				continue
			}
			envs[k] = injected
		}
	}
}

type scenarioItemRequester struct {
	scenarioItemID uint16
	sleeper        Sleeper
	requester      requester.Requester
}

// Sleeper — интерфейс для реализации различных стратегий паузы.
type Sleeper interface {
	sleep()
}

// RangeSleep — реализация функции паузы в диапазоне.
type RangeSleep struct {
	min int
	max int
}

func (rs *RangeSleep) sleep() {
	rand.Seed(time.Now().UnixNano())
	dur := rand.Intn(rs.max-rs.min+1) + rs.min
	time.Sleep(time.Duration(dur) * time.Millisecond)
}

// DurationSleep — реализация функции паузы с фиксированной длительностью.
type DurationSleep struct {
	duration int
}

func (ds *DurationSleep) sleep() {
	time.Sleep(time.Duration(ds.duration) * time.Millisecond)
}

// newSleeper — фабричный метод для реализации Sleeper.
func newSleeper(sleepStr string) Sleeper {
	if sleepStr == "" {
		return nil
	}

	var sl Sleeper

	// Поле Sleep уже проверено в types.scenario.validate(). Нет необходимости проверять ошибки парсинга здесь.
	s := strings.Split(sleepStr, "-")
	if len(s) == 2 {
		min, _ := strconv.Atoi(s[0])
		max, _ := strconv.Atoi(s[1])
		if min > max {
			min, max = max, min
		}

		sl = &RangeSleep{
			min: min,
			max: max,
		}
	} else {
		dur, _ := strconv.Atoi(s[0])

		sl = &DurationSleep{
			duration: dur,
		}
	}

	return sl
}
