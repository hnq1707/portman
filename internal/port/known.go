package port

// WellKnownPorts maps common port numbers to human-readable service labels.
// This is the single source of truth for port labels across the entire CLI.
var WellKnownPorts = map[int]string{
	21:    "FTP",
	22:    "SSH",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	443:   "HTTPS",
	1433:  "MSSQL",
	1521:  "Oracle",
	3000:  "Dev Server",
	3306:  "MySQL",
	4200:  "Angular",
	5000:  "Flask/API",
	5173:  "Vite",
	5432:  "PostgreSQL",
	5433:  "PostgreSQL",
	5672:  "RabbitMQ",
	6379:  "Redis",
	8000:  "Django/API",
	8080:  "HTTP Alt",
	8443:  "HTTPS Alt",
	8888:  "Jupyter",
	9092:  "Kafka",
	9200:  "Elasticsearch",
	9229:  "Node Debugger",
	15672: "RabbitMQ UI",
	27017: "MongoDB",
}

// GetPortLabel returns the human-readable label for a port, or empty string.
func GetPortLabel(p int) string {
	if label, ok := WellKnownPorts[p]; ok {
		return label
	}
	return ""
}

// IsKnownPort returns true if the port is a recognized well-known service port.
func IsKnownPort(p int) bool {
	_, ok := WellKnownPorts[p]
	return ok
}
