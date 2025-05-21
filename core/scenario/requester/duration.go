package requester

import (
	"sync"
	"time"
)

type duration struct {
	// Время начала чтения ответа
	resStart time.Time

	reqStart time.Time

	// Длительность DNS-запроса. Если указан IP:Port, будет 0
	dnsDur time.Duration

	// Длительность установки TCP-соединения
	connDur time.Duration

	// Длительность TLS-рукопожатия. Для HTTP будет 0
	tlsDur time.Duration

	// Длительность записи запроса
	reqDur time.Duration

	// Время от полной отправки запроса до первого байта ответа (TTFB)
	serverProcessDur time.Duration

	// Длительность чтения ответа
	resDur time.Duration

	mu sync.Mutex // Мьютекс для потокобезопасности
}

func (d *duration) setReqStart() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.reqStart = time.Now()
}

func (d *duration) setResStartTime(t time.Time) { // Установка времени начала ответа
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.resStart.IsZero() { // Только если ещё не установлено
		d.resStart = t
	}
}

func (d *duration) setDNSDur(t time.Duration) { // Установка длительности DNS
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.dnsDur == 0 { // Только если ещё не установлено
		d.dnsDur = t
	}
}

func (d *duration) getDNSDur() time.Duration { // Получение длительности DNS
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dnsDur
}

func (d *duration) setTLSDur(t time.Duration) { // Установка длительности TLS
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.tlsDur == 0 { // Только если ещё не установлено
		d.tlsDur = t
	}
}

func (d *duration) getTLSDur() time.Duration { // Получение длительности TLS
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.tlsDur
}

func (d *duration) setConnDur(t time.Duration) { // Установка длительности соединения
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.connDur == 0 { // Только если ещё не установлено
		d.connDur = t
	}
}

func (d *duration) getConnDur() time.Duration { // Получение длительности соединения
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connDur
}

func (d *duration) setReqDur(t time.Duration) { // Установка длительности отправки запроса
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.reqDur == 0 { // Только если ещё не установлено
		d.reqDur = t
	}
}

func (d *duration) getReqDur() time.Duration { // Получение длительности отправки запроса
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.reqDur
}

func (d *duration) setServerProcessDur(t time.Duration) { // Установка времени обработки сервером
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.serverProcessDur == 0 { // Только если ещё не установлено
		d.serverProcessDur = t
	}
}

func (d *duration) getServerProcessDur() time.Duration { // Получение времени обработки сервером
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.serverProcessDur
}

func (d *duration) setResDur() { // Установка длительности чтения ответа
	d.mu.Lock()
	defer d.mu.Unlock()
	d.resDur = time.Since(d.resStart) // Разница от начала чтения
}

func (d *duration) getResDur() time.Duration { // Получение длительности чтения ответа
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.resDur
}

func (d *duration) totalDuration() time.Duration { // Общая длительность всех этапов
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dnsDur + d.connDur + d.tlsDur + d.reqDur + d.serverProcessDur + d.resDur
}
