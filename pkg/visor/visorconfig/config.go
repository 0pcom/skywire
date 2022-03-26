package visorconfig

import (
	"encoding/json"
	"io/ioutil"

	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/skycoin/skywire-utilities/pkg/cipher"
	utilenv "github.com/skycoin/skywire-utilities/pkg/skyenv"
	"github.com/skycoin/skywire/pkg/app/launcher"
	"github.com/skycoin/skywire/pkg/dmsgc"
	"github.com/skycoin/skywire/pkg/restart"
	"github.com/skycoin/skywire/pkg/routing"
	"github.com/skycoin/skywire/pkg/skyenv"
	"github.com/skycoin/skywire/pkg/transport/network"
	"github.com/skycoin/skywire/pkg/visor/hypervisorconfig"
)

// MakeBaseConfig returns a visor config with 'enforced' fields only.
// This is used as default values if no config is given, or for missing *required* fields.
// This function always returns the latest config version.
func MakeBaseConfig(common *Common, testEnv bool, dmsgHTTP bool, services Services) *V1 {
	//check to see if there are values. Sometimes an empty struct is passed
	if services.DmsgDiscovery == "" {
		//fall back on skyev defaults
		if !testEnv {
			services = Services{utilenv.DefaultDmsgDiscAddr, utilenv.DefaultTpDiscAddr, utilenv.DefaultAddressResolverAddr, utilenv.DefaultRouteFinderAddr, []cipher.PubKey{utilenv.MustPK(utilenv.DefaultSetupPK)}, utilenv.DefaultUptimeTrackerAddr, utilenv.DefaultServiceDiscAddr, utilenv.GetStunServers()}
		} else {
			services = Services{utilenv.TestDmsgDiscAddr, utilenv.TestTpDiscAddr, utilenv.TestAddressResolverAddr, utilenv.TestRouteFinderAddr, []cipher.PubKey{utilenv.MustPK(utilenv.TestSetupPK)}, utilenv.TestUptimeTrackerAddr, utilenv.TestServiceDiscAddr, utilenv.GetStunServers()}
		}
	}
	conf := new(V1)
	conf.Common = common
	conf.Dmsg = &dmsgc.DmsgConfig{
		Discovery:     services.DmsgDiscovery, //utilenv.DefaultDmsgDiscAddr,
		SessionsCount: 1,
		Servers:       []*disc.Entry{},
	}
	conf.Transport = &V1Transport{
		Discovery:         services.TransportDiscovery, //utilenv.DefaultTpDiscAddr,
		AddressResolver:   services.AddressResolver,    //utilenv.DefaultAddressResolverAddr,
		PublicAutoconnect: true,
	}
	conf.Routing = &V1Routing{
		RouteFinder:        services.RouteFinder, //utilenv.DefaultRouteFinderAddr,
		SetupNodes:         services.SetupNodes,  //[]cipher.PubKey{utilenv.MustPK(utilenv.DefaultSetupPK)},
		RouteFinderTimeout: DefaultTimeout,
	}
	conf.Launcher = &V1Launcher{
		ServiceDisc: services.ServiceDiscovery, //utilenv.DefaultServiceDiscAddr,
		Apps:        nil,
		ServerAddr:  skyenv.DefaultAppSrvAddr,
		BinPath:     skyenv.DefaultAppBinPath,
	}
	conf.UptimeTracker = &V1UptimeTracker{
		Addr: services.UptimeTracker, //utilenv.DefaultUptimeTrackerAddr,
	}
	conf.CLIAddr = skyenv.DefaultRPCAddr
	conf.LogLevel = skyenv.DefaultLogLevel
	conf.LocalPath = skyenv.DefaultLocalPath
	conf.StunServers = services.StunServers //utilenv.GetStunServers()
	conf.ShutdownTimeout = DefaultTimeout
	conf.RestartCheckDelay = Duration(restart.DefaultCheckDelay)
	conf.DMSGHTTPPath = skyenv.DefaultDMSGHTTPPath

	conf.Dmsgpty = &V1Dmsgpty{
		DmsgPort: skyenv.DmsgPtyPort,
		CLINet:   skyenv.DefaultDmsgPtyCLINet,
		CLIAddr:  skyenv.DefaultDmsgPtyCLIAddr(),
	}

	conf.STCP = &network.STCPConfig{
		ListeningAddress: skyenv.DefaultSTCPAddr,
		PKTable:          nil,
	}

	return conf
}

// MakeDefaultConfig returns the default visor config from a given secret key (if specified).
// The config's 'sk' field will be nil if not specified.
// Generated config will be saved to 'confPath'.
// This function always returns the latest config version.
func MakeDefaultConfig(log *logging.MasterLogger, confPath string, sk *cipher.SecKey, pkgEnv bool, testEnv bool, dmsgHTTP bool, hypervisor bool, services Services) (*V1, error) {
	cc, err := NewCommon(log, confPath, V1Name, sk)
	if err != nil {
		return nil, err
	}
	// Enforce version and keys in 'cc'.
	cc.Version = V1Name
	if err := cc.ensureKeys(); err != nil {
		return nil, err
	}
	// Actual config generation.
	conf := MakeBaseConfig(cc, testEnv, dmsgHTTP, services)

	conf.Launcher.Apps = makeDefaultLauncherAppsConfig()

	conf.Hypervisors = make([]cipher.PubKey, 0)

	if hypervisor {
		config := hypervisorconfig.GenerateWorkDirConfig(false)
		conf.Hypervisor = &config
	}

	if pkgEnv {
		pkgconfig := skyenv.PackageConfig()
		conf.LocalPath = pkgconfig.LocalPath
		conf.Launcher.BinPath = pkgconfig.Launcher.BinPath
		conf.DMSGHTTPPath = pkgconfig.DmsghttpPath
		if conf.Hypervisor != nil {
			conf.Hypervisor.EnableAuth = pkgconfig.Hypervisor.EnableAuth
			conf.Hypervisor.DBPath = pkgconfig.Hypervisor.DbPath
		}
	}

	// Use dmsg urls for services and add dmsg-servers
	if dmsgHTTP {
		var dmsgHTTPServersList DmsgHTTPServers
		serversListJSON, err := ioutil.ReadFile(conf.DMSGHTTPPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to read dmsghttp-config.json file.")
		}
		err = json.Unmarshal(serversListJSON, &dmsgHTTPServersList)
		if err != nil {
			log.WithError(err).Fatal("Error during parsing servers list")
		}
		if testEnv {
			conf.Dmsg.Servers = dmsgHTTPServersList.Test.DMSGServers
			conf.Dmsg.Discovery = dmsgHTTPServersList.Test.DMSGDiscovery
			conf.Transport.AddressResolver = dmsgHTTPServersList.Test.AddressResolver
			conf.Transport.Discovery = dmsgHTTPServersList.Test.TransportDiscovery
			conf.UptimeTracker.Addr = dmsgHTTPServersList.Test.UptimeTracker
			conf.Routing.RouteFinder = dmsgHTTPServersList.Test.RouteFinder
			conf.Launcher.ServiceDisc = dmsgHTTPServersList.Test.ServiceDiscovery
		} else {
			conf.Dmsg.Servers = dmsgHTTPServersList.Prod.DMSGServers
			conf.Dmsg.Discovery = dmsgHTTPServersList.Prod.DMSGDiscovery
			conf.Transport.AddressResolver = dmsgHTTPServersList.Prod.AddressResolver
			conf.Transport.Discovery = dmsgHTTPServersList.Prod.TransportDiscovery
			conf.UptimeTracker.Addr = dmsgHTTPServersList.Prod.UptimeTracker
			conf.Routing.RouteFinder = dmsgHTTPServersList.Prod.RouteFinder
			conf.Launcher.ServiceDisc = dmsgHTTPServersList.Prod.ServiceDiscovery
		}
	}

	return conf, nil

}

/*
// MakeTestConfig acts like MakeDefaultConfig, however, test deployment service addresses are used instead.
func MakeTestConfig(log *logging.MasterLogger, confPath string, sk *cipher.SecKey, hypervisor bool, services Services) (*V1, error) {
	conf, err := MakeDefaultConfig(log, confPath, sk, hypervisor, services)
	if err != nil {
		return nil, err
	}
	SetDefaultTestingValues(conf)
	if conf.Hypervisor != nil {
		conf.Hypervisor.DmsgDiscovery = conf.Transport.Discovery
	}
	return conf, nil
}
*/

// SetDefaultTestingValues mutates configuration to use testing values
// makeDefaultLauncherAppsConfig creates default launcher config for apps,
// for package based installation in other platform (Darwin, Windows) it only includes
// the shipped apps for that platforms
func makeDefaultLauncherAppsConfig() []launcher.AppConfig {
	defaultConfig := []launcher.AppConfig{
		{
			Name:      skyenv.VPNClientName,
			AutoStart: false,
			Port:      routing.Port(skyenv.VPNClientPort),
		},
		{
			Name:      skyenv.SkychatName,
			AutoStart: true,
			Port:      routing.Port(skyenv.SkychatPort),
			Args:      []string{"-addr", skyenv.SkychatAddr},
		},
		{
			Name:      skyenv.SkysocksName,
			AutoStart: true,
			Port:      routing.Port(skyenv.SkysocksPort),
		},
		{
			Name:      skyenv.SkysocksClientName,
			AutoStart: false,
			Port:      routing.Port(skyenv.SkysocksClientPort),
		},
		{
			Name:      skyenv.VPNServerName,
			AutoStart: false,
			Port:      routing.Port(skyenv.VPNServerPort),
		},
	}
	return defaultConfig
}

// DmsgHTTPServers struct use to unmarshal dmsghttp file
type DmsgHTTPServers struct {
	Test DmsgHTTPServersData `json:"test"`
	Prod DmsgHTTPServersData `json:"prod"`
}

// DmsgHTTPServersData is a part of DmsgHTTPServers
type DmsgHTTPServersData struct {
	DMSGServers        []*disc.Entry `json:"dmsg_servers"`
	DMSGDiscovery      string        `json:"dmsg_discovery"`
	TransportDiscovery string        `json:"transport_discovery"`
	AddressResolver    string        `json:"address_resolver"`
	RouteFinder        string        `json:"route_finder"`
	UptimeTracker      string        `json:"uptime_tracker"`
	ServiceDiscovery   string        `json:"service_discovery"`
}

// Services are subdomains and IP addresses of the skywire services
type Services struct {
	DmsgDiscovery      string          `json:"dmsg_discovery"`
	TransportDiscovery string          `json:"transport_discovery"`
	AddressResolver    string          `json:"address_resolver"`
	RouteFinder        string          `json:"route_finder"`
	SetupNodes         []cipher.PubKey `json:"setup_nodes"`
	UptimeTracker      string          `json:"uptime_tracker"`
	ServiceDiscovery   string          `json:"service_discovery"`
	StunServers        []string        `json:"stun_servers"`
}
