package config

// Hardened defaults (fork): sessions last one day instead of 31, and the
// login brute-force lockout is on.
const (
	defaultRefreshTokenDuration = 86400
	defaultLoginLockoutDuration = 300
	defaultLoginMaxFailures     = 5
)

var defaultConfig = &Config{
	Proto: "http",
	Host:  "",
	Port: Port{
		Http:  80,
		Https: 443,
	},
	Cert: Cert{
		Crt: "server.crt",
		Key: "server.key",
	},
	Logger: Logger{
		Level: "info",
		File:  "stdout",
	},
	JWT: JWT{
		SecretKey:            "",
		RefreshTokenDuration: defaultRefreshTokenDuration,
		RevokeTokensOnLogout: true,
	},
	Stun: "stun.l.google.com:19302",
	Turn: Turn{
		TurnAddr: "",
		TurnUser: "",
		TurnCred: "",
	},
	Authentication: "enable",
	Security: Security{
		LoginLockoutDuration: defaultLoginLockoutDuration,
		LoginMaxFailures:     5,
	},
}

func checkDefaultValue() {
	if instance.JWT.SecretKey == "" {
		instance.JWT.SecretKey = generateRandomSecretKey()
		instance.JWT.RevokeTokensOnLogout = true
	}

	if instance.JWT.RefreshTokenDuration == 0 {
		instance.JWT.RefreshTokenDuration = defaultRefreshTokenDuration
	}

	// loginLockoutDuration: 0 (the upstream default, written into every
	// generated server.yaml) now means "use the default"; only a negative
	// value turns the protection off.
	if instance.Security.LoginLockoutDuration == 0 {
		instance.Security.LoginLockoutDuration = defaultLoginLockoutDuration
	}
	if instance.Security.LoginMaxFailures <= 0 {
		instance.Security.LoginMaxFailures = defaultLoginMaxFailures
	}

	if instance.Stun == "" {
		instance.Stun = "stun.l.google.com:19302"
	}

	if instance.Authentication == "" {
		instance.Authentication = "enable"
	}

	instance.Hardware = getHardware()
}
