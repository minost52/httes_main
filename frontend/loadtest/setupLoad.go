package loadtest

import (
	"fmt"
	"net/url"
	"strings"

	"httes/core/proxy"
	"httes/core/types"
)

// createScenarioStep создаёт шаг сценария на основе пользовательского ввода
func createScenarioStep(protocol, urlText, method, username, password, certPath, certKeyPath string) (types.ScenarioStep, error) {
	targetURL := strings.ToLower(protocol) + "://" + urlText
	parsedURL, err := url.ParseRequestURI(targetURL)
	if err != nil || parsedURL.Host == "" {
		return types.ScenarioStep{}, fmt.Errorf("invalid target URL: %s", targetURL)
	}

	step := types.ScenarioStep{
		ID:      1,
		Method:  strings.ToUpper(method),
		URL:     targetURL,
		Timeout: 30,
		Headers: map[string]string{},
	}

	// Настройка аутентификации
	if username != "" && password != "" {
		step.Auth = types.Auth{
			Type:     types.AuthHttpBasic,
			Username: username,
			Password: password,
		}
	}

	// Настройка сертификатов
	if certPath != "" && certKeyPath != "" {
		fmt.Println("Parsing TLS certificates:", certPath, certKeyPath)
		cert, pool, err := types.ParseTLS(certPath, certKeyPath)
		if err != nil {
			return types.ScenarioStep{}, fmt.Errorf("failed to parse TLS certificates: %v", err)
		}
		step.Cert = cert
		step.CertPool = pool
	}

	return step, nil
}

// createHeart создаёт объект Heart на основе параметров теста
func createHeart(scenario types.Scenario, proxyAddr *url.URL, iterationCount, testDuration int, loadType, reportDestination string, verbose bool) types.Heart {
	fmt.Println("Creating Heart with:", iterationCount, testDuration, loadType)
	return types.Heart{
		IterationCount: iterationCount,
		LoadType:       strings.ToLower(loadType),
		TestDuration:   testDuration,
		Scenario:       scenario,
		Proxy: proxy.Proxy{
			Strategy: proxy.ProxyTypeSingle,
			Addr:     proxyAddr,
		},
		ReportDestination: reportDestination,
		Debug:             verbose,
	}
}
