package ncsi

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/crazy-max/WindowsSpyBlocker/app/menu"
	"github.com/crazy-max/WindowsSpyBlocker/app/utils/print"
	"github.com/crazy-max/WindowsSpyBlocker/app/utils/timeu"
	"github.com/crazy-max/WindowsSpyBlocker/app/utils/windows"
	"github.com/fatih/color"
	"github.com/miekg/dns"
	"golang.org/x/sys/windows/registry"
)

// NCSI reg keys
const (
	REG_KEY                  = `SYSTEM\CurrentControlSet\Services\NlaSvc\Parameters\Internet`
	REG_WEB_PROBE_HOST       = "ActiveWebProbeHost"
	REG_WEB_PROBE_PATH       = "ActiveWebProbePath"
	REG_WEB_PROBE_CONTENT    = "ActiveWebProbeContent"
	REG_WEB_PROBE_HOST_V6    = "ActiveWebProbeHostV6"
	REG_WEB_PROBE_PATH_V6    = "ActiveWebProbePathV6"
	REG_WEB_PROBE_CONTENT_V6 = "ActiveWebProbeContentV6"
	REG_DNS_PROBE_HOST       = "ActiveDnsProbeHost"
	REG_DNS_PROBE_CONTENT    = "ActiveDnsProbeContent"
	REG_DNS_PROBE_HOST_V6    = "ActiveDnsProbeHostV6"
	REG_DNS_PROBE_CONTENT_V6 = "ActiveDnsProbeContentV6"
)

type ncsi struct {
	webHostV4    string
	webPathV4    string
	webContentV4 string
	webHostV6    string
	webPathV6    string
	webContentV6 string
	dnsHostV4    string
	dnsContentV4 string
	dnsHostV6    string
	dnsContentV6 string
}

// Menu of NCSI
func Menu(args ...string) (err error) {
	menuCommands := []menu.CommandOption{
		{
			Description: "Display your current NCSI values",
			Function:    current,
		},
		{
			Description: "Apply WindowsSpyBlocker NCSI",
			Function:    wsb,
		},
		{
			Description: "Apply Microsoft NCSI",
			Function:    microsoft,
		},
		{
			Description: "Test the internet connection",
			Function:    test,
		},
	}

	menuOptions := menu.NewOptions("NCSI", "'menu' for help [telemetry-ncsi]> ", 0, "")

	menuN := menu.NewMenu(menuCommands, menuOptions)
	menuN.Start()
	return
}

func current(args ...string) error {
	fmt.Println()
	defer timeu.Track(time.Now())

	ncsi, err := getNcsi()
	if err != nil {
		print.Error(err)
		return nil
	}
	fmt.Print("Getting current registry values... ")
	print.Ok()

	fmt.Println()
	color.New(color.FgMagenta).Println("# Web Probe IPv4")
	print.RegString(REG_WEB_PROBE_HOST, ncsi.webHostV4)
	print.RegString(REG_WEB_PROBE_PATH, ncsi.webPathV4)
	print.RegString(REG_WEB_PROBE_CONTENT, ncsi.webContentV4)

	color.New(color.FgMagenta).Println("\n# Web Probe IPv6")
	print.RegString(REG_WEB_PROBE_HOST_V6, ncsi.webHostV6)
	print.RegString(REG_WEB_PROBE_PATH_V6, ncsi.webPathV6)
	print.RegString(REG_WEB_PROBE_CONTENT_V6, ncsi.webContentV6)

	color.New(color.FgMagenta).Println("\n# DNS Probe IPv4")
	print.RegString(REG_DNS_PROBE_HOST, ncsi.dnsHostV4)
	print.RegString(REG_DNS_PROBE_CONTENT, ncsi.dnsContentV4)

	color.New(color.FgMagenta).Println("\n# DNS Probe IPv6")
	print.RegString(REG_DNS_PROBE_HOST, ncsi.dnsHostV6)
	print.RegString(REG_DNS_PROBE_CONTENT, ncsi.dnsContentV6)

	fmt.Println()
	return nil
}

func wsb(args ...string) (err error) {
	defer timeu.Track(time.Now())
	return setNcsi(ncsi{
		webHostV4:    "raw.githubusercontent.com",
		webPathV4:    "crazy-max/WindowsSpyBlocker/master/data/ncsi/ncsi.txt",
		webContentV4: "WindowsSpyBlocker",
		webHostV6:    "raw.githubusercontent.com",
		webPathV6:    "crazy-max/WindowsSpyBlocker/master/data/ncsi/ncsi.txt",
		webContentV6: "WindowsSpyBlocker",
		dnsHostV4:    "ns1.p16.dynect.net",
		dnsContentV4: "208.78.70.16",
		dnsHostV6:    "ns1.p16.dynect.net",
		dnsContentV6: "2001:500:90:1::16",
	})
}

func microsoft(args ...string) error {
	defer timeu.Track(time.Now())
	return setNcsi(ncsi{
		webHostV4:    "www.msftncsi.com",
		webPathV4:    "ncsi.txt",
		webContentV4: "Microsoft NCSI",
		webHostV6:    "ipv6.msftncsi.com",
		webPathV6:    "ncsi.txt",
		webContentV6: "Microsoft NCSI",
		dnsHostV4:    "dns.msftncsi.com",
		dnsContentV4: "131.107.255.255",
		dnsHostV6:    "dns.msftncsi.com",
		dnsContentV6: "fd3e:4f5a:5b81::1",
	})
}

func test(args ...string) (err error) {
	fmt.Println()
	defer timeu.Track(time.Now())

	current, err := getNcsi()
	if err != nil {
		return nil
	}

	fmt.Println()
	fmt.Print("Testing web request IPv4... ")
	err = testHttpProbe("http://"+current.webHostV4+"/"+current.webPathV4, current.webContentV4)
	if err != nil {
		print.Error(err)
	} else {
		print.Ok()
	}

	fmt.Print("Testing web request IPv6... ")
	err = testHttpProbe("http://"+current.webHostV6+"/"+current.webPathV6, current.webContentV6)
	if err != nil {
		print.Error(err)
	} else {
		print.Ok()
	}

	fmt.Print("Testing DNS resolution IPv4... ")
	err = testDnsProbe(current.dnsHostV4, dns.TypeA, current.dnsContentV4)
	if err != nil {
		print.Error(err)
	} else {
		print.Ok()
	}

	fmt.Print("Testing DNS resolution IPv6... ")
	err = testDnsProbe(current.dnsHostV6, dns.TypeAAAA, current.dnsContentV6)
	if err != nil {
		print.Error(err)
	} else {
		print.Ok()
	}

	return nil
}

func getNcsi() (ncsi, error) {
	key, err := windows.OpenRegKey(registry.LOCAL_MACHINE, REG_KEY, registry.QUERY_VALUE)
	if err != nil {
		return ncsi{}, err
	}
	defer key.Close()

	return ncsi{
		webHostV4:    windows.GetRegString(key, REG_WEB_PROBE_HOST),
		webPathV4:    windows.GetRegString(key, REG_WEB_PROBE_PATH),
		webContentV4: windows.GetRegString(key, REG_WEB_PROBE_CONTENT),
		webHostV6:    windows.GetRegString(key, REG_WEB_PROBE_HOST_V6),
		webPathV6:    windows.GetRegString(key, REG_WEB_PROBE_PATH_V6),
		webContentV6: windows.GetRegString(key, REG_WEB_PROBE_CONTENT_V6),
		dnsHostV4:    windows.GetRegString(key, REG_DNS_PROBE_HOST),
		dnsContentV4: windows.GetRegString(key, REG_DNS_PROBE_CONTENT),
		dnsHostV6:    windows.GetRegString(key, REG_DNS_PROBE_HOST_V6),
		dnsContentV6: windows.GetRegString(key, REG_DNS_PROBE_CONTENT_V6),
	}, nil
}

func setNcsi(aNcsi ncsi) error {
	fmt.Println()

	key, err := windows.OpenRegKey(registry.LOCAL_MACHINE, REG_KEY, registry.WRITE)
	if err != nil {
		return nil
	}
	defer key.Close()

	if err = windows.SetRegString(key, REG_WEB_PROBE_HOST, aNcsi.webHostV4); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_WEB_PROBE_PATH, aNcsi.webPathV4); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_WEB_PROBE_CONTENT, aNcsi.webContentV4); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_WEB_PROBE_HOST_V6, aNcsi.webHostV6); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_WEB_PROBE_PATH_V6, aNcsi.webPathV6); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_WEB_PROBE_CONTENT_V6, aNcsi.webContentV6); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_DNS_PROBE_HOST, aNcsi.dnsHostV4); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_DNS_PROBE_CONTENT, aNcsi.dnsContentV4); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_DNS_PROBE_HOST_V6, aNcsi.dnsHostV6); err != nil {
		return nil
	}
	if err := windows.SetRegString(key, REG_DNS_PROBE_CONTENT_V6, aNcsi.dnsContentV6); err != nil {
		return nil
	}

	fmt.Println()
	return nil
}

func testHttpProbe(url string, content string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("HTTP status code %d", response.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	bodyString := string(bodyBytes)
	if bodyString != content {
		return fmt.Errorf("Invalid content '%s'. Expected '%s'", bodyString, content)
	}

	return nil
}

func testDnsProbe(host string, dnsType uint16, content string) error {
	ipType := "IPv4"
	if dnsType == dns.TypeAAAA {
		ipType = "IPv6"
	}

	dnsTypeStr := "A"
	if dnsType == dns.TypeAAAA {
		dnsTypeStr = "AAAA"
	}

	localc := new(dns.Client)
	localc.ReadTimeout = 5 * 1e9

	localm := new(dns.Msg)
	localm.RecursionDesired = true
	localm.SetQuestion(dns.Fqdn(host), dnsType)

	ra, _, err := localc.Exchange(localm, net.JoinHostPort(host, "53"))
	if ra == nil {
		return fmt.Errorf("Error getting the %s address of %s: %s", ipType, host, err.Error())
	}
	if ra.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("Invalid answer name %s after %s query: %s", host, dnsTypeStr, dns.RcodeToString[ra.Rcode])
	}
	if dnsType == dns.TypeA && ra.Answer[0].(*dns.A).A.String() != content {
		return fmt.Errorf("Invalid content '%s'. Expected '%s'", ra.Answer[0].(*dns.A).A.String(), content)
	}
	if dnsType == dns.TypeAAAA && ra.Answer[0].(*dns.AAAA).AAAA.String() != content {
		return fmt.Errorf("Invalid content '%s'. Expected '%s'", ra.Answer[0].(*dns.AAAA).AAAA.String(), content)
	}

	return nil
}
