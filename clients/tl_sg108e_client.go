package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type TPLINKSwitchClient interface {
	GetPortStats() ([]portStats, error)
	GetSystemInfo() (systemInfo, error)
	GetPortConfig() ([]portConfig, error)
	GetHost() string
}

type TPLINKSwitch struct {
	host     string
	username string
	password string
}

type portStats struct {
	State      int
	LinkStatus int
	PktCount   map[string]int
}

type systemInfo struct {
	DeviceDescription string
	MACAddress        string
	IPAddress         string
	Netmask           string
	Gateway           string
	FirmwareVersion   string
	HardwareVersion   string
}

type portConfig struct {
	Port            int
	State           int    // 0=disabled, 1=enabled
	SpeedConfig     int    // 1=auto, 2=10MH, 3=10MF, 4=100MH, 5=100MF, 6=1000MF
	SpeedActual     int    // actual negotiated speed
	FlowControlCfg  int    // 0=off, 1=on
	FlowControlAct  int    // actual flow control state
	TrunkInfo       int    // LAG group info
}

func (client *TPLINKSwitch) GetHost() string {
	return client.host
}

func (client *TPLINKSwitch) login() error {
	resp, err := http.PostForm(fmt.Sprintf("http://%s/logon.cgi", client.host), 
		url.Values{"username": {client.username}, "password": {client.password}, "logon": {"Login"}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (client *TPLINKSwitch) GetPortStats() ([]portStats, error) {
	type allInfo struct {
		State      []int
		LinkStatus []int
		Pkts       []int
	}
	
	// Login first
	err := client.login()
	if err != nil {
		return nil, err
	}
	
	// Get port statistics
	resp2, err := http.Get(fmt.Sprintf("http://%s/PortStatisticsRpm.htm", client.host))
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	
	body, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		return nil, err
	}
	
	var jbody string = strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				string(body), "link_status", `"linkStatus"`),
			"state", `"State"`),
		"pkts", `"Pkts"`)
	
	res := regexp.MustCompile(`all_info = ({[^;]*});`).FindStringSubmatch(jbody)
	if res == nil {
		return nil, errors.New("unexpected response for port statistics http call: " + jbody)
	}
	
	var jparsed allInfo
	json.Unmarshal([]byte(res[1]), &jparsed)
	
	var portsInfos []portStats
	portcount := len(jparsed.State)
	for i := 0; i < portcount; i++ {
		var portInfo portStats
		portInfo.State = jparsed.State[i]
		portInfo.LinkStatus = jparsed.LinkStatus[i]
		if portInfo.State == 1 {
			portInfo.PktCount = make(map[string]int)
			portInfo.PktCount["TxGoodPkt"] = jparsed.Pkts[4*i]
			portInfo.PktCount["TxBadPkt"] = jparsed.Pkts[4*i+1]
			portInfo.PktCount["RxGoodPkt"] = jparsed.Pkts[4*i+2]
			portInfo.PktCount["RxBadPkt"] = jparsed.Pkts[4*i+3]
		}
		portsInfos = append(portsInfos, portInfo)
	}
	fmt.Println(portsInfos)
	return portsInfos, nil
}

func (client *TPLINKSwitch) GetSystemInfo() (systemInfo, error) {
	var sysInfo systemInfo
	
	// Login first
	err := client.login()
	if err != nil {
		return sysInfo, err
	}
	
	// Get system information page
	resp, err := http.Get(fmt.Sprintf("http://%s/SystemInfoRpm.htm", client.host))
	if err != nil {
		return sysInfo, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sysInfo, err
	}
	
	bodyStr := string(body)
	
	// Parse system info using regex patterns
	// Looking for the info_ds JavaScript object
	infoPattern := regexp.MustCompile(`info_ds\s*=\s*{([^}]*)}`)
	match := infoPattern.FindStringSubmatch(bodyStr)
	if match == nil {
		return sysInfo, errors.New("could not find info_ds in system page")
	}
	
	infoBlock := match[1]
	
	// Parse each field
	sysInfo.DeviceDescription = extractStringArray(infoBlock, "descriStr")
	sysInfo.MACAddress = extractStringArray(infoBlock, "macStr")
	sysInfo.IPAddress = extractStringArray(infoBlock, "ipStr")
	sysInfo.Netmask = extractStringArray(infoBlock, "netmaskStr")
	sysInfo.Gateway = extractStringArray(infoBlock, "gatewayStr")
	sysInfo.FirmwareVersion = extractStringArray(infoBlock, "firmwareStr")
	sysInfo.HardwareVersion = extractStringArray(infoBlock, "hardwareStr")
	
	return sysInfo, nil
}

func (client *TPLINKSwitch) GetPortConfig() ([]portConfig, error) {
	type allConfigInfo struct {
		State      []int `json:"state"`
		TrunkInfo  []int `json:"trunk_info"`
		SpdCfg     []int `json:"spd_cfg"`
		SpdAct     []int `json:"spd_act"`
		FcCfg      []int `json:"fc_cfg"`
		FcAct      []int `json:"fc_act"`
	}
	
	// Login first
	err := client.login()
	if err != nil {
		return nil, err
	}
	
	// Get port configuration page
	resp, err := http.Get(fmt.Sprintf("http://%s/PortSettingRpm.htm", client.host))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	bodyStr := string(body)
	
	// Parse the all_info JavaScript object
	configPattern := regexp.MustCompile(`all_info\s*=\s*{([^}]*)}`)
	match := configPattern.FindStringSubmatch(bodyStr)
	if match == nil {
		return nil, errors.New("could not find all_info in port config page")
	}
	
	configBlock := match[1]
	
	// Extract arrays
	var configInfo allConfigInfo
	configInfo.State = extractIntArray(configBlock, "state")
	configInfo.TrunkInfo = extractIntArray(configBlock, "trunk_info")
	configInfo.SpdCfg = extractIntArray(configBlock, "spd_cfg")
	configInfo.SpdAct = extractIntArray(configBlock, "spd_act")
	configInfo.FcCfg = extractIntArray(configBlock, "fc_cfg")
	configInfo.FcAct = extractIntArray(configBlock, "fc_act")
	
	// Convert to portConfig structs
	var portConfigs []portConfig
	portCount := len(configInfo.State)
	
	for i := 0; i < portCount; i++ {
		config := portConfig{
			Port:            i + 1,
			State:           configInfo.State[i],
			SpeedConfig:     configInfo.SpdCfg[i],
			SpeedActual:     configInfo.SpdAct[i],
			FlowControlCfg:  configInfo.FcCfg[i],
			FlowControlAct:  configInfo.FcAct[i],
			TrunkInfo:       configInfo.TrunkInfo[i],
		}
		portConfigs = append(portConfigs, config)
	}
	
	return portConfigs, nil
}

// Helper function to extract string arrays from JavaScript
func extractStringArray(text, fieldName string) string {
	pattern := regexp.MustCompile(fieldName + `:\s*\[\s*"([^"]*)"`)
	match := pattern.FindStringSubmatch(text)
	if match != nil {
		return match[1]
	}
	return ""
}

// Helper function to extract integer arrays from JavaScript
func extractIntArray(text, fieldName string) []int {
	pattern := regexp.MustCompile(fieldName + `:\s*\[([^\]]*)\]`)
	match := pattern.FindStringSubmatch(text)
	if match == nil {
		return []int{}
	}
	
	// Parse the array content
	arrayStr := strings.TrimSpace(match[1])
	if arrayStr == "" {
		return []int{}
	}
	
	// Split by comma and convert to integers
	parts := strings.Split(arrayStr, ",")
	var result []int
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			if val, err := strconv.Atoi(part); err == nil {
				result = append(result, val)
			}
		}
	}
	
	return result
}

func NewTPLinkSwitch(host string, username string, password string) *TPLINKSwitch {
	return &TPLINKSwitch{
		host:     host,
		username: username,
		password: password,
	}
}