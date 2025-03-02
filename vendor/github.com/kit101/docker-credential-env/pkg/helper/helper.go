package helper

import (
	"net/url"
	"os"
	"strings"

	docker_credentials "github.com/docker/docker-credential-helpers/credentials"
	"github.com/sirupsen/logrus"
)

const (
	defaultEnvPrefix         = "DCE" // DOCKER_CREDENTIAL_ENV
	defaultEnvUsernameSuffix = "USER"
	defaultEnvPasswordSuffix = "PASS"

	envSep = "_"
)

type (
	Config struct {
		EnvPrefix string
	}
	NotSupportedError struct {
		action string
	}
	envHelper struct {
		config Config
		logger *logrus.Logger
	}
)

var (
	defaultHelper = Helper(Config{}, logrus.StandardLogger())
)

func (m *NotSupportedError) Error() string {
	return "not supported for " + m.action
}

/* Implement docker credential helper */

func DefaultHelper() docker_credentials.Helper {
	return defaultHelper
}

func Helper(config Config, logger *logrus.Logger) docker_credentials.Helper {
	config.EnvPrefix = ifBlankThenSetDefault(config.EnvPrefix, defaultEnvPrefix)
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return &envHelper{config: config, logger: logger}
}

func (*envHelper) Add(*docker_credentials.Credentials) error {
	return &NotSupportedError{docker_credentials.ActionStore}
}

func (*envHelper) Delete(string) error {
	return &NotSupportedError{docker_credentials.ActionErase}
}

func (h *envHelper) Get(serverURL string) (username string, password string, err error) {
	return h.getCredentials(serverURL)
}

func (*envHelper) List() (map[string]string, error) {
	return nil, &NotSupportedError{docker_credentials.ActionList}
}

func (h *envHelper) getCredentials(serverURL string) (string, string, error) {
	h.logger.Warnf("xxx")
	var hostname, port string
	hostname, port, err := parseURL(serverURL)
	if err != nil {
		return "", "", err
	}

	_, envUsername, envPassword := h.getEnvironmentVariableNames(hostname, port)

	//var usernameFound, passwordFound bool
	if username, usernameFound := os.LookupEnv(envUsername); usernameFound {
		if password, passwordFound := os.LookupEnv(envPassword); passwordFound {
			h.logger.Debugf("Found credentials from environment[%s, %s] variables for %s", envUsername, envPassword, serverURL)
			return username, password, nil
		}
	}
	h.logger.Debugf(
		"Not found credentials for %s, lookup envs: %s and %s",
		serverURL, envUsername, envPassword)
	return "", "", nil
}

func (h *envHelper) getEnvironmentVariableNames(host, port string) (envHostName, envUsername, envPassword string) {
	envHostname := strings.ReplaceAll(host, "-", envSep)
	envHostname = strings.ReplaceAll(envHostname, ".", envSep)
	if port != "" {
		envHostname = strings.Join([]string{envHostname, port}, envSep)
	}
	envUsername = strings.Join([]string{h.config.EnvPrefix, envHostname, defaultEnvUsernameSuffix}, envSep)
	envPassword = strings.Join([]string{h.config.EnvPrefix, envHostname, defaultEnvPasswordSuffix}, envSep)
	return
}

/* Implement docker credential helper */

func parseURL(serverURL string) (hostname, port string, err error) {
	serverURL = strings.TrimPrefix(serverURL, "http://")
	serverURL = strings.TrimPrefix(serverURL, "https://")
	urlObj, err := url.Parse("https://" + serverURL)
	if err != nil {
		return "", "", err
	}
	return urlObj.Hostname(), urlObj.Port(), nil
}

func ifBlankThenSetDefault(i string, d string) string {
	if i == "" {
		return d
	}
	return i
}
