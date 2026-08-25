package health

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type ServiceHealth struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	URL       string    `json:"url"`
	LastCheck time.Time `json:"last_check"`
	Message   string    `json:"message,omitempty"`
}

type DependencyMap struct {
	Services  []ServiceHealth `json:"services"`
	Status    string          `json:"status"`
	LastCheck time.Time       `json:"last_check"`
}

var (
	idpURL              string
	caURL               string
	enrollmentAgentURL  string
	protectedServiceURL string
)

var dependencyMap = &DependencyMap{}
var dependencyMu sync.Mutex

// Start reads service URLs from the environment and launches the periodic
// dependency health-check goroutine.
func Start() {
	idpURL = os.Getenv("IDP_URL")
	if idpURL == "" {
		idpURL = "http://localhost:8001"
	}

	caURL = os.Getenv("CA_URL")
	if caURL == "" {
		caURL = "http://localhost:8002"
	}

	enrollmentAgentURL = os.Getenv("ENROLLMENT_AGENT_URL")
	if enrollmentAgentURL == "" {
		enrollmentAgentURL = "http://localhost:8003"
	}

	protectedServiceURL = os.Getenv("PROTECTED_SERVICE_URL")
	if protectedServiceURL == "" {
		protectedServiceURL = "https://localhost:8004"
	}

	go checkHealthPeriodically()
}

func checkHealth(url string) (string, string) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}}}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return "down", fmt.Sprintf("Error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "up", "Healthy"
	}

	return "degraded", fmt.Sprintf("Status: %d", resp.StatusCode)
}

func checkHealthPeriodically() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dependencyMu.Lock()

		services := []ServiceHealth{
			{Name: "IdP Mock", URL: idpURL},
			{Name: "CA Service", URL: caURL},
			{Name: "Enrollment Agent", URL: enrollmentAgentURL},
			{Name: "Protected Service", URL: protectedServiceURL},
		}

		for i := range services {
			status, msg := checkHealth(services[i].URL)
			services[i].Status = status
			services[i].LastCheck = time.Now()
			services[i].Message = msg
		}

		overallStatus := "ok"
		for _, svc := range services {
			if svc.Status == "down" {
				overallStatus = "error"
				break
			}
			if svc.Status == "degraded" && overallStatus != "error" {
				overallStatus = "warning"
			}
		}

		dependencyMap.Services = services
		dependencyMap.Status = overallStatus
		dependencyMap.LastCheck = time.Now()

		dependencyMu.Unlock()
	}
}

// GetDependencyMap returns a snapshot of the current dependency health map.
func GetDependencyMap() *DependencyMap {
	dependencyMu.Lock()
	defer dependencyMu.Unlock()
	snapshot := *dependencyMap
	return &snapshot
}

// EnrollmentAgentURL returns the configured enrollment-agent URL, used by
// the enroll proxy handler.
func EnrollmentAgentURL() string {
	return enrollmentAgentURL
}
