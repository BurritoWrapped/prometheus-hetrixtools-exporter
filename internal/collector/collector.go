package collector

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/prometheus/client_golang/prometheus"
)

const (
    apiBaseURL     = "https://api.hetrixtools.com/v3/uptime-monitors"
    requestTimeout = 5 * time.Second
)

type APIResponse struct {
    Monitors []struct {
        ID                 string `json:"id"`
        Name               string `json:"name"`
        Type               string `json:"type"`
        Target             string `json:"target"`
        ResolveAddress     string `json:"resolve_address"`
        ResolveAddressInfo struct {
            ASN      string `json:"ASN"`
            ISP      string `json:"ISP"`
            City     string `json:"City"`
            Region   string `json:"Region"`
            Country  string `json:"Country"`
        } `json:"resolve_address_info"`
        Port              interface{} `json:"port"`
        Keyword           interface{} `json:"keyword"`
        Category          string      `json:"category"`
        Timeout           int         `json:"timeout"`
        CheckFrequency    int         `json:"check_frequency"`
        ContactLists      []string    `json:"contact_lists"`
        CreatedAt         int64       `json:"created_at"`
        LastCheck         int64       `json:"last_check"`
        LastStatusChange  int64       `json:"last_status_change"`
        UptimeStatus      string      `json:"uptime_status"`
        MonitorStatus     string      `json:"monitor_status"`
        Uptime            string      `json:"uptime"`
        UptimeInclMaint   string      `json:"uptime_incl_maint"`
        Locations         map[string]struct {
            UptimeStatus   string `json:"uptime_status"`
            ResponseTime   int    `json:"response_time"`
            LastCheck      int64  `json:"last_check"`
        } `json:"locations"`
        SSLExpirationDate     interface{} `json:"ssl_expiration_date"`
        SSLExpirationWarn     bool        `json:"ssl_expiration_warn"`
        SSLExpirationWarnDays int         `json:"ssl_expiration_warn_days"`
        DomainExpirationDate  string      `json:"domain_expiration_date"`
        DomainExpirationWarn  bool        `json:"domain_expiration_warn"`
        DomainExpirationWarnDays int      `json:"domain_expiration_warn_days"`
        Nameservers             interface{} `json:"nameservers"`
        NameserversChangeWarn   bool        `json:"nameservers_change_warn"`
        PublicReport            bool        `json:"public_report"`
        PublicTarget            bool        `json:"public_target"`
        MaxRedirects            interface{} `json:"max_redirects"`
        HttpMethod              interface{} `json:"http_method"`
        AcceptedHTTPCodes       interface{} `json:"accepted_http_codes"`
        VerifySSLCertificate    bool        `json:"verify_ssl_certificate"`
        VerifySSLHostname       bool        `json:"verify_ssl_hostname"`
        NumberOfTries           int         `json:"number_of_tries"`
        TriggeringLocations     int         `json:"triggering_locations"`
        AlertAfterMinutes       int         `json:"alert_after_minutes"`
        RepeatAlertTimes        int         `json:"repeat_alert_times"`
        RepeatAlertFrequency    int         `json:"repeat_alert_frequency"`
        AgentID                 string      `json:"agent_id"`
    } `json:"monitors"`
    Meta struct {
        Total         int `json:"total"`
        TotalFiltered int `json:"total_filtered"`
        Returned      int `json:"returned"`
        Pagination    struct {
            Current  int `json:"current"`
            Last     int `json:"last"`
            Previous int `json:"previous"`
            Next     int `json:"next"`
        } `json:"pagination"`
    }
}


type Collector struct {
    apiKey              string
    client              http.Client
    monitorUptimeStatus *prometheus.GaugeVec
    monitorResponseTime *prometheus.GaugeVec
}

func New(namespace, apiKey string) *Collector {
    monitorUptimeStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: namespace,
        Name:      "uptime_monitor_status",
        Help:      "Uptime status of recent monitor check (1 for up, 0 for down)",
    }, []string{"id", "name", "target", "location", "status_text"})

    monitorResponseTime := prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: namespace,
        Name:      "uptime_monitor_response_time_seconds",
        Help:      "Response time of recent monitor check in seconds",
    }, []string{"id", "name", "location", "target", "timeout"})

    return &Collector{
        apiKey:              apiKey,
        client:              http.Client{Timeout: requestTimeout},
        monitorUptimeStatus: monitorUptimeStatus,
        monitorResponseTime: monitorResponseTime,
    }
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
    c.monitorUptimeStatus.Describe(ch)
    c.monitorResponseTime.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
    apiUrl := fmt.Sprintf("%s?per_page=200&page=1", apiBaseURL)
    req, err := http.NewRequest("GET", apiUrl, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

    resp, err := c.client.Do(req)
    if err != nil {
        fmt.Println("Error making API request:", err)
        return
    }
    defer resp.Body.Close()

    var apiResponse APIResponse
    decoder := json.NewDecoder(resp.Body)
    if err := decoder.Decode(&apiResponse); err != nil {
        fmt.Println("Error decoding API response:", err)
        return
    }

    for _, monitor := range apiResponse.Monitors {
        for locationKey, locationData := range monitor.Locations {
            statusText := "down"
            if locationData.UptimeStatus == "up" {
                statusText = "up"
            }
            c.monitorUptimeStatus.WithLabelValues(monitor.ID, monitor.Name, monitor.Target, locationKey, statusText).Set(1.0)
            c.monitorResponseTime.WithLabelValues(monitor.ID, monitor.Name, locationKey, monitor.Target, strconv.Itoa(monitor.Timeout)).Set(float64(locationData.ResponseTime))
        }
    }

    c.monitorUptimeStatus.Collect(ch)
    c.monitorResponseTime.Collect(ch)
}

func (c *Collector) handleRateLimiting(resp *http.Response) {
    // Implementation of rate limiting logic, if applicable
}

