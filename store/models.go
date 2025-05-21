package store

type Scenario struct {
	ID          int
	Name        string
	Description string
	Tags        []string
	Endpoints   []Endpoint
	Cert        *Cert
}

type LoadProfile struct {
	ID   int
	Name string
	Type string
}

type TestRun struct {
	ID        int
	Name      string
	StartTime string
	Status    string
}

type Endpoint struct {
	URL     string
	Method  string
	Headers string
}

type Cert struct {
	Name string
	Path string
	Key  string
}
